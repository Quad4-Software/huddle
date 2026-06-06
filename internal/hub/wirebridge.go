package hub

import (
	"huddle/internal/wire"
)

func wireType(t MessageType) (byte, bool) {
	switch t {
	case TypeCreateRoom:
		return wire.MsgCreateRoom, true
	case TypeJoin:
		return wire.MsgJoin, true
	case TypeLeave:
		return wire.MsgLeave, true
	case TypeOffer:
		return wire.MsgOffer, true
	case TypeAnswer:
		return wire.MsgAnswer, true
	case TypeICE:
		return wire.MsgICE, true
	case TypeRoomState:
		return wire.MsgRoomState, true
	case TypeMemberUpdate:
		return wire.MsgMemberUpdate, true
	case TypeAddChannel:
		return wire.MsgAddChannel, true
	case TypeError:
		return wire.MsgError, true
	case TypeJoined:
		return wire.MsgJoined, true
	case TypeCreated:
		return wire.MsgCreated, true
	case TypePeerJoined:
		return wire.MsgPeerJoined, true
	case TypeRename:
		return wire.MsgRename, true
	case TypePing:
		return wire.MsgPing, true
	case TypePong:
		return wire.MsgPong, true
	case TypeKick:
		return wire.MsgKick, true
	case TypeKicked:
		return wire.MsgKicked, true
	case TypeModerateMember:
		return wire.MsgModerateMember, true
	case TypePeerLeft:
		return wire.MsgPeerLeft, true
	default:
		return 0, false
	}
}

func typeFromWire(b byte) (MessageType, bool) {
	switch b {
	case wire.MsgCreateRoom:
		return TypeCreateRoom, true
	case wire.MsgJoin:
		return TypeJoin, true
	case wire.MsgLeave:
		return TypeLeave, true
	case wire.MsgOffer:
		return TypeOffer, true
	case wire.MsgAnswer:
		return TypeAnswer, true
	case wire.MsgICE:
		return TypeICE, true
	case wire.MsgRoomState:
		return TypeRoomState, true
	case wire.MsgMemberUpdate:
		return TypeMemberUpdate, true
	case wire.MsgAddChannel:
		return TypeAddChannel, true
	case wire.MsgError:
		return TypeError, true
	case wire.MsgJoined:
		return TypeJoined, true
	case wire.MsgCreated:
		return TypeCreated, true
	case wire.MsgPeerJoined:
		return TypePeerJoined, true
	case wire.MsgRename:
		return TypeRename, true
	case wire.MsgPing:
		return TypePing, true
	case wire.MsgPong:
		return TypePong, true
	case wire.MsgKick:
		return TypeKick, true
	case wire.MsgKicked:
		return TypeKicked, true
	case wire.MsgModerateMember:
		return TypeModerateMember, true
	case wire.MsgPeerLeft:
		return TypePeerLeft, true
	default:
		return "", false
	}
}

func encodePayload(t MessageType, payload any) ([]byte, error) {
	switch t {
	case TypeCreateRoom:
		p := payload.(CreateRoomPayload)
		return wire.EncodeCreateRoom(wire.CreateRoom{
			Name:     p.Name,
			Password: p.Password,
			Pow:      powToWire(p.Pow),
		})
	case TypeJoin:
		p := payload.(JoinPayload)
		return wire.EncodeJoin(wire.Join{
			RoomID:       p.RoomID,
			Invite:       p.Invite,
			Password:     p.Password,
			Name:         p.Name,
			ResumePeerID: p.ResumePeerID,
			ResumeToken:  p.ResumeToken,
			Pow:          powToWire(p.Pow),
		})
	case TypeCreated:
		p := payload.(CreatedPayload)
		return wire.EncodeCreated(wire.Created{
			RoomID:    p.RoomID,
			Invite:    p.Invite,
			RoomKey:   p.RoomKey,
			ExpiresAt: p.ExpiresAt,
		})
	case TypeJoined:
		p := payload.(JoinedPayload)
		return wire.EncodeJoined(joinedToWire(p))
	case TypeRoomState:
		return wire.EncodeRoomState(roomMapToWire(payload.(map[string]any)))
	case TypeOffer, TypeAnswer, TypeICE:
		return signalToWire(payload.(SignalPayload), signalKind(t))
	case TypeMemberUpdate:
		p := payload.(MemberUpdatePayload)
		return wire.EncodeMemberUpdate(wire.MemberUpdate{
			PeerID:   p.PeerID,
			Muted:    p.Muted,
			Deafened: p.Deafened,
			Speaking: p.Speaking,
		})
	case TypeAddChannel:
		p := payload.(AddChannelPayload)
		return wire.EncodeAddChannel(wire.AddChannel{ID: p.ID, Name: p.Name})
	case TypeRename:
		return wire.EncodeRename(wire.Rename{Name: payload.(RenamePayload).Name})
	case TypePing:
		return wire.EncodePing(payload.(PingPayload).T), nil
	case TypePong:
		if raw, ok := payload.([]byte); ok {
			return raw, nil
		}
		return wire.EncodePing(payload.(PingPayload).T), nil
	case TypeError:
		return wire.EncodeError(wire.ErrorMsg{Message: payload.(ErrorPayload).Message})
	case TypeKick, TypePeerJoined, TypePeerLeft:
		return wire.EncodePeerRef(wire.PeerRef{PeerID: peerIDFrom(payload)})
	case TypeKicked:
		return nil, nil
	case TypeModerateMember:
		p := payload.(ModerateMemberPayload)
		return wire.EncodeModerateMember(wire.ModerateMember{
			PeerID:   p.PeerID,
			Muted:    p.Muted,
			Deafened: p.Deafened,
		})
	case TypeLeave:
		return nil, nil
	default:
		return nil, wire.ErrInvalidFrame
	}
}

