package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

func spaHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend assets unavailable", http.StatusServiceUnavailable)
		})
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			r.URL.Path = "/index.html"
		} else if _, err := sub.Open(path); err != nil {
			r.URL.Path = "/index.html"
		}
		fileServer.ServeHTTP(w, r)
	})
}
