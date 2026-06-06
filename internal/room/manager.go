// Package room stores in-memory room state and membership.
package room

import (
	"errors"
	"sync"
	"time"

	"huddle/internal/auth"
)

var (
	// ErrRoomNotFound is returned when a room ID does not exist.
	ErrRoomNotFound = errors.New("room not found")
	// ErrRoomFull is returned when a room has reached capacity.
	ErrRoomFull = errors.New("room is full")
	// ErrTooManyRooms is returned when the room registry is at capacity.
	ErrTooManyRooms = errors.New("too many rooms")
	// ErrJoinDenied is returned when a join attempt is rejected.
	ErrJoinDenied = errors.New("unable to join room")
)

// Manager owns the in-memory room registry.
type Manager struct {
	rooms    map[string]*Room
	mu       sync.RWMutex
	secret   string
	ttl      time.Duration
	maxSize  int
	maxRooms int
}

// NewManager returns a room registry configured with invite signing and capacity limits.
func NewManager(secret string, ttl time.Duration, maxSize int, maxRooms ...int) *Manager {
	limit := 1000
	if len(maxRooms) > 0 {
		limit = maxRooms[0]
	}
	return &Manager{
		rooms:    make(map[string]*Room),
		secret:   secret,
		ttl:      ttl,
		maxSize:  maxSize,
		maxRooms: limit,
	}
}

// CreateInput describes a new room request.
type CreateInput struct {
	Name     string
	Password string
}

// CreateResult is returned after a room is created.
type CreateResult struct {
	RoomID    string
	Invite    string
	RoomKey   string
	ExpiresAt time.Time
}

func (m *Manager) Create(in CreateInput) (CreateResult, error) {
	id, err := auth.NewRoomID()
	if err != nil {
		return CreateResult{}, err
	}
	expires := time.Now().Add(m.ttl)
	invite := auth.SignInvite(m.secret, id, expires)

	var hash string
	if in.Password != "" {
		hash, err = auth.HashPassword(in.Password)
		if err != nil {
			return CreateResult{}, err
		}
	}

	roomKey, err := auth.GenerateRoomKey()
	if err != nil {
		return CreateResult{}, err
	}

	m.mu.Lock()
	m.removeExpiredLocked(time.Now())
	if m.maxRooms > 0 && len(m.rooms) >= m.maxRooms {
		m.mu.Unlock()
		return CreateResult{}, ErrTooManyRooms
	}
	m.rooms[id] = New(id, in.Name, hash, invite, expires)
	m.mu.Unlock()

	return CreateResult{
		RoomID:    id,
		Invite:    invite,
		RoomKey:   roomKey,
		ExpiresAt: expires,
	}, nil
}

func (m *Manager) Get(id string) (*Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return r, nil
}

func (m *Manager) ValidateJoin(id, invite, password string) (*Room, error) {
	if err := auth.VerifyInvite(m.secret, id, invite); err != nil {
		return nil, ErrJoinDenied
	}
	r, err := m.Get(id)
	if err != nil {
		return nil, ErrJoinDenied
	}
	if r.Password != "" && !auth.CheckPassword(r.Password, password) {
		return nil, ErrJoinDenied
	}
	if r.Size() >= m.maxSize {
		return nil, ErrJoinDenied
	}
	return r, nil
}

func (m *Manager) RemoveIfEmpty(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[id]
	if !ok {
		return
	}
	if r.Size() == 0 {
		delete(m.rooms, id)
	}
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

func (m *Manager) CleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeExpiredLocked(time.Now())
}

func (m *Manager) Cleanup() {
	m.CleanupExpired()
}

func (m *Manager) removeExpiredLocked(now time.Time) {
	for id, r := range m.rooms {
		if now.After(r.ExpiresAt) {
			delete(m.rooms, id)
		}
	}
}
