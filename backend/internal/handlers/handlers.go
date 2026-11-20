package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

type ProjectStore interface {
	ListPublishedProjects(ctx context.Context) ([]store.Project, error)
}

type CVStore interface {
	GetCV(ctx context.Context) (store.CV, error)
}

type ContactStore interface {
	CreateContactMessage(ctx context.Context, m store.NewContactMessage) (int64, error)
}

type Notifier interface {
	NotifyContact(ctx context.Context, id int64, m store.NewContactMessage) error
}

type Store interface {
	ProjectStore
	CVStore
	ContactStore
}

type Config struct {
	TrustProxy bool
	Notifier   Notifier
}

type API struct {
	db             Store
	trustProxy     bool
	notifier       Notifier
	contactLimiter *rateLimiter
	background     sync.WaitGroup
}

func New(db Store, cfg Config) *API {
	return &API{
		db:             db,
		trustProxy:     cfg.TrustProxy,
		notifier:       cfg.Notifier,
		contactLimiter: newRateLimiter(5, 10*time.Minute),
	}
}

// WaitBackground waits for background work, such as email notifications, to finish.
func (a *API) WaitBackground(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		a.background.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("background work did not finish: %w", ctx.Err())
	}
}

func (a *API) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /projects", a.listProjects)
	mux.HandleFunc("GET /cv", a.getCV)
	mux.HandleFunc("POST /contact", a.rateLimit(a.contactLimiter, a.createContact))
	return mux
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
