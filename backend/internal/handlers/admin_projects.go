package handlers

import (
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

const maxProjectBodyBytes = 128 << 10

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// projectRequest is used for both create and update. Nil fields were not sent.
// An empty repoUrl or liveUrl clears the link.
type projectRequest struct {
	Title       *string   `json:"title"`
	Slug        *string   `json:"slug"`
	Summary     *string   `json:"summary"`
	Description *string   `json:"description"`
	TechStack   *[]string `json:"techStack"`
	RepoURL     *string   `json:"repoUrl"`
	LiveURL     *string   `json:"liveUrl"`
	SortOrder   *int      `json:"sortOrder"`
	Published   *bool     `json:"published"`
}

func (p *projectRequest) normalize() {
	for _, s := range []*string{p.Title, p.Slug, p.Summary, p.Description, p.RepoURL, p.LiveURL} {
		if s != nil {
			*s = strings.TrimSpace(*s)
		}
	}
	if p.TechStack != nil {
		for i, tech := range *p.TechStack {
			(*p.TechStack)[i] = strings.TrimSpace(tech)
		}
	}
}

// validate checks the fields that were sent. With creating set, title and slug are required.
func (p projectRequest) validate(creating bool) map[string]string {
	errs := map[string]string{}

	switch {
	case p.Title == nil && creating, p.Title != nil && *p.Title == "":
		errs["title"] = "is required"
	case p.Title != nil && utf8.RuneCountInString(*p.Title) > 200:
		errs["title"] = "must be at most 200 characters"
	}

	switch {
	case p.Slug == nil && creating, p.Slug != nil && *p.Slug == "":
		errs["slug"] = "is required"
	case p.Slug != nil && len(*p.Slug) > 100:
		errs["slug"] = "must be at most 100 characters"
	case p.Slug != nil && !slugPattern.MatchString(*p.Slug):
		errs["slug"] = "must be lowercase letters, digits and single hyphens"
	}

	if p.Summary != nil && utf8.RuneCountInString(*p.Summary) > 500 {
		errs["summary"] = "must be at most 500 characters"
	}
	if p.Description != nil && utf8.RuneCountInString(*p.Description) > 20000 {
		errs["description"] = "must be at most 20000 characters"
	}

	if p.TechStack != nil {
		if len(*p.TechStack) > 30 {
			errs["techStack"] = "must have at most 30 items"
		}
		for _, tech := range *p.TechStack {
			if tech == "" || utf8.RuneCountInString(tech) > 50 {
				errs["techStack"] = "items must be 1 to 50 characters"
				break
			}
		}
	}

	if p.RepoURL != nil && *p.RepoURL != "" && !isWebURL(*p.RepoURL) {
		errs["repoUrl"] = "must be an http or https URL"
	}
	if p.LiveURL != nil && *p.LiveURL != "" && !isWebURL(*p.LiveURL) {
		errs["liveUrl"] = "must be an http or https URL"
	}

	if p.SortOrder != nil && (*p.SortOrder < math.MinInt32 || *p.SortOrder > math.MaxInt32) {
		errs["sortOrder"] = "is out of range"
	}

	return errs
}

func isWebURL(s string) bool {
	if len(s) > 2048 {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (p projectRequest) toNew() store.NewProject {
	return store.NewProject{
		Title:       valueOrZero(p.Title),
		Slug:        valueOrZero(p.Slug),
		Summary:     valueOrZero(p.Summary),
		Description: valueOrZero(p.Description),
		TechStack:   valueOrZero(p.TechStack),
		RepoURL:     nilIfEmpty(p.RepoURL),
		LiveURL:     nilIfEmpty(p.LiveURL),
		SortOrder:   valueOrZero(p.SortOrder),
		Published:   valueOrZero(p.Published),
	}
}

func valueOrZero[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func nilIfEmpty(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func (a *API) listAdminProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := a.db.ListProjects(r.Context())
	if err != nil {
		log.Printf("list admin projects: %v", err)
		writeError(w, http.StatusInternalServerError, "could not load projects")
		return
	}
	if projects == nil {
		projects = []store.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}

func (a *API) createProject(w http.ResponseWriter, r *http.Request) {
	var req projectRequest
	if !decodeJSON(w, r, &req, maxProjectBodyBytes) {
		return
	}
	req.normalize()
	if errs := req.validate(true); len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}

	project, err := a.db.CreateProject(r.Context(), req.toNew())
	if errors.Is(err, store.ErrSlugTaken) {
		writeFieldErrors(w, http.StatusConflict, map[string]string{"slug": "is already taken"})
		return
	}
	if err != nil {
		log.Printf("create project: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create project")
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (a *API) updateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req projectRequest
	if !decodeJSON(w, r, &req, maxProjectBodyBytes) {
		return
	}
	if req == (projectRequest{}) {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	req.normalize()
	if errs := req.validate(false); len(errs) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, errs)
		return
	}

	project, err := a.db.UpdateProject(r.Context(), id, store.ProjectPatch(req))
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "project not found")
	case errors.Is(err, store.ErrSlugTaken):
		writeFieldErrors(w, http.StatusConflict, map[string]string{"slug": "is already taken"})
	case err != nil:
		log.Printf("update project %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "could not update project")
	default:
		writeJSON(w, http.StatusOK, project)
	}
}

func (a *API) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := a.db.DeleteProject(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		log.Printf("delete project %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "could not delete project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
