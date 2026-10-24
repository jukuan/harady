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

	// maxLearnedPerGame limits how many cities a single room can add to the
	// DB. Prevents a spammer from filling the cities table in one sitting.
	maxLearnedPerGame = 3
)

// defaultTurnTimeout is set once at startup via SetTurnTimeout. Zero disables
// the auto-pass safety net.
var defaultTurnTimeout = 120 * time.Second

// SetTurnTimeout configures the auto-pass timer for stuck turns.
func SetTurnTimeout(d time.Duration) { defaultTurnTimeout = d }

// allowedReactions is the server-side whitelist for the /react channel.
var allowedReactions = map[string]bool{
	"👍": true, "😂": true, "🔥": true, "❤️": true, "🤔": true,
	"👏": true, "😮": true, "🙈": true, "🎉": true, "💀": true,
}

const reactCooldown = 400 * time.Millisecond

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
	startedAt time.Time

	learnedCount int

	turnTimeout time.Duration
	turnGen     uint64
	turnTimer   *time.Timer
}

func newRoom(code string, cities *store.CityStore) *Room {
	return &Room{
		Code:        code,
		Phase:       PhaseLobby,
		players:     map[string]*Player{},
		chainSet:    map[string]bool{},
		lastTouch:   time.Now(),
		Cities:      cities,
		turnTimeout: defaultTurnTimeout,
	}
}

// ---------- players ----------

func (r *Room) Join(p *Player, c ClientHandle) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p.ID != "" {
		if existing, ok := r.players[p.ID]; ok {
			existing.Nickname = p.Nickname
			existing.mu.Lock()
			existing.client = c
			existing.mu.Unlock()
			r.lastTouch = time.Now()
			return nil
		}
	}

	if len(r.players) >= MaxPlayers {
		return fmt.Errorf("Пакой поўны")
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
		return nil, fmt.Errorf("Толькі гаспадар можа дадаваць ботаў")
	}
	if len(r.players) >= MaxPlayers {
		return nil, fmt.Errorf("Пакой поўны")
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
	chain := append([]ChainEntry{}, r.chain...)
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
		return fmt.Errorf("Гульня ўжо ідзе")
	}
	if byPlayerID != r.HostID {
		r.mu.Unlock()
		return fmt.Errorf("Толькі гаспадар можа пачаць")
	}
	if len(r.players) < MinPlayers {
		r.mu.Unlock()
		return fmt.Errorf("Патрэбна мінімум %d гульцоў", MinPlayers)
	}
	for _, p := range r.players {
		p.Missed = 0
		p.Out = false
		p.LastUnknown = ""
		p.UnknownCount = 0
		p.LastReactAt = time.Time{}
	}
	r.learnedCount = 0
	r.chain = nil
	r.chainSet = map[string]bool{}
	r.required = 0
	r.Phase = PhasePlaying
	r.turnID = r.order[0]
	r.lastTouch = time.Now()
	r.startedAt = time.Now()

	firstID := r.turnID
	firstNick := ""
	if p := r.players[firstID]; p != nil {
		firstNick = p.Nickname
	}
	r.broadcastLocked(ServerMessage{Type: "turn_started", Data: TurnStarted{
		PlayerID: firstID, Nickname: firstNick, RequiredLetter: "", Round: 1,
	}})
	r.armTurnTimerLocked()
	slog.Info("game started",
		"room", r.Code, "players", len(r.players), "first", firstNick)
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
	slog.Debug("kicking bot", "room", r.Code, "bot", p.Nickname)
	go r.runBotTurn(p)
}

func (r *Room) advanceTurnLocked() {
	r.turnGen++
	if r.turnTimer != nil {
		r.turnTimer.Stop()
		r.turnTimer = nil
	}

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
	r.armTurnTimerLocked()
}

func (r *Room) armTurnTimerLocked() {
	if r.turnTimeout <= 0 {
		return
	}
	gen := r.turnGen
	timeout := r.turnTimeout
	r.turnTimer = time.AfterFunc(timeout, func() {
		r.mu.Lock()
		if r.turnGen != gen || r.Phase != PhasePlaying {
			r.mu.Unlock()
			return
		}
		stuckID := r.turnID
		stuck := r.players[stuckID]
		nick := ""
		isBot := false
		if stuck != nil {
			nick = stuck.Nickname
			isBot = stuck.IsBot
		}
		slog.Warn("turn timeout — auto-pass",
			"room", r.Code, "player", nick, "is_bot", isBot, "timeout", timeout.String())
		nextID := r.passLocked(stuckID)
		r.mu.Unlock()
		if nextID != "" {
			r.BroadcastState()
			r.maybeKickBot(nextID)
		}
	})
}

func (r *Room) endGameLocked(winnerID string) {
	wasPlaying := r.Phase == PhasePlaying
	r.Phase = PhaseEnded
	if r.turnTimer != nil {
		r.turnTimer.Stop()
		r.turnTimer = nil
	}
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

	if wasPlaying && len(r.chain) >= 2 {
		hasBots := false
		for _, p := range r.players {
			if p.IsBot {
				hasBots = true
				break
			}
		}
		rec := store.GameRecord{
			RoomCode:       r.Code,
			ChainLength:    len(r.chain),
			PlayersCount:   len(r.players),
			HasBots:        hasBots,
			WinnerNickname: winnerNick,
			StartedAt:      r.startedAt,
			EndedAt:        time.Now(),
		}
		go func() {
			if err := r.Cities.RecordGame(rec); err != nil {
				slog.Error("record game failed", "room", r.Code, "err", err)
			}
		}()
		slog.Info("game recorded",
			"room", r.Code, "chain", len(r.chain), "players", len(r.players), "has_bots", hasBots)
	}
}

