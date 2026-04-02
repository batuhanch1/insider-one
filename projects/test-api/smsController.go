package test_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-one/application/command/notification/sms"
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

type SmsController interface {
	Cancel(g *gin.Context)
	SendBatch(g *gin.Context)
	Send(g *gin.Context)
}

type smsController struct {
	client client.HttpClient
}

func NewSmsController(client client.HttpClient) SmsController {
	return &smsController{client}
}

func (c *smsController) Cancel(g *gin.Context) {
	var request = sms.CancelSmsRequest{
		Status: notification.Notification_Status_Pending,
	}
	ctx := g.Copy()
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("sms test Cancel Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/sms/cancel", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("sms test Cancel NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "1234")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("sms test Cancel Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("sms bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, nil)
}

func (c *smsController) SendBatch(g *gin.Context) {
	ctx, cancel := context.WithTimeout(g.Copy(), time.Minute*5)
	defer cancel()
	for j := 0; j < 1000; j++ {
		var request sms.SendBatchSmsRequest
		for i := 0; i < 1000; i++ {
			s := sms.SendBatchSms{
				PhoneNumber: generatePhone(),
				Sender:      fmt.Sprintf("SENDER_%d", i),
				Type:        "test_sms_type",
				Content:     fmt.Sprintf("Content Batch %d_%d", i, j),
			}
			n := rand.Intn(100)
			if n > 50 {
				nextTime := time.Now().Add(time.Hour * 1)
				s.ScheduledAt = nextTime
			}
			p := rand.Intn(3)
			switch p {
			case 1:
				s.Priority = notification.Notification_Priority_High
			case 2:
				s.Priority = notification.Notification_Priority_Low
			case 3:
				s.Priority = notification.Notification_Priority_Medium
			default:
				s.Priority = notification.Notification_Priority_Medium
			}
			request.Sms = append(request.Sms, s)
		}
		bodyBytes, err := json.Marshal(request)
		if err != nil {
			err = fmt.Errorf("sms test SendBatch Marshal error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/sms/batch", bytes.NewBuffer(bodyBytes))
		if err != nil {
			err = fmt.Errorf("sms test SendBatch NewRequest error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "1234")

		var resp *http.Response
		if resp, err = c.client.Do(ctx, req); err != nil {
			err = fmt.Errorf("sms test SendBatch Do error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("sms bad response: %d", resp.StatusCode)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
	}
	g.JSON(http.StatusOK, nil)
}

func (c *smsController) Send(g *gin.Context) {
	id, _ := uuid.NewV7()
	var request = sms.SendSmsRequest{
		Sender:      "SENDER",
		PhoneNumber: generatePhone(),
		Type:        "test_push_type",
		Content:     fmt.Sprintf("Content Batch %d", id.String()),
	}

	n := rand.Intn(100)
	if n > 50 {
		nextTime := time.Now().Add(time.Hour * 1)
		request.ScheduledAt = nextTime
	}
	p := rand.Intn(3)
	switch p {
	case 1:
		request.Priority = notification.Notification_Priority_High
	case 2:
		request.Priority = notification.Notification_Priority_Low
	case 3:
		request.Priority = notification.Notification_Priority_Medium
	default:
		request.Priority = notification.Notification_Priority_High
	}

	ctx := g.Copy()
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("sms test Send Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/sms", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("sms test Send NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "1234")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("sms test Send Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("sms  bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, nil)
}
