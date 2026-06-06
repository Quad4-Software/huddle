package hub

import (
	"encoding/json"
	"testing"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	original := Message{
		Type: TypeJoin,
		Payload: json.RawMessage(
			`{"roomId":"abc","invite":"token","name":"Ada"}`,
		),
	}

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
	if decoded.Type != original.Type {
		t.Fatalf("expected %s, got %s", original.Type, decoded.Type)
	}

	var payload JoinPayload
	if err := json.Unmarshal(decoded.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.RoomID != "abc" || payload.Name != "Ada" {
		t.Fatal("payload did not round-trip")
	}
}

func TestUnmarshalRejectsInvalidJSON(t *testing.T) {
	if _, err := Unmarshal([]byte("{")); err == nil {
		t.Fatal("expected invalid json to fail")
	}
}
