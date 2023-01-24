package rest

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(c *gin.Context, status int, err error) {
	c.AbortWithStatusJSON(status, ErrorResponse{err.Error()})
}

var (
	ErrImpossibleToBindModel = errors.New("impossible to bind request model")
	ErrInternalServer        = errors.New("internal server error")
)
