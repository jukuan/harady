package game

import (
	"math/rand"
	"strings"
	"unicode"
)

// endingSubstitutes maps a city's final letter to the letter the next city
// must actually start with. Belarusian short vowels have a "full" form:
//
//	ў (short u) → у
//	й (short i) → і
//
// so "Магілёў" → next city on "У", "Сіднэй" → next city on "І".
var endingSubstitutes = map[rune]rune{
	'ў': 'у', 'Ў': 'У',
	'й': 'і', 'Й': 'І',
}

// skipEndings are letters that essentially can't begin a city name. When a
// city ends in one of these, the required letter is the previous letter.
//
// ь (soft sign) and ъ (hard sign) are never initial in this language family;
// ы is technically possible (Ыйджонбу) but so obscure it breaks the game.
var skipEndings = map[rune]bool{
	'ь': true, 'Ь': true,
	'ъ': true, 'Ъ': true,
	'ы': true, 'Ы': true,
}

// NormalizeCity lowercases and trims a city name for comparison / storage in
// the "already used" set.
func NormalizeCity(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// FirstLetter returns the first letter of s, upper-cased for display.
func FirstLetter(s string) rune {
	for _, r := range strings.TrimSpace(s) {
		if unicode.IsLetter(r) {
			return unicode.ToUpper(r)
		}
	}
	return 0
}

// LastMeaningfulLetter returns the letter the next city must start with.
//
// Rules:
//   - If the last letter has a substitute (ў → у, й → і), return the substitute.
//   - If the last letter is a skip letter (ь, ъ, ы), step back to the previous
//     letter and use that.
//   - Otherwise return the last letter itself.
//
// Examples:
//
//	"Мінск"      -> 'К'
//	"Магілёў"    -> 'У'  (ў → у)
//	"Сіднэй"     -> 'І'  (й → і)
//	"Гомель"     -> 'Л'  (ь is skipped)
//	"Баранавічы" -> 'Ч'  (ы is skipped)
func LastMeaningfulLetter(s string) rune {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) == 0 {
		return 0
	}
	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		if !unicode.IsLetter(r) {
			continue
		}
		if sub, ok := endingSubstitutes[r]; ok {
			return unicode.ToUpper(sub)
		}
		if skipEndings[r] {
			for j := i - 1; j >= 0; j-- {
				if unicode.IsLetter(runes[j]) {
					return unicode.ToUpper(runes[j])
				}
			}
			// No previous letter — fall through and use the skip letter itself.
		}
		return unicode.ToUpper(r)
	}
	return 0
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