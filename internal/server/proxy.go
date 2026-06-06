package server

import (
	"net/http"
	"strings"
)

func withMiddleware(trustProxy bool, corsOrigins []string, limits rateLimits, next http.Handler) http.Handler {
	h := http.Handler(next)
	h = limits.middleware(trustProxy, h)
	h = corsMiddleware(corsOrigins, h)
	h = securityHeaders(h)
	if trustProxy {
		h = trustProxyHeaders(h)
	}
	h = normalizePath(h)
	return h
}

func normalizePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}

const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"font-src 'self'; " +
	"img-src 'self' blob: data:; " +
	"media-src 'self' blob:; " +
	"connect-src 'self' wss: ws:; " +
	"worker-src 'self'; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

const permissionsPolicy = "microphone=(self), display-capture=(self), camera=()"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("Permissions-Policy", permissionsPolicy)
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func trustProxyHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			r.URL.Scheme = strings.ToLower(proto)
		}
		if host := r.Header.Get("X-Forwarded-Host"); host != "" {
			r.Host = host
		}
		next.ServeHTTP(w, r)
	})
}
