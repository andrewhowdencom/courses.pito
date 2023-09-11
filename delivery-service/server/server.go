package server

import (
	"context"
	"net/http"
)

type Server struct {
	srv *http.Server
}

var defaultHandlers = map[string]http.HandlerFunc{
	"/healthz":          healthz,
	"/delivery-options": deliveryOptions,
}

// New generates a new server, appropriately configured
func New() *Server {
	mux := http.NewServeMux()
	for k, h := range defaultHandlers {
		mux.HandleFunc(k, h)
	}

	return &Server{
		srv: &http.Server{
			Handler: mux,
		},
	}
}

func (s *Server) Listen(addr string) error {
	s.srv.Addr = addr

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown() error {
	return s.srv.Shutdown(context.Background())
}
