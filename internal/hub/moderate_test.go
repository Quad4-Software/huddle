package hub

import (
	"testing"
)

func TestHubHostCanModerateMember(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host, created, _ := createAndJoinHost(t, url)

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	_ = host.readType(TypePeerJoined)

	host.send(TypeModerateMember, ModerateMemberPayload{
		PeerID: guestJoined.PeerID,
		Muted:  true,
	})
	update := guest.readType(TypeMemberUpdate)
	payload, err := decodeMemberUpdatePayload(update.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.PeerID != guestJoined.PeerID || !payload.Muted || payload.Deafened {
		t.Fatalf("unexpected member update: %+v", payload)
	}
}

func TestHubNonHostCannotModerate(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host, created, hostJoined := createAndJoinHost(t, url)

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	_ = guest.readJoined()
	_ = host.readType(TypePeerJoined)

	guest.send(TypeModerateMember, ModerateMemberPayload{
		PeerID: hostJoined.PeerID,
		Muted:  true,
	})
	msg := guest.readType(TypeError)
	payload, err := decodeErrorPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Message != "only the host can moderate members" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}
