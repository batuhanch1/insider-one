package push

import (
	"errors"
	"insider-one/application/command/notification/push"
	query "insider-one/application/query/notification/push"
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
	SendCommand             push.SendCommand
	SendBatchCommand        push.SendBatchCommand
	CancelCommand           push.CancelCommand
	GetAllQuery             query.GetAllQuery
	GetStatusByBatchIDQuery query.GetStatusByBatchIDQuery
	GetStatusByIDQuery      query.GetPushStatusByIDQuery
}

func NewController(sendCommand push.SendCommand, sendBatchCommand push.SendBatchCommand, cancelCommand push.CancelCommand, getAllQuery query.GetAllQuery, getStatusByBatchIDQuery query.GetStatusByBatchIDQuery, getPushStatusByIDQuery query.GetPushStatusByIDQuery) Controller {
	return &controller{sendCommand, sendBatchCommand, cancelCommand, getAllQuery, getStatusByBatchIDQuery, getPushStatusByIDQuery}
}

// Cancel godoc
// @Summary      Cancel push notifications
// @Description  Cancel push notifications matching the given status
// @Tags         push
// @Produce      json
// @Param        status  query  string  true  "Status to cancel"  Enums(PENDING)
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/cancel [put]
func (c *controller) Cancel(g *gin.Context) {
	var request push.CancelPushRequest
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
// @Summary      Send batch push notifications
// @Description  Queue multiple push notifications (max 1000)
// @Tags         push
// @Accept       json
// @Produce      json
// @Param        request  body      push.SendBatchPushRequest  true  "Send Batch Push Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/batch [post]
func (c *controller) SendBatch(g *gin.Context) {
	var request push.SendBatchPushRequest
	ctx := g.Copy()

	if err := g.ShouldBindBodyWithJSON(&request); err != nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, err, ""))
		return
	}

	if request.Pushes == nil {
		g.JSON(http.StatusBadRequest, errorHandling.ValidationError(ctx, errors.New("pushes is required"), "pushes"))
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
// @Summary      Send push notification
// @Description  Queue a single push notification
// @Tags         push
// @Accept       json
// @Produce      json
// @Param        request  body      push.SendPushRequest  true  "Send Push Request"
// @Success      200
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/ [post]
func (c *controller) Send(g *gin.Context) {
	var request push.SendPushRequest
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
// @Summary      List push notifications
// @Description  List push notifications with optional date and status filters
// @Tags         push
// @Produce      json
// @Param        status       query  string  true   "Status filter"        Enums(PENDING, SENT)
// @Param        page         query  int     true   "Page number"          minimum(1)
// @Param        page_size    query  int     true   "Page size"            minimum(0)  maximum(50)
// @Param        create_date  query  string  false  "Start date (RFC3339)"
// @Param        end_date     query  string  false  "End date (RFC3339)"
// @Success      200  {object}  query.GetAllPushResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/ [get]
func (c *controller) List(g *gin.Context) {
	var request query.GetAllPushRequest
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
// @Summary      Get push notification statuses by IDs
// @Description  Get the status of multiple push notifications
// @Tags         push
// @Accept       json
// @Produce      json
// @Param        request  body      query.GetStatusByBatchIDRequest  true  "Batch Status Request"
// @Success      200  {object}  query.GetStatusByBatchIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/status/batch [post]
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
// @Summary      Get push notification status by ID
// @Description  Get the status of a single push notification
// @Tags         push
// @Produce      json
// @Param        id  query  int  true  "Push ID"
// @Success      200  {object}  query.GetPushStatusByIDResponse
// @Failure      400  {object}  error_handling.Errors
// @Security     ApiKeyAuth
// @Router       /api/v1/push/status [get]
func (c *controller) GetStatusByID(g *gin.Context) {
	var request query.GetPushStatusByIDRequest
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
