package server

import (
	"net/http"

	"github.com/rocketscienceinc/tittactoe-backend/internal/config"
	"github.com/rocketscienceinc/tittactoe-backend/pkg/handlers"
)

func StartHTTPServer(cfg *config.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", handlers.PingHandler)
	return http.ListenAndServe(":"+cfg.HTTPPort, mux)
}
