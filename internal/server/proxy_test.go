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
