package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"huddle/internal/config"
)

func TestPowChallengeDisabled(t *testing.T) {
	cfg := config.Config{
		Addr:          ":0",
		InviteSecret:  "test",
		InviteTTL:     time.Hour,
		MaxRoomSize:   4,
		PowDifficulty: 0,
		PowTTL:        time.Minute,
	}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pow/challenge?action=create", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestPowChallengeIssuesChallenge(t *testing.T) {
	cfg := config.Config{
		Addr:          ":0",
		InviteSecret:  "test",
		InviteTTL:     time.Hour,
		MaxRoomSize:   4,
		PowDifficulty: 12,
		PowTTL:        time.Minute,
	}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pow/challenge?action=join", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["id"] == "" || body["prefix"] == "" {
		t.Fatalf("expected challenge fields, got %+v", body)
	}
}

func TestPowChallengeTrailingSlash(t *testing.T) {
	cfg := config.Config{
		Addr:          ":0",
		InviteSecret:  "test",
		InviteTTL:     time.Hour,
		MaxRoomSize:   4,
		PowDifficulty: 12,
		PowTTL:        time.Minute,
	}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pow/challenge/?action=join", nil)
	rec := httptest.NewRecorder()
	withMiddleware(false, nil, srv.limits, srv.mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUnknownAPIPathReturnsNotFound(t *testing.T) {
	cfg := config.Config{
		Addr:         ":0",
		InviteSecret: "test",
		InviteTTL:    time.Hour,
		MaxRoomSize:  4,
	}
	srv := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/pow/?action=join", nil)
	rec := httptest.NewRecorder()
	withMiddleware(false, nil, srv.limits, srv.mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
