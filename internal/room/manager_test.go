package room

import (
	"errors"
	"testing"
	"time"
)

func TestManagerCreateAndJoin(t *testing.T) {
	m := NewManager("secret", time.Hour, 4)

	result, err := m.Create(CreateInput{Name: "Sprint"})
	if err != nil {
		t.Fatal(err)
	}
	if result.RoomID == "" || result.Invite == "" || result.RoomKey == "" {
		t.Fatal("expected create result fields to be populated")
	}

	room, err := m.ValidateJoin(result.RoomID, result.Invite, "")
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	if room.Name != "Sprint" {
		t.Fatalf("expected Sprint room, got %s", room.Name)
	}
}

func TestManagerJoinRequiresPassword(t *testing.T) {
	m := NewManager("secret", time.Hour, 4)
	result, err := m.Create(CreateInput{Name: "Locked", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.ValidateJoin(result.RoomID, result.Invite, "")
	if !errors.Is(err, ErrJoinDenied) {
		t.Fatalf("expected join denied, got %v", err)
	}

	room, err := m.ValidateJoin(result.RoomID, result.Invite, "secret")
	if err != nil {
		t.Fatalf("expected valid password join, got %v", err)
	}
	if room.Password == "" {
		t.Fatal("expected stored password hash")
	}
}

func TestManagerJoinRejectsBadInvite(t *testing.T) {
	m := NewManager("secret", time.Hour, 4)
	result, err := m.Create(CreateInput{Name: "Open"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.ValidateJoin(result.RoomID, "bad.token.value", "")
	if !errors.Is(err, ErrJoinDenied) {
		t.Fatalf("expected join denied, got %v", err)
	}
	if _, err := m.Get(result.RoomID); err != nil {
		t.Fatal("expected room to remain after invalid invite")
	}
}

func TestManagerEnforcesRoomCapacity(t *testing.T) {
	m := NewManager("secret", time.Hour, 1)
	result, err := m.Create(CreateInput{Name: "Tiny"})
	if err != nil {
		t.Fatal(err)
	}

	room, err := m.ValidateJoin(result.RoomID, result.Invite, "")
	if err != nil {
		t.Fatal(err)
	}
	room.AddMember("peer-1", "One")

	_, err = m.ValidateJoin(result.RoomID, result.Invite, "")
	if !errors.Is(err, ErrJoinDenied) {
		t.Fatalf("expected join denied, got %v", err)
	}
}

func TestManagerRemoveIfEmpty(t *testing.T) {
	m := NewManager("secret", time.Hour, 4)
	result, err := m.Create(CreateInput{Name: "Temp"})
	if err != nil {
		t.Fatal(err)
	}

	room, err := m.Get(result.RoomID)
	if err != nil {
		t.Fatal(err)
	}
	room.AddMember("peer-1", "One")
	room.RemoveMember("peer-1")

	m.RemoveIfEmpty(result.RoomID)
	if _, err := m.Get(result.RoomID); !errors.Is(err, ErrRoomNotFound) {
		t.Fatal("expected room to be removed when empty")
	}
}
