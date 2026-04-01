package error_handling

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_ReturnsExpectedFields(t *testing.T) {
	ctx := context.Background()
	err := errors.New("something failed")
	result := Error(ctx, err)

	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.Equal(t, "Unexpected Error", result.ErrorType)
	assert.Equal(t, err, result.Err)
	assert.NotEmpty(t, result.StackTrace)
}

func TestError_WithCorrelationID_PopulatesField(t *testing.T) {
	ctx := context.WithValue(context.Background(), "CorrelationID", "abc-123")
	result := Error(ctx, errors.New("x"))
	assert.Equal(t, "abc-123", result.CorrelationID)
}

func TestError_WithoutCorrelationID_EmptyString(t *testing.T) {
	ctx := context.Background()
	result := Error(ctx, errors.New("x"))
	assert.Equal(t, "", result.CorrelationID)
}

func TestValidationError_ReturnsExpectedFields(t *testing.T) {
	ctx := context.WithValue(context.Background(), "CorrelationID", "val-id")
	err := errors.New("validation failed")
	result := ValidationError(ctx, err, "email_field")

	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.Equal(t, "Validation Error", result.ErrorType)
	assert.Equal(t, "email_field", result.Parameter)
	assert.Equal(t, "validation failed", result.Message)
	assert.Equal(t, "val-id", result.CorrelationID)
}

func TestValidationError_PanicsWithoutCorrelationIDInContext(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		ValidationError(ctx, errors.New("x"), "")
	})
}
