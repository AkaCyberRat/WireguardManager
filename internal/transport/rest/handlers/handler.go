package handlers

import (
	"net/http"

	"WireguardManager/internal/config"
	"WireguardManager/internal/services"
	"WireguardManager/internal/tools/auth"
	"WireguardManager/internal/transport/rest/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Deps struct {
	PeerService   services.PeerService
	ServerService services.ServerService
	Configuration config.Configuration
	AuthTool      auth.AuthTool
}

type Handler struct {
	PeerHandler   *PeerHandler
	ServerHandler *ServerHandler
	JwtHandler    *middlewares.JwtHandler
	GinMode       string
}

func NewHandler(deps Deps) *gin.Engine {
	h := &Handler{
		PeerHandler:   NewPeerHandler(deps.PeerService),
		ServerHandler: NewServerHandler(deps.ServerService, deps.Configuration),
		JwtHandler:    middlewares.NewJwtHandler(deps.AuthTool),
		GinMode:       deps.Configuration.RestApi.GinMode,
	}

	return h.init()
}

func (h *Handler) init() *gin.Engine {
	gin.SetMode(h.GinMode)

	router := gin.New()
	router.Use(
		gin.Recovery(),
		CustomLogger(),
		//gin.Logger(),
	)

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	h.initApi(router)

	return router
}

func (h *Handler) initApi(router *gin.Engine) {

	api := router.Group("/api", h.JwtHandler.Authenticate())
	{
		peer := api.Group("/peer", h.JwtHandler.AllowedRoles(auth.PeerManagerRole))
		{
			peer.GET("/", h.PeerHandler.Get)
			peer.POST("/", h.PeerHandler.Create)
			peer.PATCH("/", h.PeerHandler.Update)
			peer.DELETE("/", h.PeerHandler.Delete)
		}

		server := api.Group("/server", h.JwtHandler.AllowedRoles(auth.PeerManagerRole, auth.ServerManagerRole))
		{
			server.GET("/", h.ServerHandler.Get)
			server.PATCH("/", h.JwtHandler.AllowedRoles(auth.ServerManagerRole), h.ServerHandler.Update)
		}
	}
}

func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().String()

		logrus.Debugf("REST Req | %v | %v | %v | %v |", c.Request.RemoteAddr, requestId, c.Request.Method, c.Request.RequestURI)

		c.Next()

		logrus.Debugf("REST Res | %v | %v | %v | %v | %v |", c.Request.RemoteAddr, requestId, c.Request.Method, c.Request.RequestURI, c.Writer.Status())
	}
}
