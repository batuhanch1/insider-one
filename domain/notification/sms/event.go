package sms

type CreateSmsEvent struct {
	ScheduledAt    *int64 `json:"scheduled_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Sender         string `json:"sender,omitempty"`
	Type           string `json:"type,omitempty"`
	Content        string `json:"content,omitempty"`
	Priority       string `json:"priority,omitempty"`
}

type CancelSmsEvent struct {
	ID uint64 `json:"id,omitempty"`
}

type SmsCreatedEvent struct {
	ScheduledAt    *int64 `json:"scheduled_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Sender         string `json:"sender,omitempty"`
	Type           string `json:"type,omitempty"`
	Status         string `json:"status,omitempty"`
	Content        string `json:"content,omitempty"`
}
