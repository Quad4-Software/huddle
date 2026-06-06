// Package hub relays WebSocket signaling and room lifecycle events.
package hub

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strings"
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
	iceProvider   func() []ICEServer
	clients       map[string]map[string]*Client
	mu            sync.RWMutex
	unregister    chan *Client
}

// New returns a hub backed by room manager rm.
func New(
	rm *room.Manager,
	limits Limits,
	powStore *pow.Store,
	powDifficulty int,
	iceProvider ...func() []ICEServer,
) *Hub {
	var provider func() []ICEServer
	if len(iceProvider) > 0 {
		provider = iceProvider[0]
	}
	return &Hub{
		rooms:         rm,
		limits:        limits,
		pow:           powStore,
		powDifficulty: powDifficulty,
		iceProvider:   provider,
		clients:       make(map[string]map[string]*Client),
		unregister:    make(chan *Client),
	}
}

// Run processes client registration and teardown until the process exits.
func (h *Hub) Run() {
	for c := range h.unregister {
		h.safe(func() { h.removeClient(c, false, true) })
	}
}

func (h *Hub) safe(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("hub panic recovered: %v", r)
		}
	}()
	fn()
}

func (h *Hub) registerClient(c *Client) {
	h.mu.Lock()
	if h.clients[c.RoomID] == nil {
		h.clients[c.RoomID] = make(map[string]*Client)
	}
	roomClients := h.clients[c.RoomID]
	for id, client := range roomClients {
		if client == c && id != c.ID {
			delete(roomClients, id)
		}
	}
	roomClients[c.ID] = c
	h.mu.Unlock()
	h.broadcastRoomState(c.RoomID, c.ID)
}

func (h *Hub) removeClient(c *Client, keepRoom bool, notify bool) {
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

	if c.RoomID != "" && notify {
		if r, err := h.rooms.Get(c.RoomID); err == nil {
			wasHost := r.IsHost(c.ID)
			r.RemoveMember(c.ID)
			if wasHost {
				r.TransferHost(c.ID)
			}
		}
		h.broadcast(c.RoomID, TypePeerLeft, PeerLeftPayload{PeerID: c.ID}, c.ID)
		h.reconcileRoomMembers(c.RoomID)
		h.broadcastRoomState(c.RoomID, "")
		if !keepRoom {
			h.rooms.RemoveIfEmpty(c.RoomID)
		}
	}
	c.shutdown()
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
	case TypeModerateMember:
		h.handleModerateMember(c, msg.Payload)
	default:
		c.SendError("unknown message type")
	}
}

