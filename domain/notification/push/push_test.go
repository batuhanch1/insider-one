package push

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsScheduled_FutureTime_ReturnsTrue(t *testing.T) {
	p := Push{ScheduledAt: time.Now().Unix() + 3600}
	assert.True(t, p.IsScheduled())
}

func TestIsScheduled_PastTime_ReturnsFalse(t *testing.T) {
	p := Push{ScheduledAt: time.Now().Unix() - 3600}
	assert.False(t, p.IsScheduled())
}

func TestIsScheduled_ZeroValue_ReturnsFalse(t *testing.T) {
	p := Push{ScheduledAt: 0}
	assert.False(t, p.IsScheduled())
}

func TestSetStatus_Pending_WhenNotScheduled(t *testing.T) {
	p := Push{ScheduledAt: 0}
	p.SetStatus()
	assert.Equal(t, "PENDING", p.Status)
}

func TestSetStatus_Scheduled_WhenScheduledInFuture(t *testing.T) {
	p := Push{ScheduledAt: time.Now().Unix() + 3600}
	p.SetStatus()
	assert.Equal(t, "SCHEDULED", p.Status)
}

func TestSetStatus_OverwritesPreviousStatus(t *testing.T) {
	p := Push{Status: "DELIVERED", ScheduledAt: 0}
	p.SetStatus()
	assert.Equal(t, "PENDING", p.Status)
}
