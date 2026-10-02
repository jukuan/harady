package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DBPath      string
	CORSOrigins []string
	Language    string
	RoomIdleTTL time.Duration
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		Addr:        env("HARADY_ADDR", ":8080"),
		DBPath:      env("HARADY_DB_PATH", "./data/harady.db"),
		CORSOrigins: splitCSV(env("HARADY_CORS_ORIGINS", "http://localhost:5173")),
		Language:    env("HARADY_LANGUAGE", "be"),
		RoomIdleTTL: parseDur(env("HARADY_ROOM_IDLE_TTL", "2h"), 2*time.Hour),
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