func (h *Hub) handleRename(c *Client, raw []byte) {
	var p RenamePayload
	if err := decodePayload(raw, &p); err != nil {
		return
	}
	name, ok := cleanBounded(p.Name, MaxDisplayNameLength)
	if !ok {
		return
	}
	c.Name = name
	if r, err := h.rooms.Get(c.RoomID); err == nil {
		r.RenameMember(c.ID, name)
	}
	h.broadcastRoomState(c.RoomID, "")
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
	if err := decodePayload(raw, &p); err != nil {
		c.SendError("invalid create request")
		return
	}
	name, ok := cleanBounded(p.Name, MaxRoomNameLength)
	if !ok || !validateOptionalBounded(p.Password, MaxPasswordLength) {
		c.SendError("invalid create request")
		return
	}
	if err := h.verifyPow("create", p.Pow); err != nil {
		c.SendError(err.Error())
		return
	}
	result, err := h.rooms.Create(room.CreateInput{Name: name, Password: p.Password})
	if err != nil {
		c.SendError("failed to create room")
		return
	}
	c.CreatedRoomID = result.RoomID
	c.sendCriticalJSON(TypeCreated, CreatedPayload{
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
	if err := decodePayload(raw, &p); err != nil {
		c.SendError("invalid join request")
		return
	}
	name, ok := cleanBounded(p.Name, MaxDisplayNameLength)
	if !ok || p.RoomID == "" || p.Invite == "" || !validateOptionalBounded(p.Password, MaxPasswordLength) {
		c.SendError("invalid join request")
		return
	}
	if err := h.verifyPow("join", p.Pow); err != nil {
		c.SendError(err.Error())
		return
	}
	if _, err := h.rooms.Get(p.RoomID); err == nil {
		h.reconcileRoomMembers(p.RoomID)
	}
	r, err := h.rooms.ValidateJoin(p.RoomID, p.Invite, p.Password)
	if err != nil {
		c.SendError(err.Error())
		return
	}

	peerID := strings.TrimSpace(p.ResumePeerID)
	resumeToken := strings.TrimSpace(p.ResumeToken)
	resuming := false
	if peerID != "" {
		if !r.VerifyResumeToken(peerID, resumeToken) {
			peerID = ""
		} else if _, ok := r.GetMember(peerID); ok {
			h.disconnectPeer(p.RoomID, peerID, true, false)
			resuming = true
		} else {
			peerID = ""
		}
	}
	if peerID == "" {
		peerID, err = newPeerID()
		if err != nil {
			c.SendError("failed to join room")
			return
		}
		r.AddMember(peerID, name)
		if p.RoomID == c.CreatedRoomID && r.HostID() == "" {
			r.SetHost(peerID)
		}
	} else {
		r.RenameMember(peerID, name)
	}

	if c.RoomID == p.RoomID && c.ID != "" && c.ID != peerID {
		if stale, err := h.rooms.Get(p.RoomID); err == nil {
			stale.RemoveMember(c.ID)
		}
		h.mu.Lock()
		if roomClients := h.clients[p.RoomID]; roomClients != nil {
			delete(roomClients, c.ID)
		}
		h.mu.Unlock()
	}

	c.ID = peerID
	c.RoomID = p.RoomID
	c.Name = name

	resumeToken, err = newResumeToken()
	if err != nil {
		c.SendError("failed to join room")
		return
	}
	r.SetResumeToken(peerID, resumeToken)

	h.registerClient(c)
	h.reconcileRoomMembers(c.RoomID)

	peers := h.peerIDs(c.RoomID, c.ID)
	c.sendCriticalJSON(TypeJoined, JoinedPayload{
		PeerID:      c.ID,
		ResumeToken: resumeToken,
		Room:        r.Snapshot(),
		Peers:       peers,
		ICEServers:  h.iceServers(),
	})
	if !resuming {
		h.broadcast(c.RoomID, TypePeerJoined, PeerJoinedPayload{PeerID: c.ID}, c.ID)
	} else {
		h.broadcastRoomState(c.RoomID, "")
	}
}

func (h *Hub) handleSignal(c *Client, msg Message) {
	var p SignalPayload
	if err := decodePayload(msg.Payload, &p); err != nil || !validSignal(msg.Type, p) {
		return
	}
	p.From = c.ID
	h.mu.RLock()
	target, ok := h.clients[c.RoomID][p.To]
	h.mu.RUnlock()
	if !ok {
		return
	}
	target.sendCriticalJSON(msg.Type, p)
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

	target.sendCriticalJSON(TypeKicked, nil)
	h.unregister <- target
}

func (h *Hub) handleModerateMember(c *Client, raw []byte) {
	var p ModerateMemberPayload
	if err := decodePayload(raw, &p); err != nil || p.PeerID == "" {
		c.SendError("invalid moderate request")
		return
	}
	if c.RoomID == "" || p.PeerID == c.ID {
		c.SendError("invalid moderate request")
		return
	}
	r, err := h.rooms.Get(c.RoomID)
	if err != nil {
		c.SendError("room not found")
		return
	}
	if !r.IsHost(c.ID) {
		c.SendError("only the host can moderate members")
		return
	}

	h.mu.RLock()
	_, ok := h.clients[c.RoomID][p.PeerID]
	h.mu.RUnlock()
	if !ok {
		c.SendError("member not found")
		return
	}

	muted := p.Muted || p.Deafened
	r.UpdateMember(p.PeerID, muted, p.Deafened, false)
	h.broadcast(c.RoomID, TypeMemberUpdate, MemberUpdatePayload{
		PeerID:   p.PeerID,
		Muted:    muted,
		Deafened: p.Deafened,
		Speaking: false,
	}, "")
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
	if err := decodePayload(raw, &p); err != nil {
		return
	}
	id, okID := cleanBounded(p.ID, MaxChannelIDLength)
	name, okName := cleanBounded(p.Name, MaxChannelNameLength)
	if !okID || !okName {
		return
	}
	r, err := h.rooms.Get(c.RoomID)
	if err != nil || !r.IsHost(c.ID) {
		return
	}
	r.AddChannel(id, name)
	h.broadcastRoomState(c.RoomID, "")
}

func (h *Hub) iceServers() []ICEServer {
	if h.iceProvider == nil {
		return nil
	}
	return h.iceProvider()
}

func validSignal(t MessageType, p SignalPayload) bool {
	if p.To == "" || p.Nonce == "" || p.Sig == "" ||
		len(p.Nonce) > MaxSignalNonceLength || len(p.Sig) > MaxSignalSigLength {
		return false
	}
	switch t {
	case TypeOffer, TypeAnswer:
		return p.SDP != "" && len(p.SDP) <= MaxSDPLength
	case TypeICE:
		if p.Candidate == nil {
			return false
		}
		b, err := json.Marshal(p.Candidate)
		return err == nil && len(b) <= MaxCandidateLength
	default:
		return false
	}
}

func (h *Hub) broadcastRoomState(roomID, except string) {
	r, err := h.rooms.Get(roomID)
	if err != nil {
		return
	}
	h.broadcast(roomID, TypeRoomState, r.Snapshot(), except)
}

func (h *Hub) broadcast(roomID string, t MessageType, payload any, except string) {
	data, err := Marshal(t, payload)
	if err != nil {
		return
	}
	h.mu.RLock()
	targets := make([]*Client, 0, len(h.clients[roomID]))
	for id, c := range h.clients[roomID] {
		if id != except {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range targets {
		if !c.enqueue(data) {
			log.Printf("drop message to %s, disconnecting stale client", c.ID)
			h.unregister <- c
		}
	}
}

func (h *Hub) disconnectPeer(roomID, peerID string, keepRoom, notify bool) {
	h.mu.RLock()
	c := h.clients[roomID][peerID]
	h.mu.RUnlock()
	if c != nil {
		h.removeClient(c, keepRoom, notify)
	}
}

func (h *Hub) reconcileRoomMembers(roomID string) {
	r, err := h.rooms.Get(roomID)
	if err != nil {
		return
	}
	h.mu.RLock()
	active := h.clients[roomID]
	activeIDs := make(map[string]struct{}, len(active))
	for id := range active {
		activeIDs[id] = struct{}{}
	}
	h.mu.RUnlock()

	for _, m := range r.MemberList() {
		if _, ok := activeIDs[m.ID]; ok {
			continue
		}
		wasHost := r.IsHost(m.ID)
		r.RemoveMember(m.ID)
		if wasHost {
			r.TransferHost(m.ID)
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

func newResumeToken() (string, error) {
	b := make([]byte, 16)
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
		Send: make(chan []byte, 1024),
	}
	go c.writePump()
	c.readPump()
}
