package game

import "math/rand"

// maskCity replaces every letter with '_' but keeps spaces, '-' and apostrophes.
func maskCity(name string) string {
	if name == "" {
		return ""
	}
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch r {
		case ' ', '-', '\'':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

const randAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randAlphabet[rand.Intn(len(randAlphabet))]
	}
	return string(b)
}

var botNicknames = []string{
	"Bot Ales", "Bot Yas", "Bot Vasya", "Bot Hanna",
	"Bot Zmitser", "Bot Kazimir", "Bot Alesya", "Bot Minsk",
}

func pickBotName(existing map[string]*Player) string {
	used := make(map[string]bool, len(existing))
	for _, p := range existing {
		used[p.Nickname] = true
	}
	for _, n := range botNicknames {
		if !used[n] {
			return n
		}
	}
	return "Bot-" + randString(4)
}
