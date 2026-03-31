package email

type Emails []Email
type Email struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	SentAt         int64  `json:"sent_at,omitempty"`
	DeletedAt      int64  `json:"deleted_at,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	To             string `json:"to,omitempty"`
	From           string `json:"from,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Content        string `json:"content,omitempty"`
	Status         string `json:"status,omitempty"`
	Type           string `json:"type,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
}

func (e *Email) IsScheduled() bool {
	return e.ScheduledAt > 0
}
