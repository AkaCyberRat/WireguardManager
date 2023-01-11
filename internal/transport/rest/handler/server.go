package handler

import (
	"net/http"

	"WireguardManager/internal/config"
	"WireguardManager/internal/core"
	"WireguardManager/internal/service"
	"WireguardManager/internal/utility/network"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ServerHandler struct {
	Configuration config.Configuration
	serverService service.ServerService
}

func NewServerHandler(serverService service.ServerService, configuration config.Configuration) *ServerHandler {
	return &ServerHandler{serverService: serverService, Configuration: configuration}
}

// GET server
func (h *ServerHandler) Get(c *gin.Context) {

	server, err := h.serverService.Get(c.Request.Context())
	if err != nil {
		logrus.Error("Get server internal error: ", err.Error())
		newResponse(c, http.StatusInternalServerError, ErrInternalServer)

		return
	}

	responseModel := core.ResponseServer{
		HostIp:    h.Configuration.Host.Ip,
		DnsIp:     network.WgIp,
		PublicKey: server.PublicKey,
		Port:      h.Configuration.Wireguard.Port,
		Enabled:   server.Enabled,
	}

	newResponse(c, http.StatusOK, responseModel)
}

// PATCH server
func (h *ServerHandler) Update(c *gin.Context) {
	var model core.UpdateServer
	if err := c.BindJSON(&model); err != nil {
		newResponse(c, http.StatusBadRequest, err)

		return
	}

	server, err := h.serverService.Update(c.Request.Context(), &model)
	if err != nil {
		if err == core.ErrModelValidation {
			newResponse(c, http.StatusBadRequest, err)

			return
		}

		logrus.Error("Update server internal error: ", err.Error())
		newResponse(c, http.StatusInternalServerError, ErrInternalServer)

		return
	}

	responseModel := core.ResponseServer{
		HostIp:    h.Configuration.Host.Ip,
		DnsIp:     network.WgIp,
		PublicKey: server.PublicKey,
		Port:      h.Configuration.Wireguard.Port,
		Enabled:   server.Enabled,
	}

	newResponse(c, http.StatusOK, responseModel)
}
