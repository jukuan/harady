package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/db"
	"github.com/jukuan/harady/backend/internal/email"
	"github.com/jukuan/harady/backend/internal/game"
	"github.com/jukuan/harady/backend/internal/httpapi"
	"github.com/jukuan/harady/backend/internal/store"
)

func main() {
	cfg := config.Load()

	// Configure slog: JSON format, level from HARADY_LOG_LEVEL.
	// Logs go to stdout, which systemd captures into the journal. No file
	// rotation is needed — journald handles retention and rotation at the
	// system level (`journalctl -u harady` is the viewer).
	level := parseLogLevel(cfg.LogLevel)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	game.SetTurnTimeout(cfg.TurnTimeout)
	game.SetBotDelay(cfg.BotDelayMS)

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	cityStore := store.NewCityStore(conn)
	if n, _ := cityStore.Count(); n == 0 {
		slog.Warn("cities table is empty — run `harady-cli seed`")
	}

	hub := game.NewHub(cfg, cityStore)

	mailer := email.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	if cfg.SMTPHost == "" {
		slog.Warn("SMTP not configured — magic links will be logged, not emailed")
	}
	adminStore := store.NewAdminStore(conn)

	router := httpapi.NewRouter(cfg, hub, cityStore, adminStore, mailer, conn)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("harady: listening",
			"addr", cfg.Addr,
			"language", cfg.Language,
			"turn_timeout", cfg.TurnTimeout.String(),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	hub.Shutdown()
	time.Sleep(200 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	slog.Info("bye")
}

// parseLogLevel maps a string from HARADY_LOG_LEVEL to a slog.Level.
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
