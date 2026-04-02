package test_api

import (
	"bytes"
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

	req, err := http.NewRequest(http.MethodPut, "localhost:8080/api/v1/push/cancel", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("push test Cancel NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin123")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("push test Cancel Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("push bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	return
}

func generatePhone() string {
	prefixes := []string{
		"0505", "0506", "0532", "0533", "0542", "0543", "0555",
	}

	prefix := prefixes[rand.Intn(len(prefixes))]

	number := ""
	for i := 0; i < 7; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}

	return prefix + number
}

func (c *pushController) SendBatch(g *gin.Context) {
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
				p.ScheduledAt = &nextTime
			}
			prio := rand.Intn(3)
			switch prio {
			case 1:
				p.Priority = notification.Notification_Priority_High
			case 2:
				p.Priority = notification.Notification_Priority_Low
			case 3:
				p.Priority = notification.Notification_Priority_Medium
			}
			request.Pushes = append(request.Pushes, p)
		}
		ctx := g.Copy()
		bodyBytes, err := json.Marshal(request)
		if err != nil {
			err = fmt.Errorf("push test SendBatch Marshal error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req, err := http.NewRequest(http.MethodPost, "localhost:8080/api/v1/push/batch", bytes.NewBuffer(bodyBytes))
		if err != nil {
			err = fmt.Errorf("push test SendBatch NewRequest error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", "admin123")

		var resp *http.Response
		if resp, err = c.client.Do(ctx, req); err != nil {
			err = fmt.Errorf("push test SendBatch Do error %w", err)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			err = fmt.Errorf("push bad response: %d", resp.StatusCode)
			logging.Error(ctx, err)
			g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
		}
		g.JSON(http.StatusOK, errorHandling.Error(ctx, err))
	}
}

func (c *pushController) Send(g *gin.Context) {
	id, _ := uuid.NewV7()
	var request = push.SendPushRequest{
		Sender:      fmt.Sprintf("TESTSENDER_%d", id.String()),
		PhoneNumber: generatePhone(),
		Type:        "test_push_type",
		Content:     fmt.Sprintf("Content Batch %d", id.String()),
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
		err = fmt.Errorf("push test Send Marshal error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req, err := http.NewRequest(http.MethodPost, "localhost:8080/api/v1/push", bytes.NewBuffer(bodyBytes))
	if err != nil {
		err = fmt.Errorf("push test Send NewRequest error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin123")

	var resp *http.Response
	if resp, err = c.client.Do(ctx, req); err != nil {
		err = fmt.Errorf("push test Send Do error %w", err)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("push bad response: %d", resp.StatusCode)
		logging.Error(ctx, err)
		g.JSON(http.StatusBadRequest, errorHandling.Error(ctx, err))
	}
	g.JSON(http.StatusOK, errorHandling.Error(ctx, err))
}
