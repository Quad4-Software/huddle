package hub

import (
	"testing"
)

func TestHubSameConnectionRejoinReplacesStaleMember(t *testing.T) {
	url, rm := startTestHub(t, 4)

	client := dialClient(t, url)
	client.send(TypeCreateRoom, CreateRoomPayload{Name: "Rejoin"})
	created := client.readCreated()
	client.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Dave",
	})
	first := client.readJoined()

	client.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Dave",
	})
	second := client.readJoined()

	if second.PeerID == first.PeerID {
		t.Fatal("expected second join without resume to allocate a new peer id")
	}

	r, err := rm.Get(created.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Size() != 1 {
		t.Fatalf("expected 1 member after rejoin on same connection, got %d", r.Size())
	}
	for _, m := range r.MemberList() {
		if m.ID != second.PeerID {
			t.Fatalf("expected only latest peer, got %+v", m)
		}
	}
}
