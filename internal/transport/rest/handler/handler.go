package handler

import (
	"net/http"

	"WireguardManager/internal/service"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	PeerService service.PeerService
}

type Handler struct {
	Peer *PeerHandler
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		Peer: NewPeerHandler(deps.PeerService),
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
	{
		peer := api.Group("/peer")
		{
			peer.GET("/", h.Peer.Get)
			peer.POST("/", h.Peer.Create)
			peer.PATCH("/", h.Peer.Update)
			peer.DELETE("/", h.Peer.Delete)
		}
	}
}
