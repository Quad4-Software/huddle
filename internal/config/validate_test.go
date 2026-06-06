package config

import "testing"

func TestValidateRejectsDefaultSecret(t *testing.T) {
	err := Validate(Config{InviteSecret: defaultInviteSecret})
	if err == nil {
		t.Fatal("expected default secret to be rejected")
	}
}

func TestValidateRejectsShortSecret(t *testing.T) {
	err := Validate(Config{InviteSecret: "too-short"})
	if err == nil {
		t.Fatal("expected short secret to be rejected")
	}
}

func TestValidateAcceptsStrongSecret(t *testing.T) {
	secret := "01234567890123456789012345678901"
	if err := Validate(Config{InviteSecret: secret}); err != nil {
		t.Fatalf("expected valid secret, got %v", err)
	}
}
