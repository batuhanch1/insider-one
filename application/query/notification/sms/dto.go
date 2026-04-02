package sms

import (
	"insider-one/domain/notification/sms"
	"time"
)

type GetStatusByBatchIDRequest struct {
	IDs []uint64 `json:"ids,omitempty"`
}

type GetStatusByBatchIDResponse struct {
	SmsList []GetSmsStatusByIDResponse `json:"smsList,omitempty"`
}

type GetSmsStatusByIDRequest struct {
	ID uint64 `form:"id,omitempty"`
}

type GetSmsStatusByIDResponse struct {
	SmsID  uint64 `json:"sms_id,omitempty"`
	Status string `json:"status,omitempty"`
}

type GetAllSmsRequest struct {
	CreateDate *time.Time `form:"create_date"`
	EndDate    *time.Time `form:"end_date"`
	Status     string     `form:"status" binding:"required,oneof=PENDING DELIVERED SCHEDULED CANCELLED"`
	Page       int        `form:"page" binding:"required,min=1"`
	PageSize   int        `form:"page_size" binding:"required,min=0,max=50"`
}

type GetAllSmsResponse struct {
	SmsList    sms.SmsList `json:"smsList"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page,omitempty" binding:"required,min=1"`
	PageSize   int         `json:"page_size,omitempty" binding:"required,min=0,max=50"`
}
