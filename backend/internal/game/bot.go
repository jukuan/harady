package game

import (
	"math/rand"
	"time"
)

// runBotTurn: waits a beat to feel human, then either submits a city or
// (rarely, and never on the opening move) gives up.
func (r *Room) runBotTurn(p *Player) {
	time.Sleep(time.Duration(300 + rand.Intn(900)) * time.Millisecond)

	r.mu.RLock()
	if r.Phase != PhasePlaying || r.turnID != p.ID {
		r.mu.RUnlock()
		return
	}
	if bp := r.players[p.ID]; bp == nil || bp.Out {
		r.mu.RUnlock()
		return
	}
	letter := r.required
	chainLen := len(r.chain)
	used := make([]string, 0, len(r.chainSet))
	for k := range r.chainSet {
		used = append(used, k)
	}
	r.mu.RUnlock()

	// Never pass on the opening move — the first player can play any city.
	// After that, pass with a small probability so the bot is beatable.
	if chainLen > 0 && rand.Float64() < 0.05 {
		r.Pass(p.ID)
		return
	}

	var pick string
	if letter == 0 {
		cities, err := r.Cities.RandomMany(1)
		if err == nil && len(cities) > 0 {
			pick = cities[0].Name
		}
	} else {
		names, err := r.Cities.FindByFirstLetter(letter, used, 5)
		if err == nil && len(names) > 0 {
			pick = names[rand.Intn(len(names))]
		}
	}
	if pick == "" {
		r.Pass(p.ID)
		return
	}
	r.SubmitCity(p.ID, pick)
}
