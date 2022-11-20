package peers

import "github.com/gin-gonic/gin"

func AddPeersRoutes(s *gin.Engine) {

	s.HEAD("/peers", headPeersEndpoint)
	s.GET("/peers", getPeersEndpoint)
	s.POST("/peers", addPeerEndpoint)
	s.GET("/peers/:id", getPeerEndpoint)
	s.PATCH("/peers/:id", updatePeerEndpoint)
	s.DELETE("/peers/:id", removePeerEndpoint)

}
