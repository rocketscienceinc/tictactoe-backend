package socket

import (
	"bufio"
	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"log/slog"
	"net"
)

func StartSocketServer(logger *slog.Logger, cfg *config.Config) error {
	log := logger.With("component", "socket-server")
	listener, err := net.Listen("tcp", ":"+cfg.SocketPort)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("failed to accept connection")
			continue
		}
		go handleSocketConnection(logger, conn)
	}
}

func handleSocketConnection(logger *slog.Logger, conn net.Conn) {
	log := logger.With("remote", conn.RemoteAddr())

	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Error("failed to read message")
			return
		}

		log.Info("received message")
		conn.Write([]byte("Received: " + msg))
	}
}
