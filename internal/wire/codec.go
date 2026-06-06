package wire

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidFrame  = errors.New("invalid wire frame")
	ErrPayloadTooBig = errors.New("wire payload too large")
)

type Pow struct {
	ID    string
	Nonce uint64
}

type Channel struct {
	ID   string
	Name string
}

type Member struct {
	ID       string
	Name     string
	Muted    bool
	Deafened bool
	Speaking bool
}

type Room struct {
	ID       string
	Name     string
	HostID   string
	Channels []Channel
	Members  []Member
}

type ICEServer struct {
	URLs       []string
	Username   string
	Credential string
}

type Signal struct {
	To    string
	From  string
	Nonce string
	Sig   string
	Kind  byte
	Body  []byte
}

type CreateRoom struct {
	Name     string
	Password string
	Pow      *Pow
}

type Join struct {
	RoomID       string
	Invite       string
	Password     string
	Name         string
	ResumePeerID string
	ResumeToken  string
	Pow          *Pow
}

type Created struct {
	RoomID    string
	Invite    string
	RoomKey   string
	ExpiresAt int64
}

type Joined struct {
	PeerID      string
	ResumeToken string
	Room        Room
	Peers       []string
	ICEServers  []ICEServer
}

type MemberUpdate struct {
	PeerID   string
	Muted    bool
	Deafened bool
	Speaking bool
}

type AddChannel struct {
	ID   string
	Name string
}

type Rename struct {
	Name string
}

type ErrorMsg struct {
	Message string
}

type PeerRef struct {
	PeerID string
}

type ModerateMember struct {
	PeerID   string
	Muted    bool
	Deafened bool
}

func EncodeFrame(msgType byte, payload []byte) ([]byte, error) {
	if len(payload) > maxPayloadSize {
		return nil, ErrPayloadTooBig
	}
	out := make([]byte, headerSize+len(payload))
	out[0] = Magic
	out[1] = msgType
	binary.BigEndian.PutUint32(out[2:6], uint32(len(payload)))
	copy(out[headerSize:], payload)
	return out, nil
}

func DecodeFrame(data []byte) (byte, []byte, error) {
	if len(data) < headerSize || data[0] != Magic {
		return 0, nil, ErrInvalidFrame
	}
	n := int(binary.BigEndian.Uint32(data[2:6]))
	if n < 0 || headerSize+n > len(data) {
		return 0, nil, ErrInvalidFrame
	}
	if n > maxPayloadSize {
		return 0, nil, ErrPayloadTooBig
	}
	return data[1], data[headerSize : headerSize+n], nil
}

func appendString(buf []byte, s string) ([]byte, error) {
	if len(s) > maxStringSize {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(s)))
	return append(buf, s...), nil
}

func appendBytes(buf []byte, b []byte) ([]byte, error) {
	if len(b) > maxPayloadSize {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(b)))
	return append(buf, b...), nil
}

func readString(payload []byte, off int) (string, int, error) {
	if off+4 > len(payload) {
		return "", off, ErrInvalidFrame
	}
	n := int(binary.BigEndian.Uint32(payload[off : off+4]))
	off += 4
	if n < 0 || off+n > len(payload) || n > maxStringSize {
		return "", off, ErrInvalidFrame
	}
	return string(payload[off : off+n]), off + n, nil
}

func readBytes(payload []byte, off int) ([]byte, int, error) {
	if off+4 > len(payload) {
		return nil, off, ErrInvalidFrame
	}
	n := int(binary.BigEndian.Uint32(payload[off : off+4]))
	off += 4
	if n < 0 || off+n > len(payload) || n > maxPayloadSize {
		return nil, off, ErrInvalidFrame
	}
	out := make([]byte, n)
	copy(out, payload[off:off+n])
	return out, off + n, nil
}

func appendPow(buf []byte, p *Pow) ([]byte, error) {
	if p == nil {
		return append(buf, 0), nil
	}
	buf = append(buf, 1)
	var err error
	buf, err = appendString(buf, p.ID)
	if err != nil {
		return nil, err
	}
	return binary.BigEndian.AppendUint64(buf, p.Nonce), nil
}

func readPow(payload []byte, off int) (*Pow, int, error) {
	if off >= len(payload) {
		return nil, off, ErrInvalidFrame
	}
	if payload[off] == 0 {
		return nil, off + 1, nil
	}
	off++
	var id string
	var err error
	id, off, err = readString(payload, off)
	if err != nil {
		return nil, off, err
	}
	if off+8 > len(payload) {
		return nil, off, ErrInvalidFrame
	}
	nonce := binary.BigEndian.Uint64(payload[off : off+8])
	return &Pow{ID: id, Nonce: nonce}, off + 8, nil
}

func appendMemberFlags(m Member) byte {
	var f byte
	if m.Muted {
		f |= flagMuted
	}
	if m.Deafened {
		f |= flagDeafened
	}
	if m.Speaking {
		f |= flagSpeaking
	}
	return f
}

func memberFromFlags(id, name string, f byte) Member {
	return Member{
		ID:       id,
		Name:     name,
		Muted:    f&flagMuted != 0,
		Deafened: f&flagDeafened != 0,
		Speaking: f&flagSpeaking != 0,
	}
}
