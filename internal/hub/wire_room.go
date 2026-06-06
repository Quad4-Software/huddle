package hub

import (
	"huddle/internal/room"
	"huddle/internal/wire"
)

func roomMapToWire(m map[string]any) wire.Room {
	wr := wire.Room{
		ID:     stringVal(m["id"]),
		Name:   stringVal(m["name"]),
		HostID: stringVal(m["hostId"]),
	}
	if chs, ok := m["channels"].([]room.Channel); ok {
		wr.Channels = make([]wire.Channel, len(chs))
		for i, ch := range chs {
			wr.Channels[i] = wire.Channel{ID: ch.ID, Name: ch.Name}
		}
	}
	if mems, ok := m["members"].([]room.Member); ok {
		wr.Members = make([]wire.Member, len(mems))
		for i, mem := range mems {
			wr.Members[i] = wire.Member{
				ID:       mem.ID,
				Name:     mem.Name,
				Muted:    mem.Muted,
				Deafened: mem.Deafened,
				Speaking: mem.Speaking,
			}
		}
	}
	return wr
}

func joinedToWire(p JoinedPayload) wire.Joined {
	servers := make([]wire.ICEServer, len(p.ICEServers))
	for i, s := range p.ICEServers {
		servers[i] = wire.ICEServer{
			URLs:       append([]string(nil), s.URLs...),
			Username:   s.Username,
			Credential: s.Credential,
		}
	}
	return wire.Joined{
		PeerID:      p.PeerID,
		ResumeToken: p.ResumeToken,
		Room:        roomMapToWire(p.Room),
		Peers:       append([]string(nil), p.Peers...),
		ICEServers:  servers,
	}
}

func stringVal(v any) string {
	s, _ := v.(string)
	return s
}
