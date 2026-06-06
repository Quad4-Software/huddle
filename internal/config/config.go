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
	Addr            string
	InviteSecret    string
	InviteTTL       time.Duration
	MaxRoomSize     int
	MaxRooms        int
	TrustProxy      bool
	CORSOrigins     []string
	RateLimitWindow time.Duration
	RateLimitHTTP   int
	RateLimitWS     int
	RateLimitCreate int
	RateLimitJoin   int
	PowDifficulty   int
	PowTTL          time.Duration
	TURNEnabled     bool
	TURNListenAddr  string
	TURNPublicAddr  string
	TURNRealm       string
	TURNCredTTL     time.Duration
}

// Load parses flags and environment variables into Config.
func Load() Config {
	addr := flag.String("addr", ":8080", "listen address")
	ttl := flag.Duration("invite-ttl", 24*time.Hour, "invite token lifetime")
	maxSize := flag.Int("max-room-size", 12, "maximum peers per room")
	maxRooms := flag.Int("max-rooms", 1000, "maximum active rooms")
	rateWindow := flag.Duration("rate-limit-window", time.Minute, "rate limit window")
	rateHTTP := flag.Int("rate-limit-http", 180, "HTTP requests per IP per window")
	rateWS := flag.Int("rate-limit-ws", 20, "WebSocket upgrades per IP per window")
	rateCreate := flag.Int("rate-limit-create", 10, "room creates per IP per window")
	rateJoin := flag.Int("rate-limit-join", 30, "room joins per IP per window")
	powDifficulty := flag.Int("pow-difficulty", 0, "proof-of-work leading zero bits (0 disables)")
	powTTL := flag.Duration("pow-ttl", 5*time.Minute, "proof-of-work challenge lifetime")
	turnListenAddr := flag.String("turn-listen-addr", ":3478", "TURN UDP listen address")
	turnPublicAddr := flag.String("turn-public-addr", "", "public host:port advertised for built-in TURN")
	turnRealm := flag.String("turn-realm", "huddle", "TURN authentication realm")
	turnCredTTL := flag.Duration("turn-credential-ttl", 4*time.Hour, "TURN credential lifetime")
	flag.Parse()

	secret := os.Getenv("HUDDLE_INVITE_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}

	trustProxy := os.Getenv("HUDDLE_TRUST_PROXY")
	trust := trustProxy == "1" || strings.EqualFold(trustProxy, "true")

	corsOrigins := parseCSV(os.Getenv("HUDDLE_CORS_ORIGINS"))
	turnEnabled := envBool("HUDDLE_TURN_ENABLED")

	return Config{
		Addr:            *addr,
		InviteSecret:    secret,
		InviteTTL:       *ttl,
		MaxRoomSize:     *maxSize,
		MaxRooms:        *maxRooms,
		TrustProxy:      trust,
		CORSOrigins:     corsOrigins,
		RateLimitWindow: *rateWindow,
		RateLimitHTTP:   *rateHTTP,
		RateLimitWS:     *rateWS,
		RateLimitCreate: *rateCreate,
		RateLimitJoin:   *rateJoin,
		PowDifficulty:   *powDifficulty,
		PowTTL:          *powTTL,
		TURNEnabled:     turnEnabled,
		TURNListenAddr:  envOr("HUDDLE_TURN_LISTEN_ADDR", *turnListenAddr),
		TURNPublicAddr:  envOr("HUDDLE_TURN_PUBLIC_ADDR", *turnPublicAddr),
		TURNRealm:       envOr("HUDDLE_TURN_REALM", *turnRealm),
		TURNCredTTL:     *turnCredTTL,
	}
}

func envBool(name string) bool {
	value := os.Getenv(name)
	return value == "1" || strings.EqualFold(value, "true")
}

func envOr(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
