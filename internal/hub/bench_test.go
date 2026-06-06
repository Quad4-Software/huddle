package hub

import (
	"testing"
	"time"

	"huddle/internal/room"
	"huddle/internal/wire"
)

func BenchmarkMarshalSignal(b *testing.B) {
	payload := SignalPayload{
		To:    "peer-b",
		From:  "peer-a",
		SDP:   "v=0",
		Nonce: "nonce-123",
		Sig:   "sig-456",
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Marshal(TypeOffer, payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalICE(b *testing.B) {
	payload := SignalPayload{
		To:        "peer-b",
		From:      "peer-a",
		Nonce:     "nonce-123",
		Sig:       "sig-456",
		Candidate: []byte(`{"candidate":"candidate:1 1 udp 2122260223 192.168.0.2 54321 typ host","sdpMid":"0","sdpMLineIndex":0}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Marshal(TypeICE, payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidSignalICE(b *testing.B) {
	payload := SignalPayload{
		To:        "peer-b",
		Nonce:     "nonce-123",
		Sig:       "sig-456",
		Candidate: []byte(`{"candidate":"candidate:1 1 udp 2122260223 192.168.0.2 54321 typ host","sdpMid":"0","sdpMLineIndex":0}`),
	}
	b.ReportAllocs()
	for b.Loop() {
		if !validSignal(TypeICE, payload) {
			b.Fatal("expected valid signal")
		}
	}
}

func BenchmarkMarshalPingPong(b *testing.B) {
	raw := wire.EncodePing(1234567890)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := marshalPong(raw); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBroadcastMemberUpdate(b *testing.B) {
	h, host, guest := benchHubWithPeers(b)
	update := MemberUpdatePayload{PeerID: host.ID, Muted: true, Deafened: false, Speaking: true}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		h.broadcast(host.RoomID, TypeMemberUpdate, update, host.ID)
		drainClient(guest)
	}
}

func BenchmarkHandleSignal(b *testing.B) {
	h, host, guest := benchHubWithPeers(b)
	msg, err := Marshal(TypeOffer, SignalPayload{
		To:    guest.ID,
		SDP:   "v=0",
		Nonce: "nonce",
		Sig:   "sig",
	})
	if err != nil {
		b.Fatal(err)
	}
	wire, err := Unmarshal(msg)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		h.handleSignal(host, wire)
		drainClient(guest)
	}
}

func testHubWithPeers(tb testing.TB) (*Hub, *Client, *Client) {
	tb.Helper()
	rm := room.NewManager("bench-secret", time.Hour, 32)
	h := New(rm, testLimits(), nil, 0)
	go h.Run()
	tb.Cleanup(func() {
		close(h.unregister)
	})
	host := registerBenchClient(h, "host")
	guest := registerBenchClient(h, "guest")
	return h, host, guest
}

func benchHubWithPeers(tb testing.TB) (*Hub, *Client, *Client) {
	return testHubWithPeers(tb)
}

func registerBenchClient(h *Hub, id string) *Client {
	c := &Client{
		Hub:  h,
		IP:   "127.0.0.1",
		Send: make(chan []byte, 1024),
	}
	c.ID = id
	c.RoomID = "bench-room"
	c.Name = id

	h.mu.Lock()
	if h.clients[c.RoomID] == nil {
		h.clients[c.RoomID] = make(map[string]*Client)
	}
	h.clients[c.RoomID][c.ID] = c
	h.mu.Unlock()
	return c
}

func drainClient(c *Client) {
	for {
		select {
		case <-c.Send:
		default:
			return
		}
	}
}
