package game

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/yourname/harady/backend/internal/models"
	"github.com/yourname/harady/backend/internal/store"
)

const (
	PhaseLobby         = "lobby"
	PhaseRoundActive   = "round_active"
	PhaseRoundRevealed = "round_revealed"
	PhaseEnded         = "ended"

	MinPlayers     = 2
	MaxPlayers     = 10
	RoundDuration  = 90 * time.Second
	RevealDuration = 4 * time.Second
	MaxTextLength  = 120
)

// Room is a single game instance. All mutations are guarded by mu.
//
// Concurrency contract:
//   - Callers that hold r.mu (R or W) must use broadcastLocked.
//   - Callers that don't hold r.mu must use BroadcastState or broadcast.
type Room struct {
	Code   string
	HostID string
	Phase  string
	Round  int

	Cities *store.CityStore

	mu         sync.RWMutex
	players    map[string]*Player
	order      []string // join order; used for actor rotation
	actorIdx   int
	secret     *models.City
	roundTimer *time.Timer
	roundDone  chan struct{}
	lastTouch  time.Time
}

func newRoom(code string, cities *store.CityStore) *Room {
	return &Room{
		Code:      code,
		Phase:     PhaseLobby,
		players:   map[string]*Player{},
		actorIdx:  -1,
		lastTouch: time.Now(),
		Cities:    cities,
	}
}

// ---------- player management ----------

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
	p, ok := r.players[playerID]
	if !ok {
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
	wasActor := r.Phase == PhaseRoundActive && r.actorIDLocked() == playerID
	if r.HostID == playerID && len(r.order) > 0 {
		r.HostID = r.order[0]
		if hp := r.players[r.HostID]; hp != nil {
			hp.IsHost = true
		}
	}
	if wasActor {
		r.endRoundLocked("")
	}
	_ = p
	r.mu.Unlock()
	r.BroadcastState()
}

// AddBotBy adds a bot, but only if hostID is the current host.
func (r *Room) AddBotBy(hostID, nickname string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if hostID != r.HostID {
		return nil, fmt.Errorf("only host can add bots")
	}
	return r.addBotLocked(nickname)
}

func (r *Room) addBotLocked(nickname string) (*Player, error) {
	if len(r.players) >= MaxPlayers {
		return nil, fmt.Errorf("room full")
	}
	if nickname == "" {
		nickname = pickBotName(r.players)
	}
	p := &Player{
		ID:       "bot-" + randString(6),
		Nickname: nickname,
		IsBot:    true,
	}
	r.players[p.ID] = p
	r.order = append(r.order, p.ID)
	r.lastTouch = time.Now()
	return p, nil
}

func (r *Room) actorIDLocked() string {
	if r.actorIdx < 0 || r.actorIdx >= len(r.order) {
		return ""
	}
	return r.order[r.actorIdx]
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

// ---------- snapshots ----------

func (r *Room) snapshotLocked() RoomState {
	actorID := r.actorIDLocked()
	actorNick := ""
	if p := r.players[actorID]; p != nil {
		actorNick = p.Nickname
	}
	masked := ""
	if r.secret != nil {
		masked = maskCity(r.secret.Name)
	}
	return RoomState{
		Code:       r.Code,
		HostID:     r.HostID,
		Phase:      r.Phase,
		Round:      r.Round,
		ActorID:    actorID,
		ActorNick:  actorNick,
		MaskedCity: masked,
		Players:    r.playerListLocked(),
		MinPlayers: MinPlayers,
		MaxPlayers: MaxPlayers,
		Duration:   int(RoundDuration / time.Second),
	}
}

func (r *Room) Snapshot() RoomState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshotLocked()
}

// BroadcastState sends the current room_state to every human player.
func (r *Room) BroadcastState() {
	r.mu.RLock()
	snap := r.snapshotLocked()
	r.broadcastLocked(ServerMessage{Type: "room_state", Data: snap})
	r.mu.RUnlock()
}

// PlayerCount / IdleSince are used by the hub GC.
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

// broadcastLocked sends m to every non-bot player. Caller must hold r.mu.
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

// ---------- game loop ----------

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
		p.Score = 0
	}
	r.Round = 0
	r.actorIdx = rand.Intn(len(r.order))
	r.mu.Unlock()

	r.BroadcastState()
	go r.startRound()
	return nil
}

func (r *Room) startRound() {
	r.mu.Lock()

	if len(r.order) < MinPlayers {
		r.Phase = PhaseEnded
		snap := r.snapshotLocked()
		r.broadcastLocked(ServerMessage{Type: "game_ended", Data: snap})
		r.mu.Unlock()
		return
	}

	city, err := r.Cities.Random()
	if err != nil {
		r.Phase = PhaseEnded
		r.mu.Unlock()
		r.broadcast(ServerMessage{
			Type: "error",
			Data: ErrorData{Message: "no cities in DB — run `harady-cli seed`"},
		})
		return
	}

	r.Round++
	r.Phase = PhaseRoundActive
	r.secret = city

	actorID := r.actorIDLocked()
	actor := r.players[actorID]
	actorNick := ""
	if actor != nil {
		actorNick = actor.Nickname
	}

	r.broadcastLocked(ServerMessage{
		Type: "round_started",
		Data: RoundStarted{
			ActorID:    actorID,
			ActorNick:  actorNick,
			MaskedCity: maskCity(city.Name),
			Round:      r.Round,
			Duration:   int(RoundDuration / time.Second),
		},
	})

	// The secret is sent ONLY to the human actor.
	if actor != nil && !actor.IsBot {
		actor.Send(ServerMessage{
			Type: "your_word",
			Data: YourWord{City: city.Name, Region: city.Region},
		})
	}

	// Schedule round timeout.
	if r.roundTimer != nil {
		r.roundTimer.Stop()
	}
	if r.roundDone != nil {
		close(r.roundDone)
	}
	r.roundDone = make(chan struct{})
	tickDone := r.roundDone
	r.roundTimer = time.AfterFunc(RoundDuration, func() {
		r.mu.Lock()
		if r.Phase == PhaseRoundActive {
			r.endRoundLocked("")
		}
		r.mu.Unlock()
	})
	go r.tickLoop(tickDone)

	// Snapshot for bot goroutines.
	playersCopy := make([]*Player, 0, len(r.players))
	for _, id := range r.order {
		if p := r.players[id]; p != nil {
			playersCopy = append(playersCopy, p)
		}
	}
	secretCopy := *city

	r.mu.Unlock()

	for _, p := range playersCopy {
		if !p.IsBot {
			continue
		}
		if p.ID == actorID {
			go r.runActorBot(p, &secretCopy)
		} else {
			go r.runGuesserBot(p)
		}
	}

	r.BroadcastState()
}

