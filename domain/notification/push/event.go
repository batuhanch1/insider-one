package push

type CreatePushEvent struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	Sender         string `json:"sender,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Type           string `json:"type,omitempty"`
	Content        string `json:"content,omitempty"`
	Priority       string `json:"priority,omitempty"`
}

type CancelPushEvent struct {
	ID uint64 `json:"id,omitempty"`
}

type PushCreatedEvent struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	Sender         string `json:"sender,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Type           string `json:"type,omitempty"`
	Status         string `json:"status,omitempty"`
	Content        string `json:"content,omitempty"`
}
