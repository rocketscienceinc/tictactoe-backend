package rest

import (
	"fmt"
	"net/http"
	"time"
)

func Start(port string) error {
	mux := http.NewServeMux()

	// All endpoints should have a prefix /api, because our nginx will proxy all requests to /api to this service.
	mux.HandleFunc("/api/ping", pingHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
