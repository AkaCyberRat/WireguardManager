package models

import "gorm.io/gorm"

type Server struct {
	gorm.Model

	IpAddress  string
	PrivateKey string
	PublicKey  string
	PeerLimit  int64
}
