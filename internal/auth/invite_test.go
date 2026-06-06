package auth

import (
	"strings"
	"testing"
	"time"
)

func TestSignAndVerifyInvite(t *testing.T) {
	secret := "test-secret"
	roomID := "abc123"
	expires := time.Now().Add(time.Hour)

	token := SignInvite(secret, roomID, expires)
	if err := VerifyInvite(secret, roomID, token); err != nil {
		t.Fatalf("expected valid invite, got %v", err)
	}
}

func TestVerifyInviteRejectsWrongSecret(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	token := SignInvite("secret-a", "room1", expires)

	if err := VerifyInvite("secret-b", "room1", token); err == nil {
		t.Fatal("expected tampered secret to be rejected")
	}
}

func TestVerifyInviteRejectsWrongRoom(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	token := SignInvite("secret", "room-a", expires)

	if err := VerifyInvite("secret", "room-b", token); err == nil {
		t.Fatal("expected room mismatch to be rejected")
	}
}

func TestVerifyInviteRejectsExpired(t *testing.T) {
	expires := time.Now().Add(-time.Minute)
	token := SignInvite("secret", "room1", expires)

	if err := VerifyInvite("secret", "room1", token); err == nil {
		t.Fatal("expected expired invite to be rejected")
	}
}

func TestVerifyInviteRejectsMalformedToken(t *testing.T) {
	cases := []string{"", "no-dot", "bad.sig", "only.one.part.extra"}
	for _, token := range cases {
		if err := VerifyInvite("secret", "room1", token); err == nil {
			t.Fatalf("expected malformed token %q to be rejected", token)
		}
	}
}

func TestVerifyInviteRejectsTamperedPayload(t *testing.T) {
	expires := time.Now().Add(time.Hour)
	token := SignInvite("secret", "room1", expires)
	parts := strings.Split(token, ".")
	tampered := parts[0] + "x." + parts[1]

	if err := VerifyInvite("secret", "room1", tampered); err == nil {
		t.Fatal("expected tampered payload to be rejected")
	}
}

func TestGenerateRoomKey(t *testing.T) {
	a, err := GenerateRoomKey()
	if err != nil {
		t.Fatal(err)
	}
	b, err := GenerateRoomKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("expected non-empty room keys")
	}
	if a == b {
		t.Fatal("expected unique room keys")
	}
}

func TestNewRoomID(t *testing.T) {
	id, err := NewRoomID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 16 {
		t.Fatalf("expected 16 hex chars, got %d", len(id))
	}
}
