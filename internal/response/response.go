package response

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/logger"
	responseTypes "github.com/taking/kubemigrate/pkg/response"
)

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
