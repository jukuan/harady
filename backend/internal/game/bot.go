package game

import (
	"math/rand"
	"time"

	"github.com/yourname/harady/backend/internal/models"
)

// runActorBot: emits the city's curated clues one by one.
func (r *Room) runActorBot(p *Player, city *models.City) {
	clues := append([]string(nil), city.Clues...)
	if len(clues) == 0 {
		return
	}
	rand.Shuffle(len(clues), func(i, j int) { clues[i], clues[j] = clues[j], clues[i] })

	for _, c := range clues {
		time.Sleep(time.Duration(3500+rand.Intn(3000)) * time.Millisecond)
		if !r.botShouldAct(p.ID, true) {
			return
		}
		r.HandleClue(p.ID, c)
	}
}

// runGuesserBot: tries a random city from the DB every 8–14s.
// Statistically, with N cities, it gets the answer right about every N tries.
func (r *Room) runGuesserBot(p *Player) {
	time.Sleep(time.Duration(6000+rand.Intn(4000)) * time.Millisecond)
	for {
		if !r.botShouldAct(p.ID, false) {
			return
		}
		candidates, err := r.Cities.RandomMany(1)
		if err == nil && len(candidates) > 0 {
			r.HandleGuess(p.ID, candidates[0].Name)
		}
		time.Sleep(time.Duration(8000+rand.Intn(6000)) * time.Millisecond)
	}
}

// botShouldAct reports whether the bot is still in the room and the round is
// active. asActor additionally requires the bot to still be the actor.
func (r *Room) botShouldAct(botID string, asActor bool) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.Phase != PhaseRoundActive {
		return false
	}
	if _, ok := r.players[botID]; !ok {
		return false
	}
	if asActor {
		return r.actorIDLocked() == botID
	}
	return r.actorIDLocked() != botID
}
