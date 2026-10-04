package game

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jukuan/harady/backend/internal/store"
)

const (
	PhaseLobby   = "lobby"
	PhasePlaying = "playing"
	PhaseEnded   = "ended"

	MinPlayers = 2
	MaxPlayers = 10
	MaxMisses  = 2
	MaxCityLen = 80

	// maxLearnedPerGame limits how many cities a single room can add to
	// the DB. Prevents a spammer from filling the cities table in one sitting.
	maxLearnedPerGame = 3
)

type Room struct {
	Code   string
	HostID string
	Phase  string

	Cities *store.CityStore

	mu        sync.RWMutex
	players   map[string]*Player
	order     []string
	chain     []ChainEntry
	chainSet  map[string]bool
	required  rune
	turnID    string
	lastTouch time.Time

	// Three-strikes cap: max learned cities per room per game.
	learnedCount int
}

func newRoom(code string, cities *store.CityStore) *Room {
	return &Room{
		Code:      code,
		Phase:     PhaseLobby,
		players:   map[string]*Player{},
		chainSet:  map[string]bool{},
		lastTouch: time.Now(),
		Cities:    cities,
	}
}

// ---------- players ----------

func (r *Room) Join(p *Player, c ClientHandle) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.players) >= MaxPlayers {
		return fmt.Errorf("room full")
	}
	if r.HostID == "" {
		p.IsHost = true
		r.HostID = p.ID
	}
	p.mu.Lock()
	p.client = c
	p.mu.Unlock()
	r.players[p.ID] = p
	r.order = append(r.order, p.ID)
	r.lastTouch = time.Now()
	return nil
}

func (r *Room) Leave(playerID string) {
	r.mu.Lock()
	if _, ok := r.players[playerID]; !ok {
		r.mu.Unlock()
		return
	}
	delete(r.players, playerID)
	for i, id := range r.order {
		if id == playerID {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	if r.HostID == playerID && len(r.order) > 0 {
		r.HostID = r.order[0]
		if hp := r.players[r.HostID]; hp != nil {
			hp.IsHost = true
		}
	}
	if r.Phase == PhasePlaying && r.turnID == playerID {
		r.advanceTurnLocked()
	}
	r.mu.Unlock()
	r.BroadcastState()
}

func (r *Room) AddBotBy(hostID string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if hostID != r.HostID {
		return nil, fmt.Errorf("only host can add bots")
	}
	if len(r.players) >= MaxPlayers {
		return nil, fmt.Errorf("room full")
	}
	p := &Player{ID: "bot-" + randString(6), Nickname: pickBotName(r.players), IsBot: true}
	r.players[p.ID] = p
	r.order = append(r.order, p.ID)
	r.lastTouch = time.Now()
	return p, nil
}

func (r *Room) playerListLocked() []PlayerView {
	out := make([]PlayerView, 0, len(r.players))
	for _, id := range r.order {
		if p := r.players[id]; p != nil {
			out = append(out, p.View())
		}
	}
	return out
}

func (r *Room) activeOrderLocked() []string {
	out := make([]string, 0, len(r.order))
	for _, id := range r.order {
		if p := r.players[id]; p != nil && !p.Out {
			out = append(out, id)
		}
	}
	return out
}

// ---------- snapshots ----------

func (r *Room) snapshotLocked() RoomState {
	nick := ""
	if p := r.players[r.turnID]; p != nil {
		nick = p.Nickname
	}
	letter := ""
	if r.required != 0 {
		letter = string(r.required)
	}
	players := r.playerListLocked()
	if players == nil {
		players = []PlayerView{}
	}
	chain := append([]ChainEntry{}, r.chain...) // non-nil even when empty
	return RoomState{
		Code:            r.Code,
		HostID:          r.HostID,
		Phase:           r.Phase,
		Players:         players,
		Chain:           chain,
		RequiredLetter:  letter,
		CurrentTurnID:   r.turnID,
		CurrentTurnNick: nick,
		MinPlayers:      MinPlayers,
		MaxPlayers:      MaxPlayers,
		MaxMisses:       MaxMisses,
	}
}

func (r *Room) Snapshot() RoomState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshotLocked()
}

func (r *Room) BroadcastState() {
	r.mu.RLock()
	snap := r.snapshotLocked()
	r.broadcastLocked(ServerMessage{Type: "room_state", Data: snap})
	r.mu.RUnlock()
}

func (r *Room) PlayerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.players)
}

