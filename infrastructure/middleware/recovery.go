package middleware

import (
	"fmt"
	errorHandling "insider-one/infrastructure/error-handling"
	"insider-one/infrastructure/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ctx := ginContext.Copy()
				switch r.(type) {
				case errorHandling.Errors:
					errors := r.(errorHandling.Errors)
					logging.Error(ctx, errors.Err)
					ginContext.JSON(errors.StatusCode, errors)
				default:
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					logging.Error(ctx, err)
					ginContext.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"message": err.Error(),
					})
				}
			}
		}()

		ginContext.Next()
	}
}
