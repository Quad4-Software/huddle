package hub

import (
	"sync"
	"testing"
)

func TestValidSignalICEUsesRawCandidate(t *testing.T) {
	candidate := []byte(`{"candidate":"candidate:1","sdpMid":"0","sdpMLineIndex":0}`)
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
		Candidate: []byte(`not-json`),
	}) {
		t.Fatal("expected invalid candidate json to fail")
	}
}

func TestHubConcurrentHandleSignal(t *testing.T) {
	h, host, guest := testHubWithPeers(t)
	msg, err := Marshal(TypeICE, SignalPayload{
		To:        guest.ID,
		Nonce:     "nonce",
		Sig:       "sig",
		Candidate: []byte(`{"candidate":"candidate:1","sdpMid":"0","sdpMLineIndex":0}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := Unmarshal(msg)
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
				h.handleSignal(host, frame)
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
	pong, err := decodePingPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if pong.T != 99 {
		t.Fatalf("expected ping/pong to keep working, got %d", pong.T)
	}
	_ = rm
}
