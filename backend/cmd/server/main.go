package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/harady/backend/internal/config"
	"github.com/yourname/harady/backend/internal/db"
	"github.com/yourname/harady/backend/internal/game"
	"github.com/yourname/harady/backend/internal/httpapi"
	"github.com/yourname/harady/backend/internal/store"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer conn.Close()

	cityStore := store.NewCityStore(conn)
	if n, _ := cityStore.Count(); n == 0 {
		log.Printf("warning: cities table is empty — run `harady-cli seed`")
	}

	hub := game.NewHub(cfg, cityStore)
	router := httpapi.NewRouter(cfg, hub, cityStore)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("harady: listening on %s (language=%s)", cfg.Addr, cfg.Language)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("bye")
}
