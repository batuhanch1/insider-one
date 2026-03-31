package middleware

import (
	"insider-one/infrastructure/logging"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(ginContext *gin.Context) {

		logging.ExternalLogStart(ginContext.Copy(), ginContext.Request)
		// Pre-handler phase
		ginContext.Next()

		//logging.ExternalLogFinish(ginContext.Copy(), ginContext.Request, ginContext.Writer)
	}
}
