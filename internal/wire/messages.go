package wire

import (
	"encoding/binary"
	"encoding/json"
)

func EncodeCreateRoom(p CreateRoom) ([]byte, error) {
	buf := make([]byte, 0, 128)
	var err error
	buf, err = appendString(buf, p.Name)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Password)
	if err != nil {
		return nil, err
	}
	return appendPow(buf, p.Pow)
}

func DecodeCreateRoom(payload []byte) (CreateRoom, error) {
	off := 0
	var out CreateRoom
	var err error
	out.Name, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Password, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Pow, off, err = readPow(payload, off)
	return out, err
}

func EncodeJoin(p Join) ([]byte, error) {
	buf := make([]byte, 0, 256)
	var err error
	buf, err = appendString(buf, p.RoomID)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Invite)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Password)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Name)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.ResumePeerID)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.ResumeToken)
	if err != nil {
		return nil, err
	}
	return appendPow(buf, p.Pow)
}

func DecodeJoin(payload []byte) (Join, error) {
	off := 0
	var out Join
	var err error
	out.RoomID, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Invite, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Password, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Name, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.ResumePeerID, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.ResumeToken, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Pow, off, err = readPow(payload, off)
	return out, err
}

func EncodeCreated(p Created) ([]byte, error) {
	buf := make([]byte, 0, 256)
	var err error
	buf, err = appendString(buf, p.RoomID)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Invite)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.RoomKey)
	if err != nil {
		return nil, err
	}
	return binary.BigEndian.AppendUint64(buf, uint64(p.ExpiresAt)), nil
}

func DecodeCreated(payload []byte) (Created, error) {
	off := 0
	var out Created
	var err error
	out.RoomID, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Invite, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.RoomKey, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	if off+8 > len(payload) {
		return out, ErrInvalidFrame
	}
	out.ExpiresAt = int64(binary.BigEndian.Uint64(payload[off : off+8]))
	return out, nil
}

func appendRoom(buf []byte, r Room) ([]byte, error) {
	var err error
	buf, err = appendString(buf, r.ID)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, r.Name)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, r.HostID)
	if err != nil {
		return nil, err
	}
	if len(r.Channels) > 0xffff {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(r.Channels)))
	for _, ch := range r.Channels {
		buf, err = appendString(buf, ch.ID)
		if err != nil {
			return nil, err
		}
		buf, err = appendString(buf, ch.Name)
		if err != nil {
			return nil, err
		}
	}
	if len(r.Members) > 0xffff {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(r.Members)))
	for _, m := range r.Members {
		buf, err = appendString(buf, m.ID)
		if err != nil {
			return nil, err
		}
		buf, err = appendString(buf, m.Name)
		if err != nil {
			return nil, err
		}
		buf = append(buf, appendMemberFlags(m))
	}
	return buf, nil
}

func readRoom(payload []byte, off int) (Room, int, error) {
	var r Room
	var err error
	r.ID, off, err = readString(payload, off)
	if err != nil {
		return r, off, err
	}
	r.Name, off, err = readString(payload, off)
	if err != nil {
		return r, off, err
	}
	r.HostID, off, err = readString(payload, off)
	if err != nil {
		return r, off, err
	}
	if off+2 > len(payload) {
		return r, off, ErrInvalidFrame
	}
	chCount := int(binary.BigEndian.Uint16(payload[off : off+2]))
	off += 2
	r.Channels = make([]Channel, 0, chCount)
	for range chCount {
		var ch Channel
		ch.ID, off, err = readString(payload, off)
		if err != nil {
			return r, off, err
		}
		ch.Name, off, err = readString(payload, off)
		if err != nil {
			return r, off, err
		}
		r.Channels = append(r.Channels, ch)
	}
	if off+2 > len(payload) {
		return r, off, ErrInvalidFrame
	}
	mCount := int(binary.BigEndian.Uint16(payload[off : off+2]))
	off += 2
	r.Members = make([]Member, 0, mCount)
	for range mCount {
		var id, name string
		id, off, err = readString(payload, off)
		if err != nil {
			return r, off, err
		}
		name, off, err = readString(payload, off)
		if err != nil {
			return r, off, err
		}
		if off >= len(payload) {
			return r, off, ErrInvalidFrame
		}
		r.Members = append(r.Members, memberFromFlags(id, name, payload[off]))
		off++
	}
	return r, off, nil
}

