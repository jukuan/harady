package game

import (
	"log/slog"
	"math/rand"
	"time"
)

// botDelayBase is the min delay before a bot acts, in milliseconds.
// Actual delay is base + rand[0..base]. Configurable via HARADY_BOT_DELAY_MS.
var botDelayBase = 600

// SetBotDelay is called once at startup from main.
func SetBotDelay(ms int) {
	if ms < 0 {
		ms = 0
	}
	botDelayBase = ms
}

// runBotTurn: waits a beat to feel human, then tries a list of candidate
// cities until one is accepted. Falls back to Pass if none work. Always
// releases the turn.
func (r *Room) runBotTurn(p *Player) {
	// A panic must never leave the turn stuck.
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("bot panic", "bot", p.Nickname, "panic", rec)
			r.Pass(p.ID)
		}
	}()

	d := 0
	if botDelayBase > 0 {
		d = botDelayBase + rand.Intn(botDelayBase+1)
	}
	if d > 0 {
		time.Sleep(time.Duration(d) * time.Millisecond)
	}

	r.mu.RLock()
	if r.Phase != PhasePlaying || r.turnID != p.ID {
		r.mu.RUnlock()
		slog.Debug("bot aborted (turn moved on)", "bot", p.Nickname)
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
		slog.Debug("bot chose to pass", "bot", p.Nickname)
		r.Pass(p.ID)
		return
	}

	// Build a candidate pool. We ask for more than one so we can retry if
	// our first pick is rejected (race with another player, duplicate
	// canonicalization, etc.).
	var candidates []string
	if letter == 0 {
		cities, err := r.Cities.RandomMany(5)
		if err == nil {
			for _, c := range cities {
				candidates = append(candidates, c.Name)
			}
		}
	} else {
		names, err := r.Cities.FindByFirstLetter(letter, used, 8)
		if err == nil {
			candidates = names
		}
	}

	if len(candidates) == 0 {
		slog.Warn("bot has no candidates — passing",
			"room", r.Code, "bot", p.Nickname, "letter", string(letter))
		r.Pass(p.ID)
		return
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for _, name := range candidates {
		slog.Debug("bot tries", "bot", p.Nickname, "city", name)
		if r.SubmitCity(p.ID, name) {
			return
		}
		// If the turn has already moved on, stop retrying.
		r.mu.RLock()
		stillMine := r.Phase == PhasePlaying && r.turnID == p.ID
		r.mu.RUnlock()
		if !stillMine {
			return
		}
	}

	slog.Warn("bot exhausted candidates — passing",
		"room", r.Code, "bot", p.Nickname, "tried", len(candidates))
	r.Pass(p.ID)
}
