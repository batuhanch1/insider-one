package sms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsScheduled_FutureTime_ReturnsTrue(t *testing.T) {
	s := Sms{ScheduledAt: time.Now().Unix() + 3600}
	assert.True(t, s.IsScheduled())
}

func TestIsScheduled_PastTime_ReturnsFalse(t *testing.T) {
	s := Sms{ScheduledAt: time.Now().Unix() - 3600}
	assert.False(t, s.IsScheduled())
}

func TestIsScheduled_ZeroValue_ReturnsFalse(t *testing.T) {
	s := Sms{ScheduledAt: 0}
	assert.False(t, s.IsScheduled())
}

func TestSetStatus_Pending_WhenNotScheduled(t *testing.T) {
	s := Sms{ScheduledAt: 0}
	s.SetStatus()
	assert.Equal(t, "PENDING", s.Status)
}

func TestSetStatus_Scheduled_WhenScheduledInFuture(t *testing.T) {
	s := Sms{ScheduledAt: time.Now().Unix() + 3600}
	s.SetStatus()
	assert.Equal(t, "SCHEDULED", s.Status)
}

func TestSetStatus_OverwritesPreviousStatus(t *testing.T) {
	s := Sms{Status: "DELIVERED", ScheduledAt: 0}
	s.SetStatus()
	assert.Equal(t, "PENDING", s.Status)
}
