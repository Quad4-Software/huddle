package hub

import (
	"testing"

	"huddle/internal/wire"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	data, err := Marshal(TypeJoin, JoinPayload{
		RoomID: "abc",
		Invite: "token",
		Name:   "Ada",
	})
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != TypeJoin {
		t.Fatalf("expected %s, got %s", TypeJoin, decoded.Type)
	}

	var payload JoinPayload
	if err := decodePayloadTyped(TypeJoin, decoded.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.RoomID != "abc" || payload.Name != "Ada" {
		t.Fatal("payload did not round-trip")
	}
}

func TestUnmarshalRejectsInvalidFrame(t *testing.T) {
	if _, err := Unmarshal([]byte("not-a-frame")); err == nil {
		t.Fatal("expected invalid frame to fail")
	}
	if _, _, err := wire.DecodeFrame([]byte{wire.Magic, wire.MsgPing, 0, 0, 0, 10}); err == nil {
		t.Fatal("expected truncated frame to fail")
	}
}
