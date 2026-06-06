package pow

import (
	"testing"
	"time"
)

func TestVerifyFindsSolution(t *testing.T) {
	prefix := "abc123"
	difficulty := 12
	var nonce uint64
	for nonce = 0; nonce < 1_000_000; nonce++ {
		if Verify(prefix, nonce, difficulty) {
			break
		}
	}
	if !Verify(prefix, nonce, difficulty) {
		t.Fatal("expected to find a valid nonce")
	}
	if Verify(prefix, nonce+1, difficulty) {
		t.Fatal("did not expect adjacent nonce to pass")
	}
}

func TestStoreVerifyConsumesChallenge(t *testing.T) {
	store := NewStore()
	ch, err := store.Issue("create", 8, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	var nonce uint64
	for nonce = 0; nonce < 1_000_000; nonce++ {
		if Verify(ch.Prefix, nonce, ch.Difficulty) {
			break
		}
	}

	if err := store.Verify("create", ch.ID, nonce); err != nil {
		t.Fatalf("expected verify success, got %v", err)
	}
	if err := store.Verify("create", ch.ID, nonce); err != ErrChallengeNotFound {
		t.Fatalf("expected replay to fail, got %v", err)
	}
}

func TestStoreRejectsWrongAction(t *testing.T) {
	store := NewStore()
	ch, err := store.Issue("create", 8, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	var nonce uint64
	for nonce = 0; nonce < 1_000_000; nonce++ {
		if Verify(ch.Prefix, nonce, ch.Difficulty) {
			break
		}
	}
	if err := store.Verify("join", ch.ID, nonce); err != ErrInvalidAction {
		t.Fatalf("expected invalid action, got %v", err)
	}
}
