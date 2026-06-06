package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"huddle/internal/config"
)

func TestRateLimitBlocksWebSocketUpgrades(t *testing.T) {
	cfg := config.Config{
		Addr:            ":0",
		InviteSecret:    "test",
		InviteTTL:       time.Hour,
		MaxRoomSize:     4,
		RateLimitWindow: time.Minute,
		RateLimitHTTP:   1000,
		RateLimitWS:     2,
		RateLimitCreate: 1000,
		RateLimitJoin:   1000,
	}
	srv := New(cfg)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		rec := httptest.NewRecorder()
		withMiddleware(false, srv.limits, srv.mux).ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("expected upgrade attempt %d to pass, got 429", i+1)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	withMiddleware(false, srv.limits, srv.mux).ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimitSkipsHealthEndpoint(t *testing.T) {
	rl := newRateLimits(1, 1, time.Minute)
	h := rl.middleware(false, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected health check %d to pass, got %d", i+1, rec.Code)
		}
	}
}

func TestClientIPUsesForwardedForWhenTrustProxyEnabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")

	if got := clientIP(req, true); got != "203.0.113.5" {
		t.Fatalf("expected forwarded client IP, got %q", got)
	}
	if got := clientIP(req, false); got != "10.0.0.1" {
		t.Fatalf("expected remote addr IP, got %q", got)
	}
}
