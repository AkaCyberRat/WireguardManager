package middlewares

import (
	"WireguardManager/internal/tools/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JwtAuth() func(c *gin.Context) {
	return func(c *gin.Context) {

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

		claims, err := auth.ValidateToken(headerParts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})

			return
		}

		c.Set("Claims", *claims)

		c.Next()
	}
}

func AllowWithAny(allowedScopes ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		claims := c.Keys["Claims"].(auth.JwtClaims)

		if !anyMatched(allowedScopes, claims.Scopes) {
			c.AbortWithStatusJSON(http.StatusForbidden, nil)
			return
		}

		c.Next()
	}
}

func anyMatched(allowedScopes, receivedScopes []string) bool {
	if len(allowedScopes) > len(receivedScopes) {
		return false
	}
	for _, e := range allowedScopes {
		if !contains(receivedScopes, e) {
			return false
		}
	}
	return true
}

func contains(list []string, element string) bool {
	for _, subElement := range list {
		if subElement == element {
			return true
		}
	}
	return false
}
