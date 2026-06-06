package hub

import (
	"encoding/json"
	"testing"
)

func TestHubResumeRequiresToken(t *testing.T) {
	url, rm := startTestHub(t, 4)
	_, created, hostJoined := createAndJoinHost(t, url)

	attacker := dialClient(t, url)
	attacker.send(TypeJoin, JoinPayload{
		RoomID:       created.RoomID,
		Invite:       created.Invite,
		Name:         "Attacker",
		ResumePeerID: hostJoined.PeerID,
	})
	joined := attacker.readJoined()

	if joined.PeerID == hostJoined.PeerID {
		t.Fatal("resume without token must not reuse another peer id")
	}

	r, err := rm.Get(created.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsHost(hostJoined.PeerID) {
		t.Fatal("host identity must remain unchanged")
	}
}

func TestHubNonHostCannotAddChannel(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host, created, _ := createAndJoinHost(t, url)

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guest.readJoined()
	_ = host.readType(TypePeerJoined)

	guest.send(TypeAddChannel, AddChannelPayload{ID: "malicious", Name: "malicious"})
	state := host.readType(TypeRoomState)

	var roomState struct {
		Channels []struct {
			ID string `json:"id"`
		} `json:"channels"`
	}
	if err := json.Unmarshal(state.Payload, &roomState); err != nil {
		t.Fatal(err)
	}
	for _, ch := range roomState.Channels {
		if ch.ID == "malicious" {
			t.Fatal("non-host must not add channels")
		}
	}
}
