package email

type CreateEmailEvent struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	To             string `json:"to,omitempty"`
	From           string `json:"from,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Content        string `json:"content,omitempty"`
	Type           string `json:"type,omitempty"`
	Priority       string `json:"priority,omitempty"`
}

type CancelEmailEvent struct {
	ID uint64 `json:"id,omitempty"`
}

type EmailCreatedEvent struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	SentAt         int64  `json:"sent_at,omitempty"`
	DeletedAt      int64  `json:"deleted_at,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	To             string `json:"to,omitempty"`
	From           string `json:"from,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Content        string `json:"content,omitempty"`
	Status         string `json:"status,omitempty"`
	Type           string `json:"type,omitempty"`
}
