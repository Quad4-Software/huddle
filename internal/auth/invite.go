// Package auth issues invite tokens, room keys, and password hashes.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrInvalidInvite is returned when an invite token fails validation.
var ErrInvalidInvite = errors.New("invalid invite token")

// Invite is the decoded payload of a signed invite token.
type Invite struct {
	RoomID    string
	ExpiresAt time.Time
}

// NewRoomID returns a random hex-encoded room identifier.
func NewRoomID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SignInvite returns an HMAC-signed invite token for roomID.
func SignInvite(secret, roomID string, expiresAt time.Time) string {
	payload := fmt.Sprintf("%s:%d", roomID, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := mac.Sum(nil)
	token := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(sig)
	return token
}

// VerifyInvite checks that token is valid for roomID and not expired.
func VerifyInvite(secret, roomID, token string) error {
	payload, err := parseToken(secret, token)
	if err != nil {
		return err
	}
	if payload.RoomID != roomID {
		return ErrInvalidInvite
	}
	if time.Now().After(payload.ExpiresAt) {
		return ErrInvalidInvite
	}
	return nil
}

func parseToken(secret, token string) (Invite, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Invite{}, ErrInvalidInvite
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Invite{}, ErrInvalidInvite
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Invite{}, ErrInvalidInvite
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	if !hmac.Equal(mac.Sum(nil), sig) {
		return Invite{}, ErrInvalidInvite
	}
	segments := strings.Split(string(raw), ":")
	if len(segments) != 2 {
		return Invite{}, ErrInvalidInvite
	}
	exp, err := strconv.ParseInt(segments[1], 10, 64)
	if err != nil {
		return Invite{}, ErrInvalidInvite
	}
	return Invite{RoomID: segments[0], ExpiresAt: time.Unix(exp, 0)}, nil
}

// GenerateRoomKey returns a base64url-encoded AES-256 key for client-side E2E encryption.
func GenerateRoomKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
