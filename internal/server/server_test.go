package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"huddle/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Config{Addr: ":0", InviteSecret: "test", InviteTTL: 0, MaxRoomSize: 4}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected ok status, got %q", body["status"])
	}
}
