package hub

import (
	"huddle/internal/wire"
)

type MessageType string

const (
	TypeCreateRoom     MessageType = "create_room"
	TypeJoin           MessageType = "join"
	TypeLeave          MessageType = "leave"
	TypeOffer          MessageType = "offer"
	TypeAnswer         MessageType = "answer"
	TypeICE            MessageType = "ice"
	TypeRoomState      MessageType = "room_state"
	TypeMemberUpdate   MessageType = "member_update"
	TypeAddChannel     MessageType = "add_channel"
	TypeError          MessageType = "error"
	TypeJoined         MessageType = "joined"
	TypeCreated        MessageType = "created"
	TypePeerJoined     MessageType = "peer_joined"
	TypeRename         MessageType = "rename"
	TypePing           MessageType = "ping"
	TypePong           MessageType = "pong"
	TypeKick           MessageType = "kick"
	TypeKicked         MessageType = "kicked"
	TypeModerateMember MessageType = "moderate_member"
	TypePeerLeft       MessageType = "peer_left"
)

type Message struct {
	Type    MessageType
	Payload []byte
}

type PowPayload struct {
	ID    string
	Nonce uint64
}

type CreateRoomPayload struct {
	Name     string
	Password string
	Pow      *PowPayload
}

type CreatedPayload struct {
	RoomID    string
	Invite    string
	RoomKey   string
	ExpiresAt int64
}

type ICEServer struct {
	URLs       []string
	Username   string
	Credential string
}

type JoinPayload struct {
	RoomID       string
	Invite       string
	Password     string
	Name         string
	ResumePeerID string
	ResumeToken  string
	Pow          *PowPayload
}

type KickPayload struct {
	PeerID string
}

type ModerateMemberPayload struct {
	PeerID   string
	Muted    bool
	Deafened bool
}

type PeerLeftPayload struct {
	PeerID string
}

type JoinedPayload struct {
	PeerID      string
	Room        map[string]any
	ResumeToken string
	Peers       []string
	ICEServers  []ICEServer
}

type SignalPayload struct {
	To        string
	From      string
	SDP       string
	Candidate []byte
	Nonce     string
	Sig       string
}

type MemberUpdatePayload struct {
	PeerID   string
	Muted    bool
	Deafened bool
	Speaking bool
}

type AddChannelPayload struct {
	ID   string
	Name string
}

type PeerJoinedPayload struct {
	PeerID string
}

type RenamePayload struct {
	Name string
}

type PingPayload struct {
	T int64
}

type ErrorPayload struct {
	Message string
}

func Marshal(t MessageType, payload any) ([]byte, error) {
	wt, ok := wireType(t)
	if !ok {
		return nil, wire.ErrInvalidFrame
	}
	body, err := encodePayload(t, payload)
	if err != nil {
		return nil, err
	}
	return wire.EncodeFrame(wt, body)
}

func Unmarshal(data []byte) (Message, error) {
	wt, payload, err := wire.DecodeFrame(data)
	if err != nil {
		return Message{}, err
	}
	t, ok := typeFromWire(wt)
	if !ok {
		return Message{}, wire.ErrInvalidFrame
	}
	return Message{Type: t, Payload: payload}, nil
}

func marshalPong(raw []byte) ([]byte, error) {
	return wire.EncodeFrame(wire.MsgPong, raw)
}
