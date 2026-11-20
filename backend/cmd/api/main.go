package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/config"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/email"
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

	var notifier handlers.Notifier
	if cfg.ResendAPIKey != "" && cfg.ContactNotifyTo != "" {
		notifier = email.NewResend(cfg.ResendAPIKey, cfg.ContactNotifyFrom, cfg.ContactNotifyTo)
	} else {
		log.Print("email notifications disabled: RESEND_API_KEY or CONTACT_NOTIFY_TO not set")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlers.New(db, handlers.Config{TrustProxy: cfg.TrustProxy, Notifier: notifier}).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	return srv.ListenAndServe()
}
