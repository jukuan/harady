package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/yourname/harady/backend/internal/config"
	"github.com/yourname/harady/backend/internal/db"
	"github.com/yourname/harady/backend/internal/game"
	"github.com/yourname/harady/backend/internal/httpapi"
	"github.com/yourname/harady/backend/internal/models"
	"github.com/yourname/harady/backend/internal/store"
)

// =============================================================================
// Fixture: real router + real httptest.Server + real WebSocket dialer
// =============================================================================

func setupServer(t *testing.T) (httpBase, wsBase string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	cities := store.NewCityStore(conn)
	for _, c := range []models.City{
		{Name: "Мінск", Region: "Мінская", Clues: []string{"Сталіца", "Няміга", "Трэцяя падсказка"}},
		{Name: "Гомель", Region: "Гомельская", Clues: []string{"Сож", "Другі горад"}},
		{Name: "Брэст", Region: "Брэсцкая", Clues: []string{"Крэпасць", "Заходняя брама"}},
	} {
		if _, err := cities.Create(&c); err != nil {
			t.Fatalf("seed %s: %v", c.Name, err)
		}
	}

	cfg := &config.Config{
		Addr:        ":0",
		DBPath:      filepath.Join(dir, "test.db"),
		CORSOrigins: []string{"*"},
		Language:    "be",
		RoomIdleTTL: time.Hour,
	}
	hub := game.NewHub(cfg, cities)
	router := httpapi.NewRouter(cfg, hub, cities)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return srv.URL, strings.Replace(srv.URL, "http://", "ws://", 1)
}

func createRoom(t *testing.T, httpBase string) string {
	t.Helper()
	resp, err := http.Post(httpBase+"/api/rooms", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/rooms: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("POST /api/rooms status=%d", resp.StatusCode)
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode rooms: %v", err)
	}
	if body.Code == "" {
		t.Fatal("empty room code")
	}
	return body.Code
}

// =============================================================================
// wsClient: buffered reader with polling helpers
// =============================================================================

type wsClient struct {
	t    *testing.T
	conn *websocket.Conn

	mu    sync.Mutex
	inbox []game.ServerMessage
}

func dial(t *testing.T, url string) *wsClient {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	c := &wsClient{t: t, conn: conn}
	go c.readLoop()
	t.Cleanup(func() { _ = conn.Close() })
	return c
}

func (c *wsClient) readLoop() {
	for {
		var m game.ServerMessage
		if err := c.conn.ReadJSON(&m); err != nil {
			return
		}
		c.mu.Lock()
		c.inbox = append(c.inbox, m)
		c.mu.Unlock()
	}
}

func (c *wsClient) send(typ string, data any) {
	c.t.Helper()
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			c.t.Fatalf("marshal %s: %v", typ, err)
		}
		raw = b
	}
	if err := c.conn.WriteJSON(game.ClientMessage{Type: typ, Data: raw}); err != nil {
		c.t.Fatalf("ws write %s: %v", typ, err)
	}
}

// wait consumes the first buffered frame that matches pred, polling up to timeout.
func (c *wsClient) wait(pred func(game.ServerMessage) bool, timeout time.Duration) game.ServerMessage {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		for i, m := range c.inbox {
			if pred(m) {
				c.inbox = append(c.inbox[:i], c.inbox[i+1:]...)
				c.mu.Unlock()
				return m
			}
		}
		c.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	c.mu.Lock()
	types := make([]string, len(c.inbox))
	for i, m := range c.inbox {
		types[i] = m.Type
	}
	c.mu.Unlock()
	c.t.Fatalf("timeout waiting for frame; buffered=%v", types)
	return game.ServerMessage{}
}

func (c *wsClient) waitFor(typ string, timeout time.Duration) game.ServerMessage {
	c.t.Helper()
	return c.wait(func(m game.ServerMessage) bool { return m.Type == typ }, timeout)
}

func (c *wsClient) waitRoomState(players int, timeout time.Duration) game.RoomState {
	c.t.Helper()
	msg := c.wait(func(m game.ServerMessage) bool {
		if m.Type != "room_state" {
			return false
		}
		return len(decodeData[game.RoomState](c.t, m).Players) == players
	}, timeout)
	return decodeData[game.RoomState](c.t, msg)
}

// =============================================================================
// Helpers for typed decoding of ServerMessage.Data (which is `any`)
// =============================================================================

func decodeData[T any](t *testing.T, m game.ServerMessage) T {
	t.Helper()
	var out T
	b, err := json.Marshal(m.Data)
	if err != nil {
		t.Fatalf("remarshal %s: %v", m.Type, err)
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal %s: %v (raw=%s)", m.Type, err, string(b))
	}
	return out
}

func joinAs(t *testing.T, c *wsClient, nick string) string {
	t.Helper()
	c.send("join", game.JoinData{Nickname: nick})
	m := c.waitFor("joined", 2*time.Second)
	data := decodeData[map[string]any](t, m)
	id, _ := data["player_id"].(string)
	if id == "" {
		t.Fatalf("joined missing player_id: %v", data)
	}
	return id
}

