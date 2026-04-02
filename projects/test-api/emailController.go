package test_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"insider-one/application/command/notification/email"
	"insider-one/domain/notification"
	"insider-one/infrastructure/adapters/client"
	errorHandling "insider-one/infrastructure/error-handling"
	"insider-one/infrastructure/logging"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmailController interface {
	Cancel(g *gin.Context)
	SendBatch(g *gin.Context)
	Send(g *gin.Context)
}

type emailController struct {
	client client.HttpClient
}

func NewEmailController(client client.HttpClient) EmailController {
	return &emailController{client}
}

func (c *emailController) Cancel(g *gin.Context) {
	var request = email.CancelEmailRequest{
		Status: notification.Notification_Status_Pending,
	}
	ctx := g.Copy()
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("email test Cancel Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPut, "localhost:8080/api/v1/email/cancel", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("email test Cancel NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin123")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("email test Cancel Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("email bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	return
}

func (c *emailController) SendBatch(g *gin.Context) {
	for j := 0; j < 1000; j++ {
		var request email.SendBatchEmailRequest
		for i := 0; i < 1000; i++ {
			e := email.SendBatchEmail{
				To:      fmt.Sprintf("to_%d_%d@test.com", i, j),
				From:    fmt.Sprintf("from_%d_%d@test.com", i, j),
				Subject: fmt.Sprintf("Subject Batch %d_%d", i, j),
				Content: fmt.Sprintf("Content Batch %d_%d", i, j),
				Type:    "test_email_type",
			}
			n := rand.Intn(100)
			if n > 50 {
				nextTime := time.Now().Add(time.Hour * 1)
				e.ScheduledAt = &nextTime
			}
			p := rand.Intn(3)
			switch p {
			case 1:
				e.Priority = notification.Notification_Priority_High
			case 2:
				e.Priority = notification.Notification_Priority_Low
			case 3:
				e.Priority = notification.Notification_Priority_Medium
			}
			request.Emails = append(request.Emails, e)
		}
		ctx := g.Copy()
		bodyBytes, err := json.Marshal(request)
		if err != nil {
			err = fmt.Errorf("email test SendBatch Marshal error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req, err := http.NewRequest(http.MethodPost, "localhost:8080/api/v1/email/batch", bytes.NewBuffer(bodyBytes))
		if err != nil {
			err = fmt.Errorf("email test SendBatch NewRequest error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "admin123")

		var resp *http.Response
		if resp, err = c.client.Do(ctx, req); err != nil {
			err = fmt.Errorf("email test SendBatch Do error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			err = fmt.Errorf("email bad response: %d", resp.StatusCode)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		g.JSON(http.StatusOK, errorHandling.Error(ctx, err))
	}
}

func (c *emailController) Send(g *gin.Context) {
	id, _ := uuid.NewV7()
	var request = email.SendEmailRequest{
		To:      fmt.Sprintf("to_single_%s@test.com", id.String()),
		From:    fmt.Sprintf("from_single_%s@test.com", id.String()),
		Subject: fmt.Sprintf("Subject Single %s", id.String()),
		Content: fmt.Sprintf("Content Single %s", id.String()),
		Type:    "test_email_type",
	}

	n := rand.Intn(100)
	if n > 50 {
		nextTime := time.Now().Add(time.Hour * 1)
		request.ScheduledAt = &nextTime
	}
	p := rand.Intn(3)
	switch p {
	case 1:
		request.Priority = notification.Notification_Priority_High
	case 2:
		request.Priority = notification.Notification_Priority_Low
	case 3:
		request.Priority = notification.Notification_Priority_Medium
	}

	ctx := g.Copy()
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("email test Send Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPost, "localhost:8080/api/v1/email", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("email test Send NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin123")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("email test Send Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("email bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, errorHandling.Error(ctx, err))
}
