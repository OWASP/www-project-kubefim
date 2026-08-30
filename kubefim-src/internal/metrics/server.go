package metrics

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// Server owns the metrics listener so address errors are returned before the
// eBPF collector starts.
type Server struct {
	listener net.Listener
	server   *http.Server
}

func Listen(address string, handler http.Handler) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
		server: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}, nil
}

func (s *Server) Address() net.Addr { return s.listener.Addr() }

func (s *Server) Serve() error {
	err := s.server.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error { return s.server.Shutdown(ctx) }
