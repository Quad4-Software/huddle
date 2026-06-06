package hub

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"huddle/internal/pow"
	"huddle/internal/ratelimit"
	"huddle/internal/room"
)

func testLimits() Limits {
	return Limits{
		Create: ratelimit.New(1000, time.Hour),
		Join:   ratelimit.New(1000, time.Hour),
	}
}

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type wsClient struct {
	conn *websocket.Conn
	t    *testing.T
}

func startTestHub(t *testing.T, maxSize int) (string, *room.Manager) {
	t.Helper()
	return startTestHubWithLimits(t, maxSize, testLimits())
}

func startTestHubWithLimits(t *testing.T, maxSize int, limits Limits) (string, *room.Manager) {
	t.Helper()

	rm := room.NewManager("hub-test-secret", time.Hour, maxSize)
	h := New(rm, limits, pow.NewStore(), 0)
	go h.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		h.ServeClient(conn, host)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return strings.Replace(srv.URL, "http://", "ws://", 1) + "/ws", rm
}

func dialClient(t *testing.T, url string) *wsClient {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &wsClient{conn: conn, t: t}
}

func (c *wsClient) send(t MessageType, payload any) {
	data, err := Marshal(t, payload)
	if err != nil {
		c.t.Fatalf("marshal failed: %v", err)
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		c.t.Fatalf("write failed: %v", err)
	}
}

func (c *wsClient) read() Message {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		c.t.Fatalf("read failed: %v", err)
	}
	msg, err := Unmarshal(data)
	if err != nil {
		c.t.Fatalf("unmarshal failed: %v", err)
	}
	return msg
}

func (c *wsClient) readCreated() CreatedPayload {
	msg := c.read()
	if msg.Type != TypeCreated {
		c.t.Fatalf("expected created, got %s", msg.Type)
	}
	payload, err := decodeCreatedPayload(msg.Payload)
	if err != nil {
		c.t.Fatalf("decode created: %v", err)
	}
	return payload
}

func (c *wsClient) readJoined() JoinedPayload {
	msg := c.readType(TypeJoined)
	payload, err := decodeJoinedPayload(msg.Payload)
	if err != nil {
		c.t.Fatalf("decode joined: %v", err)
	}
	return payload
}

func (c *wsClient) readType(want MessageType) Message {
	for {
		msg := c.read()
		if msg.Type == want {
			return msg
		}
		if msg.Type == TypeError {
			errPayload, err := decodeErrorPayload(msg.Payload)
			if err != nil {
				c.t.Fatalf("decode error: %v", err)
			}
			c.t.Fatalf("unexpected error: %s", errPayload.Message)
		}
	}
}

func TestHubCreateAndJoin(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host := dialClient(t, url)

	created := func() CreatedPayload {
		host.send(TypeCreateRoom, CreateRoomPayload{Name: "War Room"})
		return host.readCreated()
	}()
	if created.RoomID == "" || created.Invite == "" {
		t.Fatal("expected room id and invite")
	}

	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	joined := host.readJoined()
	if joined.PeerID == "" {
		t.Fatal("expected peer id")
	}
	if joined.Room["name"] != "War Room" {
		t.Fatalf("unexpected room name: %v", joined.Room["name"])
	}

	guest := dialClient(t, url)
	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	if len(guestJoined.Peers) != 1 || guestJoined.Peers[0] != joined.PeerID {
		t.Fatalf("expected existing peer list, got %v", guestJoined.Peers)
	}

	_ = host.readType(TypePeerJoined)
}

func TestHubJoinRejectsInvalidPassword(t *testing.T) {
	url, _ := startTestHub(t, 4)
	client := dialClient(t, url)

	client.send(TypeCreateRoom, CreateRoomPayload{Name: "Locked", Password: "huddle"})
	created := client.readCreated()

	client.send(TypeJoin, JoinPayload{
		RoomID:   created.RoomID,
		Invite:   created.Invite,
		Password: "wrong",
		Name:     "Intruder",
	})

	msg := client.read()
	if msg.Type != TypeError {
		t.Fatalf("expected error, got %s", msg.Type)
	}
	payload, err := decodeErrorPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Message != "unable to join room" {
		t.Fatalf("unexpected error: %s", payload.Message)
	}
}

