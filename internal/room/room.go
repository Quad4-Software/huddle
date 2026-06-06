package room

import (
	"sort"
	"sync"
	"time"
)

// Channel is a named chat channel inside a room.
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Member is a connected peer in a room.
type Member struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Muted    bool   `json:"muted"`
	Deafened bool   `json:"deafened"`
	Speaking bool   `json:"speaking"`
}

// Room holds membership and metadata for a single session.
type Room struct {
	ID         string
	Name       string
	Password   string
	Invite     string
	ExpiresAt  time.Time
	HostPeerID string
	Channels   []Channel
	Members    map[string]*Member
	mu         sync.RWMutex
}

func New(id, name, password, invite string, expiresAt time.Time) *Room {
	return &Room{
		ID:        id,
		Name:      name,
		Password:  password,
		Invite:    invite,
		ExpiresAt: expiresAt,
		Channels: []Channel{
			{ID: "general", Name: "general"},
		},
		Members: make(map[string]*Member),
	}
}

func (r *Room) AddMember(id, name string) *Member {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := &Member{ID: id, Name: name}
	r.Members[id] = m
	return m
}

func (r *Room) RemoveMember(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Members, id)
}

func (r *Room) MemberList() []Member {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Member, 0, len(r.Members))
	for _, m := range r.Members {
		out = append(out, *m)
	}
	return out
}

func (r *Room) RenameMember(id, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.Members[id]; ok {
		m.Name = name
	}
}

func (r *Room) UpdateMember(id string, muted, deafened, speaking bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.Members[id]; ok {
		m.Muted = muted
		m.Deafened = deafened
		m.Speaking = speaking
	}
}

func (r *Room) SetHost(peerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.HostPeerID = peerID
}

func (r *Room) IsHost(peerID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.HostPeerID != "" && r.HostPeerID == peerID
}

func (r *Room) HostID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.HostPeerID
}

func (r *Room) TransferHost(except string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.HostPeerID != "" && r.HostPeerID != except {
		return
	}
	for id := range r.Members {
		if id != except {
			r.HostPeerID = id
			return
		}
	}
	r.HostPeerID = ""
}

func (r *Room) AddChannel(id, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.Channels {
		if c.ID == id {
			return
		}
	}
	r.Channels = append(r.Channels, Channel{ID: id, Name: name})
}

func (r *Room) Snapshot() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return map[string]any{
		"id":       r.ID,
		"name":     r.Name,
		"hostId":   r.HostPeerID,
		"channels": r.Channels,
		"members":  r.memberSlice(),
	}
}

func (r *Room) memberSlice() []Member {
	out := make([]Member, 0, len(r.Members))
	for _, m := range r.Members {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (r *Room) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Members)
}
