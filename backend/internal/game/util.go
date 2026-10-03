package game

import (
	"math/rand"
	"strings"
	"unicode"
)

// exceptionEndings are letters that are hard or impossible to begin a city
// name with. If a city ends in one of these, the required next letter is the
// previous character instead.
//
// Belarusian + Russian Cyrillic: ь ъ ы й ў і
// (Deliberately excludes а / е — they are both common endings AND common
// starting letters, so treating them as exceptions would break most chains.)
var exceptionEndings = map[rune]bool{
	'ь': true, 'Ь': true,
	'ъ': true, 'Ъ': true,
	'ы': true, 'Ы': true,
	'й': true, 'Й': true,
	'ў': true, 'Ў': true,
	'і': true, 'І': true,
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

// LastMeaningfulLetter returns the letter the next city must start with,
// applying the exception-endings rule.
//
//	"Мінск"   -> 'К'
//	"Гомель"  -> 'Л'  (exception: Ь)
//	"Магілёў" -> 'Ё'  (exception: Ў)
//	"Баранавічы" -> 'Ч' (exception: Ы)
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
		if exceptionEndings[r] && i > 0 {
			for j := i - 1; j >= 0; j-- {
				if unicode.IsLetter(runes[j]) {
					return unicode.ToUpper(runes[j])
				}
			}
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
