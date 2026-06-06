package config

import (
	"errors"
	"fmt"
)

const defaultInviteSecret = "change-me-in-production"

// Validate checks production-critical settings before startup.
func Validate(cfg Config) error {
	if cfg.InviteSecret == "" || cfg.InviteSecret == defaultInviteSecret {
		return errors.New("set HUDDLE_INVITE_SECRET to a long random secret")
	}
	if len(cfg.InviteSecret) < 32 {
		return fmt.Errorf("HUDDLE_INVITE_SECRET must be at least 32 characters, got %d", len(cfg.InviteSecret))
	}
	return nil
}
