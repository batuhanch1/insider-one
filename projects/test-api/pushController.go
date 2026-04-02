package test_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-one/application/command/notification/push"
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

type PushController interface {
	Cancel(g *gin.Context)
	SendBatch(g *gin.Context)
	Send(g *gin.Context)
}

type pushController struct {
	client client.HttpClient
}

func NewPushController(client client.HttpClient) PushController {
	return &pushController{client}
}

func (c *pushController) Cancel(g *gin.Context) {
	var request = push.CancelPushRequest{
		Status: notification.Notification_Status_Pending,
	}
	ctx := g.Copy()
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		err = fmt.Errorf("push test Cancel Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost:8080/api/v1/push/cancel", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("push test Cancel NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "1234")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("push test Cancel Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("push bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, nil)
}

func generatePhone() string {
	prefixes := []string{
		"90505", "90506", "90532", "90533", "90542", "90543", "90555",
	}

	prefix := prefixes[rand.Intn(len(prefixes))]

	number := ""
	for i := 0; i < 7; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}

	return "+" + prefix + number
}

func (c *pushController) SendBatch(g *gin.Context) {
	ctx, cancel := context.WithTimeout(g.Copy(), time.Minute*5)
	defer cancel()
	for j := 0; j < 1000; j++ {
		var request push.SendBatchPushRequest
		for i := 0; i < 1000; i++ {
			p := push.SendBatchPush{
				Sender:      fmt.Sprintf("TESTSENDER_%d", i),
				PhoneNumber: generatePhone(),
				Type:        "test_push_type",
				Content:     fmt.Sprintf("Content Batch %d_%d", i, j),
			}

			n := rand.Intn(100)
			if n > 50 {
				nextTime := time.Now().Add(time.Hour * 1)
				p.ScheduledAt = nextTime
			}
			prio := rand.Intn(3)
			switch prio {
			case 1:
				p.Priority = notification.Notification_Priority_High
			case 2:
				p.Priority = notification.Notification_Priority_Low
			case 3:
				p.Priority = notification.Notification_Priority_Medium
			default:
				p.Priority = notification.Notification_Priority_Medium
			}
			request.Pushes = append(request.Pushes, p)
		}
		bodyBytes, err := json.Marshal(request)
		if err != nil {
			err = fmt.Errorf("push test SendBatch Marshal error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/push/batch", bytes.NewBuffer(bodyBytes))
		if err != nil {
			err = fmt.Errorf("push test SendBatch NewRequest error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "1234")

		var resp *http.Response
		if resp, err = c.client.Do(ctx, req); err != nil {
			err = fmt.Errorf("push test SendBatch Do error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("push bad response: %d", resp.StatusCode)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
	}
	g.JSON(http.StatusOK, nil)
}

func (c *pushController) Send(g *gin.Context) {
	id, _ := uuid.NewV7()
	var request = push.SendPushRequest{
		Sender:      fmt.Sprintf("TESTSENDER_%s", id.String()),
		PhoneNumber: generatePhone(),
		Type:        "test_push_type",
		Content:     fmt.Sprintf("Content Batch %s", id.String()),
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
		err = fmt.Errorf("push test Send Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v1/push", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("push test Send NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "1234")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("push test Send Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("push bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, nil)
}
