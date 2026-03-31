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
