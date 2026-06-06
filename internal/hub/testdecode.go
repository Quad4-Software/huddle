package hub

import (
	"huddle/internal/room"
	"huddle/internal/wire"
)

func decodeCreatedPayload(raw []byte) (CreatedPayload, error) {
	p, err := wire.DecodeCreated(raw)
	if err != nil {
		return CreatedPayload{}, err
	}
	return CreatedPayload{
		RoomID:    p.RoomID,
		Invite:    p.Invite,
		RoomKey:   p.RoomKey,
		ExpiresAt: p.ExpiresAt,
	}, nil
}

func decodeJoinedPayload(raw []byte) (JoinedPayload, error) {
	p, err := wire.DecodeJoined(raw)
	if err != nil {
		return JoinedPayload{}, err
	}
	return JoinedPayload{
		PeerID:      p.PeerID,
		ResumeToken: p.ResumeToken,
		Room:        wireRoomToMap(p.Room),
		Peers:       p.Peers,
		ICEServers:  iceServersFromWire(p.ICEServers),
	}, nil
}

func decodeErrorPayload(raw []byte) (ErrorPayload, error) {
	p, err := wire.DecodeError(raw)
	return ErrorPayload{Message: p.Message}, err
}

func decodeSignalPayload(raw []byte) (SignalPayload, error) {
	p, err := wire.DecodeSignal(raw)
	if err != nil {
		return SignalPayload{}, err
	}
	return signalFromWire(p), nil
}

func decodeMemberUpdatePayload(raw []byte) (MemberUpdatePayload, error) {
	p, err := wire.DecodeMemberUpdate(raw)
	if err != nil {
		return MemberUpdatePayload{}, err
	}
	return MemberUpdatePayload{
		PeerID:   p.PeerID,
		Muted:    p.Muted,
		Deafened: p.Deafened,
		Speaking: p.Speaking,
	}, nil
}

func decodePingPayload(raw []byte) (PingPayload, error) {
	t, err := wire.DecodePing(raw)
	return PingPayload{T: t}, err
}

func decodeRoomStatePayload(raw []byte) (map[string]any, error) {
	r, err := wire.DecodeRoomState(raw)
	if err != nil {
		return nil, err
	}
	return wireRoomToMap(r), nil
}

func wireRoomToMap(r wire.Room) map[string]any {
	channels := make([]room.Channel, len(r.Channels))
	for i, ch := range r.Channels {
		channels[i] = room.Channel{ID: ch.ID, Name: ch.Name}
	}
	members := make([]room.Member, len(r.Members))
	for i, m := range r.Members {
		members[i] = room.Member{
			ID:       m.ID,
			Name:     m.Name,
			Muted:    m.Muted,
			Deafened: m.Deafened,
			Speaking: m.Speaking,
		}
	}
	return map[string]any{
		"id":       r.ID,
		"name":     r.Name,
		"hostId":   r.HostID,
		"channels": channels,
		"members":  members,
	}
}

func iceServersFromWire(servers []wire.ICEServer) []ICEServer {
	out := make([]ICEServer, len(servers))
	for i, s := range servers {
		out[i] = ICEServer{
			URLs:       append([]string(nil), s.URLs...),
			Username:   s.Username,
			Credential: s.Credential,
		}
	}
	return out
}
