package game

import "encoding/json"

type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type ServerMessage struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type JoinData struct {
	Nickname string `json:"nickname"`
}

type CityData struct {
	City string `json:"city"`
}

type PlayerView struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	IsBot    bool   `json:"is_bot"`
	IsHost   bool   `json:"is_host"`
	Missed   int    `json:"missed"`
	Out      bool   `json:"out"`
	Online   bool   `json:"online"`
}

type ChainEntry struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	City     string `json:"city"`
	IsBot    bool   `json:"is_bot"`
}

type RoomState struct {
	Code            string       `json:"code"`
	HostID          string       `json:"host_id"`
	Phase           string       `json:"phase"` // lobby | playing | ended
	Players         []PlayerView `json:"players"`
	Chain           []ChainEntry `json:"chain"`
	RequiredLetter  string       `json:"required_letter"`
	CurrentTurnID   string       `json:"current_turn_id"`
	CurrentTurnNick string       `json:"current_turn_nickname"`
	MinPlayers      int          `json:"min_players"`
	MaxPlayers      int          `json:"max_players"`
	MaxMisses       int          `json:"max_misses"`
}

type TurnStarted struct {
	PlayerID       string `json:"player_id"`
	Nickname       string `json:"nickname"`
	RequiredLetter string `json:"required_letter"`
	Round          int    `json:"round"`
}

type ChainAdded struct {
	Entry              ChainEntry `json:"entry"`
	NextRequiredLetter string     `json:"next_required_letter"`
}

type ChainRejected struct {
	Reason string `json:"reason"` // wrong_letter | already_used | not_in_db | empty
	Text   string `json:"text"`
}

type PlayerOut struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	Missed   int    `json:"missed"`
}

type PlayerPassed struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	Missed   int    `json:"missed"`
	IsBot    bool   `json:"is_bot"`
}


type GameEnded struct {
	WinnerID   string       `json:"winner_id,omitempty"`
	WinnerNick string       `json:"winner_nickname,omitempty"`
	Players    []PlayerView `json:"players"`
	Chain      []ChainEntry `json:"chain"`
}

type CityLearned struct {
	City   string `json:"city"`
	Region string `json:"region,omitempty"`
}

type ReactData struct {
	Emoji string `json:"emoji"`
}

type ReactionData struct {
	PlayerID   string `json:"player_id"`
	Nickname   string `json:"nickname"`
	Emoji      string `json:"emoji"`
	ChainIndex int    `json:"chain_index"`
	Ts         int64  `json:"ts"`
}

type ErrorData struct {
	Message string `json:"message"`
}
