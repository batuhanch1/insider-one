package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsScheduled_FutureTime_ReturnsTrue(t *testing.T) {
	v := time.Now().Unix() + 3600
	e := Email{ScheduledAt: &v}
	assert.True(t, e.IsScheduled())
}

func TestIsScheduled_PastTime_ReturnsFalse(t *testing.T) {
	v := time.Now().Unix() - 3600
	e := Email{ScheduledAt: &v}
	assert.False(t, e.IsScheduled())
}

func TestIsScheduled_ZeroValue_ReturnsFalse(t *testing.T) {
	e := Email{ScheduledAt: nil}
	assert.False(t, e.IsScheduled())
}

func TestSetStatus_Pending_WhenNotScheduled(t *testing.T) {
	e := Email{ScheduledAt: nil}
	e.SetStatus()
	assert.Equal(t, "PENDING", e.Status)
}

func TestSetStatus_Scheduled_WhenScheduledInFuture(t *testing.T) {
	v := time.Now().Unix() + 3600
	e := Email{ScheduledAt: &v}
	e.SetStatus()
	assert.Equal(t, "SCHEDULED", e.Status)
}

func TestSetStatus_OverwritesPreviousStatus(t *testing.T) {
	e := Email{Status: "DELIVERED", ScheduledAt: nil}
	e.SetStatus()
	assert.Equal(t, "PENDING", e.Status)
}
