package store

import (
	"strings"
	"unicode"
)

// levenshtein returns the edit distance between two strings, compared
// case-insensitively and rune-wise (so Cyrillic counts correctly).
func levenshtein(a, b string) int {
	ra := []rune(strings.ToLower(a))
	rb := []rune(strings.ToLower(b))
	n, m := len(ra), len(rb)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	prev := make([]int, m+1)
	cur := make([]int, m+1)
	for j := 0; j <= m; j++ {
		prev[j] = j
	}
	for i := 1; i <= n; i++ {
		cur[0] = i
		for j := 1; j <= m; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[m]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// firstLetterLower returns the lowercased first letter of s, skipping
// non-letter characters and leading whitespace. Returns 0 if none found.
func firstLetterLower(s string) rune {
	for _, r := range strings.TrimSpace(s) {
		if unicode.IsLetter(r) {
			return unicode.ToLower(r)
		}
	}
	return 0
}
