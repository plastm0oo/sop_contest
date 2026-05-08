package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type CORS struct {
	allowedOrigin string
}

func NewCORS(allowedOrigin string) *CORS {
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return &CORS{
		allowedOrigin: allowedOrigin,
	}
}

func (c *CORS) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if c.allowedOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin == c.allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", c.allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	attempts int
	window   time.Duration

	mu      sync.Mutex
	clients map[string]*clientBucket
}

type clientBucket struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(attempts int, window time.Duration) *RateLimiter {
	if attempts <= 0 {
		attempts = 5
	}

	if window <= 0 {
		window = time.Minute
	}

	return &RateLimiter{
		attempts: attempts,
		window:   window,
		clients:  make(map[string]*clientBucket),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthLimitedRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		key := clientIP(r) + ":" + r.URL.Path

		if !rl.allow(key) {
			writeRateLimitError(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clients[key]
	if !exists || now.After(bucket.resetTime) {
		rl.clients[key] = &clientBucket{
			count:     1,
			resetTime: now.Add(rl.window),
		}

		rl.cleanupExpired(now)
		return true
	}

	if bucket.count >= rl.attempts {
		return false
	}

	bucket.count++
	return true
}

func (rl *RateLimiter) cleanupExpired(now time.Time) {
	for key, bucket := range rl.clients {
		if now.After(bucket.resetTime) {
			delete(rl.clients, key)
		}
	}
}

func isAuthLimitedRoute(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}

	return r.URL.Path == "/api/auth/register" || r.URL.Path == "/api/auth/login"
}

func clientIP(r *http.Request) string {
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func writeRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)

	_, _ = w.Write([]byte(`{"error":"слишком много попыток"}` + "\n"))
}
