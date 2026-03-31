package middleware

import (
	"insider-one/infrastructure/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func CorrelationID() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		var headerValue = ginContext.Request.Header.Get(utils.Header_CorrelationID)
		if len(headerValue) == 0 {
			var correlationId, _ = uuid.NewV4()
			ginContext.Request.Header.Add(utils.Header_CorrelationID, correlationId.String())
		}

		ginContext.Next()
	}
}

func StartTime() gin.HandlerFunc {
	return func(ginContext *gin.Context) {

		var headerValue = ginContext.Request.Header.Get(utils.Header_ExternalRequestStartTime)
		if len(headerValue) == 0 {
			var startTime = time.Now()
			ginContext.Request.Header.Add(utils.Header_ExternalRequestStartTime, startTime.Format(utils.Layout_TimeWithNano))
		}

		ginContext.Next()
	}
}
