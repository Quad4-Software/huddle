package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"huddle/internal/config"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	var served bool
	h := withMiddleware(false, []string{"http://localhost:5173"}, newRateLimits(1000, 1000, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/pow/challenge?action=create", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("expected request to reach handler")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected ACAO header, got %q", got)
	}
}

func TestCORSMiddlewareRejectsUnknownOrigin(t *testing.T) {
	var served bool
	h := withMiddleware(false, []string{"http://localhost:5173"}, newRateLimits(1000, 1000, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/pow/challenge?action=create", nil)
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("expected request to reach handler")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no ACAO header, got %q", got)
	}
}

func TestCORSMiddlewareAllowsSameOriginWithoutConfig(t *testing.T) {
	h := withMiddleware(false, nil, newRateLimits(1000, 1000, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "http://huddle.example/api/health", nil)
	req.Host = "huddle.example"
	req.Header.Set("Origin", "http://huddle.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://huddle.example" {
		t.Fatalf("expected same-origin ACAO header, got %q", got)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	h := withMiddleware(false, []string{"http://localhost:5173"}, newRateLimits(1000, 1000, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("preflight should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/pow/challenge", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != corsAllowMethods {
		t.Fatalf("expected allow-methods header, got %q", got)
	}
}

func TestWebSocketCheckOriginRejectsCrossOriginWithoutConfig(t *testing.T) {
	srv := New(config.Config{
		Addr:         ":0",
		InviteSecret: "test",
		InviteTTL:    time.Hour,
		MaxRoomSize:  4,
	})

	req := websocketUpgradeRequest("http://evil.example", "huddle.example")
	rec := httptest.NewRecorder()
	srv.handleWS(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestWebSocketCheckOriginAllowsConfiguredOrigin(t *testing.T) {
	srv := New(config.Config{
		Addr:         ":0",
		InviteSecret: "test",
		InviteTTL:    time.Hour,
		MaxRoomSize:  4,
		CORSOrigins:  []string{"http://localhost:5173"},
	})

	req := websocketUpgradeRequest("http://localhost:5173", "huddle.example")
	rec := httptest.NewRecorder()
	srv.handleWS(rec, req)

	if rec.Code == http.StatusForbidden {
		t.Fatalf("expected upgrade attempt to pass origin check, got %d", rec.Code)
	}
}

func websocketUpgradeRequest(origin, host string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Host = host
	req.Header.Set("Origin", origin)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	return req
}
