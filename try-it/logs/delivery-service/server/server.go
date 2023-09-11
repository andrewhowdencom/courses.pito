package server

import "fmt"

type Server struct{}

// New generates a new server, appropriately configured
func New() *Server {
	return &Server{}
}

func (s *Server) Listen(addr string) error {
	return fmt.Errorf("not implemented")
}

func (s *Server) Shutdown() error {
	return fmt.Errorf("not implemetned")
}