func (r *Room) IdleSince() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastTouch
}

// ---------- broadcasting ----------

func (r *Room) broadcastLocked(m ServerMessage) {
	for _, id := range r.order {
		if p := r.players[id]; p != nil {
			p.Send(m)
		}
	}
}

func (r *Room) broadcast(m ServerMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	r.broadcastLocked(m)
}

// ---------- game flow ----------

func (r *Room) Start(byPlayerID string) error {
	r.mu.Lock()
	if r.Phase != PhaseLobby && r.Phase != PhaseEnded {
		r.mu.Unlock()
		return fmt.Errorf("game already running")
	}
	if byPlayerID != r.HostID {
		r.mu.Unlock()
		return fmt.Errorf("only host can start")
	}
	if len(r.players) < MinPlayers {
		r.mu.Unlock()
		return fmt.Errorf("need at least %d players", MinPlayers)
	}
	for _, p := range r.players {
		p.Missed = 0
		p.Out = false
		p.LastUnknown = ""
		p.UnknownCount = 0
	}
	r.learnedCount = 0
	r.chain = nil
	r.chainSet = map[string]bool{}
	r.required = 0
	r.Phase = PhasePlaying
	r.turnID = r.order[0]
	r.lastTouch = time.Now()

	firstID := r.turnID
	firstNick := ""
	if p := r.players[firstID]; p != nil {
		firstNick = p.Nickname
	}
	r.broadcastLocked(ServerMessage{Type: "turn_started", Data: TurnStarted{
		PlayerID: firstID, Nickname: firstNick, RequiredLetter: "", Round: 1,
	}})
	r.mu.Unlock()

	r.BroadcastState()
	r.maybeKickBot(firstID)
	return nil
}

func (r *Room) maybeKickBot(playerID string) {
	r.mu.RLock()
	p := r.players[playerID]
	active := r.Phase == PhasePlaying && r.turnID == playerID && p != nil && p.IsBot
	r.mu.RUnlock()
	if !active {
		return
	}
	go r.runBotTurn(p)
}

func (r *Room) advanceTurnLocked() {
	active := r.activeOrderLocked()
	if len(active) == 0 {
		r.endGameLocked("")
		return
	}
	if len(active) == 1 {
		r.endGameLocked(active[0])
		return
	}
	cur := -1
	for i, id := range active {
		if id == r.turnID {
			cur = i
			break
		}
	}
	next := 0
	if cur >= 0 {
		next = (cur + 1) % len(active)
	}
	r.turnID = active[next]
	nick := ""
	if p := r.players[r.turnID]; p != nil {
		nick = p.Nickname
	}
	letter := ""
	if r.required != 0 {
		letter = string(r.required)
	}
	r.broadcastLocked(ServerMessage{Type: "turn_started", Data: TurnStarted{
		PlayerID: r.turnID, Nickname: nick, RequiredLetter: letter, Round: len(r.chain) + 1,
	}})
}

func (r *Room) endGameLocked(winnerID string) {
	r.Phase = PhaseEnded
	winnerNick := ""
	if winnerID != "" {
		if p := r.players[winnerID]; p != nil {
			winnerNick = p.Nickname
		}
	}
	players := r.playerListLocked()
	if players == nil {
		players = []PlayerView{}
	}
	chain := append([]ChainEntry{}, r.chain...)
	r.broadcastLocked(ServerMessage{Type: "game_ended", Data: GameEnded{
		WinnerID:   winnerID,
		WinnerNick: winnerNick,
		Players:    players,
		Chain:      chain,
	}})
}

func (r *Room) sendRejectedLocked(playerID, reason, text string) {
	if p := r.players[playerID]; p != nil {
		p.Send(ServerMessage{Type: "chain_rejected", Data: ChainRejected{Reason: reason, Text: text}})
	}
}

