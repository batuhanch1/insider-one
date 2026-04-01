package email

import "time"

type SendEmailRequest struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" binding:"omitempty,gt_now,within_one_year"`
	To          string     `json:"to" binding:"required,email,max=255"`
	From        string     `json:"from" binding:"required,email,max=255"`
	Subject     string     `json:"subject" binding:"required,min=1,max=150"`
	Content     string     `json:"content" binding:"required,min=1,max=10000"`
	Type        string     `json:"type" binding:"required,min=1,max=150"`
	Priority    string     `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH"`
}

type SendBatchEmailRequest struct {
	Emails []SendBatchEmail `json:"emails" binding:"min=1,max=1000"`
}

type SendBatchEmail struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" binding:"omitempty,gt_now,within_one_year"`
	To          string     `json:"to" binding:"required,email,max=255"`
	From        string     `json:"from" binding:"required,email,max=255"`
	Subject     string     `json:"subject" binding:"required,min=1,max=150"`
	Content     string     `json:"content" binding:"required,min=1,max=10000"`
	Type        string     `json:"type" binding:"required,min=1,max=150"`
	Priority    string     `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH"`
}

type CancelEmailRequest struct {
	Status string `form:"status" binding:"required,oneof=PENDING"`
}
