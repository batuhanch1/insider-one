package logging

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type BodyWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (w *BodyWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}
