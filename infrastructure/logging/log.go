package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"insider-one/infrastructure/utils"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gofrs/uuid"
)

type logModel struct {
	HttpStatusCode int    `json:"HttpStatusCode"`
	Duration       int64  `json:"Duration"`
	Time           string `json:"Time"`
	CorrelationID  string `json:"CorrelationID"`
	RequestBody    string `json:"RequestBody"`
	ResponseBody   string `json:"ResponseBody"`
	HttpMethod     string `json:"HttpMethod"`
	Host           string `json:"Host"`
	Url            string `json:"Url"`
	QueryParameter string `json:"QueryParameter"`
	Exception      string `json:"Exception"`
	Message        string `json:"Message"`
	Level          string `json:"Level"`
	DBQuery        string `json:"DBQuery"`
	QueueMessage   string `json:"QueueMessage"`
	Queue          string `json:"Queue"`
	Key            string `json:"Key"`
}

func InternalLogStart(ctx context.Context, request *http.Request) {
	var logModel = logModel{
		Time:           time.Now().Format(utils.Layout_Time),
		Level:          utils.LogLevel_Information,
		Message:        utils.LogStatus_StartRequest,
		HttpStatusCode: http.StatusProcessing,
	}
	logModel.addCorrelationId(ctx).addRequestBody(request).addRequestDetails(request).do()

}

func InternalLogFinish(ctx context.Context, response *http.Response) {

	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_FinishRequest,
	}

	logModel.addCorrelationId(ctx).addResponseBody(response).addResponseDetails(response).addRequestDetails(response.Request).setDuration(ctx, utils.Header_InternalRequestStartTime).do()

}

func ExternalLogStart(ctx context.Context, request *http.Request) {
	var logModel = logModel{
		Time:           time.Now().Format(utils.Layout_Time),
		Level:          utils.LogLevel_Information,
		Message:        utils.LogStatus_Executing,
		HttpStatusCode: http.StatusProcessing,
	}
	logModel.addCorrelationId(ctx).addExternalRequestBody(request).addExternalRequestDetails(request).do()

}

func ExternalLogFinish(ctx context.Context, request *http.Request, response *BodyWriter) {
	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_Executed,
	}
	logModel.addCorrelationId(ctx).addExternalResponseBody(response).addExternalResponseDetails(response.ResponseWriter).addExternalRequestDetails(request).setDuration(ctx, utils.Header_ExternalRequestStartTime).do()

}

func DbQueryStart(ctx context.Context, query string) {
	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_DbQueryStart,
		DBQuery: query,
	}

	logModel.addCorrelationId(ctx).do()
}

func DbQueryFinish(ctx context.Context) {

	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_DbQueryFinish,
	}

	logModel.addCorrelationId(ctx).setDuration(ctx, utils.Header_QueryStartTime).do()
}

func Error(ctx context.Context, err error) {
	var logModel = logModel{
		Time:           time.Now().Format(utils.Layout_Time),
		Level:          utils.LogLevel_Error,
		Message:        "Unexpected Error",
		HttpStatusCode: http.StatusInternalServerError,
	}

	logModel.addCorrelationId(ctx).addException(err).do()
}

func Info(ctx context.Context, message string) {
	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: message,
	}

	logModel.addCorrelationId(ctx).do()
}

func (l *logModel) addRequestBody(request *http.Request) *logModel {
	if request == nil || request.Body == nil {
		return l
	}

	contents, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(contents))

	l.RequestBody = string(contents)
	return l
}

func (l *logModel) addQueueMessage(message []byte) *logModel {
	l.RequestBody = string(message)
	return l
}

func (l *logModel) addQueue(queue string) *logModel {
	l.Queue = queue
	return l
}

func (l *logModel) addKey(key string) *logModel {
	l.Key = key
	return l
}

func (l *logModel) addExternalRequestBody(request *http.Request) *logModel {
	if request == nil {
		return l
	}

	contents, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(contents))

	l.RequestBody = string(contents)

	return l
}

func (l *logModel) addResponseBody(response *http.Response) *logModel {
	if response == nil || response.Body == nil {
		return l
	}

	contents, _ := io.ReadAll(response.Body)
	response.Body = io.NopCloser(bytes.NewBuffer(contents))

	l.ResponseBody = string(contents)
	return l
}

func (l *logModel) addExternalResponseBody(response *BodyWriter) *logModel {
	if response == nil {
		return l
	}

	l.ResponseBody = response.Body.String()
	return l
}

func (l *logModel) addRequestDetails(request *http.Request) *logModel {
	if request == nil {
		return l
	}

	l.HttpMethod = request.Method
	l.Host = request.Host

	if request.URL == nil {
		return l
	}

	l.Url = request.URL.Path
	l.QueryParameter = request.URL.RawQuery
	return l
}

func (l *logModel) addExternalRequestDetails(request *http.Request) *logModel {
	if request == nil {
		return l
	}

	url := *request.URL
	l.Host = request.Host
	l.Url = url.Path
	l.QueryParameter = url.RawQuery
	return l
}

func (l *logModel) addResponseDetails(response *http.Response) *logModel {
	if response == nil {
		return l
	}

	if response.StatusCode != http.StatusOK {
		l.Level = utils.LogLevel_Error
	}

	l.HttpStatusCode = response.StatusCode
	return l
}

func (l *logModel) addExternalResponseDetails(response gin.ResponseWriter) *logModel {
	if response == nil {
		return l
	}

	if response.Status() != http.StatusOK {
		l.Level = utils.LogLevel_Error
	}

	l.HttpStatusCode = response.Status()
	return l
}

func (l *logModel) addException(err error) *logModel {
	l.Exception = utils.GetStackTrace(err)
	return l
}

func (l *logModel) addCorrelationId(ctx context.Context) *logModel {
	var correlationIdString, _ = uuid.NewV4()

	if tempCorrelationId, ok := ctx.Value(utils.Header_CorrelationID).(uuid.UUID); ok {
		correlationIdString = tempCorrelationId
	}

	l.CorrelationID = correlationIdString.String()
	return l
}

func (l *logModel) do() {
	logMessage, _ := json.Marshal(l)
	fmt.Println(string(logMessage))
}

func (l *logModel) setDuration(ctx context.Context, contextKey string) *logModel {
	queryStartTimeString, ok := utils.InterfaceToString(ctx.Value(contextKey))
	if !ok {
		return l
	}
	queryStartTime, _ := time.Parse(utils.Layout_TimeWithNano, queryStartTimeString)
	l.Duration = time.Now().Sub(queryStartTime).Milliseconds()
	return l

}

func WriteMessageToQueue(ctx context.Context, message []byte, queue, key string) {
	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_MessageWritedQueue,
	}

	logModel.addCorrelationId(ctx).addQueue(queue).addKey(key).addQueueMessage(message).do()
}

func ReadMessageFromQueue(ctx context.Context, message []byte, queue, key string) {
	var logModel = logModel{
		Time:    time.Now().Format(utils.Layout_Time),
		Level:   utils.LogLevel_Information,
		Message: utils.LogStatus_MessageReadedQueue,
	}

	logModel.addCorrelationId(ctx).addQueue(queue).addKey(key).addQueueMessage(message).do()
}
