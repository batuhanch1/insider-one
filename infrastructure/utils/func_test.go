package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceToString_StringValue_ReturnsValue(t *testing.T) {
	value, ok := InterfaceToString("hello")
	assert.True(t, ok)
	assert.Equal(t, "hello", value)
}

func TestInterfaceToString_NilValue_ReturnsFalse(t *testing.T) {
	value, ok := InterfaceToString(nil)
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestInterfaceToString_NonStringType_ReturnsFalse(t *testing.T) {
	value, ok := InterfaceToString(42)
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestInterfaceToString_EmptyString_ReturnsTrue(t *testing.T) {
	value, ok := InterfaceToString("")
	assert.True(t, ok)
	assert.Equal(t, "", value)
}

func TestGetStackTrace_ReturnsStringContainingErrorMessage(t *testing.T) {
	err := errors.New("test error message")
	result := GetStackTrace(err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "test error message")
}

func TestGetStackTrace_ContainsStackFrames(t *testing.T) {
	result := GetStackTrace(errors.New("x"))
	assert.Contains(t, result, "goroutine")
}
