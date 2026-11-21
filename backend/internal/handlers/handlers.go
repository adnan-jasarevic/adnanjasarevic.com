package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/github"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

type ProjectStore interface {
	ListPublishedProjects(ctx context.Context) ([]store.Project, error)
	ListProjects(ctx context.Context) ([]store.Project, error)
	CreateProject(ctx context.Context, p store.NewProject) (store.Project, error)
	UpdateProject(ctx context.Context, id int64, p store.ProjectPatch) (store.Project, error)
	DeleteProject(ctx context.Context, id int64) error
}

type CVStore interface {
	GetCV(ctx context.Context) (store.CV, error)
}

type ContactStore interface {
	CreateContactMessage(ctx context.Context, m store.NewContactMessage) (int64, error)
	ListContactMessages(ctx context.Context, unreadOnly bool) ([]store.ContactMessage, error)
	SetContactMessageRead(ctx context.Context, id int64, read bool) error
	DeleteContactMessage(ctx context.Context, id int64) error
}

type AdminSessionStore interface {
	CreateAdminSession(ctx context.Context, tokenHash []byte, githubID int64, githubLogin string, expiresAt time.Time) error
	GetAdminSession(ctx context.Context, tokenHash []byte) (store.AdminSession, error)
	DeleteAdminSession(ctx context.Context, tokenHash []byte) error
}

type GitHubAuth interface {
	LoginURL(state string) string
	UserFromCode(ctx context.Context, code string) (github.User, error)
}

type Notifier interface {
	NotifyContact(ctx context.Context, id int64, m store.NewContactMessage) error
}

type Store interface {
	ProjectStore
	CVStore
	ContactStore
	AdminSessionStore
}

type Config struct {
	TrustProxy     bool
	Notifier       Notifier
	AllowedOrigins []string
	GitHub         GitHubAuth
	AdminGithubID  int64
	AdminURL       string
	SecureCookies  bool
}

type API struct {
	db             Store
	trustProxy     bool
	notifier       Notifier
	allowedOrigins map[string]bool
	github         GitHubAuth
	adminGithubID  int64
	adminURL       string
	secureCookies  bool
	contactLimiter *rateLimiter
	background     sync.WaitGroup
}

func New(db Store, cfg Config) *API {
	origins := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		origins[o] = true
	}
	return &API{
		db:             db,
		trustProxy:     cfg.TrustProxy,
		notifier:       cfg.Notifier,
		allowedOrigins: origins,
		github:         cfg.GitHub,
		adminGithubID:  cfg.AdminGithubID,
		adminURL:       cfg.AdminURL,
		secureCookies:  cfg.SecureCookies,
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

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /projects", a.listProjects)
	mux.HandleFunc("GET /cv", a.getCV)
	mux.HandleFunc("POST /contact", a.rateLimit(a.contactLimiter, a.createContact))

	mux.HandleFunc("GET /auth/github/login", a.githubLogin)
	mux.HandleFunc("GET /auth/github/callback", a.githubCallback)
	mux.HandleFunc("POST /auth/logout", a.logout)
	mux.HandleFunc("GET /auth/me", a.requireAdmin(a.me))

	mux.HandleFunc("GET /admin/messages", a.requireAdmin(a.listMessages))
	mux.HandleFunc("PATCH /admin/messages/{id}", a.requireAdmin(a.updateMessage))
	mux.HandleFunc("DELETE /admin/messages/{id}", a.requireAdmin(a.deleteMessage))

	mux.HandleFunc("GET /admin/projects", a.requireAdmin(a.listAdminProjects))
	mux.HandleFunc("POST /admin/projects", a.requireAdmin(a.createProject))
	mux.HandleFunc("PATCH /admin/projects/{id}", a.requireAdmin(a.updateProject))
	mux.HandleFunc("DELETE /admin/projects/{id}", a.requireAdmin(a.deleteProject))
	return a.cors(mux)
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

func writeFieldErrors(w http.ResponseWriter, status int, fields map[string]string) {
	writeJSON(w, status, map[string]any{"error": "validation failed", "fields": fields})
}

// decodeJSON reads a JSON body of at most maxBytes into dst. On failure it writes the error response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return false
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}

// pathID parses the {id} path value. On failure it writes a 404 and returns false.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, "not found")
		return 0, false
	}
	return id, true
}
