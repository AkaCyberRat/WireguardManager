package handler

import (
	"net/http"

	"WireguardManager/internal/config"
	"WireguardManager/internal/service"
	"WireguardManager/internal/transport/rest/middlewares"
	"WireguardManager/internal/utility/auth"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	PeerService   service.PeerService
	ServerService service.ServerService
	Configuration config.Configuration
}

type Handler struct {
	Peer   *PeerHandler
	Server *ServerHandler
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		Peer:   NewPeerHandler(deps.PeerService),
		Server: NewServerHandler(deps.ServerService, deps.Configuration),
	}
}

func (h *Handler) Init(ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.Default()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	h.initApi(router)

	return router
}

func (h *Handler) initApi(router *gin.Engine) {

	api := router.Group("/api")
	api.Use(middlewares.JwtAuth())
	{
		peer := api.Group("/peer", middlewares.AllowWithAny(auth.PeerManager))
		{
			peer.GET("/", h.Peer.Get)
			peer.POST("/", h.Peer.Create)
			peer.PATCH("/", h.Peer.Update)
			peer.DELETE("/", h.Peer.Delete)
		}

		server := api.Group("/server")
		{
			server.GET("/", middlewares.AllowWithAny(auth.PeerManager, auth.ServerManager), h.Server.Get)
			server.PATCH("/", middlewares.AllowWithAny(auth.ServerManager), h.Server.Update)
		}
	}

}
