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

type API struct {
	projects ProjectStore
}

func New(projects ProjectStore) *API {
	return &API{projects: projects}
}

func (a *API) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /projects", a.listProjects)
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
