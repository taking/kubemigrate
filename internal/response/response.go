package response

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/logger"
)

// SuccessResponse 표준 성공 응답 구조체
type SuccessResponse struct {
	Status    string      `json:"status"`               // 항상 "success"
	Message   string      `json:"message,omitempty"`    // 성공 메시지 (선택적)
	Data      interface{} `json:"data,omitempty"`       // 응답 데이터 (선택적)
	Timestamp time.Time   `json:"timestamp"`            // 응답 생성 시각
	RequestID string      `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}

// RespondWithSuccessModel : 새로운 SuccessResponse 모델을 사용하여 성공 응답을 보냅니다
func RespondWithSuccessModel(ctx echo.Context, statusCode int, message string, data interface{}) error {
	response := &SuccessResponse{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// 로그 기록
	logger.Info("성공 응답 전송",
		logger.Int("status_code", statusCode),
		logger.String("message", message),
		logger.String("request_id", response.RequestID),
	)

	return ctx.JSON(statusCode, response)
}

// RespondWithData : 데이터만 포함하는 간단한 성공 응답 (중복 방지)
func RespondWithData(ctx echo.Context, statusCode int, data interface{}) error {
	response := &SuccessResponse{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// 로그 기록
	logger.Info("데이터 응답 전송",
		logger.Int("status_code", statusCode),
		logger.String("request_id", response.RequestID),
	)

	return ctx.JSON(statusCode, response)
}

// RespondWithMessage : 메시지만 포함하는 간단한 성공 응답
func RespondWithMessage(ctx echo.Context, statusCode int, message string) error {
	response := &SuccessResponse{
		Status:    "success",
		Message:   message,
		Timestamp: time.Now().UTC(),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// 로그 기록
	logger.Info("메시지 응답 전송",
		logger.Int("status_code", statusCode),
		logger.String("message", message),
		logger.String("request_id", response.RequestID),
	)

	return ctx.JSON(statusCode, response)
}

// RespondWithError : 에러 응답을 보냅니다
func RespondWithError(ctx echo.Context, statusCode int, code, message, details string) error {
	response := map[string]interface{}{
		"status":    "error",
		"code":      code,
		"message":   message,
		"details":   details,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response["request_id"] = requestID
	}

	// 로그 기록
	requestID := ""
	if rid, ok := response["request_id"].(string); ok {
		requestID = rid
	}
	logger.Error("에러 응답 전송",
		logger.Int("status_code", statusCode),
		logger.String("code", code),
		logger.String("message", message),
		logger.String("request_id", requestID),
	)

	return ctx.JSON(statusCode, response)
}
