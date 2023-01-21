package middlewares

import (
	"WireguardManager/internal/tools/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type JwtHandler struct {
	authTool auth.AuthTool
}

func NewJwtHandler(authTool auth.AuthTool) *JwtHandler {
	return &JwtHandler{authTool: authTool}
}

func (h *JwtHandler) Authenticate() func(c *gin.Context) {
	return func(c *gin.Context) {

		if !h.authTool.IsEnabled() {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "request does not contain an access token"})

			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "request contains invalid authorization header format"})

			return
		}

		claims, err := h.authTool.ValidateToken(headerParts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})

			return
		}

		c.Set("Claims", *claims)

		c.Next()
	}
}

func (h *JwtHandler) AllowedRoles(roles ...string) func(c *gin.Context) {
	return func(c *gin.Context) {

		if !h.authTool.IsEnabled() {
			c.Next()
			return
		}

		claims := c.Keys["Claims"].(auth.JwtClaims)

		if !h.authTool.ValidateRoles(roles, claims.Roles) {
			c.AbortWithStatusJSON(http.StatusForbidden, nil)

			return
		}

		c.Next()
	}
}
