package hub

import (
	"testing"
	"time"
)

func TestHubRefreshReplacesSameIPConnection(t *testing.T) {
	url, rm := startTestHub(t, 4)

	first := dialClient(t, url)
	first.send(TypeCreateRoom, CreateRoomPayload{Name: "Refresh"})
	created := first.readCreated()
	first.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Dave",
	})
	firstJoined := first.readJoined()

	second := dialClient(t, url)
	second.send(TypeJoin, JoinPayload{
		RoomID:       created.RoomID,
		Invite:       created.Invite,
		Name:         "Dave",
		ResumePeerID: firstJoined.PeerID,
		ResumeToken:  firstJoined.ResumeToken,
	})
	secondJoined := second.readJoined()

	if secondJoined.PeerID != firstJoined.PeerID {
		t.Fatalf("expected resume to keep peer id, got %s then %s", firstJoined.PeerID, secondJoined.PeerID)
	}

	time.Sleep(50 * time.Millisecond)

	r, err := rm.Get(created.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Size() != 1 {
		t.Fatalf("expected 1 member after refresh, got %d", r.Size())
	}
	for _, m := range r.MemberList() {
		if m.ID != firstJoined.PeerID {
			t.Fatalf("expected only refreshed peer, got %+v", m)
		}
	}
}

func TestHubJoinDoesNotDuplicateMembersForHostAndGuest(t *testing.T) {
	url, rm := startTestHub(t, 4)

	host := dialClient(t, url)
	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Pair"})
	created := host.readCreated()
	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Dave",
	})
	host.readJoined()

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	_ = host.readType(TypePeerJoined)

	r, err := rm.Get(created.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Size() != 2 {
		t.Fatalf("expected 2 members, got %d", r.Size())
	}

	members, ok := guestJoined.Room["members"].([]any)
	if !ok || len(members) != 2 {
		t.Fatalf("expected 2 members in joined snapshot, got %+v", guestJoined.Room["members"])
	}

	ids := map[string]bool{guestJoined.PeerID: true}
	for _, raw := range members {
		m, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected member shape: %+v", raw)
		}
		name, _ := m["name"].(string)
		id, _ := m["id"].(string)
		if name == "Dave" {
			ids[id] = true
		}
	}
	if len(ids) != 2 {
		t.Fatalf("expected distinct host and guest ids, got %+v", members)
	}
}

func TestHubBroadcastSurvivesDisconnectRace(t *testing.T) {
	url, _ := startTestHub(t, 4)

	host := dialClient(t, url)
	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Race"})
	created := host.readCreated()
	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	host.readJoined()

	for range 20 {
		client := dialClient(t, url)
		client.send(TypeJoin, JoinPayload{
			RoomID: created.RoomID,
			Invite: created.Invite,
			Name:   "Guest",
		})
		client.readJoined()
		host.readType(TypePeerJoined)
		_ = client.conn.Close()
	}

	time.Sleep(100 * time.Millisecond)
	host.send(TypeRename, RenamePayload{Name: "Host-Renamed"})
	_ = host.readType(TypeRoomState)
}
