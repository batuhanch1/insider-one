package push

import "time"

type SendPushRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"gt_now,within_one_year"`
	Sender      string    `json:"sender" binding:"required,min=2,max=100"`
	PhoneNumber string    `json:"phone_number" binding:"required,e164"`
	Type        string    `json:"type" binding:"required"`
	Content     string    `json:"content" binding:"required,min=1,max=1000"`
	Priority    string    `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH"`
}

type SendBatchPushRequest struct {
	Pushes []SendBatchPush `json:"pushes" binding:"min=1,max=1000"`
}

type SendBatchPush struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"gt_now,within_one_year"`
	Sender      string    `json:"sender" binding:"required,min=2,max=100"`
	PhoneNumber string    `json:"phone_number" binding:"required,e164"`
	Type        string    `json:"type" binding:"required"`
	Content     string    `json:"content" binding:"required,min=1,max=1000"`
	Priority    string    `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH"`
}

type CancelPushRequest struct {
	Status string `form:"status" binding:"required,oneof=PENDING"`
}
