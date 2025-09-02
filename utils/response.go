package utils

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Status: "success",
		Data:   data,
	}
}

func ErrorResponse(message string) *APIResponse {
	return &APIResponse{
		Status: "error",
		Error:  message,
	}
}

func StatusResponse(status, message string) *APIResponse {
	return &APIResponse{
		Status:  status,
		Message: message,
	}
}

func RespondJSON(ctx echo.Context, statusCode int, response *APIResponse) error {
	return ctx.JSON(statusCode, response)
}

// RespondError : 데이터 성공 응답
func RespondError(ctx echo.Context, statusCode int, message string) error {
	return RespondJSON(ctx, statusCode, ErrorResponse(message))
}

// RespondStatus : 상태 메시지 응답
func RespondStatus(ctx echo.Context, status, message string) error {
	return RespondJSON(ctx, http.StatusOK, StatusResponse(status, message))
}

// RespondSuccess : 데이터 성공 응답
func RespondSuccess(ctx echo.Context, data interface{}) error {
	return RespondJSON(ctx, http.StatusOK, SuccessResponse(data))
}
