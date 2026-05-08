package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterLimitsAuthLogin(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	nextCalls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Middleware(next)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}

	if nextCalls != 2 {
		t.Fatalf("expected next handler to be called 2 times, got %d", nextCalls)
	}
}

func TestRateLimiterDoesNotLimitTeachersEndpoint(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	nextCalls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		w.WriteHeader(http.StatusOK)
	})

	handler := limiter.Middleware(next)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/teachers", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	if nextCalls != 3 {
		t.Fatalf("expected next handler to be called 3 times, got %d", nextCalls)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	cors := NewCORS("*")

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := cors.Middleware(next)

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	if nextCalled {
		t.Fatal("expected next handler not to be called for OPTIONS preflight")
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=*, got %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected Access-Control-Allow-Methods: %q", got)
	}
}
