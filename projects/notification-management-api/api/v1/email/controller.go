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
	CancelCommand           command.CancelCommand
	GetAllQuery             query.GetAllQuery
	GetStatusByBatchIDQuery query.GetStatusByBatchIDQuery
	GetStatusByIDQuery      query.GetEmailStatusByIDQuery
}

func NewController(sendCommand command.SendCommand, sendBatchCommand command.SendBatchCommand, cancelCommand command.CancelCommand, getAllQuery query.GetAllQuery, getStatusByBatchIDQuery query.GetStatusByBatchIDQuery, getEmailStatusByIDQuery query.GetEmailStatusByIDQuery) Controller {
	return &controller{sendCommand, sendBatchCommand, cancelCommand, getAllQuery, getStatusByBatchIDQuery, getEmailStatusByIDQuery}
}

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
