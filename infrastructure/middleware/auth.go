package middleware

import "github.com/gin-gonic/gin"

func AuthRequired() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		// ... check token, session, etc.
		ginContext.Next()
	}
}
