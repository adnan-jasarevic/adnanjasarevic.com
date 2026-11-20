package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/config"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/handlers"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	cfg := config.Load()

	db, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlers.New(db).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	return srv.ListenAndServe()
}
