package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsScheduled_FutureTime_ReturnsTrue(t *testing.T) {
	e := Email{ScheduledAt: time.Now().Unix() + 3600}
	assert.True(t, e.IsScheduled())
}

func TestIsScheduled_PastTime_ReturnsFalse(t *testing.T) {
	e := Email{ScheduledAt: time.Now().Unix() - 3600}
	assert.False(t, e.IsScheduled())
}

func TestIsScheduled_ZeroValue_ReturnsFalse(t *testing.T) {
	e := Email{ScheduledAt: 0}
	assert.False(t, e.IsScheduled())
}

func TestSetStatus_Pending_WhenNotScheduled(t *testing.T) {
	e := Email{ScheduledAt: 0}
	e.SetStatus()
	assert.Equal(t, "PENDING", e.Status)
}

func TestSetStatus_Scheduled_WhenScheduledInFuture(t *testing.T) {
	e := Email{ScheduledAt: time.Now().Unix() + 3600}
	e.SetStatus()
	assert.Equal(t, "SCHEDULED", e.Status)
}

func TestSetStatus_OverwritesPreviousStatus(t *testing.T) {
	e := Email{Status: "DELIVERED", ScheduledAt: 0}
	e.SetStatus()
	assert.Equal(t, "PENDING", e.Status)
}
