package game

import "sync"

// ClientHandle is implemented by the ws.Client. Kept as an interface so the
// game package has no dependency on the transport layer.
type ClientHandle interface {
	Send(msg ServerMessage)
}

type Player struct {
	ID       string
	Nickname string
	IsBot    bool
	IsHost   bool
	Score    int // guarded by Room.mu

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

// View builds a PlayerView; only call while holding Room.mu.
func (p *Player) View() PlayerView {
	p.mu.RLock()
	online := p.client != nil
	p.mu.RUnlock()
	return PlayerView{
		ID:       p.ID,
		Nickname: p.Nickname,
		IsBot:    p.IsBot,
		IsHost:   p.IsHost,
		Score:    p.Score,
		Online:   p.IsBot || online,
	}
}
