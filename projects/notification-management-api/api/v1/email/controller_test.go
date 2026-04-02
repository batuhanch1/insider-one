package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	command "insider-one/application/command/notification/email"
	query "insider-one/application/query/notification/email"
	email_domain "insider-one/domain/notification/email"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock commands and queries ---

type mockSendCommand struct{ mock.Mock }

func (m *mockSendCommand) Execute(ctx context.Context, req command.SendEmailRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockSendBatchCommand struct{ mock.Mock }

func (m *mockSendBatchCommand) Execute(ctx context.Context, req command.SendBatchEmailRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockCancelCommand struct{ mock.Mock }

func (m *mockCancelCommand) Execute(ctx context.Context, req command.CancelEmailRequest) error {
	return m.Called(ctx, req).Error(0)
}

type mockGetAllQuery struct{ mock.Mock }

func (m *mockGetAllQuery) Execute(ctx context.Context, req query.GetAllEmailRequest) (*query.GetAllEmailResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetAllEmailResponse), args.Error(1)
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

func (m *mockGetStatusByIDQuery) Execute(ctx context.Context, req query.GetEmailStatusByIDRequest) (*query.GetEmailStatusByIDResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*query.GetEmailStatusByIDResponse), args.Error(1)
}

// --- Test helpers ---

func setupRouter(ctrl Controller) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("CorrelationID", "test-correlation-id")
		c.Next()
	})
	r.POST("/api/v1/email/", ctrl.Send)
	r.POST("/api/v1/email/batch", ctrl.SendBatch)
	r.PUT("/api/v1/email/cancel", ctrl.Cancel)
	r.GET("/api/v1/email/", ctrl.List)
	r.POST("/api/v1/email/status/batch", ctrl.GetStatusByIDs)
	r.GET("/api/v1/email/status", ctrl.GetStatusByID)
	return r
}

func newController(
	sendCmd query.SendCommand,
	sendBatchCmd query.SendBatchCommand,
	cancelCmd query.CancelCommand,
	getAllQ query.GetAllQuery,
	getBatchQ query.GetStatusByBatchIDQuery,
	getByIDQ query.GetEmailStatusByIDQuery,
) Controller {
	return NewController(sendCmd, sendBatchCmd, cancelCmd, getAllQ, getBatchQ, getByIDQ)
}

// --- Tests ---

func TestEmailController_Send_Success(t *testing.T) {
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

	ctrl := newController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"to": "to@example.com", "from": "from@example.com",
		"subject": "Hello", "content": "World", "type": "PROMO", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	sendCmd.AssertExpectations(t)
}

func TestEmailController_Send_ValidationError_MissingFields(t *testing.T) {
	sendCmd := &mockSendCommand{}
	ctrl := newController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	sendCmd.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestEmailController_Send_CommandError(t *testing.T) {
	sendCmd := &mockSendCommand{}
	sendCmd.On("Execute", mock.Anything, mock.Anything).Return(errors.New("broker down"))

	ctrl := newController(sendCmd, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"to": "to@example.com", "from": "from@example.com",
		"subject": "Sub", "content": "Body", "type": "T", "priority": "HIGH",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_SendBatch_Success(t *testing.T) {
	sendBatchCmd := &mockSendBatchCommand{}
	sendBatchCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := newController(&mockSendCommand{}, sendBatchCmd, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{
		"emails": []map[string]interface{}{
			{"to": "a@a.com", "from": "b@b.com", "subject": "S", "content": "C", "type": "T", "priority": "HIGH"},
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmailController_SendBatch_NilEmails_Returns400(t *testing.T) {
	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"emails": nil})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_Cancel_Success(t *testing.T) {
	cancelCmd := &mockCancelCommand{}
	cancelCmd.On("Execute", mock.Anything, mock.Anything).Return(nil)

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, cancelCmd,
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/email/cancel?status=PENDING", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmailController_Cancel_ValidationError_InvalidStatus(t *testing.T) {
	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/email/cancel?status=INVALID", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_Cancel_CommandError(t *testing.T) {
	cancelCmd := &mockCancelCommand{}
	cancelCmd.On("Execute", mock.Anything, mock.Anything).Return(errors.New("db error"))

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, cancelCmd,
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/email/cancel?status=PENDING", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_List_Success(t *testing.T) {
	getAllQ := &mockGetAllQuery{}
	getAllQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetAllEmailResponse{
			Emails:     email_domain.Emails{{ID: 1}},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}, nil,
	)

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		getAllQ, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/email/?status=PENDING&page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp query.GetAllEmailResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, resp.TotalCount)
}

func TestEmailController_List_ValidationError(t *testing.T) {
	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/email/", nil) // missing required params
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_List_QueryError(t *testing.T) {
	getAllQ := &mockGetAllQuery{}
	getAllQ.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		getAllQ, &mockGetStatusByBatchIDQuery{}, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/email/?status=PENDING&page=1&page_size=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_GetStatusByIDs_Success(t *testing.T) {
	getBatchQ := &mockGetStatusByBatchIDQuery{}
	getBatchQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetStatusByBatchIDResponse{
			Emails: []query.GetEmailStatusByIDResponse{{EmailID: 1, Status: "PENDING"}},
		}, nil,
	)

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, getBatchQ, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"ids": []uint64{1}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/status/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmailController_GetStatusByIDs_QueryError(t *testing.T) {
	getBatchQ := &mockGetStatusByBatchIDQuery{}
	getBatchQ.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, getBatchQ, &mockGetStatusByIDQuery{})
	r := setupRouter(ctrl)

	body, _ := json.Marshal(map[string]interface{}{"ids": []uint64{1}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/email/status/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmailController_GetStatusByID_Success(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(
		&query.GetEmailStatusByIDResponse{EmailID: 42, Status: "PENDING"}, nil,
	)

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/email/status?id=42", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmailController_GetStatusByID_QueryError(t *testing.T) {
	getByIDQ := &mockGetStatusByIDQuery{}
	getByIDQ.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	ctrl := newController(&mockSendCommand{}, &mockSendBatchCommand{}, &mockCancelCommand{},
		&mockGetAllQuery{}, &mockGetStatusByBatchIDQuery{}, getByIDQ)
	r := setupRouter(ctrl)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/email/status?id=42", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
