package util

import "strings"

// SplitList splits "a;b;c" into []string{"a","b","c"} and trims empties.
// Returns nil if the input is empty or contains no non-empty items.
func SplitList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	var out []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
