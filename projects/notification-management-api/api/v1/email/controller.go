package email

import (
	"errors"
	command "insider-one/application/command/notification/email"
	query "insider-one/application/query/notification/email"
	errorHandling "insider-one/infrastructure/error-handling"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller interface {
	Cancel(g *gin.Context)
	SendBatch(g *gin.Context)
	Send(g *gin.Context)
	List(g *gin.Context)
	GetStatusByIDs(g *gin.Context)
	GetStatusByID(g *gin.Context)
}

type controller struct {
	SendCommand             command.SendCommand
	SendBatchCommand        command.SendBatchCommand
	CancelCommand           command.EnqueueCancelCommand
	GetAllQuery             query.GetAllQuery
	GetStatusByBatchIDQuery query.GetStatusByBatchIDQuery
	GetStatusByIDQuery      query.GetEmailStatusByIDQuery
}

func NewController(sendCommand command.SendCommand, sendBatchCommand command.SendBatchCommand, cancelCommand command.EnqueueCancelCommand, getAllQuery query.GetAllQuery, getStatusByBatchIDQuery query.GetStatusByBatchIDQuery, getEmailStatusByIDQuery query.GetEmailStatusByIDQuery) Controller {
	return &controller{sendCommand, sendBatchCommand, cancelCommand, getAllQuery, getStatusByBatchIDQuery, getEmailStatusByIDQuery}
}

// Cancel godoc
// @Summary      Cancel emails
// @Description  Cancel email notifications matching the given status
// @Tags         email
// @Produce      json
// @Param        status  query  string  true  "Status to cancel"  Enums(PENDING)
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/cancel [put]
func (c *controller) Cancel(g *gin.Context) {
	var request command.CancelEmailRequest
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
// @Summary      Send batch emails
// @Description  Queue multiple email notifications (max 1000)
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request  body      command.SendBatchEmailRequest  true  "Send Batch Email Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/batch [post]
func (c *controller) SendBatch(g *gin.Context) {
	var request command.SendBatchEmailRequest
	ctx := g.Copy()

	if err := g.ShouldBindBodyWithJSON(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}
	if request.Emails == nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, errors.New("emails is required"), "emails"))
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
// @Summary      Send email
// @Description  Queue a single email notification
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request  body      command.SendEmailRequest  true  "Send Email Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/ [post]
func (c *controller) Send(g *gin.Context) {
	var request command.SendEmailRequest
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
// @Summary      List emails
// @Description  List email notifications with optional date and status filters
// @Tags         email
// @Produce      json
// @Param        status       query  string  true   "Status filter"        Enums(PENDING, DELIVERED, SCHEDULED, CANCELLED)
// @Param        page         query  int     true   "Page number"          minimum(1)
// @Param        page_size    query  int     true   "Page size"            minimum(0)  maximum(50)
// @Param        create_date  query  string  false  "Start date (RFC3339)"
// @Param        end_date     query  string  false  "End date (RFC3339)"
// @Success      200  {object}  query.GetAllEmailResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/ [get]
func (c *controller) List(g *gin.Context) {
	var request query.GetAllEmailRequest
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
// @Summary      Get email statuses by IDs
// @Description  Get the status of multiple email notifications
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request  body      query.GetStatusByBatchIDRequest  true  "Batch Status Request"
// @Success      200  {object}  query.GetStatusByBatchIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/status/batch [post]
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
// @Summary      Get email status by ID
// @Description  Get the status of a single email notification
// @Tags         email
// @Produce      json
// @Param        id  query  int  true  "Email ID"
// @Success      200  {object}  query.GetEmailStatusByIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/email/status [get]
func (c *controller) GetStatusByID(g *gin.Context) {
	var request query.GetEmailStatusByIDRequest
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
