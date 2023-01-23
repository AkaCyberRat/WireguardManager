package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrInternalServer = errors.New("internal server error")
)

func newResponse(c *gin.Context, status int, obj interface{}) {

	if err, ok := obj.(error); ok {
		c.JSON(status, ErrorResponse{err.Error()})
	} else {
		c.JSON(status, obj)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