func (r *Room) sendRejectedLocked(playerID, reason, text string) {
	if p := r.players[playerID]; p != nil {
		p.Send(ServerMessage{Type: "chain_rejected", Data: ChainRejected{Reason: reason, Text: text}})
	}
}

func (r *Room) SubmitCity(playerID, raw string) bool {
	r.mu.Lock()
	if r.Phase != PhasePlaying || playerID != r.turnID {
		r.mu.Unlock()
		slog.Debug("submit ignored (not your turn)",
			"room", r.Code, "player", playerID, "turn", r.turnID)
		return false
	}
	p := r.players[playerID]
	if p == nil || p.Out {
		r.mu.Unlock()
		return false
	}
	city := strings.TrimSpace(raw)
	if city == "" || len([]rune(city)) > MaxCityLen {
		r.mu.Unlock()
		return false
	}
	if r.required != 0 && FirstLetter(city) != r.required {
		r.sendRejectedLocked(playerID, "wrong_letter", city)
		r.mu.Unlock()
		return false
	}
	r.mu.Unlock()

	matched, err := r.Cities.FindSimilar(city)
	var cityID int64
	canonical := ""
	if err == nil && matched != nil {
		canonical = matched.Name
		cityID = matched.ID
		r.resetUnknown(playerID)
	} else {
		id, learned, ok := r.tryLearnCity(playerID, city)
		if !ok {
			r.mu.Lock()
			if r.Phase == PhasePlaying && playerID == r.turnID {
				r.sendRejectedLocked(playerID, "not_in_db", city)
			}
			r.mu.Unlock()
			return false
		}
		canonical = learned
		cityID = id
	}

	norm := NormalizeCity(canonical)

	r.mu.Lock()
	if r.Phase != PhasePlaying || playerID != r.turnID {
		r.mu.Unlock()
		return false
	}
	p = r.players[playerID]
	if p == nil || p.Out {
		r.mu.Unlock()
		return false
	}
	if r.chainSet[norm] {
		slog.Info("city rejected",
			"room", r.Code, "player", p.Nickname, "city", canonical, "reason", "already_used")
		r.sendRejectedLocked(playerID, "already_used", canonical)
		r.mu.Unlock()
		return false
	}
	slog.Info("city accepted",
		"room", r.Code, "player", p.Nickname, "is_bot", p.IsBot, "city", canonical)
	if !p.IsBot {
		r.Cities.RecordSubmission(cityID, r.Code)
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
	return true
}

func (r *Room) resetUnknown(playerID string) {
	r.mu.Lock()
	if p := r.players[playerID]; p != nil {
		p.LastUnknown = ""
		p.UnknownCount = 0
	}
	r.mu.Unlock()
}

func (r *Room) tryLearnCity(playerID, city string) (int64, string, bool) {
	if !validLearnedCity(city) {
		r.resetUnknown(playerID)
		return 0, "", false
	}
	norm := NormalizeCity(city)

	r.mu.Lock()
	p := r.players[playerID]
	if p == nil {
		r.mu.Unlock()
		return 0, "", false
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
		return 0, "", false
	}
	if capReached {
		return 0, "", false
	}

	canonical := capitalizeFirst(city)
	id, err := r.Cities.AddLearned(canonical, "")
	if err != nil && !errors.Is(err, store.ErrDuplicate) {
		return 0, "", false
	}
	if errors.Is(err, store.ErrDuplicate) {
		if existing, ferr := r.Cities.FindSimilar(canonical); ferr == nil && existing != nil {
			id = existing.ID
		}
	}

	r.mu.Lock()
	if pp := r.players[playerID]; pp != nil {
		pp.LastUnknown = ""
		pp.UnknownCount = 0
	}
	if err == nil {
		r.learnedCount++
	}
	r.broadcastLocked(ServerMessage{Type: "city_learned", Data: CityLearned{
		City: canonical,
	}})
	r.mu.Unlock()

	return id, canonical, true
}

func (r *Room) Pass(playerID string) {
	r.mu.Lock()
	nextID := r.passLocked(playerID)
	r.mu.Unlock()
	if nextID == "" {
		return
	}
	r.BroadcastState()
	r.maybeKickBot(nextID)
}

func (r *Room) passLocked(playerID string) string {
	if r.Phase != PhasePlaying || playerID != r.turnID {
		return ""
	}
	p := r.players[playerID]
	if p == nil || p.Out {
		return ""
	}
	p.Missed++
	slog.Info("player passed",
		"room", r.Code, "player", p.Nickname, "is_bot", p.IsBot, "missed", p.Missed)
	r.broadcastLocked(ServerMessage{Type: "player_passed", Data: PlayerPassed{
		PlayerID: p.ID, Nickname: p.Nickname, Missed: p.Missed, IsBot: p.IsBot,
	}})
	if p.Missed >= MaxMisses {
		p.Out = true
		slog.Info("player out",
			"room", r.Code, "player", p.Nickname, "missed", p.Missed)
		r.broadcastLocked(ServerMessage{Type: "player_out", Data: PlayerOut{
			PlayerID: p.ID, Nickname: p.Nickname, Missed: p.Missed,
		}})
	}
	r.advanceTurnLocked()
	return r.turnID
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
