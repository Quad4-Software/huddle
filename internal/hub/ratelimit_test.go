package hub

import (
	"encoding/json"
	"testing"
	"time"

	"huddle/internal/ratelimit"
)

func TestHubRateLimitsCreateRoom(t *testing.T) {
	url, _ := startTestHubWithLimits(t, 4, Limits{
		Create: ratelimit.New(1, time.Minute),
		Join:   ratelimit.New(100, time.Minute),
	})
	client := dialClient(t, url)

	client.send(TypeCreateRoom, CreateRoomPayload{Name: "First"})
	if client.read().Type != TypeCreated {
		t.Fatal("expected first create to succeed")
	}

	client.send(TypeCreateRoom, CreateRoomPayload{Name: "Second"})
	msg := client.read()
	if msg.Type != TypeError {
		t.Fatalf("expected error, got %s", msg.Type)
	}
	var payload ErrorPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Message != "rate limit exceeded" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}

func TestHubRateLimitsJoin(t *testing.T) {
	url, _ := startTestHubWithLimits(t, 4, Limits{
		Create: ratelimit.New(100, time.Minute),
		Join:   ratelimit.New(1, time.Minute),
	})
	client := dialClient(t, url)

	client.send(TypeCreateRoom, CreateRoomPayload{Name: "Room"})
	created := client.readCreated()

	client.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	if client.read().Type != TypeJoined {
		t.Fatal("expected first join to succeed")
	}

	client.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Again",
	})
	msg := client.readType(TypeError)
	var payload ErrorPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Message != "rate limit exceeded" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}

