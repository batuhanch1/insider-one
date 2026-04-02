package push

import (
	"insider-one/domain/notification/push"
	"time"
)

type GetStatusByBatchIDRequest struct {
	IDs []uint64 `json:"ids,omitempty"`
}

type GetStatusByBatchIDResponse struct {
	Pushes []GetPushStatusByIDResponse `json:"pushes,omitempty"`
}
type GetPushStatusByIDRequest struct {
	ID uint64 `form:"id,omitempty"`
}
type GetPushStatusByIDResponse struct {
	PushID uint64 `json:"push_id,omitempty"`
	Status string `json:"status,omitempty"`
}

type GetAllPushRequest struct {
	CreateDate *time.Time `form:"create_date"`
	EndDate    *time.Time `form:"end_date"`
	Status     string     `form:"status" binding:"required,oneof=PENDING DELIVERED SCHEDULED CANCELLED"`
	Page       int        `form:"page" binding:"required,min=1"`
	PageSize   int        `form:"page_size" binding:"required,min=0,max=50"`
}

type GetAllPushResponse struct {
	Pushes     push.Pushes `json:"pushes"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page,omitempty" binding:"required,min=1"`
	PageSize   int         `json:"page_size,omitempty" binding:"required,min=0,max=50"`
}
