package api

import (
	"WireguardManager/src/api/middlewares"
	"WireguardManager/src/api/routes/peers"

	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func Configure() {
	server = (gin.New())

	addMiddlewares(server)

	addRouting(server)
}

func addMiddlewares(s *gin.Engine) {

	s.Use(middlewares.JwtMiddleware())
}

func addRouting(s *gin.Engine) {
	peers.AddPeersRoutes(server)
}
