package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("huddle-pass")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "huddle-pass") {
		t.Fatal("expected password to match hash")
	}
	if CheckPassword(hash, "wrong-pass") {
		t.Fatal("expected wrong password to be rejected")
	}
}

func TestCheckPasswordRejectsEmptyHash(t *testing.T) {
	if CheckPassword("", "anything") {
		t.Fatal("expected empty hash to be rejected")
	}
}
