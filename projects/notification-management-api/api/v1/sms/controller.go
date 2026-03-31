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
	CancelCommand           sms.CancelCommand
	GetAllQuery             query.GetAllQuery
	GetStatusByBatchIDQuery query.GetStatusByBatchIDQuery
	GetStatusByIDQuery      query.GetSmsStatusByIDQuery
}

func NewController(sendCommand sms.SendCommand, sendBatchCommand sms.SendBatchCommand, cancelCommand sms.CancelCommand, getAllQuery query.GetAllQuery, getStatusByBatchIDQuery query.GetStatusByBatchIDQuery, getSmsStatusByIDQuery query.GetSmsStatusByIDQuery) Controller {
	return &controller{sendCommand, sendBatchCommand, cancelCommand, getAllQuery, getStatusByBatchIDQuery, getSmsStatusByIDQuery}
}

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
