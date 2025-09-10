package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/logger"
	responseTypes "github.com/taking/kubemigrate/pkg/response"
)

// RespondError : 에러 응답 전송 (utils에서 사용)
func RespondError(ctx echo.Context, statusCode int, message string) error {
	response := &responseTypes.ErrorResponse{
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
	response := &responseTypes.ErrorResponse{
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
	response := &responseTypes.SuccessResponse{
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
	response := &responseTypes.SuccessResponse{
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
	response := &responseTypes.SuccessResponse{
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
func RespondWithValidationError(ctx echo.Context, errors []responseTypes.ValidationError) error {
	response := &responseTypes.ValidationErrorResponse{
		ErrorResponse: responseTypes.ErrorResponse{
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
