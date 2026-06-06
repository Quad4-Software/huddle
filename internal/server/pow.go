package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"huddle/internal/pow"
)

type powHandler struct {
	store      *pow.Store
	difficulty int
	ttl        time.Duration
}

func (h *powHandler) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if h.difficulty <= 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	action := strings.TrimSpace(r.URL.Query().Get("action"))
	if action != "create" && action != "join" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	ch, err := h.store.Issue(action, h.difficulty, h.ttl)
	if err != nil {
		http.Error(w, "challenge unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":         ch.ID,
		"prefix":     ch.Prefix,
		"difficulty": ch.Difficulty,
		"expiresAt":  ch.ExpiresAt.Unix(),
	})
}
