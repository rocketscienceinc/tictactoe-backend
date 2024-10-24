package rest

import "net/http"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (that *Handler) PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
