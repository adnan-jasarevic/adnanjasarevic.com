package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/github"
	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
)

const (
	sessionCookie   = "admin_session"
	stateCookie     = "oauth_state"
	sessionDuration = 7 * 24 * time.Hour
	stateDuration   = 10 * time.Minute
)

type contextKey int

const sessionKey contextKey = iota

func newToken() string {
	return rand.Text()
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (a *API) setCookie(w http.ResponseWriter, name, value, path string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   a.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) clearCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *API) githubLogin(w http.ResponseWriter, r *http.Request) {
	if a.github == nil {
		writeError(w, http.StatusServiceUnavailable, "admin login is not configured")
		return
	}
	state := newToken()
	a.setCookie(w, stateCookie, state, "/auth/github", stateDuration)
	http.Redirect(w, r, a.github.LoginURL(state), http.StatusFound)
}

func (a *API) githubCallback(w http.ResponseWriter, r *http.Request) {
	if a.github == nil {
		writeError(w, http.StatusServiceUnavailable, "admin login is not configured")
		return
	}

	cookie, err := r.Cookie(stateCookie)
	a.clearCookie(w, stateCookie, "/auth/github")
	state := r.URL.Query().Get("state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		a.redirectToAdmin(w, r, "invalid_state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		a.redirectToAdmin(w, r, "cancelled")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	user, err := a.github.UserFromCode(ctx, code)
	if err != nil {
		log.Printf("github login: %v", err)
		a.redirectToAdmin(w, r, "error")
		return
	}
	if user.ID != a.adminGithubID {
		log.Printf("github login refused for %s (id %d)", user.Login, user.ID)
		a.redirectToAdmin(w, r, "not_allowed")
		return
	}

	token := newToken()
	err = a.db.CreateAdminSession(r.Context(), hashToken(token), user.ID, user.Login, time.Now().Add(sessionDuration))
	if err != nil {
		log.Printf("create admin session: %v", err)
		a.redirectToAdmin(w, r, "error")
		return
	}
	a.setCookie(w, sessionCookie, token, "/", sessionDuration)
	a.redirectToAdmin(w, r, "")
}

func (a *API) redirectToAdmin(w http.ResponseWriter, r *http.Request, loginError string) {
	target := a.adminURL
	if loginError != "" {
		target += "?login_error=" + url.QueryEscape(loginError)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := a.db.DeleteAdminSession(r.Context(), hashToken(cookie.Value)); err != nil {
			log.Printf("delete admin session: %v", err)
			writeError(w, http.StatusInternalServerError, "could not log out")
			return
		}
	}
	a.clearCookie(w, sessionCookie, "/")
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(sessionKey).(store.AdminSession)
	writeJSON(w, http.StatusOK, map[string]any{
		"githubId":  session.GithubID,
		"login":     session.GithubLogin,
		"expiresAt": session.ExpiresAt,
	})
}

func (a *API) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		session, err := a.db.GetAdminSession(r.Context(), hashToken(cookie.Value))
		if errors.Is(err, store.ErrNotFound) || (err == nil && session.GithubID != a.adminGithubID) {
			a.clearCookie(w, sessionCookie, "/")
			writeError(w, http.StatusUnauthorized, "not logged in")
			return
		}
		if err != nil {
			log.Printf("get admin session: %v", err)
			writeError(w, http.StatusInternalServerError, "could not check session")
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, session)
		next(w, r.WithContext(ctx))
	}
}

var _ GitHubAuth = (*github.OAuth)(nil)
