package core

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

//
// Main model
//

type Peer struct {
	Id            string
	Ip            string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	Status        Status
}

type Status string

const (
	Unused   Status = "unused"
	Enabled  Status = "enabled"
	Disabled Status = "disabled"
)

//
// Operation models
//

type GetPeer struct {
	Id string `validate:"required"`
	//PublicKey string `validate:"required,base64"`
}

func (p *GetPeer) Validate() bool {
	err := validator.New().Struct(p)
	return err == nil
}

type CreatePeer struct {
	PublicKey     string `validate:"required,base64"`
	PresharedKey  string `validate:"required,base64"`
	DownloadSpeed int    `validate:"required,numeric,min=1"`
	UploadSpeed   int    `validate:"required,numeric,min=1"`
	Enabled       *bool  `validate:"required"`
}

func (p *CreatePeer) Validate() bool {
	err := validator.New().Struct(p)
	return err == nil
}

type UpdatePeer struct {
	Id            string  `validate:"required"`
	PublicKey     *string `validate:"omitempty,base64"`
	PresharedKey  *string `validate:"omitempty,base64"`
	DownloadSpeed *int    `validate:"omitempty,numeric,min=1"`
	UploadSpeed   *int    `validate:"omitempty,numeric,min=1"`
	Enabled       *bool   `validate:"omitempty"`
}

func (p *UpdatePeer) Validate() bool {
	if p.PublicKey == nil && p.PresharedKey == nil && p.DownloadSpeed == nil && p.UploadSpeed == nil && p.Enabled == nil {
		return false
	}

	err := validator.New().Struct(p)
	return err == nil
}

type DeletePeer struct {
	Id string `validate:"required"`
	//PublicKey string
}

func (p *DeletePeer) Validate() bool {
	err := validator.New().Struct(p)

	return err == nil
}

//
// Errors
//

var (
	ErrPeerNotFound     = errors.New("peer not found")
	ErrPeerLimitReached = errors.New("peer limit reached")
	ErrModelValidation  = errors.New("model validation fail")
)
