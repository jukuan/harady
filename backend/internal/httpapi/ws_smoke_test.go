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

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/game"
	"github.com/jukuan/harady/backend/internal/httpapi"
	"github.com/jukuan/harady/backend/internal/models"
	"github.com/jukuan/harady/backend/internal/store"
)

// =============================================================================
// Fixture: real router + real httptest.Server + real WebSocket dialer
// =============================================================================

func setupServer(t *testing.T) (string, string) {
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
		{Name: "Мінск"}, {Name: "Кіеў"}, {Name: "Варшава"},
		{Name: "Амстэрдам"}, {Name: "Магілёў"}, {Name: "Гомель"},
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

func createRoom(t *testing.T, base string) string {
	t.Helper()
	resp, err := http.Post(base+"/api/rooms", "application/json", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer resp.Body.Close()
	var b struct{ Code string `json:"code"` }
	_ = json.NewDecoder(resp.Body).Decode(&b)
	if b.Code == "" {
		t.Fatal("empty room code")
	}
	return b.Code
}

type client struct {
	t    *testing.T
	conn *websocket.Conn
	mu   sync.Mutex
	box  []game.ServerMessage
}

func dial(t *testing.T, url string) *client {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	c := &client{t: t, conn: conn}
	go c.read()
	t.Cleanup(func() { _ = conn.Close() })
	return c
}

func (c *client) read() {
	for {
		var m game.ServerMessage
		if err := c.conn.ReadJSON(&m); err != nil {
			return
		}
		c.mu.Lock()
		c.box = append(c.box, m)
		c.mu.Unlock()
	}
}

func (c *client) send(typ string, data any) {
	var raw json.RawMessage
	if data != nil {
		b, _ := json.Marshal(data)
		raw = b
	}
	if err := c.conn.WriteJSON(game.ClientMessage{Type: typ, Data: raw}); err != nil {
		c.t.Fatalf("send %s: %v", typ, err)
	}
}

func (c *client) wait(typ string, timeout time.Duration) game.ServerMessage {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		for i, m := range c.box {
			if m.Type == typ {
				c.box = append(c.box[:i], c.box[i+1:]...)
				c.mu.Unlock()
				return m
			}
		}
		c.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	c.mu.Lock()
	types := make([]string, len(c.box))
	for i, m := range c.box {
		types[i] = m.Type
	}
	c.mu.Unlock()
	c.t.Fatalf("timeout waiting for %q; buffered=%v", typ, types)
	return game.ServerMessage{}
}

func decode[T any](t *testing.T, m game.ServerMessage) T {
	t.Helper()
	var out T
	b, _ := json.Marshal(m.Data)
	_ = json.Unmarshal(b, &out)
	return out
}

func joinAs(t *testing.T, c *client, nick string) string {
	c.send("join", game.JoinData{Nickname: nick})
	m := c.wait("joined", 2*time.Second)
	id, _ := decode[map[string]any](t, m)["player_id"].(string)
	return id
}

// =============================================================================
// Tests
// =============================================================================

func TestWS_JoinThenRoomState(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)
	c := dial(t, wsBase+"/ws/rooms/"+code)
	if joinAs(t, c, "Ales") == "" {
		t.Fatal("no id")
	}
	st := decode[game.RoomState](t, c.wait("room_state", 2*time.Second))
	if st.Phase != game.PhaseLobby {
		t.Errorf("phase=%q", st.Phase)
	}
	if st.Players[0].Nickname != "Ales" {
		t.Errorf("nick=%q", st.Players[0].Nickname)
	}
}

func TestWS_ValidChainAdvances(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)
	a := dial(t, wsBase+"/ws/rooms/"+code)
	aID := joinAs(t, a, "Ales")
	b := dial(t, wsBase+"/ws/rooms/"+code)
	joinAs(t, b, "Yana")
	a.wait("room_state", time.Second) // 2 players

	a.send("start", nil)
	turn := decode[game.TurnStarted](t, a.wait("turn_started", 2*time.Second))
	actor, other := a, b
	if turn.PlayerID != aID {
		actor, other = b, a
	}
	// First move: any city. Actor submits "Мінск".
	actor.send("submit_city", game.CityData{City: "Мінск"})
	added := decode[game.ChainAdded](t, other.wait("chain_added", 3*time.Second))
	if added.Entry.City != "Мінск" {
		t.Errorf("entry city=%q", added.Entry.City)
	}
	if added.NextRequiredLetter != "К" {
		t.Errorf("next letter=%q want К", added.NextRequiredLetter)
	}
}

