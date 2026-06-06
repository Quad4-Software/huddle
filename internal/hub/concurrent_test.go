package hub

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestValidSignalICEUsesRawCandidate(t *testing.T) {
	candidate := json.RawMessage(`{"candidate":"candidate:1","sdpMid":"0","sdpMLineIndex":0}`)
	if !validSignal(TypeICE, SignalPayload{
		To:        "peer-b",
		Nonce:     "nonce",
		Sig:       "sig",
		Candidate: candidate,
	}) {
		t.Fatal("expected valid ice signal")
	}
	if validSignal(TypeICE, SignalPayload{
		To:        "peer-b",
		Nonce:     "nonce",
		Sig:       "sig",
		Candidate: json.RawMessage(`not-json`),
	}) {
		t.Fatal("expected invalid candidate json to fail")
	}
}

func TestMarshalWithPayload(t *testing.T) {
	raw := json.RawMessage(`{"t":42}`)
	data, err := marshalWithPayload(TypePong, raw)
	if err != nil {
		t.Fatal(err)
	}
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != TypePong {
		t.Fatalf("expected pong, got %s", msg.Type)
	}
	if string(msg.Payload) != string(raw) {
		t.Fatalf("payload mismatch: %s", msg.Payload)
	}
}

func TestHubConcurrentHandleSignal(t *testing.T) {
	h, host, guest := testHubWithPeers(t)
	msg, err := Marshal(TypeICE, SignalPayload{
		To:        guest.ID,
		Nonce:     "nonce",
		Sig:       "sig",
		Candidate: json.RawMessage(`{"candidate":"candidate:1","sdpMid":"0","sdpMLineIndex":0}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := Unmarshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	const workers = 16
	const perWorker = 64
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for range perWorker {
				h.handleSignal(host, wire)
				drainClient(guest)
			}
		}()
	}
	wg.Wait()
}

func TestHubBroadcastEvictsStaleClientWithoutDeadlock(t *testing.T) {
	url, rm := startTestHub(t, 4)
	host := dialClient(t, url)
	guest := dialClient(t, url)

	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Evict"})
	created := host.readCreated()
	host.send(TypeJoin, JoinPayload{RoomID: created.RoomID, Invite: created.Invite, Name: "Host"})
	host.readJoined()
	guest.send(TypeJoin, JoinPayload{RoomID: created.RoomID, Invite: created.Invite, Name: "Guest"})
	guest.readJoined()
	_ = host.readType(TypePeerJoined)

	for range 1100 {
		host.send(TypeMemberUpdate, MemberUpdatePayload{Muted: true, Deafened: false, Speaking: false})
	}
	guest.readType(TypeMemberUpdate)

	host.send(TypePing, PingPayload{T: 99})
	msg := host.readType(TypePong)
	var pong PingPayload
	if err := json.Unmarshal(msg.Payload, &pong); err != nil {
		t.Fatal(err)
	}
	if pong.T != 99 {
		t.Fatalf("expected ping/pong to keep working, got %d", pong.T)
	}
	_ = rm
}
