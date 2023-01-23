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
	//PublicKey string `validate:"requred,base64"`
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

type ResponsePeer struct {
	Id            string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	Enabled       bool
}

func (p *ResponsePeer) BindFrom(model *Peer) {
	p.Id = model.Id
	p.PublicKey = model.PublicKey
	p.PresharedKey = model.PresharedKey
	p.DownloadSpeed = model.DownloadSpeed
	p.UploadSpeed = model.UploadSpeed

	if model.Status == Enabled {
		p.Enabled = true
	} else {
		p.Enabled = false
	}
}

//
// Errors
//

var (
	ErrPeerNotFound     = errors.New("peer not found")
	ErrPeerLimitReached = errors.New("peer limit reached")
	ErrModelValidation  = errors.New("model validation fail")
)
