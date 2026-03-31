package sms

type SmsList []Sms
type Sms struct {
	ScheduledAt    int64  `json:"scheduled_at,omitempty"`
	SentAt         int64  `json:"sent_at,omitempty"`
	DeletedAt      int64  `json:"deleted_at,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	ID             uint64 `json:"id,omitempty"`
	IdempotencyKey uint64 `json:"idempotency_key,omitempty"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Sender         string `json:"sender,omitempty"`
	Type           string `json:"type,omitempty"`
	Status         string `json:"status,omitempty"`
	Content        string `json:"content,omitempty"`
}

func (s *Sms) IsScheduled() bool {
	return s.ScheduledAt > 0
}
