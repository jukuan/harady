package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/jukuan/harady/backend/internal/game"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 4096
	sendBuffer     = 64
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Permit all origins; tighten via a reverse proxy in production if needed.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client is the per-connection WebSocket handler. It implements
// game.ClientHandle so the game package can push messages to it.
type Client struct {
	conn   *websocket.Conn
	send   chan game.ServerMessage
	room   *game.Room
	player *game.Player
}

func ServeRoomWS(hub *game.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		room, err := hub.GetRoom(c.Param("code"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		cl := &Client{conn: conn, send: make(chan game.ServerMessage, sendBuffer), room: room}
		go cl.writePump()
		cl.readPump()
	}
}

func (c *Client) Send(m game.ServerMessage) {
	select {
	case c.send <- m:
	default:
	}
}

func (c *Client) readPump() {
	defer func() {
		if c.player != nil {
			c.room.Leave(c.player.ID)
		}
		close(c.send)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg game.ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: "bad json"}})
			continue
		}
		c.dispatch(msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) dispatch(msg game.ClientMessage) {
	switch msg.Type {
	case "join":
		c.handleJoin(msg)
	case "start":
		if c.player == nil {
			return
		}
		if err := c.room.Start(c.player.ID); err != nil {
			c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: err.Error()}})
		}
	case "add_bot":
		if c.player == nil {
			return
		}
		if _, err := c.room.AddBotBy(c.player.ID); err != nil {
			c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: err.Error()}})
			return
		}
		c.room.BroadcastState()
	case "end_game":
		if c.player != nil {
			c.room.ForceEnd(c.player.ID)
		}
	case "submit_city":
		if c.player == nil {
			return
		}
		var d game.CityData
		_ = json.Unmarshal(msg.Data, &d)
		c.room.SubmitCity(c.player.ID, d.City)
	case "pass":
		if c.player != nil {
			c.room.Pass(c.player.ID)
		}
	default:
		c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: "unknown message: " + msg.Type}})
	}
}

func (c *Client) handleJoin(msg game.ClientMessage) {
	if c.player != nil {
		c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: "already joined"}})
		return
	}
	var d game.JoinData
	_ = json.Unmarshal(msg.Data, &d)
	nick := sanitizeNick(d.Nickname)
	if nick == "" {
		c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: "nickname required"}})
		return
	}
	p := &game.Player{ID: uuid.NewString(), Nickname: nick}
	if err := c.room.Join(p, c); err != nil {
		c.Send(game.ServerMessage{Type: "error", Data: game.ErrorData{Message: err.Error()}})
		return
	}
	c.player = p
	c.Send(game.ServerMessage{Type: "joined", Data: map[string]any{
		"player_id": p.ID, "room": c.room.Code, "is_host": p.IsHost,
	}})
	c.room.BroadcastState()
}

func sanitizeNick(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > 20 {
		r = r[:20]
	}
	return string(r)
}
