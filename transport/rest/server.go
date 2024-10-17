package rest

import (
	"fmt"
	"net/http"
	"time"
)

func Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)

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
