package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSPAHandlerRoot(t *testing.T) {
	h := spaHandler()

	for _, path := range []string{"/", "/index.html", "/r/test"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusFound {
			t.Fatalf("GET %s: got redirect %d to %q", path, rec.Code, rec.Header().Get("Location"))
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: expected 200, got %d", path, rec.Code)
		}
	}
}

func TestSPAHandlerServesAsset(t *testing.T) {
	h := spaHandler()
	req := httptest.NewRequest(http.MethodGet, "/assets/index-BSFGxzOq.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for asset, got %d", rec.Code)
	}
}
