package middleware

import (
	"fmt"
	prometheus2 "insider-one/infrastructure/prometheus"
	"time"

	"github.com/gin-gonic/gin"
)

func PromMiddleware(prometheusWrapper prometheus2.Prometheus) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		status := fmt.Sprintf("%d", c.Writer.Status())

		prometheusWrapper.HttpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		prometheusWrapper.HttpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
