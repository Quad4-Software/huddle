// Package server serves the embedded web UI, HTTP API, and WebSocket endpoint.
package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"huddle/internal/config"
	"huddle/internal/hub"
	"huddle/internal/pow"
	"huddle/internal/ratelimit"
	"huddle/internal/room"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Server is the HTTP front end for Huddle.
type Server struct {
	cfg    config.Config
	hub    *hub.Hub
	mux    *http.ServeMux
	limits rateLimits
	pow    *powHandler
}

// New constructs a server from cfg and starts the signaling hub.
func New(cfg config.Config) *Server {
	rm := room.NewManager(cfg.InviteSecret, cfg.InviteTTL, cfg.MaxRoomSize)
	powStore := pow.NewStore()
	limits := hub.Limits{
		Create: ratelimit.New(cfg.RateLimitCreate, cfg.RateLimitWindow),
		Join:   ratelimit.New(cfg.RateLimitJoin, cfg.RateLimitWindow),
	}
	h := hub.New(rm, limits, powStore, cfg.PowDifficulty)
	go h.Run()

	rl := newRateLimits(cfg.RateLimitHTTP, cfg.RateLimitWS, cfg.RateLimitWindow)
	startRateLimitCleanup(cfg.RateLimitWindow, rl.http, rl.ws, limits.Create, limits.Join, powStore)

	s := &Server{
		cfg:    cfg,
		hub:    h,
		mux:    http.NewServeMux(),
		limits: rl,
		pow: &powHandler{
			store:      powStore,
			difficulty: cfg.PowDifficulty,
			ttl:        cfg.PowTTL,
		},
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/pow/challenge", s.pow.handleChallenge)
	s.mux.HandleFunc("GET /ws", s.handleWS)
	s.mux.Handle("/", spaHandler())
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	s.hub.ServeClient(conn, clientIP(r, s.cfg.TrustProxy))
}

// ListenAndServe starts the HTTP server on cfg.Addr.
func (s *Server) ListenAndServe() error {
	log.Printf("huddle listening on %s", s.cfg.Addr)
	if s.cfg.TrustProxy {
		log.Printf("trusting reverse-proxy forwarded headers")
	}
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           withMiddleware(s.cfg.TrustProxy, s.limits, s.mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	return srv.ListenAndServe()
}
