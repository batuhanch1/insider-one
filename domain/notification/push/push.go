package push

import (
	"insider-one/domain/notification"
	"time"
)

type Pushes []Push
type Push struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	SentAt         int64  `json:"sent_at,omitempty"`
	DeletedAt      int64  `json:"deleted_at,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	Sender         string `json:"sender,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Type           string `json:"type,omitempty"`
	Status         string `json:"status,omitempty"`
	Content        string `json:"content,omitempty"`
	Priority       string `json:"priority,omitempty"`
}

func (p *Push) IsScheduled() bool {
	return p.ScheduledAt > time.Now().Unix()
}

func (p *Push) SetStatus() {
	p.Status = notification.Notification_Status_Pending
	if p.IsScheduled() {
		p.Status = notification.Notification_Status_Scheduled
	}
}
