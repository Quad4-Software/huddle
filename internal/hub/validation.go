package hub

import (
	"errors"
	"strings"
)

const (
	MaxRoomNameLength    = 80
	MaxDisplayNameLength = 40
	MaxPasswordLength    = 256
	MaxChannelIDLength   = 64
	MaxChannelNameLength = 80
	MaxSDPLength         = 32768
	MaxCandidateLength   = 8192
	MaxSignalNonceLength = 64
	MaxSignalSigLength   = 128
)

var errInvalidInput = errors.New("invalid input")

func cleanBounded(value string, max int) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > max {
		return "", false
	}
	return value, true
}

func validateOptionalBounded(value string, max int) bool {
	return len(value) <= max
}
