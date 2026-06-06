// Package config loads runtime settings from flags and environment variables.
package config

import (
	"flag"
	"os"
	"strings"
	"time"
)

// Config holds server runtime settings.
type Config struct {
	Addr             string
	InviteSecret     string
	InviteTTL        time.Duration
	MaxRoomSize      int
	TrustProxy       bool
	RateLimitWindow  time.Duration
	RateLimitHTTP    int
	RateLimitWS      int
	RateLimitCreate  int
	RateLimitJoin    int
	PowDifficulty    int
	PowTTL           time.Duration
}

// Load parses flags and environment variables into Config.
func Load() Config {
	addr := flag.String("addr", ":8080", "listen address")
	ttl := flag.Duration("invite-ttl", 24*time.Hour, "invite token lifetime")
	maxSize := flag.Int("max-room-size", 12, "maximum peers per room")
	rateWindow := flag.Duration("rate-limit-window", time.Minute, "rate limit window")
	rateHTTP := flag.Int("rate-limit-http", 180, "HTTP requests per IP per window")
	rateWS := flag.Int("rate-limit-ws", 20, "WebSocket upgrades per IP per window")
	rateCreate := flag.Int("rate-limit-create", 10, "room creates per IP per window")
	rateJoin := flag.Int("rate-limit-join", 30, "room joins per IP per window")
	powDifficulty := flag.Int("pow-difficulty", 18, "proof-of-work leading zero bits (0 disables)")
	powTTL := flag.Duration("pow-ttl", 5*time.Minute, "proof-of-work challenge lifetime")
	flag.Parse()

	secret := os.Getenv("HUDDLE_INVITE_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}

	trustProxy := os.Getenv("HUDDLE_TRUST_PROXY")
	trust := trustProxy == "1" || strings.EqualFold(trustProxy, "true")

	return Config{
		Addr:            *addr,
		InviteSecret:    secret,
		InviteTTL:       *ttl,
		MaxRoomSize:     *maxSize,
		TrustProxy:      trust,
		RateLimitWindow: *rateWindow,
		RateLimitHTTP:   *rateHTTP,
		RateLimitWS:     *rateWS,
		RateLimitCreate: *rateCreate,
		RateLimitJoin:   *rateJoin,
		PowDifficulty:   *powDifficulty,
		PowTTL:          *powTTL,
	}
}
