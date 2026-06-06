package server

import (
	"net/http"
	"slices"
	"strings"
)

const (
	corsAllowMethods = "GET, OPTIONS"
	corsAllowHeaders = "Content-Type"
	corsMaxAge       = "86400"
)

func corsMiddleware(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && originAllowed(r, origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			appendVary(w, "Origin")
			w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
			w.Header().Set("Access-Control-Max-Age", corsMaxAge)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func originAllowed(r *http.Request, origin string, allowed []string) bool {
	if origin == "" {
		return true
	}
	if sameOrigin(r, origin) {
		return true
	}
	return slices.Contains(allowed, origin)
}

func sameOrigin(r *http.Request, origin string) bool {
	return origin == requestOrigin(r)
}

func requestOrigin(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}
	return scheme + "://" + r.Host
}

func appendVary(w http.ResponseWriter, value string) {
	existing := w.Header().Get("Vary")
	if existing == "" {
		w.Header().Set("Vary", value)
		return
	}
	for part := range strings.SplitSeq(existing, ",") {
		if strings.TrimSpace(part) == value {
			return
		}
	}
	w.Header().Set("Vary", existing+", "+value)
}
