package push

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsScheduled_FutureTime_ReturnsTrue(t *testing.T) {
	v := time.Now().Unix() + 3600
	p := Push{ScheduledAt: &v}
	assert.True(t, p.IsScheduled())
}

func TestIsScheduled_PastTime_ReturnsFalse(t *testing.T) {
	v := time.Now().Unix() - 3600
	p := Push{ScheduledAt: &v}
	assert.False(t, p.IsScheduled())
}

func TestIsScheduled_ZeroValue_ReturnsFalse(t *testing.T) {
	p := Push{ScheduledAt: nil}
	assert.False(t, p.IsScheduled())
}

func TestSetStatus_Pending_WhenNotScheduled(t *testing.T) {
	p := Push{ScheduledAt: nil}
	p.SetStatus()
	assert.Equal(t, "PENDING", p.Status)
}

func TestSetStatus_Scheduled_WhenScheduledInFuture(t *testing.T) {
	v := time.Now().Unix() + 3600
	p := Push{ScheduledAt: &v}
	p.SetStatus()
	assert.Equal(t, "SCHEDULED", p.Status)
}

func TestSetStatus_OverwritesPreviousStatus(t *testing.T) {
	p := Push{Status: "DELIVERED", ScheduledAt: nil}
	p.SetStatus()
	assert.Equal(t, "PENDING", p.Status)
}
