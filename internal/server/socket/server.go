package socket

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"log/slog"
	"net/http"
	"time"
)

// StartSocketServer - starts WebSocket server
func StartSocketServer(logger *slog.Logger, port string) error {
	log := logger.With("component", "socket-server")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgradeConnection(w, r, log)
	})

	return http.ListenAndServe(":"+port, nil)
}

// upgradeConnection - upgrades the connection to WebSocket
func upgradeConnection(w http.ResponseWriter, r *http.Request, log *slog.Logger) {
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "not a websocket upgrade", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("user_session")
	if err != nil {
		cookie = &http.Cookie{
			Name:    "user_session",
			Value:   generateNewSessionID(),
			Expires: time.Now().Add(24 * time.Hour),
			Path:    "/ws",
		}
		http.SetCookie(w, cookie)
		log.Info("session cookie not found, new one created", "cookie", cookie.Value)
	} else {
		log.Info("session cookie found", "cookie", cookie.Value)
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	acceptKey := generateAcceptKey(key)

	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Sec-WebSocket-Accept", acceptKey)
	w.WriteHeader(http.StatusSwitchingProtocols)

	log.Info("WebSocket connection established")
}

// generateAcceptKey - generates key for WebSocket handshake
func generateAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// generateNewSessionID - generates a new unique session identifier
func generateNewSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "error-generating-session-id"
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
