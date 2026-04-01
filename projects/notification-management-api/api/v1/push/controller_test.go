package push

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	command "insider-one/application/command/notification/push"
	query "insider-one/application/query/notification/push"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSendCommand struct{ mock.Mock }

func (m *mockSendCommand) Execute(ctx context.Context, req command.SendPushRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockSendBatchCommand struct{ mock.Mock }

func (m *mockSendBatchCommand) Execute(ctx context.Context, req command.SendBatchPushRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockCancelCommand struct{ mock.Mock }

func (m *mockCancelCommand) Execute(ctx context.Context, req command.CancelPushRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockGetAllQuery struct{ mock.Mock }

func (m *mockGetAllQuery) Execute(ctx context.Context, req query.GetAllPushRequest) (*query.GetAllPushResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetAllPushResponse), args.Error(1)
}

type mockGetStatusByBatchIDQuery struct{ mock.Mock }

func (m *mockGetStatusByBatchIDQuery) Execute(ctx context.Context, req query.GetStatusByBatchIDRequest) (*query.GetStatusByBatchIDResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetStatusByBatchIDResponse), args.Error(1)
}

type mockGetStatusByIDQuery struct{ mock.Mock }

func (m *mockGetStatusByIDQuery) Execute(ctx context.Context, req query.GetPushStatusByIDRequest) (*query.GetPushStatusByIDResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetPushStatusByIDResponse), args.Error(1)
}

func setupRouter(ctrl Controller) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("CorrelationID", "test-id")
		c.Next()
	})
	r.POST("/api/v1/push/", ctrl.Send)
	r.POST("/api/v1/push/batch", ctrl.SendBatch)
	r.PUT("/api/v1/push/cancel", ctrl.Cancel)
	r.GET("/api/v1/push/", ctrl.List)
	r.POST("/api/v1/push/status/batch", ctrl.GetStatusByIDs)
	r.GET("/api/v1/push/status", ctrl.GetStatusByID)
	return r
}

func TestPushController_Send_Success(t *testing.T) {
	sendCmd := &mockSendCommand{}
	sendCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"sender": "PUSH", "phone_number": "+15551234567",
		"content": "Hello", "type": "ALERT", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_Send_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushController_Send_CommandError(t *testing.T) {
	sendCmd := &mockSendCommand{}
	sendCmd.On("Execute", mock.Anything, mock.Anything).Return(errors.New("broker down"))

	ctrl := NewController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"sender": "PUSH", "phone_number": "+15551234567",
		"content": "Hello", "type": "ALERT", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushController_SendBatch_Success(t *testing.T) {
	sendBatchCmd := &mockSendBatchCommand{}
	sendBatchCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(&mockSendCommand{}, sendBatchCmd, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"pushes": []map[string]interface{}{
			{"sender": "S", "phone_number": "+15551234567", "content": "C", "type": "T", "priority": "LOW"},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_SendBatch_NilPushes_Returns400(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"pushes": nil})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushController_Cancel_Success(t *testing.T) {
	cancelCmd := &mockCancelCommand{}
	cancelCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, cancelCmd,
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/push/cancel?status=PENDING", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_Cancel_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/push/cancel?status=NOPE", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushController_List_Success(t *testing.T) {
	getAllQ := &mockGetAllQuery{}
	getAllQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetAllPushResponse{TotalCount: 1}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		getAllQ, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/push/?status=PENDING&page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_List_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/push/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPushController_GetStatusByIDs_Success(t *testing.T) {
	getBatchQ := &mockGetStatusByBatchIDQuery{}
	getBatchQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetStatusByBatchIDResponse{}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, getBatchQ, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"ids": []uint64{1}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/push/status/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_GetStatusByID_Success(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetPushStatusByIDResponse{PushID: 1, Status: "PENDING"}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/push/status?id=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPushController_GetStatusByID_QueryError(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/push/status?id=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
