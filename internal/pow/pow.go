// Package pow implements proof-of-work challenges for room create and join.
package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrChallengeExpired  = errors.New("challenge expired")
	ErrInvalidSolution   = errors.New("invalid proof of work")
	ErrInvalidAction     = errors.New("invalid challenge action")
)

// Challenge is a single-use proof-of-work puzzle.
type Challenge struct {
	ID         string
	Prefix     string
	Difficulty int
	Action     string
	ExpiresAt  time.Time
}

// Store issues and verifies proof-of-work challenges.
type Store struct {
	mu         sync.Mutex
	challenges map[string]Challenge
}

// NewStore returns an empty challenge store.
func NewStore() *Store {
	return &Store{challenges: make(map[string]Challenge)}
}

// Issue creates a challenge for action with the given difficulty and lifetime.
func (s *Store) Issue(action string, difficulty int, ttl time.Duration) (Challenge, error) {
	if difficulty <= 0 {
		return Challenge{}, errors.New("pow disabled")
	}

	prefixBytes := make([]byte, 16)
	if _, err := rand.Read(prefixBytes); err != nil {
		return Challenge{}, err
	}
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return Challenge{}, err
	}

	ch := Challenge{
		ID:         hex.EncodeToString(idBytes),
		Prefix:     hex.EncodeToString(prefixBytes),
		Difficulty: difficulty,
		Action:     action,
		ExpiresAt:  time.Now().Add(ttl),
	}

	s.mu.Lock()
	s.challenges[ch.ID] = ch
	s.mu.Unlock()
	return ch, nil
}

// Verify checks and consumes a challenge solution.
func (s *Store) Verify(action, id string, nonce uint64) error {
	s.mu.Lock()
	ch, ok := s.challenges[id]
	if ok {
		delete(s.challenges, id)
	}
	s.mu.Unlock()

	if !ok {
		return ErrChallengeNotFound
	}
	if ch.Action != action {
		return ErrInvalidAction
	}
	if time.Now().After(ch.ExpiresAt) {
		return ErrChallengeExpired
	}
	if !Verify(ch.Prefix, nonce, ch.Difficulty) {
		return ErrInvalidSolution
	}
	return nil
}

func (s *Store) Cleanup() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, ch := range s.challenges {
		if now.After(ch.ExpiresAt) {
			delete(s.challenges, id)
		}
	}
}

// Verify reports whether nonce satisfies the proof-of-work for prefix.
func Verify(prefix string, nonce uint64, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}
	sum := sha256.Sum256(fmt.Appendf(nil, "%s:%d", prefix, nonce))
	return hasLeadingZeroBits(sum[:], difficulty)
}

func hasLeadingZeroBits(hash []byte, bits int) bool {
	if bits <= 0 {
		return true
	}
	fullBytes := bits / 8
	remBits := bits % 8
	for i := range fullBytes {
		if hash[i] != 0 {
			return false
		}
	}
	if remBits == 0 {
		return true
	}
	mask := byte(0xFF << (8 - remBits))
	return hash[fullBytes]&mask == 0
}