// =============================================================================
// Tests
// =============================================================================

func TestWS_JoinThenRoomState(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	c := dial(t, wsBase+"/ws/rooms/"+code)
	id := joinAs(t, c, "Alice")
	if id == "" {
		t.Fatal("empty id")
	}

	state := c.waitRoomState(1, 2*time.Second)
	if state.Code != code {
		t.Errorf("state.Code=%q want %q", state.Code, code)
	}
	if state.Phase != game.PhaseLobby {
		t.Errorf("phase=%q want %q", state.Phase, game.PhaseLobby)
	}
	if state.Players[0].Nickname != "Alice" {
		t.Errorf("player nickname=%q", state.Players[0].Nickname)
	}
	if !state.Players[0].IsHost {
		t.Errorf("first joiner should be host")
	}
}

func TestWS_RoomNotFound(t *testing.T) {
	_, wsBase := setupServer(t)
	_, resp, err := websocket.DefaultDialer.Dial(wsBase+"/ws/rooms/nope", nil)
	if err == nil {
		t.Fatal("expected dial to fail")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got resp=%v err=%v", resp, err)
	}
}

func TestWS_UnknownMessageYieldsError(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	c := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, c, "Alice")
	c.waitRoomState(1, 2*time.Second)

	c.send("xyzzy", nil)
	errMsg := decodeData[game.ErrorData](t, c.waitFor("error", 2*time.Second))
	if !strings.Contains(errMsg.Message, "unknown") {
		t.Errorf("error message=%q", errMsg.Message)
	}
}

func TestWS_NonHostCannotStart(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, alice, "Alice")
	alice.waitRoomState(1, 2*time.Second)

	bob := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, bob, "Bob")
	alice.waitRoomState(2, 2*time.Second)

	bob.send("start", nil)
	errMsg := decodeData[game.ErrorData](t, bob.waitFor("error", 2*time.Second))
	if !strings.Contains(errMsg.Message, "host") {
		t.Errorf("error message=%q", errMsg.Message)
	}
}

func TestWS_TwoPlayersRoundStarts(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	aliceID := joinAs(t, alice, "Alice")

	bob := dial(t, wsBase+"/ws/rooms/"+code)
	bobID := joinAs(t, bob, "Bob")

	alice.waitRoomState(2, 2*time.Second)

	alice.send("start", nil)

	aRound := decodeData[game.RoundStarted](t, alice.waitFor("round_started", 3*time.Second))
	bRound := decodeData[game.RoundStarted](t, bob.waitFor("round_started", 3*time.Second))

	if aRound.ActorID == "" {
		t.Fatal("empty actor id")
	}
	if aRound.ActorID != bRound.ActorID {
		t.Errorf("actor mismatch: a=%q b=%q", aRound.ActorID, bRound.ActorID)
	}
	if aRound.ActorID != aliceID && aRound.ActorID != bobID {
		t.Errorf("actor id %q matches neither player", aRound.ActorID)
	}
	if aRound.MaskedCity == "" {
		t.Error("expected a masked city on round_started")
	}
	if aRound.Round != 1 {
		t.Errorf("round=%d want 1", aRound.Round)
	}
}

// The actor receives the secret; the guesser sends it back; the round ends
// with the correct-guess event visible to both players.
func TestWS_CorrectGuessEndsRound(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	aliceID := joinAs(t, alice, "Alice")

	bob := dial(t, wsBase+"/ws/rooms/"+code)
	bobID := joinAs(t, bob, "Bob")

	alice.waitRoomState(2, 2*time.Second)
	alice.send("start", nil)

	round := decodeData[game.RoundStarted](t, alice.waitFor("round_started", 3*time.Second))
	bob.waitFor("round_started", 3*time.Second)

	var actor, guesser *wsClient
	switch round.ActorID {
	case aliceID:
		actor, guesser = alice, bob
	case bobID:
		actor, guesser = bob, alice
	default:
		t.Fatalf("unknown actor %q", round.ActorID)
	}

	word := decodeData[game.YourWord](t, actor.waitFor("your_word", 2*time.Second))
	if word.City == "" {
		t.Fatal("actor did not receive a city")
	}

	// Actor gives a benign clue.
	actor.send("clue", game.TextData{Text: "Гэта горад"})

	// Guesser submits the correct city.
	guesser.send("guess", game.TextData{Text: word.City})

	// Both parties should see a chat echo of the guess.
	var sawChat bool
	for i := 0; i < 5 && !sawChat; i++ {
		msg := guesser.wait(func(m game.ServerMessage) bool { return m.Type == "chat" }, 2*time.Second)
		chat := decodeData[game.ChatMessage](t, msg)
		if chat.Kind == "guess" && chat.Text == word.City {
			sawChat = true
		}
	}
	if !sawChat {
		t.Error("guesser never saw their guess echoed")
	}

	cg := decodeData[game.CorrectGuess](t, guesser.waitFor("correct_guess", 3*time.Second))
	if cg.City != word.City {
		t.Errorf("correct_guess city=%q want %q", cg.City, word.City)
	}
	guesser.waitFor("round_ended", 3*time.Second)

	actorCG := decodeData[game.CorrectGuess](t, actor.waitFor("correct_guess", 3*time.Second))
	if actorCG.City != word.City {
		t.Errorf("actor saw wrong city %q", actorCG.City)
	}
	actor.waitFor("round_ended", 3*time.Second)
}

