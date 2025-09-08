package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	responseTypes "github.com/taking/kubemigrate/pkg/response"
	"go.uber.org/zap"
)

// SuccessResponse : 성공 응답 생성
func SuccessResponse(data interface{}) *responseTypes.SuccessResponse {
	return &responseTypes.SuccessResponse{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// ErrorResponse : 에러 응답 생성
func ErrorResponse(message string) *responseTypes.ErrorResponse {
	return &responseTypes.ErrorResponse{
		Status:    "error",
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}

// ErrorResponseWithCode : 에러 코드가 있는 에러 응답 생성
func ErrorResponseWithCode(message, code string) *responseTypes.ErrorResponse {
	return &responseTypes.ErrorResponse{
		Status:    "error",
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC(),
	}
}

// StatusResponse : 상태 응답 생성
func StatusResponse(status, message string) *responseTypes.SuccessResponse {
	return &responseTypes.SuccessResponse{
		Status:    status,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}

// RespondJSON : JSON 응답 전송
func RespondJSON(ctx echo.Context, statusCode int, response interface{}) error {
	// Add request ID if available
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		// Try to add request ID to models that support it
		switch r := response.(type) {
		case *responseTypes.SuccessResponse:
			r.RequestID = requestID
		case *responseTypes.ErrorResponse:
			r.RequestID = requestID
		}
	}

	// Log response for debugging
	zap.L().Debug("sending response",
		zap.Int("status_code", statusCode),
		zap.String("request_id", ctx.Response().Header().Get(echo.HeaderXRequestID)),
	)

	return ctx.JSON(statusCode, response)
}

// RespondError : 에러 응답 전송
func RespondError(ctx echo.Context, statusCode int, message string) error {
	return RespondJSON(ctx, statusCode, ErrorResponse(message))
}

// RespondErrorWithCode : 에러 코드가 있는 에러 응답 전송
func RespondErrorWithCode(ctx echo.Context, statusCode int, message, code string) error {
	return RespondJSON(ctx, statusCode, ErrorResponseWithCode(message, code))
}

// RespondStatus : 상태 응답 전송
func RespondStatus(ctx echo.Context, status, message string) error {
	return RespondJSON(ctx, http.StatusOK, StatusResponse(status, message))
}

// RespondSuccess : 성공 응답 전송
func RespondSuccess(ctx echo.Context, data interface{}) error {
	return RespondJSON(ctx, http.StatusOK, SuccessResponse(data))
}

// RespondCreated : 생성 응답 전송
func RespondCreated(ctx echo.Context, data interface{}) error {
	return RespondJSON(ctx, http.StatusCreated, SuccessResponse(data))
}

// RespondAccepted : 수락 응답 전송
func RespondAccepted(ctx echo.Context, message string) error {
	return RespondJSON(ctx, http.StatusAccepted, StatusResponse("accepted", message))
}

// RespondNoContent : 콘텐츠 없음 응답 전송
func RespondNoContent(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

// RespondBadRequest : 잘못된 요청 응답 전송
func RespondBadRequest(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusBadRequest, message)
}

// RespondUnauthorized : 권한 없음 응답 전송
func RespondUnauthorized(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusUnauthorized, message)
}

// RespondForbidden : 금지된 요청 응답 전송
func RespondForbidden(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusForbidden, message)
}

// RespondNotFound : 찾을 수 없음 응답 전송
func RespondNotFound(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusNotFound, message)
}

// RespondInternalError : 내부 서버 오류 응답 전송
func RespondInternalError(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusInternalServerError, message)
}

// RespondServiceUnavailable : 서비스 사용 불가 응답 전송
func RespondServiceUnavailable(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusServiceUnavailable, message)
}

// RespondTooManyRequests : 너무 많은 요청 응답 전송
func RespondTooManyRequests(ctx echo.Context, message string) error {
	return RespondError(ctx, http.StatusTooManyRequests, message)
}

// HealthCheckResponse : 헬스체크 응답 구조체
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Message   string            `json:"message"`
	Features  []string          `json:"features"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
}

// RespondHealthCheck : 헬스체크 응답 전송
func RespondHealthCheck(ctx echo.Context, status, version, message string, features []string, services map[string]string) error {
	response := &HealthCheckResponse{
		Status:    status,
		Version:   version,
		Message:   message,
		Features:  features,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	// 요청 ID 추가 (필요시 HealthCheckResponse에 RequestID 필드 추가 가능)
	_ = ctx.Response().Header().Get(echo.HeaderXRequestID)

	return ctx.JSON(http.StatusOK, response)
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
	zap.L().Error("에러 응답 전송",
		zap.Int("status_code", statusCode),
		zap.String("code", code),
		zap.String("message", message),
		zap.String("details", details),
		zap.String("request_id", response.RequestID),
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
	zap.L().Info("성공 응답 전송",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("request_id", response.RequestID),
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
	zap.L().Warn("유효성 검사 실패",
		zap.Int("error_count", len(errors)),
		zap.String("request_id", response.RequestID),
	)

	return ctx.JSON(http.StatusBadRequest, response)
}
