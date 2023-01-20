package rest

import (
	"context"
	"crypto/tls"
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

type SslDeps struct {
	CrtPath *string
	KeyPath *string
}

func (s *Server) LoadSSL(deps SslDeps) error {

	if deps.CrtPath != nil && deps.KeyPath != nil {
		cert, err := tls.LoadX509KeyPair(*deps.CrtPath, *deps.KeyPath)
		if err != nil {
			return err
		}

		s.httpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	return nil
}

func (s *Server) ListenAndServe() error {
	if s.httpServer.TLSConfig != nil {
		return s.httpServer.ListenAndServeTLS("", "")
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
