package handlers

import (
	"log"
	"net/http"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

func (a *API) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := a.db.ListPublishedProjects(r.Context())
	if err != nil {
		log.Printf("list projects: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load projects")
		return
	}
	if projects == nil {
		projects = []store.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}
