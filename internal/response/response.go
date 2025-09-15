package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/logger"
)

// ErrorResponse : 표준 에러 응답 구조체
type ErrorResponse struct {
	Status    string    `json:"status"`               // 항상 "error"
	Code      string    `json:"code"`                 // 에러 코드 (예: "VALIDATION_FAILED", "CONNECTION_FAILED")
	Message   string    `json:"message"`              // 사용자 친화적인 에러 메시지
	Details   string    `json:"details,omitempty"`    // 상세 에러 정보 (선택적)
	Timestamp time.Time `json:"timestamp"`            // 에러 발생 시각
	RequestID string    `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}

// SuccessResponse : 표준 성공 응답 구조체
type SuccessResponse struct {
	Status    string      `json:"status"`               // 항상 "success"
	Message   string      `json:"message,omitempty"`    // 성공 메시지 (선택적)
	Data      interface{} `json:"data,omitempty"`       // 응답 데이터 (선택적)
	Timestamp time.Time   `json:"timestamp"`            // 응답 생성 시각
	RequestID string      `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}

// ValidationError : 유효성 검사 에러 구조체
type ValidationError struct {
	Field   string `json:"field"`           // 에러가 발생한 필드명
	Message string `json:"message"`         // 에러 메시지
	Value   string `json:"value,omitempty"` // 잘못된 값 (선택적)
}

// ValidationErrorResponse : 유효성 검사 에러 응답 구조체
type ValidationErrorResponse struct {
	ErrorResponse
	Errors []ValidationError `json:"errors"` // 유효성 검사 에러 목록
}

// RespondError : 에러 응답 전송 (utils에서 사용)
func RespondError(ctx echo.Context, statusCode int, message string) error {
	response := &ErrorResponse{
		Status:    "error",
		Message:   message,
		Timestamp: time.Now().UTC(),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	return ctx.JSON(statusCode, response)
}

// RespondWithErrorModel : 새로운 ErrorResponse 모델을 사용하여 에러 응답을 보냅니다
func RespondWithErrorModel(ctx echo.Context, statusCode int, code, message, details string) error {
	response := &ErrorResponse{
		Status:    "error",
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// 로그 기록
	logger.Error("에러 응답 전송",
		logger.Int("status_code", statusCode),
		logger.String("code", code),
		logger.String("message", message),
		logger.String("details", details),
		logger.String("request_id", response.RequestID),
	)

	return ctx.JSON(statusCode, response)
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

// RespondWithValidationError : 유효성 검사 에러 응답
func RespondWithValidationError(ctx echo.Context, errors []ValidationError) error {
	response := &ValidationErrorResponse{
		ErrorResponse: ErrorResponse{
			Status:    "error",
			Code:      "VALIDATION_FAILED",
			Message:   "유효성 검사에 실패했습니다",
			Timestamp: time.Now().UTC(),
		},
		Errors: errors,
	}

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		response.RequestID = requestID
	}

	// 로그 기록
	logger.Warn("유효성 검사 실패",
		logger.Int("error_count", len(errors)),
		logger.String("request_id", response.RequestID),
	)

	return ctx.JSON(http.StatusBadRequest, response)
}
