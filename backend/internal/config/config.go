package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DBPath      string
	LogPath     string
	LogLevel    string
	CORSOrigins []string
	Language    string
	RoomIdleTTL time.Duration
	TurnTimeout time.Duration
	BotDelayMS  int
	AdminToken  string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Addr:        env("HARADY_ADDR", ":8080"),
		DBPath:      env("HARADY_DB_PATH", "./data/harady.db"),
		LogPath:     env("HARADY_LOG_PATH", "./data/harady.log"),
		LogLevel:    env("HARADY_LOG_LEVEL", "info"),
		CORSOrigins: splitCSV(env("HARADY_CORS_ORIGINS", "http://localhost:5173")),
		Language:    env("HARADY_LANGUAGE", "be"),
		RoomIdleTTL: parseDur(env("HARADY_ROOM_IDLE_TTL", "2h"), 2*time.Hour),
		TurnTimeout: parseDur(env("HARADY_TURN_TIMEOUT", "120s"), 120*time.Second),
		BotDelayMS:  parseInt(env("HARADY_BOT_DELAY_MS", "600"), 600),
		AdminToken:  env("HARADY_ADMIN_TOKEN", ""),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDur(s string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

func parseInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}
