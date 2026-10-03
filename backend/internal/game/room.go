package game

import (
	"fmt"
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
	}
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

	matched, err := r.Cities.FindSimilar(city)
	if err != nil || matched == nil {
		r.mu.Lock()
		if r.Phase == PhasePlaying && playerID == r.turnID {
			r.sendRejectedLocked(playerID, "not_in_db", city)
		}
		r.mu.Unlock()
		return
	}

	canonical := matched.Name
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
