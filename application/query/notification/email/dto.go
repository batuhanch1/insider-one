package email

import (
	"insider-one/domain/notification/email"
	"time"
)

type GetStatusByBatchIDRequest struct {
	IDs []uint64 `json:"ids,omitempty"`
}

type GetStatusByBatchIDResponse struct {
	Emails []GetEmailStatusByIDResponse `json:"emails,omitempty"`
}
type GetEmailStatusByIDRequest struct {
	ID uint64 `form:"id,omitempty"`
}
type GetEmailStatusByIDResponse struct {
	EmailID uint64 `json:"email_id,omitempty"`
	Status  string `json:"status,omitempty"`
}

type GetAllEmailRequest struct {
	CreateDate *time.Time `form:"create_date"`
	EndDate    *time.Time `form:"end_date"`
	Status     string     `form:"status" binding:"required,oneof=PENDING SENT"`
	Page       int        `form:"page" binding:"required,min=1"`
	PageSize   int        `form:"page_size" binding:"required,min=0,max=50"`
}

type GetAllEmailResponse struct {
	Emails     email.Emails `json:"emails,omitempty"`
	TotalCount int          `json:"total_count,omitempty"`
	Page       int          `json:"page,omitempty" binding:"required,min=1"`
	PageSize   int          `json:"page_size,omitempty" binding:"required,min=0,max=50"`
}