func appendICEServers(buf []byte, servers []ICEServer) ([]byte, error) {
	if len(servers) > 0xffff {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(servers)))
	for _, s := range servers {
		if len(s.URLs) > 0xffff {
			return nil, ErrPayloadTooBig
		}
		buf = binary.BigEndian.AppendUint16(buf, uint16(len(s.URLs)))
		var err error
		for _, u := range s.URLs {
			buf, err = appendString(buf, u)
			if err != nil {
				return nil, err
			}
		}
		buf, err = appendString(buf, s.Username)
		if err != nil {
			return nil, err
		}
		buf, err = appendString(buf, s.Credential)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func readICEServers(payload []byte, off int) ([]ICEServer, int, error) {
	if off+2 > len(payload) {
		return nil, off, ErrInvalidFrame
	}
	count := int(binary.BigEndian.Uint16(payload[off : off+2]))
	off += 2
	out := make([]ICEServer, 0, count)
	for range count {
		if off+2 > len(payload) {
			return nil, off, ErrInvalidFrame
		}
		urlCount := int(binary.BigEndian.Uint16(payload[off : off+2]))
		off += 2
		s := ICEServer{URLs: make([]string, 0, urlCount)}
		var err error
		for range urlCount {
			var u string
			u, off, err = readString(payload, off)
			if err != nil {
				return nil, off, err
			}
			s.URLs = append(s.URLs, u)
		}
		s.Username, off, err = readString(payload, off)
		if err != nil {
			return nil, off, err
		}
		s.Credential, off, err = readString(payload, off)
		if err != nil {
			return nil, off, err
		}
		out = append(out, s)
	}
	return out, off, nil
}

func EncodeJoined(p Joined) ([]byte, error) {
	buf := make([]byte, 0, 512)
	var err error
	buf, err = appendString(buf, p.PeerID)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.ResumeToken)
	if err != nil {
		return nil, err
	}
	buf, err = appendRoom(buf, p.Room)
	if err != nil {
		return nil, err
	}
	if len(p.Peers) > 0xffff {
		return nil, ErrPayloadTooBig
	}
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(p.Peers)))
	for _, peer := range p.Peers {
		buf, err = appendString(buf, peer)
		if err != nil {
			return nil, err
		}
	}
	return appendICEServers(buf, p.ICEServers)
}

func DecodeJoined(payload []byte) (Joined, error) {
	off := 0
	var out Joined
	var err error
	out.PeerID, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.ResumeToken, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Room, off, err = readRoom(payload, off)
	if err != nil {
		return out, err
	}
	if off+2 > len(payload) {
		return out, ErrInvalidFrame
	}
	peerCount := int(binary.BigEndian.Uint16(payload[off : off+2]))
	off += 2
	out.Peers = make([]string, 0, peerCount)
	for range peerCount {
		var peer string
		peer, off, err = readString(payload, off)
		if err != nil {
			return out, err
		}
		out.Peers = append(out.Peers, peer)
	}
	out.ICEServers, off, err = readICEServers(payload, off)
	return out, err
}

func EncodeRoomState(r Room) ([]byte, error) {
	return appendRoom(make([]byte, 0, 256), r)
}

func DecodeRoomState(payload []byte) (Room, error) {
	r, _, err := readRoom(payload, 0)
	return r, err
}

func EncodeSignal(p Signal) ([]byte, error) {
	buf := make([]byte, 0, 256)
	var err error
	buf, err = appendString(buf, p.To)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.From)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Nonce)
	if err != nil {
		return nil, err
	}
	buf, err = appendString(buf, p.Sig)
	if err != nil {
		return nil, err
	}
	buf = append(buf, p.Kind)
	return appendBytes(buf, p.Body)
}

