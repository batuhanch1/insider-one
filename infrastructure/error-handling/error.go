package error_handling

import (
	"context"
	"insider-one/infrastructure/utils"
	"net/http"
)

type Errors struct {
	StatusCode    int    `json:"StatusCode,omitempty"`
	Err           error  `json:"Err,omitempty"`
	ErrorType     string `json:"ErrorType,omitempty"`
	Parameter     string `json:"Parameter,omitempty"`
	CorrelationID string `json:"CorrelationID,omitempty"`
	Message       string `json:"Message,omitempty"`
	StackTrace    string `json:"StackTrace,omitempty"`
}

func Error(ctx context.Context, err error) Errors {
	correlationId, _ := utils.InterfaceToString(ctx.Value("CorrelationID"))
	return Errors{
		StatusCode:    http.StatusInternalServerError,
		Err:           err,
		ErrorType:     "Unexpected Error",
		CorrelationID: correlationId,
		StackTrace:    utils.GetStackTrace(err),
	}
}

func ValidationError(ctx context.Context, err error, parameter string) Errors {
	correlationId := ctx.Value("CorrelationID").(string)
	return Errors{
		StatusCode:    http.StatusBadRequest,
		Err:           err,
		ErrorType:     "Validation Error",
		Parameter:     parameter,
		CorrelationID: correlationId,
		Message:       err.Error(),
	}
}
