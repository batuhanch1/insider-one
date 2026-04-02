package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	command "insider-one/application/command/notification/sms"
	query "insider-one/application/query/notification/sms"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSendCommand struct{ mock.Mock }

func (m *mockSendCommand) Execute(ctx context.Context, req command.SendSmsRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockSendBatchCommand struct{ mock.Mock }

func (m *mockSendBatchCommand) Execute(ctx context.Context, req command.SendBatchSmsRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockCancelCommand struct{ mock.Mock }

func (m *mockCancelCommand) Execute(ctx context.Context, req command.CancelSmsRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockGetAllQuery struct{ mock.Mock }

func (m *mockGetAllQuery) Execute(ctx context.Context, req query.GetAllSmsRequest) (*query.GetAllSmsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetAllSmsResponse), args.Error(1)
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

func (m *mockGetStatusByIDQuery) Execute(ctx context.Context, req query.GetSmsStatusByIDRequest) (*query.GetSmsStatusByIDResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetSmsStatusByIDResponse), args.Error(1)
}

func setupRouter(ctrl Controller) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("CorrelationID", "test-id")
		c.Next()
	})
	r.POST("/api/v1/sms/", ctrl.Send)
	r.POST("/api/v1/sms/batch", ctrl.SendBatch)
	r.PUT("/api/v1/sms/cancel", ctrl.Cancel)
	r.GET("/api/v1/sms/", ctrl.List)
	r.POST("/api/v1/sms/status/batch", ctrl.GetStatusByIDs)
	r.GET("/api/v1/sms/status", ctrl.GetStatusByID)
	return r
}

func TestSmsController_Send_Success(t *testing.T) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// gt_now
		_ = v.RegisterValidation("gt_now", func(fl validator.FieldLevel) bool {
			val, ok := fl.Field().Interface().(*time.Time)
			if !ok || val == nil {
				return true // omitempty olduğu için nil geçsin
			}
			return val.After(time.Now())
		})

		// within_one_year
		_ = v.RegisterValidation("within_one_year", func(fl validator.FieldLevel) bool {
			val, ok := fl.Field().Interface().(*time.Time)
			if !ok || val == nil {
				return true
			}
			return val.Before(time.Now().AddDate(1, 0, 0))
		})
	}
	sendCmd := &mockSendCommand{}
	sendCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"phone_number": "+15551234567", "sender": "TEST",
		"content": "Hello", "type": "OTP", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_Send_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmsController_Send_CommandError(t *testing.T) {
	sendCmd := &mockSendCommand{}
	sendCmd.On("Execute", mock.Anything, mock.Anything).Return(errors.New("broker down"))

	ctrl := NewController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"phone_number": "+15551234567", "sender": "TEST",
		"content": "Hello", "type": "OTP", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmsController_SendBatch_Success(t *testing.T) {
	sendBatchCmd := &mockSendBatchCommand{}
	sendBatchCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(&mockSendCommand{}, sendBatchCmd, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"sms": []map[string]interface{}{
			{"phone_number": "+15551234567", "sender": "S", "content": "C", "type": "T", "priority": "LOW"},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_SendBatch_NilSms_Returns400(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"sms": nil})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmsController_Cancel_Success(t *testing.T) {
	cancelCmd := &mockCancelCommand{}
	cancelCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, cancelCmd,
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/sms/cancel?status=PENDING", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_Cancel_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/sms/cancel?status=NOPE", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmsController_List_Success(t *testing.T) {
	getAllQ := &mockGetAllQuery{}
	getAllQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetAllSmsResponse{TotalCount: 2}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		getAllQ, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sms/?status=PENDING&page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_List_ValidationError(t *testing.T) {
	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sms/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSmsController_GetStatusByIDs_Success(t *testing.T) {
	getBatchQ := &mockGetStatusByBatchIDQuery{}
	getBatchQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetStatusByBatchIDResponse{}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, getBatchQ, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"ids": []uint64{1}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sms/status/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_GetStatusByID_Success(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetSmsStatusByIDResponse{SmsID: 1, Status: "PENDING"}, nil,
	)

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sms/status?id=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSmsController_GetStatusByID_QueryError(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	ctrl := NewController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sms/status?id=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
