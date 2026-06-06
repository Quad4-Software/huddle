package hub

import (
	"encoding/json"
	"testing"
)

func createAndJoinHost(t *testing.T, url string) (*wsClient, CreatedPayload, JoinedPayload) {
	t.Helper()
	host := dialClient(t, url)
	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Hosted"})
	created := host.readCreated()
	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	joined := host.readJoined()
	return host, created, joined
}

func TestHubCreatorBecomesHost(t *testing.T) {
	url, _ := startTestHub(t, 4)
	_, _, joined := createAndJoinHost(t, url)
	if joined.Room["hostId"] != joined.PeerID {
		t.Fatalf("expected creator to be host, got %v", joined.Room["hostId"])
	}
}

func TestHubHostCanKickMember(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host, created, hostJoined := createAndJoinHost(t, url)

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	_ = host.readType(TypePeerJoined)

	host.send(TypeKick, KickPayload{PeerID: guestJoined.PeerID})
	_ = guest.readType(TypeKicked)
	_ = host.readType(TypePeerLeft)

	state := host.readType(TypeRoomState)
	var roomState struct {
		Members []struct {
			ID string `json:"id"`
		} `json:"members"`
	}
	if err := json.Unmarshal(state.Payload, &roomState); err != nil {
		t.Fatal(err)
	}
	if len(roomState.Members) != 1 || roomState.Members[0].ID != hostJoined.PeerID {
		t.Fatalf("expected only host to remain, got %+v", roomState.Members)
	}
}

func TestHubNonHostCannotKick(t *testing.T) {
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

	guest.send(TypeKick, KickPayload{PeerID: hostJoined.PeerID})
	msg := guest.readType(TypeError)
	var payload ErrorPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Message != "only the host can kick members" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}
