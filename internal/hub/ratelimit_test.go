package hub

import (
	"testing"
	"time"

	"huddle/internal/ratelimit"
	"huddle/internal/wire"
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
	payload, err := decodeErrorPayload(msg.Payload)
	if err != nil {
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
	if client.readType(TypeJoined).Type != TypeJoined {
		t.Fatal("expected first join to succeed")
	}

	client.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Again",
	})
	msg := client.readType(TypeError)
	payload, err := decodeErrorPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Message != "rate limit exceeded" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}

func TestMarshalPongFrame(t *testing.T) {
	raw := wire.EncodePing(42)
	data, err := marshalPong(raw)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Type != TypePong {
		t.Fatalf("expected pong, got %s", msg.Type)
	}
	pong, err := decodePingPayload(msg.Payload)
	if err != nil || pong.T != 42 {
		t.Fatalf("unexpected pong payload: %+v err=%v", pong, err)
	}
}
