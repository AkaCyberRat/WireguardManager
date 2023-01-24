package handlers

import (
	"net/http"

	"WireguardManager/internal/core"
	"WireguardManager/internal/services"
	"WireguardManager/internal/transport/rest"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ServerHandler struct {
	serverService services.ServerService
}

func NewServerHandler(serverService services.ServerService) *ServerHandler {
	return &ServerHandler{serverService: serverService}
}

// GET server
func (h *ServerHandler) Get(c *gin.Context) {

	server, err := h.serverService.Get(c.Request.Context())
	if err != nil {
		logrus.Error("Get server internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)

		return
	}

	c.JSON(http.StatusOK, server)
}

// PATCH server
func (h *ServerHandler) Update(c *gin.Context) {
	var model core.UpdateServer
	if err := c.BindJSON(&model); err != nil {
		rest.NewErrorResponse(c, http.StatusBadRequest, rest.ErrImpossibleToBindModel)

		return
	}

	server, err := h.serverService.Update(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			rest.NewErrorResponse(c, http.StatusBadRequest, err)

			return
		}

		logrus.Error("Update server internal error: ", err.Error())
		rest.NewErrorResponse(c, http.StatusInternalServerError, rest.ErrInternalServer)

		return
	}

	c.JSON(http.StatusOK, server)
}
