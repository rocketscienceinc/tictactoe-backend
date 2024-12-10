package rest

import "net/http"

type PingHandler interface {
	PingHandler(w http.ResponseWriter, _ *http.Request)
}

type pingHandler struct{}

func NewPingHandler() PingHandler {
	return &pingHandler{}
}

func (that *pingHandler) PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
