package server

import (
	"net/http"
	"time"

	"huddle/internal/ratelimit"
)

type rateLimits struct {
	http *ratelimit.Limiter
	ws   *ratelimit.Limiter
}

func newRateLimits(httpLimit, wsLimit int, window time.Duration) rateLimits {
	return rateLimits{
		http: ratelimit.New(httpLimit, window),
		ws:   ratelimit.New(wsLimit, window),
	}
}

func (rl rateLimits) middleware(trustProxy bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/ws" {
			if !rl.ws.Allow(clientIP(r, trustProxy)) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if !rl.http.Allow(clientIP(r, trustProxy)) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type powCleaner interface {
	Cleanup()
}

func startRateLimitCleanup(interval time.Duration, limiters ...any) {
	if interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			for _, item := range limiters {
				switch v := item.(type) {
				case *ratelimit.Limiter:
					v.Cleanup()
				case powCleaner:
					v.Cleanup()
				}
			}
		}
	}()
}
