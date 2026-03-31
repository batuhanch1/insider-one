package sms

import "time"

type SendSmsRequest struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" binding:"omitempty,gt_now,within_one_year"`
	PhoneNumber string     `json:"phone_number"  binding:"required,e164"`
	Sender      string     `json:"sender" binding:"required,min=2,max=11"`
	Type        string     `json:"type" binding:"required"`
	Content     string     `json:"content" binding:"required,min=1,max=160"`
	Priority    string     `json:"priority" binding:"required,oneof=low medium high"`
}
type SendBatchSmsRequest struct {
	Sms []SendBatchSms `json:"sms" binding:"min=1,max=1000"`
}

type SendBatchSms struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" binding:"omitempty,gt_now,within_one_year"`
	PhoneNumber string     `json:"phone_number"  binding:"required,e164"`
	Sender      string     `json:"sender" binding:"required,min=2,max=11"`
	Type        string     `json:"type" binding:"required"`
	Content     string     `json:"content" binding:"required,min=1,max=160"`
	Priority    string     `json:"priority" binding:"required,oneof=low medium high"`
}

type CancelSmsRequest struct {
	Status string `form:"status" binding:"required,oneof=PENDING SENT"`
}
