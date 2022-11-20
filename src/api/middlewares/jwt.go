package middlewares

import "github.com/gin-gonic/gin"

func JwtMiddleware() gin.HandlerFunc {
	// Init and Prepare

	// Returns meddleware
	return func(ctx *gin.Context) {

		ctx.Next()
	}
}
