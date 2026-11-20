package handlers

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	clients     map[string]*clientWindow
	lastCleanup time.Time
	now         func() time.Time
}

type clientWindow struct {
	count int
	start time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:   limit,
		window:  window,
		clients: map[string]*clientWindow{},
		now:     time.Now,
	}
}

// allow reports whether the client may make a request, and if not, how long until it may.
func (rl *rateLimiter) allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	if now.Sub(rl.lastCleanup) > rl.window {
		for k, c := range rl.clients {
			if now.Sub(c.start) >= rl.window {
				delete(rl.clients, k)
			}
		}
		rl.lastCleanup = now
	}

	c, ok := rl.clients[key]
	if !ok || now.Sub(c.start) >= rl.window {
		rl.clients[key] = &clientWindow{count: 1, start: now}
		return true, 0
	}
	if c.count >= rl.limit {
		return false, c.start.Add(rl.window).Sub(now)
	}
	c.count++
	return true, 0
}

func (a *API) rateLimit(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, retryAfter := rl.allow(a.clientIP(r))
		if !ok {
			seconds := int(retryAfter.Seconds()) + 1
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			writeError(w, http.StatusTooManyRequests, "too many requests, try again later")
			return
		}
		next(w, r)
	}
}

func (a *API) clientIP(r *http.Request) string {
	if a.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
