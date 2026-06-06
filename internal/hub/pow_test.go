package hub

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"huddle/internal/pow"
	"huddle/internal/room"
)

func TestHubRequiresPowWhenEnabled(t *testing.T) {
	store := pow.NewStore()
	url, _ := startTestHubWithPow(t, 4, store, 10)

	client := dialClient(t, url)
	client.send(TypeCreateRoom, CreateRoomPayload{Name: "NoPow"})
	msg := client.readType(TypeError)
	if string(msg.Payload) == "" {
		t.Fatal("expected error payload")
	}

	ch, err := store.Issue("create", 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var nonce uint64
	for nonce = 0; nonce < 1_000_000; nonce++ {
		if pow.Verify(ch.Prefix, nonce, ch.Difficulty) {
			break
		}
	}

	client.send(TypeCreateRoom, CreateRoomPayload{
		Name: "WithPow",
		Pow:  &PowPayload{ID: ch.ID, Nonce: nonce},
	})
	if client.read().Type != TypeCreated {
		t.Fatal("expected create to succeed with valid pow")
	}
}

func startTestHubWithPow(t *testing.T, maxSize int, store *pow.Store, difficulty int) (string, *room.Manager) {
	t.Helper()
	rm := room.NewManager("hub-test-secret", time.Hour, maxSize)
	h := New(rm, testLimits(), store, difficulty)
	go h.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		h.ServeClient(conn, host)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws", rm
}
