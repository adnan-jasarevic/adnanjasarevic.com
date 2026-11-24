package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/config"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/email"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/github"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/handlers"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

// Cloud Run gives a container 10 seconds between SIGTERM and SIGKILL.
const shutdownTimeout = 9 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	var githubAuth handlers.GitHubAuth
	if cfg.GithubClientID != "" && cfg.GithubClientSecret != "" {
		githubAuth = github.NewOAuth(cfg.GithubClientID, cfg.GithubClientSecret, cfg.GithubRedirectURL)
	} else {
		log.Print("admin login disabled: GITHUB_CLIENT_ID or GITHUB_CLIENT_SECRET not set")
	}

	api := handlers.New(db, handlers.Config{
		ClientIPHeader: cfg.ClientIPHeader,
		Notifier:       notifier,
		AllowedOrigins: cfg.AllowedOrigins,
		GitHub:         githubAuth,
		AdminGithubID:  cfg.AdminGithubID,
		AdminURL:       cfg.AdminURL,
		SecureCookies:  strings.HasPrefix(cfg.GithubRedirectURL, "https://"),
	})
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	log.Print("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := api.WaitBackground(shutdownCtx); err != nil {
		return err
	}
	if err := <-serverErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Print("shutdown complete")
	return nil
}
