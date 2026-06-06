// Package turnserver runs Huddle's optional embedded TURN relay.
package turnserver

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pion/turn/v4"
)

// Config controls the embedded TURN relay.
type Config struct {
	ListenAddr string
	PublicAddr string
	Realm      string
	Secret     string
	CredTTL    time.Duration
}

// ICEServer is the browser-facing WebRTC ICE server shape.
type ICEServer struct {
	URLs       []string
	Username   string
	Credential string
}

// Server owns the embedded TURN listener.
type Server struct {
	server *turn.Server
	conn   net.PacketConn
}

// Start starts a UDP TURN listener.
func Start(cfg Config) (*Server, error) {
	if cfg.ListenAddr == "" || cfg.PublicAddr == "" || cfg.Realm == "" || cfg.Secret == "" {
		return nil, errors.New("turn config is incomplete")
	}
	host, _, err := net.SplitHostPort(cfg.PublicAddr)
	if err != nil {
		return nil, err
	}
	publicIP := net.ParseIP(host)
	if publicIP == nil {
		return nil, errors.New("turn public address host must be an IP")
	}

	conn, err := net.ListenPacket("udp4", cfg.ListenAddr)
	if err != nil {
		return nil, err
	}
	srv, err := turn.NewServer(turn.ServerConfig{
		Realm: cfg.Realm,
		AuthHandler: func(username, realm string, _ net.Addr) ([]byte, bool) {
			if !validUsername(username) {
				return nil, false
			}
			credential := Credential(cfg.Secret, username)
			return turn.GenerateAuthKey(username, realm, credential), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: conn,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: publicIP,
					Address:      "0.0.0.0",
				},
			},
		},
	})
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Server{server: srv, conn: conn}, nil
}

// Close shuts down the TURN listener.
func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	err := s.server.Close()
	if closeErr := s.conn.Close(); err == nil {
		err = closeErr
	}
	return err
}

// NewICEServer creates short-lived TURN credentials for a client.
func NewICEServer(cfg Config) (ICEServer, error) {
	username, err := Username(cfg.CredTTL)
	if err != nil {
		return ICEServer{}, err
	}
	return ICEServer{
		URLs:       []string{fmt.Sprintf("turn:%s?transport=udp", cfg.PublicAddr)},
		Username:   username,
		Credential: Credential(cfg.Secret, username),
	}, nil
}

// Username returns an expiring TURN username.
func Username(ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = time.Hour
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d:%s", time.Now().Add(ttl).Unix(), hex.EncodeToString(b)), nil
}

// Credential returns the TURN password for username.
func Credential(secret, username string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(username))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validUsername(username string) bool {
	exp, _, ok := strings.Cut(username, ":")
	if !ok {
		return false
	}
	unix, err := strconv.ParseInt(exp, 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Unix() <= unix
}