func TestWS_WrongLetterRejected(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)
	a := dial(t, wsBase+"/ws/rooms/"+code)
	aID := joinAs(t, a, "Ales")
	b := dial(t, wsBase+"/ws/rooms/"+code)
	bID := joinAs(t, b, "Yana")
	a.wait("room_state", time.Second)

	a.send("start", nil)
	turn := decode[game.TurnStarted](t, a.wait("turn_started", 2*time.Second))

	var actor, other *client
	var otherID string
	if turn.PlayerID == aID {
		actor, other, otherID = a, b, bID
	} else {
		actor, other, otherID = b, a, aID
	}

	actor.send("submit_city", game.CityData{City: "\u041c\u0456\u043d\u0441\u043a"})
	// Wait until the server has processed the actor's move and the turn
	// actually belongs to `other`. Without this, both submissions race and
	// the server may drop the second one (it arrives while it's still the
	// actor's turn).
	other.waitTurnFor(otherID, 3*time.Second)

	// \u041c\u0456\u043d\u0441\u043a ends in \u041a; the other player submits a city starting with \u041c.
	other.send("submit_city", game.CityData{City: "\u041c\u0430\u0433\u0456\u043b\u0451\u045e"})
	rej := decode[game.ChainRejected](t, other.wait("chain_rejected", 2*time.Second))
	if rej.Reason != "wrong_letter" {
		t.Errorf("reason=%q", rej.Reason)
	}
}

func TestWS_UnknownCityRejected(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)
	a := dial(t, wsBase+"/ws/rooms/"+code)
	aID := joinAs(t, a, "Ales")
	b := dial(t, wsBase+"/ws/rooms/"+code)
	bID := joinAs(t, b, "Yana")
	a.wait("room_state", time.Second)

	a.send("start", nil)
	turn := decode[game.TurnStarted](t, a.wait("turn_started", 2*time.Second))

	var actor, other *client
	var otherID string
	if turn.PlayerID == aID {
		actor, other, otherID = a, b, bID
	} else {
		actor, other, otherID = b, a, aID
	}

	actor.send("submit_city", game.CityData{City: "\u041c\u0456\u043d\u0441\u043a"})
	other.waitTurnFor(otherID, 3*time.Second)

	// "\u041a\u0430\u043d\u0430\u0434\u0430" starts with \u041a (correct letter) but isn't a city in our DB.
	other.send("submit_city", game.CityData{City: "\u041a\u0430\u043d\u0430\u0434\u0430"})
	rej := decode[game.ChainRejected](t, other.wait("chain_rejected", 2*time.Second))
	if rej.Reason != "not_in_db" {
		t.Errorf("reason=%q", rej.Reason)
	}
}

// waitTurnFor consumes turn_started frames until one names the given player.
// This is the correct way to wait: the server broadcasts turn_started to
// everyone, so a naive type-only match can grab a stale frame for a different
// player and let the test race ahead of the server.
func (c *client) waitTurnFor(id string, timeout time.Duration) game.TurnStarted {
	c.t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		m := c.wait("turn_started", remaining)
		ts := decode[game.TurnStarted](c.t, m)
		if ts.PlayerID == id {
			return ts
		}
		// Stale frame for another player; keep waiting.
	}
	c.t.Fatalf("timeout waiting for turn of %s", id)
	return game.TurnStarted{}
}

func TestWS_PassTwiceEliminates(t *testing.T) {
	httpBase, wsBase := setupServer(t)
	code := createRoom(t, httpBase)

	a := dial(t, wsBase+"/ws/rooms/"+code)
	aID := joinAs(t, a, "Ales")
	b := dial(t, wsBase+"/ws/rooms/"+code)
	bID := joinAs(t, b, "Yana")
	a.wait("room_state", time.Second) // 2 players

	a.send("start", nil)

	// Figure out who goes first.
	first := decode[game.TurnStarted](t, a.wait("turn_started", 2*time.Second))
	var actor, other *client
	var actorID, otherID string
	if first.PlayerID == aID {
		actor, actorID, other, otherID = a, aID, b, bID
	} else if first.PlayerID == bID {
		actor, actorID, other, otherID = b, bID, a, aID
	} else {
		t.Fatalf("unknown first player %q", first.PlayerID)
	}

	// Pass #1 — actor misses once, turn goes to the other player.
	actor.send("pass", nil)
	other.waitTurnFor(otherID, 3*time.Second)

	// Other player submits a valid city, turn comes back to actor.
	other.send("submit_city", game.CityData{City: "Мінск"})
	actor.waitTurnFor(actorID, 3*time.Second)

	// Pass #2 — actor hits MaxMisses, gets eliminated, game ends.
	actor.send("pass", nil)

	// Either side should see player_out for the actor, then game_ended.
	po := decode[game.PlayerOut](t, actor.wait("player_out", 3*time.Second))
	if po.PlayerID != actorID {
		t.Errorf("player_out for %q, want %q", po.PlayerID, actorID)
	}
	if po.Missed < 2 {
		t.Errorf("missed=%d, want >= 2", po.Missed)
	}

	ended := decode[game.GameEnded](t, other.wait("game_ended", 3*time.Second))
	if ended.WinnerID != otherID {
		t.Errorf("winner=%q, want %q", ended.WinnerID, otherID)
	}
}
