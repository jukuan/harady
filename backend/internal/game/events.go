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

type TextData struct {
	Text string `json:"text"`
}

type PlayerView struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	IsBot    bool   `json:"is_bot"`
	IsHost   bool   `json:"is_host"`
	Score    int    `json:"score"`
	Online   bool   `json:"online"`
}

type RoomState struct {
	Code       string       `json:"code"`
	HostID     string       `json:"host_id"`
	Phase      string       `json:"phase"`
	Round      int          `json:"round"`
	ActorID    string       `json:"actor_id"`
	ActorNick  string       `json:"actor_nickname"`
	MaskedCity string       `json:"masked_city"`
	Players    []PlayerView `json:"players"`
	MinPlayers int          `json:"min_players"`
	MaxPlayers int          `json:"max_players"`
	Duration   int          `json:"duration"`
}

type RoundStarted struct {
	ActorID    string `json:"actor_id"`
	ActorNick  string `json:"actor_nickname"`
	MaskedCity string `json:"masked_city"`
	Round      int    `json:"round"`
	Duration   int    `json:"duration"`
}

type TickData struct {
	SecondsLeft int `json:"seconds_left"`
	Round       int `json:"round"`
	Duration    int `json:"duration"`
}

type YourWord struct {
	City   string `json:"city"`
	Region string `json:"region,omitempty"`
}

type ChatMessage struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
	Kind     string `json:"kind"` // clue | guess
	IsBot    bool   `json:"is_bot"`
	Ts       int64  `json:"ts"`
}

type CorrectGuess struct {
	PlayerID string `json:"player_id"`
	Nickname string `json:"nickname"`
	City     string `json:"city"`
}

type RoundEnded struct {
	WinnerID string       `json:"winner_id,omitempty"`
	City     string       `json:"city"`
	Scores   []PlayerView `json:"scores"`
}

type ErrorData struct {
	Message string `json:"message"`
}
