package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

type ProjectStore interface {
	ListPublishedProjects(ctx context.Context) ([]store.Project, error)
}

type CVStore interface {
	GetCV(ctx context.Context) (store.CV, error)
}

type ContactStore interface {
	CreateContactMessage(ctx context.Context, m store.NewContactMessage) error
}

type Store interface {
	ProjectStore
	CVStore
	ContactStore
}

type API struct {
	db Store
}

func New(db Store) *API {
	return &API{db: db}
}

func (a *API) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /projects", a.listProjects)
	mux.HandleFunc("GET /cv", a.getCV)
	mux.HandleFunc("POST /contact", a.createContact)
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