func decodePayloadTyped(t MessageType, raw []byte, out any) error {
	switch t {
	case TypeCreateRoom:
		p, err := wire.DecodeCreateRoom(raw)
		if err != nil {
			return err
		}
		*out.(*CreateRoomPayload) = CreateRoomPayload{
			Name:     p.Name,
			Password: p.Password,
			Pow:      powFromWire(p.Pow),
		}
	case TypeJoin:
		p, err := wire.DecodeJoin(raw)
		if err != nil {
			return err
		}
		*out.(*JoinPayload) = JoinPayload{
			RoomID:       p.RoomID,
			Invite:       p.Invite,
			Password:     p.Password,
			Name:         p.Name,
			ResumePeerID: p.ResumePeerID,
			ResumeToken:  p.ResumeToken,
			Pow:          powFromWire(p.Pow),
		}
	case TypeMemberUpdate:
		p, err := wire.DecodeMemberUpdate(raw)
		if err != nil {
			return err
		}
		*out.(*MemberUpdatePayload) = MemberUpdatePayload{
			PeerID:   p.PeerID,
			Muted:    p.Muted,
			Deafened: p.Deafened,
			Speaking: p.Speaking,
		}
	case TypeAddChannel:
		p, err := wire.DecodeAddChannel(raw)
		if err != nil {
			return err
		}
		*out.(*AddChannelPayload) = AddChannelPayload{ID: p.ID, Name: p.Name}
	case TypeRename:
		p, err := wire.DecodeRename(raw)
		if err != nil {
			return err
		}
		*out.(*RenamePayload) = RenamePayload{Name: p.Name}
	case TypeKick:
		p, err := wire.DecodePeerRef(raw)
		if err != nil {
			return err
		}
		*out.(*KickPayload) = KickPayload{PeerID: p.PeerID}
	case TypeModerateMember:
		p, err := wire.DecodeModerateMember(raw)
		if err != nil {
			return err
		}
		*out.(*ModerateMemberPayload) = ModerateMemberPayload{
			PeerID:   p.PeerID,
			Muted:    p.Muted,
			Deafened: p.Deafened,
		}
	case TypeOffer, TypeAnswer, TypeICE:
		p, err := wire.DecodeSignal(raw)
		if err != nil {
			return err
		}
		*out.(*SignalPayload) = signalFromWire(p)
	default:
		return wire.ErrInvalidFrame
	}
	return nil
}

func powToWire(p *PowPayload) *wire.Pow {
	if p == nil {
		return nil
	}
	return &wire.Pow{ID: p.ID, Nonce: p.Nonce}
}

func powFromWire(p *wire.Pow) *PowPayload {
	if p == nil {
		return nil
	}
	return &PowPayload{ID: p.ID, Nonce: p.Nonce}
}

func signalKind(t MessageType) byte {
	switch t {
	case TypeICE:
		return wire.SignalCandidate
	default:
		return wire.SignalSDP
	}
}

func signalToWire(p SignalPayload, kind byte) ([]byte, error) {
	body := []byte(p.SDP)
	if kind == wire.SignalCandidate {
		body = append([]byte(nil), p.Candidate...)
	}
	return wire.EncodeSignal(wire.Signal{
		To:    p.To,
		From:  p.From,
		Nonce: p.Nonce,
		Sig:   p.Sig,
		Kind:  kind,
		Body:  body,
	})
}

func signalFromWire(p wire.Signal) SignalPayload {
	out := SignalPayload{
		To:    p.To,
		From:  p.From,
		Nonce: p.Nonce,
		Sig:   p.Sig,
	}
	if p.Kind == wire.SignalCandidate {
		out.Candidate = append([]byte(nil), p.Body...)
	} else {
		out.SDP = string(p.Body)
	}
	return out
}

func peerIDFrom(payload any) string {
	switch p := payload.(type) {
	case PeerJoinedPayload:
		return p.PeerID
	case PeerLeftPayload:
		return p.PeerID
	case KickPayload:
		return p.PeerID
	default:
		return ""
	}
}
