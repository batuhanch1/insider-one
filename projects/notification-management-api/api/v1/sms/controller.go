package sms

import (
	"errors"
	"insider-one/application/command/notification/sms"
	query "insider-one/application/query/notification/sms"
	errorHandling "insider-one/infrastructure/error-handling"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller interface {
	Cancel(ctx *gin.Context)
	SendBatch(ctx *gin.Context)
	Send(ctx *gin.Context)
	List(g *gin.Context)
	GetStatusByIDs(g *gin.Context)
	GetStatusByID(g *gin.Context)
}

type controller struct {
	SendCommand             sms.SendCommand
	SendBatchCommand        sms.SendBatchCommand
	CancelCommand           sms.EnqueueCancelCommand
	GetAllQuery             query.GetAllQuery
	GetStatusByBatchIDQuery query.GetStatusByBatchIDQuery
	GetStatusByIDQuery      query.GetSmsStatusByIDQuery
}

func NewController(sendCommand sms.SendCommand, sendBatchCommand sms.SendBatchCommand, cancelCommand sms.EnqueueCancelCommand, getAllQuery query.GetAllQuery, getStatusByBatchIDQuery query.GetStatusByBatchIDQuery, getSmsStatusByIDQuery query.GetSmsStatusByIDQuery) Controller {
	return &controller{sendCommand, sendBatchCommand, cancelCommand, getAllQuery, getStatusByBatchIDQuery, getSmsStatusByIDQuery}
}

// Cancel godoc
// @Summary      Cancel SMS
// @Description  Cancel SMS notifications matching the given status
// @Tags         sms
// @Produce      json
// @Param        status  query  string  true  "Status to cancel"  Enums(PENDING)
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/cancel [put]
func (c *controller) Cancel(g *gin.Context) {
	var request sms.CancelSmsRequest
	ctx := g.Copy()

	if err := g.ShouldBindQuery(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	err := c.CancelCommand.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, nil)
}

// SendBatch godoc
// @Summary      Send batch SMS
// @Description  Queue multiple SMS notifications (max 1000)
// @Tags         sms
// @Accept       json
// @Produce      json
// @Param        request  body      sms.SendBatchSmsRequest  true  "Send Batch SMS Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/batch [post]
func (c *controller) SendBatch(g *gin.Context) {
	var request sms.SendBatchSmsRequest
	ctx := g.Copy()

	if err := g.ShouldBindBodyWithJSON(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}

	if request.Sms == nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, errors.New("sms is required"), "sms"))
		return
	}

	err := c.SendBatchCommand.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, nil)
}

// Send godoc
// @Summary      Send SMS
// @Description  Queue a single SMS notification
// @Tags         sms
// @Accept       json
// @Produce      json
// @Param        request  body      sms.SendSmsRequest  true  "Send SMS Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/ [post]
func (c *controller) Send(g *gin.Context) {
	var request sms.SendSmsRequest
	ctx := g.Copy()

	if err := g.ShouldBindBodyWithJSON(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	err := c.SendCommand.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, nil)
}

// List godoc
// @Summary      List SMS
// @Description  List SMS notifications with optional date and status filters
// @Tags         sms
// @Produce      json
// @Param        status       query  string  true   "Status filter"        Enums(PENDING, DELIVERED, SCHEDULED, CANCELLED)
// @Param        page         query  int     true   "Page number"          minimum(1)
// @Param        page_size    query  int     true   "Page size"            minimum(0)  maximum(50)
// @Param        create_date  query  string  false  "Start date (RFC3339)"
// @Param        end_date     query  string  false  "End date (RFC3339)"
// @Success      200  {object}  query.GetAllSmsResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/ [get]
func (c *controller) List(g *gin.Context) {
	var request query.GetAllSmsRequest
	ctx := g.Copy()

	if err := g.ShouldBindQuery(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	response, err := c.GetAllQuery.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, response)
}

// GetStatusByIDs godoc
// @Summary      Get SMS statuses by IDs
// @Description  Get the status of multiple SMS notifications
// @Tags         sms
// @Accept       json
// @Produce      json
// @Param        request  body      query.GetStatusByBatchIDRequest  true  "Batch Status Request"
// @Success      200  {object}  query.GetStatusByBatchIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/status/batch [post]
func (c *controller) GetStatusByIDs(g *gin.Context) {
	var request query.GetStatusByBatchIDRequest
	ctx := g.Copy()

	if err := g.ShouldBindBodyWithJSON(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	response, err := c.GetStatusByBatchIDQuery.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, response)
}

// GetStatusByID godoc
// @Summary      Get SMS status by ID
// @Description  Get the status of a single SMS notification
// @Tags         sms
// @Produce      json
// @Param        id  query  int  true  "SMS ID"
// @Success      200  {object}  query.GetSmsStatusByIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/sms/status [get]
func (c *controller) GetStatusByID(g *gin.Context) {
	var request query.GetSmsStatusByIDRequest
	ctx := g.Copy()

	if err := g.ShouldBindQuery(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	response, err := c.GetStatusByIDQuery.Execute(ctx, request)
	if err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		return
	}

	g.JSON(http.StatusOK, response)
}
