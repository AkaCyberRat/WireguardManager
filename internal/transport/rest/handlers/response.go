package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrInternalServer = errors.New("internal server error")
)

func newErrorResponse(c *gin.Context, status int, err error) {
	c.JSON(status, ErrorResponse{err.Error()})
}

type ErrorResponse struct {
	Error string `json:"error"`
}
