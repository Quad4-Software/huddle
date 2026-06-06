package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 65536
)

// Client is a single WebSocket peer attached to the hub.
type Client struct {
	ID            string
	RoomID        string
	Name          string
	IP            string
	CreatedRoomID string
	Hub           *Hub
	Conn          *websocket.Conn
	Send          chan []byte
	removed       sync.Once
	sendMu        sync.Mutex
	closed        bool
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMsgSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws read: %v", err)
			}
			break
		}
		c.Hub.handleMessage(c, data)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) enqueue(data []byte) bool {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.closed {
		return false
	}
	select {
	case c.Send <- data:
		return true
	default:
		return false
	}
}

func (c *Client) shutdown() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	close(c.Send)
}

func (c *Client) SendJSON(t MessageType, payload any) {
	data, err := Marshal(t, payload)
	if err != nil {
		return
	}
	if !c.enqueue(data) {
		log.Printf("drop message to %s", c.ID)
	}
}

func (c *Client) SendError(msg string) {
	c.SendJSON(TypeError, ErrorPayload{Message: msg})
}

func decodePayload[T any](raw json.RawMessage, out *T) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}
