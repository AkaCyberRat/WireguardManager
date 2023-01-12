package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	ReadTimeout  = time.Second * 10
	WriteTimeout = time.Second * 10
	MaxHeadersMb = 1
)

type Server struct {
	httpServer *http.Server
}

func NewServer(port int, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        handler,
			ReadTimeout:    ReadTimeout,
			WriteTimeout:   WriteTimeout,
			MaxHeaderBytes: MaxHeadersMb << 20,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
