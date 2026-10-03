package models

import "time"

// City is one playable answer.
// The whole app is single-language, so there is one `Name` field only.
type City struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Region    string    `json:"region,omitempty"`
	Clues     []string  `json:"clues"`
	Source    string    `json:"source,omitempty"` // "seed" | "learned"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
