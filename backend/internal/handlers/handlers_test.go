package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

type fakeStore struct {
	projects []store.Project
	err      error
}

func (f *fakeStore) ListPublishedProjects(ctx context.Context) ([]store.Project, error) {
	return f.projects, f.err
}

func serve(t *testing.T, fs *fakeStore, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	New(fs).Routes().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestHealth(t *testing.T) {
	rec := serve(t, &fakeStore{}, http.MethodGet, "/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestHealthRejectsPost(t *testing.T) {
	rec := serve(t, &fakeStore{}, http.MethodPost, "/health")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestListProjects(t *testing.T) {
	fs := &fakeStore{projects: []store.Project{{ID: 1, Title: "Portfolio", Slug: "portfolio"}}}
	rec := serve(t, fs, http.MethodGet, "/projects")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got []store.Project
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(got) != 1 || got[0].Slug != "portfolio" {
		t.Errorf("projects = %+v, want one project with slug portfolio", got)
	}
}

func TestListProjectsEmptyReturnsArray(t *testing.T) {
	rec := serve(t, &fakeStore{}, http.MethodGet, "/projects")

	if got, want := rec.Body.String(), "[]\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestListProjectsStoreError(t *testing.T) {
	rec := serve(t, &fakeStore{err: errors.New("db down")}, http.MethodGet, "/projects")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
