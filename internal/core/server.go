package core

import (
	"errors"

	validator "github.com/go-playground/validator/v10"
)

//
// Основная модель
//

// Server представляет сервер, хранящий открытый и закрытый ключи, а также статус включения.
type Server struct {
	PublicKey  string
	PrivateKey string
	Enabled    bool
}

//
// Модели операций
//

// UpdateServer — модель запроса на обновление сервера.
type UpdateServer struct {
	PrivateKey *string `validate:"omitempty,base64"`
	Enabled    *bool   `validate:"omitempty"`
}

// Validate проверяет, что хотя бы одно поле задано, и что структура валидна.
func (p *UpdateServer) Validate() bool {
	if p.PrivateKey == nil && p.Enabled == nil {
		return false
	}

	err := validator.New().Struct(p)
	return err == nil
}

// ResponseServer — DTO, возвращаемый после операций с сервером.
type ResponseServer struct {
	HostIp    string
	DnsIp     string
	PublicKey string
	Port      int
	Enabled   bool
}

var (
	// ErrServerNotFound возвращается, когда сервер не найден.
	ErrServerNotFound = errors.New("server not found")
	// ErrServerAlreadyExists возвращается, когда сервер с таким ключом уже существует.
	ErrServerAlreadyExists = errors.New("server alredy exists")
	// ErrIncorrectPrivateKey возвращается, если передан неверный закрытый ключ.
	ErrIncorrectPrivateKey = errors.New("incorrect private key")
)