func TestHubRelaysSignalBetweenPeers(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host := dialClient(t, url)
	guest := dialClient(t, url)

	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Signal"})
	created := host.readCreated()

	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	hostJoined := host.readJoined()

	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	_ = host.readType(TypePeerJoined)

	host.send(TypeOffer, SignalPayload{
		To:    guestJoined.PeerID,
		SDP:   "v=0",
		Nonce: "nonce",
		Sig:   "sig",
	})

	msg := guest.readType(TypeOffer)
	relayed, err := decodeSignalPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if relayed.From != hostJoined.PeerID || relayed.SDP != "v=0" || relayed.Nonce != "nonce" || relayed.Sig != "sig" {
		t.Fatalf("unexpected relayed offer: %+v", relayed)
	}
}

func TestHubRelaysICEWithRawCandidate(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host := dialClient(t, url)
	guest := dialClient(t, url)

	host.send(TypeCreateRoom, CreateRoomPayload{Name: "ICE"})
	created := host.readCreated()

	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	hostJoined := host.readJoined()

	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guestJoined := guest.readJoined()
	_ = host.readType(TypePeerJoined)

	candidate := []byte(`{"candidate":"candidate:1 1 udp 2122260223 192.168.0.2 54321 typ host","sdpMid":"0","sdpMLineIndex":0}`)
	host.send(TypeICE, SignalPayload{
		To:        guestJoined.PeerID,
		Nonce:     "nonce",
		Sig:       "sig",
		Candidate: candidate,
	})

	msg := guest.readType(TypeICE)
	relayed, err := decodeSignalPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if relayed.From != hostJoined.PeerID {
		t.Fatalf("unexpected sender: %s", relayed.From)
	}
	if string(relayed.Candidate) != string(candidate) {
		t.Fatalf("candidate not relayed: %s", relayed.Candidate)
	}
}

func TestHubPingPong(t *testing.T) {
	url, _ := startTestHub(t, 4)
	client := dialClient(t, url)

	client.send(TypeCreateRoom, CreateRoomPayload{Name: "Latency"})
	created := client.readCreated()
	client.send(TypeJoin, JoinPayload{RoomID: created.RoomID, Invite: created.Invite, Name: "Host"})
	client.readJoined()

	client.send(TypePing, PingPayload{T: 12345})
	msg := client.readType(TypePong)
	pong, err := decodePingPayload(msg.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if pong.T != 12345 {
		t.Fatalf("expected echoed timestamp 12345, got %d", pong.T)
	}
}

func TestHubRenameBroadcasts(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host := dialClient(t, url)
	guest := dialClient(t, url)

	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Names"})
	created := host.readCreated()
	host.send(TypeJoin, JoinPayload{RoomID: created.RoomID, Invite: created.Invite, Name: "Host"})
	host.readJoined()

	guest.send(TypeJoin, JoinPayload{RoomID: created.RoomID, Invite: created.Invite, Name: "Guest"})
	guest.readJoined()
	host.readType(TypePeerJoined)

	host.send(TypeRename, RenamePayload{Name: "Hostess"})

	for range 5 {
		msg := guest.readType(TypeRoomState)
		state, err := decodeRoomStatePayload(msg.Payload)
		if err != nil {
			t.Fatal(err)
		}
		members, _ := state["members"].([]room.Member)
		for _, m := range members {
			if m.Name == "Hostess" {
				return
			}
		}
	}
	t.Fatal("expected renamed member in room state")
}

func TestHubBroadcastsChannelUpdates(t *testing.T) {
	url, _ := startTestHub(t, 4)
	host := dialClient(t, url)
	guest := dialClient(t, url)

	host.send(TypeCreateRoom, CreateRoomPayload{Name: "Channels"})
	created := host.readCreated()

	host.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Host",
	})
	host.readJoined()

	guest.send(TypeJoin, JoinPayload{
		RoomID: created.RoomID,
		Invite: created.Invite,
		Name:   "Guest",
	})
	guest.readJoined()
	_ = host.readType(TypePeerJoined)

	host.send(TypeAddChannel, AddChannelPayload{ID: "ops", Name: "ops"})

	_ = guest.readType(TypeRoomState)
}
