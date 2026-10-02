package game

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/yourname/harady/backend/internal/config"
	"github.com/yourname/harady/backend/internal/store"
)

var ErrRoomNotFound = errors.New("room not found")

type Hub struct {
	cfg    *config.Config
	cities *store.CityStore

	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub(cfg *config.Config, cities *store.CityStore) *Hub {
	h := &Hub{
		cfg:    cfg,
		cities: cities,
		rooms:  map[string]*Room{},
	}
	go h.cleanupLoop()
	return h
}

func (h *Hub) CreateRoom() *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := 0; i < 6; i++ {
		code := newRoomCode()
		if _, exists := h.rooms[code]; !exists {
			r := newRoom(code, h.cities)
			h.rooms[code] = r
			return r
		}
	}
	// Extremely unlikely; widen the code.
	code := newRoomCode() + newRoomCode()[:4]
	r := newRoom(code, h.cities)
	h.rooms[code] = r
	return r
}

func (h *Hub) GetRoom(code string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[code]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

func (h *Hub) cleanupLoop() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-h.cfg.RoomIdleTTL)
		h.mu.Lock()
		for code, r := range h.rooms {
			if r.PlayerCount() == 0 || r.IdleSince().Before(cutoff) {
				delete(h.rooms, code)
			}
		}
		h.mu.Unlock()
	}
}

func newRoomCode() string {
	return uuid.New().String()[:6]
}
