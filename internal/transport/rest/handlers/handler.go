package handlers

import (
	"fmt"
	"net/http"
	"time"

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
	AuthTool      auth.AuthTool
	Config        config.Configuration
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
		ServerHandler: NewServerHandler(deps.ServerService),
		JwtHandler:    middlewares.NewJwtHandler(deps.AuthTool),
		GinMode:       deps.Config.RestApi.GinMode,
	}

	return h.init()
}

func (h *Handler) init() *gin.Engine {
	gin.SetMode(h.GinMode)

	router := gin.New()
	router.Use(
		CustomLogger(),
		gin.Recovery(),
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

		start := time.Now()
		requestId := uuid.New().String()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logrus.Debugf("REST Req | %v | %v | %v | %v |", clientIP, requestId, method, c.Request.RequestURI)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}

		msg := fmt.Sprintf("REST Res | %v | %v | %v | %v | %v | %v | %v ", clientIP, requestId, method, c.Request.RequestURI, latency, statusCode, errorMessage)

		if c.Writer.Status() >= 500 {
			logrus.Error(msg)

			return
		}

		logrus.Debug(msg)
	}
}
