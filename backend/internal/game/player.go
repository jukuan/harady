package game

import (
	"sync"
	"time"
)

type ClientHandle interface {
	Send(msg ServerMessage)
}

type Player struct {
	ID       string
	Nickname string
	IsBot    bool
	IsHost   bool
	Missed   int
	Out      bool

	// Three-strikes bookkeeping. See Room.SubmitCity.
	// Guarded by Room.mu (not the client mutex).
	LastUnknown  string // normalized name last seen as not_in_db
	UnknownCount int    // how many times in a row the player typed it

	// Reaction rate limit. Guarded by Room.mu.
	LastReactAt time.Time

	mu     sync.RWMutex
	client ClientHandle
}

func (p *Player) Send(m ServerMessage) {
	if p.IsBot {
		return
	}
	p.mu.RLock()
	c := p.client
	p.mu.RUnlock()
	if c != nil {
		c.Send(m)
	}
}

func (p *Player) View() PlayerView {
	p.mu.RLock()
	online := p.client != nil
	p.mu.RUnlock()
	return PlayerView{
		ID:       p.ID,
		Nickname: p.Nickname,
		IsBot:    p.IsBot,
		IsHost:   p.IsHost,
		Missed:   p.Missed,
		Out:      p.Out,
		Online:   p.IsBot || online,
	}
}
