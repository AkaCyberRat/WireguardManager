package core

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

//
// Main model
//

type Server struct {
	PublicKey  string
	PrivateKey string
	Enabled    bool
}

//
// Operation models
//

type UpdateServer struct {
	PrivateKey *string `validate:"omitempty,base64"`
	Enabled    *bool   `validate:"omitempty"`
}

func (p *UpdateServer) Validate() bool {
	if p.PrivateKey == nil && p.Enabled == nil {
		return false
	}

	err := validator.New().Struct(p)
	return err == nil
}

type ResponseServer struct {
	HostIp    string
	DnsIp     string
	PublicKey string
	Port      int
	Enabled   bool
}

var (
	ErrServerNotFound      = errors.New("server not found")
	ErrServerAlreadyExists = errors.New("server alredy exists")
	ErrIncorrectPrivateKey = errors.New("incorrect private key")
)
