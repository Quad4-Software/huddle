// Package hub relays WebSocket signaling and room lifecycle events.
package hub

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"

	"huddle/internal/pow"
	"huddle/internal/ratelimit"
	"huddle/internal/room"
)

// Limits holds per-action rate limiters for hub operations.
type Limits struct {
	Create *ratelimit.Limiter
	Join   *ratelimit.Limiter
}

// Hub routes WebSocket messages between clients in each room.
type Hub struct {
	rooms         *room.Manager
	limits        Limits
	pow           *pow.Store
	powDifficulty int
	clients       map[string]map[string]*Client
	mu            sync.RWMutex
	register      chan *Client
	unregister    chan *Client
}

// New returns a hub backed by room manager rm.
func New(rm *room.Manager, limits Limits, powStore *pow.Store, powDifficulty int) *Hub {
	return &Hub{
		rooms:         rm,
		limits:        limits,
		pow:           powStore,
		powDifficulty: powDifficulty,
		clients:       make(map[string]map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

// Run processes client registration and teardown until the process exits.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.clients[c.RoomID] == nil {
				h.clients[c.RoomID] = make(map[string]*Client)
			}
			h.clients[c.RoomID][c.ID] = c
			h.mu.Unlock()
			h.broadcastRoomState(c.RoomID)

		case c := <-h.unregister:
			h.removeClient(c)
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	shouldRemove := false
	c.removed.Do(func() {
		shouldRemove = true
	})
	if !shouldRemove {
		return
	}

	h.mu.Lock()
	roomClients, ok := h.clients[c.RoomID]
	if ok {
		delete(roomClients, c.ID)
		if len(roomClients) == 0 {
			delete(h.clients, c.RoomID)
		}
	}
	h.mu.Unlock()

	if c.RoomID != "" {
		if r, err := h.rooms.Get(c.RoomID); err == nil {
			wasHost := r.IsHost(c.ID)
			r.RemoveMember(c.ID)
			if wasHost {
				r.TransferHost(c.ID)
			}
		}
		h.broadcast(c.RoomID, TypePeerLeft, PeerLeftPayload{PeerID: c.ID}, c.ID)
		h.broadcastRoomState(c.RoomID)
		h.rooms.RemoveIfEmpty(c.RoomID)
	}
	close(c.Send)
}

func (h *Hub) handleMessage(c *Client, data []byte) {
	msg, err := Unmarshal(data)
	if err != nil {
		c.SendError("invalid message")
		return
	}

	switch msg.Type {
	case TypeCreateRoom:
		h.handleCreate(c, msg.Payload)
	case TypeJoin:
		h.handleJoin(c, msg.Payload)
	case TypeOffer, TypeAnswer, TypeICE:
		h.handleSignal(c, msg)
	case TypeMemberUpdate:
		h.handleMemberUpdate(c, msg.Payload)
	case TypeAddChannel:
		h.handleAddChannel(c, msg.Payload)
	case TypeRename:
		h.handleRename(c, msg.Payload)
	case TypePing:
		h.handlePing(c, msg.Payload)
	case TypeLeave:
		h.unregister <- c
	case TypeKick:
		h.handleKick(c, msg.Payload)
	default:
		c.SendError("unknown message type")
	}
}

func (h *Hub) handleRename(c *Client, raw []byte) {
	var p RenamePayload
	if err := decodePayload(raw, &p); err != nil || p.Name == "" {
		return
	}
	c.Name = p.Name
	if r, err := h.rooms.Get(c.RoomID); err == nil {
		r.RenameMember(c.ID, p.Name)
	}
	h.broadcastRoomState(c.RoomID)
}

func (h *Hub) handlePing(c *Client, raw []byte) {
	var p PingPayload
	if err := decodePayload(raw, &p); err != nil {
		return
	}
	c.SendJSON(TypePong, p)
}

func (h *Hub) handleCreate(c *Client, raw []byte) {
	if !h.limits.Create.Allow(c.IP) {
		c.SendError("rate limit exceeded")
		return
	}
	var p CreateRoomPayload
	if err := decodePayload(raw, &p); err != nil || p.Name == "" {
		c.SendError("invalid create request")
		return
	}
	if err := h.verifyPow("create", p.Pow); err != nil {
		c.SendError(err.Error())
		return
	}
	result, err := h.rooms.Create(room.CreateInput{Name: p.Name, Password: p.Password})
	if err != nil {
		c.SendError("failed to create room")
		return
	}
	c.CreatedRoomID = result.RoomID
	c.SendJSON(TypeCreated, CreatedPayload{
		RoomID:    result.RoomID,
		Invite:    result.Invite,
		RoomKey:   result.RoomKey,
		ExpiresAt: result.ExpiresAt.Unix(),
	})
}

func (h *Hub) handleJoin(c *Client, raw []byte) {
	if !h.limits.Join.Allow(c.IP) {
		c.SendError("rate limit exceeded")
		return
	}
	var p JoinPayload
	if err := decodePayload(raw, &p); err != nil || p.RoomID == "" || p.Invite == "" || p.Name == "" {
		c.SendError("invalid join request")
		return
	}
	if err := h.verifyPow("join", p.Pow); err != nil {
		c.SendError(err.Error())
		return
	}
	r, err := h.rooms.ValidateJoin(p.RoomID, p.Invite, p.Password)
	if err != nil {
		c.SendError(err.Error())
		return
	}

	peerID, err := newPeerID()
	if err != nil {
		c.SendError("failed to join room")
		return
	}
	c.ID = peerID
	c.RoomID = p.RoomID
	c.Name = p.Name
	r.AddMember(c.ID, p.Name)
	if p.RoomID == c.CreatedRoomID && r.HostID() == "" {
		r.SetHost(c.ID)
	}

	h.register <- c

	peers := h.peerIDs(c.RoomID, c.ID)
	c.SendJSON(TypeJoined, JoinedPayload{
		PeerID: c.ID,
		Room:   r.Snapshot(),
		Peers:  peers,
	})
	h.broadcast(c.RoomID, TypePeerJoined, PeerJoinedPayload{PeerID: c.ID}, c.ID)
}

func (h *Hub) handleSignal(c *Client, msg Message) {
	var p SignalPayload
	if err := decodePayload(msg.Payload, &p); err != nil || p.To == "" {
		return
	}
	p.From = c.ID
	h.mu.RLock()
	target, ok := h.clients[c.RoomID][p.To]
	h.mu.RUnlock()
	if !ok {
		return
	}
	target.SendJSON(msg.Type, p)
}

func (h *Hub) handleMemberUpdate(c *Client, raw []byte) {
	var p MemberUpdatePayload
	if err := decodePayload(raw, &p); err != nil {
		return
	}
	if r, err := h.rooms.Get(c.RoomID); err == nil {
		r.UpdateMember(c.ID, p.Muted, p.Deafened, p.Speaking)
	}
	p.PeerID = c.ID
	h.broadcast(c.RoomID, TypeMemberUpdate, p, c.ID)
}

func (h *Hub) handleKick(c *Client, raw []byte) {
	var p KickPayload
	if err := decodePayload(raw, &p); err != nil || p.PeerID == "" {
		c.SendError("invalid kick request")
		return
	}
	if c.RoomID == "" || p.PeerID == c.ID {
		c.SendError("invalid kick request")
		return
	}
	r, err := h.rooms.Get(c.RoomID)
	if err != nil {
		c.SendError("room not found")
		return
	}
	if !r.IsHost(c.ID) {
		c.SendError("only the host can kick members")
		return
	}

	h.mu.RLock()
	target, ok := h.clients[c.RoomID][p.PeerID]
	h.mu.RUnlock()
	if !ok {
		c.SendError("member not found")
		return
	}

	target.SendJSON(TypeKicked, nil)
	h.unregister <- target
}

func (h *Hub) verifyPow(action string, payload *PowPayload) error {
	if h.powDifficulty <= 0 || h.pow == nil {
		return nil
	}
	if payload == nil || payload.ID == "" {
		return errors.New("proof of work required")
	}
	if err := h.pow.Verify(action, payload.ID, payload.Nonce); err != nil {
		return errors.New("invalid proof of work")
	}
	return nil
}

func (h *Hub) handleAddChannel(c *Client, raw []byte) {
	var p AddChannelPayload
	if err := decodePayload(raw, &p); err != nil || p.ID == "" || p.Name == "" {
		return
	}
	if r, err := h.rooms.Get(c.RoomID); err == nil {
		r.AddChannel(p.ID, p.Name)
	}
	h.broadcastRoomState(c.RoomID)
}

func (h *Hub) broadcastRoomState(roomID string) {
	r, err := h.rooms.Get(roomID)
	if err != nil {
		return
	}
	h.broadcast(roomID, TypeRoomState, r.Snapshot(), "")
}

func (h *Hub) broadcast(roomID string, t MessageType, payload any, except string) {
	data, err := Marshal(t, payload)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, c := range h.clients[roomID] {
		if id == except {
			continue
		}
		select {
		case c.Send <- data:
		default:
			log.Printf("drop message to %s", id)
		}
	}
}

func (h *Hub) peerIDs(roomID, except string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0)
	for id := range h.clients[roomID] {
		if id != except {
			ids = append(ids, id)
		}
	}
	return ids
}

func newPeerID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ServeClient handles a WebSocket connection until it closes.
func (h *Hub) ServeClient(conn *websocket.Conn, ip string) {
	c := &Client{
		Hub:  h,
		Conn: conn,
		IP:   ip,
		Send: make(chan []byte, 64),
	}
	go c.writePump()
	c.readPump()
}
