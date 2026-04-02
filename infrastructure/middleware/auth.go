package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		username, password, ok := ginContext.Request.BasicAuth()

		if !ok || !validateUser(username, password) {
			ginContext.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		ginContext.Next()
	}
}

func validateUser(username, password string) bool {
	return username == "admin" && password == "1234"
}
