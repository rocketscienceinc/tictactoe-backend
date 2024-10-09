package socket

import (
	"crypto/rand"
	"crypto/sha1" //nolint: gosec // idk how to fix that
	"encoding/base64"
	"math/big"
)

// Static GUID defined in RFC 6455 for WebSocket.
const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// GenerateAcceptKey - generates key for WebSocket handshake.
func GenerateAcceptKey(key string) string {
	h := sha1.New() //nolint: gosec // RFC 6455 requires the use of SHA-1 for WebSocket

	h.Write([]byte(key + websocketGUID))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GenerateNewSessionID - generates a new unique sessionID.
func GenerateNewSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "error-generating-session-id"
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateGameID - generates a unique identifier for the room.
func GenerateGameID() string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err != nil {
		return ""
	}
	return n.String()
}
