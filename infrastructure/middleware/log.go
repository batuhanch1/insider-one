package middleware

import (
	"bytes"
	"insider-one/infrastructure/logging"
	"strings"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		if strings.Contains(ginContext.Request.RequestURI, "swagger") {
			return
		}
		bw := &logging.BodyWriter{
			ResponseWriter: ginContext.Writer,
			Body:           bytes.NewBufferString(""),
		}

		ginContext.Writer = bw

		logging.ExternalLogStart(ginContext.Copy(), ginContext.Request)
		ginContext.Next()
		logging.ExternalLogFinish(ginContext.Copy(), ginContext.Request, bw)
	}
}
