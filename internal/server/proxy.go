package server

import (
	"net/http"
	"strings"
)

func withMiddleware(trustProxy bool, limits rateLimits, next http.Handler) http.Handler {
	h := http.Handler(next)
	h = limits.middleware(trustProxy, h)
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

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
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
