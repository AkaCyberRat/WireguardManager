package core

import (
	"errors"

	validator "github.com/go-playground/validator/v10"
)

//
// Основная модель
//

// Peer представляет сущность пира VPN с деталями подключения и статусом.
type Peer struct {
	Id            string
	Ip            string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	Status        Status
}

// Status представляет текущее состояние пира.
type Status string

const (
	Unused   Status = "unused"
	Enabled  Status = "enabled"
	Disabled Status = "disabled"
)

//
// Модели операций
//

// GetPeer — модель запроса на получение одного пира.
type GetPeer struct {
	Id string `validate:"required"`
	// PublicKey string `validate:"required,base64"`
}

// Validate проверяет, что в структуре GetPeer заданы все обязательные поля.
func (p *GetPeer) Validate() bool {
	err := validator.New().Struct(p)
	return err == nil
}

// CreatePeer — модель запроса на создание нового пира.
type CreatePeer struct {
	PublicKey     string `validate:"required,base64"`
	PresharedKey  string `validate:"required,base64"`
	DownloadSpeed int    `validate:"required,numeric,min=1"`
	UploadSpeed   int    `validate:"required,numeric,min=1"`
	Enabled       *bool  `validate:"required"`
}

// Validate проверяет, что в структуре CreatePeer заданы все обязательные поля и она правильно отформатирована.
func (p *CreatePeer) Validate() bool {
	err := validator.New().Struct(p)
	return err == nil
}

// UpdatePeer — модель запроса на обновление существующего пира.
type UpdatePeer struct {
	Id            string  `validate:"required"`
	PublicKey     *string `validate:"omitempty,base64"`
	PresharedKey  *string `validate:"omitempty,base64"`
	DownloadSpeed *int    `validate:"omitempty,numeric,min=1"`
	UploadSpeed   *int    `validate:"omitempty,numeric,min=1"`
	Enabled       *bool   `validate:"omitempty"`
}

// Validate проверяет, что хотя бы одно обновляемое поле задано, и что структура в остальном валидна.
func (p *UpdatePeer) Validate() bool {
	if p.PublicKey == nil && p.PresharedKey == nil && p.DownloadSpeed == nil && p.UploadSpeed == nil && p.Enabled == nil {
		return false
	}

	err := validator.New().Struct(p)
	return err == nil
}

// DeletePeer — модель запроса на удаление пира.
type DeletePeer struct {
	Id string `validate:"required"`
	// PublicKey string
}

// Validate проверяет, что в структуре DeletePeer заданы все обязательные поля.
func (p *DeletePeer) Validate() bool {
	err := validator.New().Struct(p)

	return err == nil
}

// ResponsePeer — DTO, возвращаемый после операций с пиром.
type ResponsePeer struct {
	Id            string
	PublicKey     string
	PresharedKey  string
	DownloadSpeed int
	UploadSpeed   int
	Enabled       bool
}

// BindFrom мапит модель Peer в DTO ResponsePeer.
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
// Ошибки
//

// ErrPeerNotFound возвращается, когда запрошенный пир не найден.
var ErrPeerNotFound = errors.New("peer not found")

// ErrPeerLimitReached возвращается, когда достигнут максимальный лимит пиров.
var ErrPeerLimitReached = errors.New("peer limit reached")

// ErrModelValidation возвращается, когда модель не проходит валидацию.
var ErrModelValidation = errors.New("model validation fail")