// The actor cannot guess their own word; server silently ignores it.
// We assert nothing is broadcast to the other player.
func TestWS_ActorCannotGuess(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	aliceID := joinAs(t, alice, "Alice")

	bob := dial(t, wsBase+"/ws/rooms/"+code)
	bobID := joinAs(t, bob, "Bob")

	alice.waitRoomState(2, 2*time.Second)
	alice.send("start", nil)

	round := decodeData[game.RoundStarted](t, alice.waitFor("round_started", 3*time.Second))
	bob.waitFor("round_started", 3*time.Second)

	var actor, other *wsClient
	switch round.ActorID {
	case aliceID:
		actor, other = alice, bob
	case bobID:
		actor, other = bob, alice
	default:
		t.Fatalf("unknown actor %q", round.ActorID)
	}

	actor.send("guess", game.TextData{Text: "спам-тэст"})

	// Drain any chat messages for a short period; there should be none
	// carrying our sentinel text.
	deadline := time.Now().Add(700 * time.Millisecond)
	for time.Now().Before(deadline) {
		other.mu.Lock()
		found := false
		for _, m := range other.inbox {
			if m.Type == "chat" {
				ch := decodeData[game.ChatMessage](t, m)
				if ch.Text == "спам-тэст" {
					found = true
					break
				}
			}
		}
		other.mu.Unlock()
		if found {
			t.Fatal("actor's guess was broadcast; it should have been dropped")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestWS_LeaveBroadcastsUpdatedState(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, alice, "Alice")
	alice.waitRoomState(1, 2*time.Second)

	bob := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, bob, "Bob")
	alice.waitRoomState(2, 2*time.Second)

	_ = bob.conn.Close()

	state := alice.waitRoomState(1, 3*time.Second)
	if state.Players[0].Nickname != "Alice" {
		t.Errorf("remaining player=%q", state.Players[0].Nickname)
	}
}

// Bot as clue-giver or bot as guesser — either way a chat frame should
// eventually arrive from the bot side. Marked slow under -short.
func TestWS_BotProducesChat(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: waits for bot timer")
	}
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	alice := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, alice, "Alice")
	alice.waitRoomState(1, 2*time.Second)

	alice.send("add_bot", nil)
	state := alice.waitRoomState(2, 2*time.Second)

	var botSeen bool
	for _, p := range state.Players {
		if p.IsBot {
			botSeen = true
			if p.Nickname == "" {
				t.Error("bot has empty nickname")
			}
		}
	}
	if !botSeen {
		t.Fatal("no bot in room after add_bot")
	}

	alice.send("start", nil)

	// Wait up to 15s for any chat frame from the bot.
	msg := alice.wait(func(m game.ServerMessage) bool {
		if m.Type != "chat" {
			return false
		}
		return decodeData[game.ChatMessage](t, m).IsBot
	}, 15*time.Second)

	chat := decodeData[game.ChatMessage](t, msg)
	if chat.Text == "" {
		t.Error("bot chat with empty text")
	}
	if chat.Kind != "clue" && chat.Kind != "guess" {
		t.Errorf("unexpected bot chat kind=%q", chat.Kind)
	}
}

// Frames emitted in one room never appear on another room's socket.
func TestWS_RoomsAreIsolated(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code1 := createRoom(t, httpBase)
	code2 := createRoom(t, httpBase)

	a := dial(t, wsBase+"/ws/rooms/"+code1)
	joinAs(t, a, "Alice")
	a.waitRoomState(1, 2*time.Second)

	b := dial(t, wsBase+"/ws/rooms/"+code2)
	joinAs(t, b, "Bob")
	b.waitRoomState(1, 2*time.Second)

	// Alice sends a clue; only she's in room1 and no round is running,
	// so nothing should propagate anywhere. Then we churn Bob in room2.
	a.send("clue", game.TextData{Text: "should-be-ignored"})

	time.Sleep(300 * time.Millisecond)

	b.mu.Lock()
	for _, m := range b.inbox {
		if m.Type == "chat" {
			t.Fatalf("chat leaked into other room: %+v", m)
		}
	}
	b.mu.Unlock()

	a.mu.Lock()
	for _, m := range a.inbox {
		if m.Type == "chat" {
			t.Fatalf("chat from lobby-phase clue: %+v", m)
		}
	}
	a.mu.Unlock()
}
