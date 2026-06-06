package wire

const (
	Magic byte = 0x48

	MsgCreateRoom     byte = 1
	MsgJoin           byte = 2
	MsgLeave          byte = 3
	MsgOffer          byte = 4
	MsgAnswer         byte = 5
	MsgICE            byte = 6
	MsgRoomState      byte = 7
	MsgMemberUpdate   byte = 8
	MsgAddChannel     byte = 9
	MsgError          byte = 10
	MsgJoined         byte = 11
	MsgCreated        byte = 12
	MsgPeerJoined     byte = 13
	MsgRename         byte = 14
	MsgPing           byte = 15
	MsgPong           byte = 16
	MsgKick           byte = 17
	MsgKicked         byte = 18
	MsgModerateMember byte = 19
	MsgPeerLeft       byte = 20
)

const (
	headerSize     = 6
	maxPayloadSize = 1 << 20
	maxStringSize  = 65536

	SignalSDP       byte = 1
	SignalCandidate byte = 2

	flagMuted    byte = 1
	flagDeafened byte = 2
	flagSpeaking byte = 4
)