func DecodeSignal(payload []byte) (Signal, error) {
	off := 0
	var out Signal
	var err error
	out.To, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.From, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Nonce, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Sig, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	if off >= len(payload) {
		return out, ErrInvalidFrame
	}
	out.Kind = payload[off]
	off++
	out.Body, off, err = readBytes(payload, off)
	return out, err
}

func EncodeMemberUpdate(p MemberUpdate) ([]byte, error) {
	buf, err := appendString(make([]byte, 0, 32), p.PeerID)
	if err != nil {
		return nil, err
	}
	return append(buf, appendMemberFlags(Member{
		Muted:    p.Muted,
		Deafened: p.Deafened,
		Speaking: p.Speaking,
	})), nil
}

func DecodeMemberUpdate(payload []byte) (MemberUpdate, error) {
	off := 0
	id, off, err := readString(payload, off)
	if err != nil {
		return MemberUpdate{}, err
	}
	if off >= len(payload) {
		return MemberUpdate{}, ErrInvalidFrame
	}
	m := memberFromFlags(id, "", payload[off])
	return MemberUpdate{
		PeerID:   m.ID,
		Muted:    m.Muted,
		Deafened: m.Deafened,
		Speaking: m.Speaking,
	}, nil
}

func EncodeAddChannel(p AddChannel) ([]byte, error) {
	buf, err := appendString(make([]byte, 0, 64), p.ID)
	if err != nil {
		return nil, err
	}
	return appendString(buf, p.Name)
}

func DecodeAddChannel(payload []byte) (AddChannel, error) {
	off := 0
	var out AddChannel
	var err error
	out.ID, off, err = readString(payload, off)
	if err != nil {
		return out, err
	}
	out.Name, off, err = readString(payload, off)
	return out, err
}

func EncodeRename(p Rename) ([]byte, error) {
	return appendString(make([]byte, 0, 32), p.Name)
}

func DecodeRename(payload []byte) (Rename, error) {
	name, _, err := readString(payload, 0)
	return Rename{Name: name}, err
}

func EncodeError(p ErrorMsg) ([]byte, error) {
	return appendString(make([]byte, 0, 64), p.Message)
}

func DecodeError(payload []byte) (ErrorMsg, error) {
	msg, _, err := readString(payload, 0)
	return ErrorMsg{Message: msg}, err
}

func EncodePeerRef(p PeerRef) ([]byte, error) {
	return appendString(make([]byte, 0, 16), p.PeerID)
}

func DecodePeerRef(payload []byte) (PeerRef, error) {
	id, _, err := readString(payload, 0)
	return PeerRef{PeerID: id}, err
}

func EncodeModerateMember(p ModerateMember) ([]byte, error) {
	buf, err := appendString(make([]byte, 0, 32), p.PeerID)
	if err != nil {
		return nil, err
	}
	var flags byte
	if p.Muted {
		flags |= flagMuted
	}
	if p.Deafened {
		flags |= flagDeafened
	}
	return append(buf, flags), nil
}

func DecodeModerateMember(payload []byte) (ModerateMember, error) {
	off := 0
	id, off, err := readString(payload, off)
	if err != nil {
		return ModerateMember{}, err
	}
	if off >= len(payload) {
		return ModerateMember{}, ErrInvalidFrame
	}
	f := payload[off]
	return ModerateMember{
		PeerID:   id,
		Muted:    f&flagMuted != 0,
		Deafened: f&flagDeafened != 0,
	}, nil
}

func EncodePing(t int64) []byte {
	return binary.BigEndian.AppendUint64(make([]byte, 0, 8), uint64(t))
}

func DecodePing(payload []byte) (int64, error) {
	if len(payload) != 8 {
		return 0, ErrInvalidFrame
	}
	return int64(binary.BigEndian.Uint64(payload)), nil
}

func ValidCandidateBody(body []byte) bool {
	return len(body) > 0 && len(body) <= maxStringSize && json.Valid(body)
}
