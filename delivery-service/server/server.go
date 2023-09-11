package server

import (
	"context"
	"net/http"

	"github.com/andrewhowdencom/courses.pito/delivery-service/carriers"
)

type Server struct {
	srv *http.Server

	// carriers are the carriers that can provide the shipping method.
	carriers []carriers.Carrier
}

// New generates a new server, appropriately configured
func New() *Server {
	srv := &Server{}
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", srv.healthz)
	mux.HandleFunc("/delivery-options", srv.deliveryOptions)

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
