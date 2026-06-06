package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustProxyHeaders(t *testing.T) {
	var gotScheme, gotHost string
	h := trustProxyHeaders(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotScheme = r.URL.Scheme
		gotHost = r.Host
	}))

	req := httptest.NewRequest(http.MethodGet, "http://ignored/ws", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "huddle.example.com")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if gotScheme != "https" {
		t.Fatalf("expected https scheme, got %q", gotScheme)
	}
	if gotHost != "huddle.example.com" {
		t.Fatalf("expected forwarded host, got %q", gotHost)
	}
}

func TestSecurityHeaders(t *testing.T) {
	var served bool
	h := securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("expected request to reach handler")
	}
	if got := rec.Header().Get("Content-Security-Policy"); got != contentSecurityPolicy {
		t.Fatalf("unexpected CSP: %q", got)
	}
	if got := rec.Header().Get("Permissions-Policy"); got != permissionsPolicy {
		t.Fatalf("unexpected permissions policy: %q", got)
	}
	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", got)
	}
}