func (r *Room) HandleClue(playerID, text string) {
	r.mu.RLock()
	phase := r.Phase
	actorID := r.actorIDLocked()
	secret := r.secret
	p := r.players[playerID]
	r.mu.RUnlock()

	if phase != PhaseRoundActive || p == nil {
		return
	}
	if playerID != actorID {
		p.Send(ServerMessage{Type: "error", Data: ErrorData{Message: "only the actor can give clues"}})
		return
	}

	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxTextLength {
		return
	}
	if secret != nil {
		low := strings.ToLower(text)
		if strings.Contains(low, strings.ToLower(secret.Name)) {
			p.Send(ServerMessage{Type: "error", Data: ErrorData{Message: "clue cannot contain the city name"}})
			return
		}
	}

	r.broadcast(ServerMessage{Type: "chat", Data: ChatMessage{
		PlayerID: p.ID, Nickname: p.Nickname, Text: text, Kind: "clue",
		IsBot: p.IsBot, Ts: time.Now().UnixMilli(),
	}})
}

func (r *Room) HandleGuess(playerID, text string) {
	r.mu.RLock()
	phase := r.Phase
	actorID := r.actorIDLocked()
	secret := r.secret
	p := r.players[playerID]
	r.mu.RUnlock()

	if phase != PhaseRoundActive || p == nil || playerID == actorID {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxTextLength {
		return
	}

	r.broadcast(ServerMessage{Type: "chat", Data: ChatMessage{
		PlayerID: p.ID, Nickname: p.Nickname, Text: text, Kind: "guess",
		IsBot: p.IsBot, Ts: time.Now().UnixMilli(),
	}})

	if secret == nil || !strings.EqualFold(text, secret.Name) {
		return
	}

	r.mu.Lock()
	if r.Phase == PhaseRoundActive {
		if w := r.players[playerID]; w != nil {
			w.Score += 100
		}
		if a := r.players[actorID]; a != nil {
			a.Score += 50
		}
		r.endRoundLocked(playerID)
	}
	r.mu.Unlock()
}

// endRoundLocked must be called with r.mu held.
func (r *Room) endRoundLocked(winnerID string) {
	if r.roundTimer != nil {
		r.roundTimer.Stop()
		r.roundTimer = nil
	}
	if r.roundDone != nil {
		close(r.roundDone)
		r.roundDone = nil
	}
	city := r.secret
	if city == nil {
		r.Phase = PhaseEnded
		return
	}
	r.Phase = PhaseRoundRevealed

	r.broadcastLocked(ServerMessage{
		Type: "round_ended",
		Data: RoundEnded{
			WinnerID: winnerID,
			City:     city.Name,
			Scores:   r.playerListLocked(),
		},
	})

	if winnerID != "" {
		winnerNick := ""
		if w := r.players[winnerID]; w != nil {
			winnerNick = w.Nickname
		}
		r.broadcastLocked(ServerMessage{
			Type: "correct_guess",
			Data: CorrectGuess{PlayerID: winnerID, Nickname: winnerNick, City: city.Name},
		})
	}

	if len(r.order) > 0 {
		r.actorIdx = (r.actorIdx + 1) % len(r.order)
	}

	time.AfterFunc(RevealDuration, func() {
		r.mu.Lock()
		cont := r.Phase == PhaseRoundRevealed && len(r.players) >= MinPlayers
		r.mu.Unlock()
		if cont {
			r.startRound()
		}
	})
}

func (r *Room) ForceEnd(byPlayerID string) {
	r.mu.Lock()
	if byPlayerID != r.HostID {
		r.mu.Unlock()
		return
	}
	if r.roundTimer != nil {
		r.roundTimer.Stop()
		r.roundTimer = nil
	}
	r.Phase = PhaseEnded
	snap := r.snapshotLocked()
	r.mu.Unlock()

	r.broadcast(ServerMessage{Type: "game_ended", Data: snap})
}

// tickLoop sends a tick frame every second until the round ends or the
// provided channel is closed. Duration is fixed for the whole round.
func (r *Room) tickLoop(done <-chan struct{}) {
	deadline := time.Now().Add(RoundDuration)
	dur := int(RoundDuration / time.Second)
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-done:
			return
		case now := <-t.C:
			left := int(deadline.Sub(now).Seconds())
			if left < 0 {
				left = 0
			}
			r.mu.RLock()
			round := r.Round
			active := r.Phase == PhaseRoundActive
			r.mu.RUnlock()
			if !active {
				return
			}
			r.broadcast(ServerMessage{
				Type: "tick",
				Data: TickData{SecondsLeft: left, Round: round, Duration: dur},
			})
			if left == 0 {
				return
			}
		}
	}
}
