package game

import (
	"math/rand"
	"time"
)

// runBotTurn: waits a beat to feel human, then either submits a valid city or
// gives up. Occasionally passes even when it could play, so it's beatable.
func (r *Room) runBotTurn(p *Player) {
	time.Sleep(time.Duration(2000+rand.Intn(3000)) * time.Millisecond)

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
	used := make([]string, 0, len(r.chainSet))
	for k := range r.chainSet {
		used = append(used, k)
	}
	r.mu.RUnlock()

	if rand.Float64() < 0.15 {
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
