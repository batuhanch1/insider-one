package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchPublisher_Publish_OverLimit_ReturnsError(t *testing.T) {
	bp := NewBatchPublisher(nil)
	items := make([]any, 1001)

	err := bp.Publish(context.Background(), items, BatchPublisherOptions{})

	assert.Error(t, err)
	assert.Equal(t, "maks len 1000", err.Error())
}
