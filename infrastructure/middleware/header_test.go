package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"insider-one/infrastructure/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCorrelationID_WhenMissing_GeneratesValue(t *testing.T) {
	router := gin.New()
	router.Use(CorrelationID())

	var capturedValue string
	router.GET("/test", func(c *gin.Context) {
		capturedValue = c.Request.Header.Get(utils.Header_CorrelationID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedValue)
}

func TestCorrelationID_WhenPresent_PreservesExistingValue(t *testing.T) {
	router := gin.New()
	router.Use(CorrelationID())

	var capturedValue string
	router.GET("/test", func(c *gin.Context) {
		capturedValue = c.Request.Header.Get(utils.Header_CorrelationID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(utils.Header_CorrelationID, "existing-id-123")
	router.ServeHTTP(w, req)

	assert.Equal(t, "existing-id-123", capturedValue)
}

func TestStartTime_WhenMissing_AddsHeader(t *testing.T) {
	router := gin.New()
	router.Use(StartTime())

	var capturedValue string
	router.GET("/test", func(c *gin.Context) {
		capturedValue = c.Request.Header.Get(utils.Header_ExternalRequestStartTime)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedValue)
}

func TestStartTime_WhenPresent_PreservesExistingValue(t *testing.T) {
	router := gin.New()
	router.Use(StartTime())

	var capturedValue string
	router.GET("/test", func(c *gin.Context) {
		capturedValue = c.Request.Header.Get(utils.Header_ExternalRequestStartTime)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(utils.Header_ExternalRequestStartTime, "2026-01-01T00:00:00Z")
	router.ServeHTTP(w, req)

	assert.Equal(t, "2026-01-01T00:00:00Z", capturedValue)
}
