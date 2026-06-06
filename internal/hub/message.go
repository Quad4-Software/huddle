package hub

import "encoding/json"

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
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type PowPayload struct {
	ID    string `json:"id"`
	Nonce uint64 `json:"nonce"`
}

type CreateRoomPayload struct {
	Name     string      `json:"name"`
	Password string      `json:"password,omitempty"`
	Pow      *PowPayload `json:"pow,omitempty"`
}

type CreatedPayload struct {
	RoomID    string `json:"roomId"`
	Invite    string `json:"invite"`
	RoomKey   string `json:"roomKey"`
	ExpiresAt int64  `json:"expiresAt"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type JoinPayload struct {
	RoomID       string      `json:"roomId"`
	Invite       string      `json:"invite"`
	Password     string      `json:"password,omitempty"`
	Name         string      `json:"name"`
	ResumePeerID string      `json:"resumePeerId,omitempty"`
	ResumeToken  string      `json:"resumeToken,omitempty"`
	Pow          *PowPayload `json:"pow,omitempty"`
}

type KickPayload struct {
	PeerID string `json:"peerId"`
}

type ModerateMemberPayload struct {
	PeerID   string `json:"peerId"`
	Muted    bool   `json:"muted"`
	Deafened bool   `json:"deafened"`
}

type PeerLeftPayload struct {
	PeerID string `json:"peerId"`
}

type JoinedPayload struct {
	PeerID      string         `json:"peerId"`
	ResumeToken string         `json:"resumeToken"`
	Room        map[string]any `json:"room"`
	Peers       []string       `json:"peers"`
	ICEServers  []ICEServer    `json:"iceServers,omitempty"`
}

type SignalPayload struct {
	To        string          `json:"to"`
	From      string          `json:"from,omitempty"`
	SDP       string          `json:"sdp,omitempty"`
	Candidate json.RawMessage `json:"candidate,omitempty"`
	Nonce     string          `json:"nonce,omitempty"`
	Sig       string          `json:"sig,omitempty"`
}

type MemberUpdatePayload struct {
	PeerID   string `json:"peerId"`
	Muted    bool   `json:"muted"`
	Deafened bool   `json:"deafened"`
	Speaking bool   `json:"speaking"`
}

type AddChannelPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PeerJoinedPayload struct {
	PeerID string `json:"peerId"`
}

type RenamePayload struct {
	Name string `json:"name"`
}

type PingPayload struct {
	T int64 `json:"t"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

func marshalWithPayload(t MessageType, payload json.RawMessage) ([]byte, error) {
	return json.Marshal(Message{Type: t, Payload: payload})
}

func Marshal(t MessageType, payload any) ([]byte, error) {
	if payload == nil {
		return json.Marshal(Message{Type: t})
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return marshalWithPayload(t, b)
}

func Unmarshal(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return m, err
}