// SubmitCity is called when the player whose turn it is proposes a city.
// The input is matched against the DB with typo tolerance; the canonical
// DB name is what goes into the chain.
//
// Three-strikes self-healing: if the same unknown name is typed three times
// in a row by the same player (and passes validation), it is inserted into
// the DB with source='learned' and the move is accepted.
func (r *Room) SubmitCity(playerID, raw string) {
	r.mu.Lock()
	if r.Phase != PhasePlaying || playerID != r.turnID {
		r.mu.Unlock()
		return
	}
	p := r.players[playerID]
	if p == nil || p.Out {
		r.mu.Unlock()
		return
	}
	city := strings.TrimSpace(raw)
	if city == "" || len([]rune(city)) > MaxCityLen {
		r.mu.Unlock()
		return
	}
	if r.required != 0 && FirstLetter(city) != r.required {
		r.sendRejectedLocked(playerID, "wrong_letter", city)
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()

	// Try the DB (with typo tolerance) first.
	matched, err := r.Cities.FindSimilar(city)
	canonical := ""
	if err == nil && matched != nil {
		canonical = matched.Name
		r.resetUnknown(playerID)
	} else {
		// Not in DB — try the three-strikes learning path.
		learned, ok := r.tryLearnCity(playerID, city)
		if !ok {
			r.mu.Lock()
			if r.Phase == PhasePlaying && playerID == r.turnID {
				r.sendRejectedLocked(playerID, "not_in_db", city)
			}
			r.mu.Unlock()
			return
		}
		canonical = learned
	}

	norm := NormalizeCity(canonical)

	r.mu.Lock()
	if r.Phase != PhasePlaying || playerID != r.turnID {
		r.mu.Unlock()
		return
	}
	p = r.players[playerID]
	if p == nil || p.Out {
		r.mu.Unlock()
		return
	}
	if r.chainSet[norm] {
		r.sendRejectedLocked(playerID, "already_used", canonical)
		r.mu.Unlock()
		return
	}
	entry := ChainEntry{PlayerID: p.ID, Nickname: p.Nickname, City: canonical, IsBot: p.IsBot}
	r.chain = append(r.chain, entry)
	r.chainSet[norm] = true
	r.required = LastMeaningfulLetter(canonical)

	letter := ""
	if r.required != 0 {
		letter = string(r.required)
	}
	r.broadcastLocked(ServerMessage{Type: "chain_added", Data: ChainAdded{
		Entry: entry, NextRequiredLetter: letter,
	}})

	r.advanceTurnLocked()
	nextID := r.turnID
	r.mu.Unlock()

	r.BroadcastState()
	r.maybeKickBot(nextID)
}

// resetUnknown clears the three-strikes counter for a player.
func (r *Room) resetUnknown(playerID string) {
	r.mu.Lock()
	if p := r.players[playerID]; p != nil {
		p.LastUnknown = ""
		p.UnknownCount = 0
	}
	r.mu.Unlock()
}

// tryLearnCity implements the three-strikes rule. Returns the canonical name
// to use if the city was learned (or a duplicate insert raced), or ("", false)
// if the move should be rejected as not_in_db.
func (r *Room) tryLearnCity(playerID, city string) (string, bool) {
	if !validLearnedCity(city) {
		r.resetUnknown(playerID)
		return "", false
	}
	norm := NormalizeCity(city)

	r.mu.Lock()
	p := r.players[playerID]
	if p == nil {
		r.mu.Unlock()
		return "", false
	}
	if p.LastUnknown == norm {
		p.UnknownCount++
	} else {
		p.LastUnknown = norm
		p.UnknownCount = 1
	}
	count := p.UnknownCount
	capReached := r.learnedCount >= maxLearnedPerGame
	r.mu.Unlock()

	if count < 3 {
		return "", false
	}
	if capReached {
		return "", false
	}

	canonical := capitalizeFirst(city)
	_, err := r.Cities.AddLearned(canonical, "")
	if err != nil && !errors.Is(err, store.ErrDuplicate) {
		// Real DB error (disk full, etc.) — do not accept the move.
		return "", false
	}
	// ErrDuplicate means someone else just inserted it. Either way, use the
	// canonical name and proceed.

	r.mu.Lock()
	if p := r.players[playerID]; p != nil {
		p.LastUnknown = ""
		p.UnknownCount = 0
	}
	// Only count it against the game cap when we actually inserted.
	if err == nil {
		r.learnedCount++
	}
	r.broadcastLocked(ServerMessage{Type: "city_learned", Data: CityLearned{
		City: canonical,
	}})
	r.mu.Unlock()

	return canonical, true
}

// Pass is when a player gives up their turn.
func (r *Room) Pass(playerID string) {
	r.mu.Lock()
	if r.Phase != PhasePlaying || playerID != r.turnID {
		r.mu.Unlock()
		return
	}
	p := r.players[playerID]
	if p == nil || p.Out {
		r.mu.Unlock()
		return
	}
	p.Missed++
	r.broadcastLocked(ServerMessage{Type: "player_passed", Data: PlayerPassed{
		PlayerID: p.ID, Nickname: p.Nickname, Missed: p.Missed, IsBot: p.IsBot,
	}})
	if p.Missed >= MaxMisses {
		p.Out = true
		r.broadcastLocked(ServerMessage{Type: "player_out", Data: PlayerOut{
			PlayerID: p.ID, Nickname: p.Nickname, Missed: p.Missed,
		}})
	}
	r.advanceTurnLocked()
	nextID := r.turnID
	r.mu.Unlock()

	r.BroadcastState()
	r.maybeKickBot(nextID)
}

func (r *Room) ForceEnd(byPlayerID string) {
	r.mu.Lock()
	if byPlayerID != r.HostID {
		r.mu.Unlock()
		return
	}
	r.endGameLocked("")
	r.mu.Unlock()
	r.BroadcastState()
}

// CloseAll sends a "service restart" close to every human player.
// Called by Hub.Shutdown on SIGTERM.
func (r *Room) CloseAll() {
	r.mu.RLock()
	players := make([]*Player, 0, len(r.players))
	for _, id := range r.order {
		if p := r.players[id]; p != nil && !p.IsBot {
			players = append(players, p)
		}
	}
	r.mu.RUnlock()
	for _, p := range players {
		p.Send(ServerMessage{Type: "server_shutdown", Data: nil})
	}
}


// allowedReactions is the server-side whitelist. A modified client cannot
// inject arbitrary text into the reaction stream — only these exact strings
// are accepted.
var allowedReactions = map[string]bool{
	"👍": true, "😂": true, "🔥": true, "❤️": true, "🤔": true,
	"👏": true, "😮": true, "🙈": true, "🎉": true, "💀": true,
}

// reactCooldown limits how often a single player can react.
const reactCooldown = 400 * time.Millisecond

// React broadcasts a floating emoji attached to the latest chain entry.
// Rate-limited per player; silent no-op if the whitelist, cooldown, or
// game state reject it. Never fails the caller.
func (r *Room) React(playerID, emoji string) {
	if !allowedReactions[emoji] {
		return
	}

	r.mu.Lock()
	if r.Phase != PhasePlaying && r.Phase != PhaseEnded {
		r.mu.Unlock()
		return
	}
	p := r.players[playerID]
	if p == nil {
		r.mu.Unlock()
		return
	}
	now := time.Now()
	if !p.LastReactAt.IsZero() && now.Sub(p.LastReactAt) < reactCooldown {
		r.mu.Unlock()
		return
	}
	p.LastReactAt = now

	// Attach to the last chain entry. If the chain is empty, still broadcast
	// with chain_index = -1 — the client renders it at the bottom.
	idx := len(r.chain) - 1
	if idx < 0 {
		idx = -1
	}
	payload := ReactionData{
		PlayerID:   p.ID,
		Nickname:   p.Nickname,
		Emoji:      emoji,
		ChainIndex: idx,
		Ts:         now.UnixMilli(),
	}
	r.broadcastLocked(ServerMessage{Type: "reaction", Data: payload})
	r.mu.Unlock()

	slog.Debug("reaction", "room", r.Code, "player", p.Nickname, "emoji", emoji, "index", idx)
}
