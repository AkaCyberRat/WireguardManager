package models

import (
	"WireguardManager/internal/core"
)

type Server struct {
	Id         uint `gorm:"primarykey"`
	PublicKey  string
	PrivateKey string
	Enabled    bool
}

func (s *Server) ToCore() *core.Server {
	var server core.Server

	server.PublicKey = s.PublicKey
	server.PrivateKey = s.PrivateKey
	server.Enabled = s.Enabled

	return &server
}

func (s *Server) FromCore(server *core.Server) {
	if server == nil {
		panic("nil argument exception")
	}

	s.PublicKey = server.PublicKey
	s.PrivateKey = server.PrivateKey
	s.Enabled = server.Enabled
}
