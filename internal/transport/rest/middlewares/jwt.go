package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"WireguardManager/internal/tools/auth"
	"WireguardManager/internal/transport/rest"

	"github.com/gin-gonic/gin"
)

type JwtHandler struct {
	authTool auth.AuthTool
}

func NewJwtHandler(authTool auth.AuthTool) *JwtHandler {
	return &JwtHandler{authTool: authTool}
}

func (h *JwtHandler) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {

		if !h.authTool.IsEnabled() {
			c.Next()

			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			rest.NewErrorResponse(c, http.StatusUnauthorized, ErrNoAccessToken)

			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			rest.NewErrorResponse(c, http.StatusUnauthorized, ErrInvalidAuthHeaderFormat)

			return
		}

		claims, err := h.authTool.ValidateToken(headerParts[1])
		if err != nil {
			rest.NewErrorResponse(c, http.StatusUnauthorized, err)

			return
		}

		c.Set("Claims", *claims)

		c.Next()
	}
}

func (h *JwtHandler) AllowedRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if !h.authTool.IsEnabled() {
			c.Next()

			return
		}

		claims := c.Keys["Claims"].(auth.JwtClaims)

		if !h.authTool.ValidateRoles(roles, claims.Roles) {
			rest.NewErrorResponse(c, http.StatusForbidden, ErrNotEnoughRights)

			return
		}

		c.Next()
	}
}

var (
	ErrNoAccessToken           = errors.New("request does not contain an access token")
	ErrInvalidAuthHeaderFormat = errors.New("request contains invalid authorization header format")
	ErrNotEnoughRights         = errors.New("not enough rights to access")
)
