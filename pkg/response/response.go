package response

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"taking.kr/velero/pkg/models"
)

// SuccessResponse : 성공 응답 생성
func SuccessResponse(data interface{}) *models.SuccessResponse {
	return &models.SuccessResponse{
		Status:    "success",
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// ErrorResponse : 에러 응답 생성
func ErrorResponse(message string) *models.ErrorResponse {
	return &models.ErrorResponse{
		Status:    "error",
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
}

// ErrorResponseWithCode : 에러 코드가 있는 에러 응답 생성
func ErrorResponseWithCode(message, code string) *models.ErrorResponse {
	return &models.ErrorResponse{
		Status:    "error",
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC(),
	}
}

// StatusResponse : 상태 응답 생성
func StatusResponse(status, message string) *models.SuccessResponse {
	return &models.SuccessResponse{
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
		case *models.SuccessResponse:
			r.RequestID = requestID
		case *models.ErrorResponse:
			r.RequestID = requestID
		case *models.HealthResponse:
			// HealthResponse doesn't have RequestID field, skip
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

	// 요청 ID 추가
	if requestID := ctx.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		// We can extend HealthCheckResponse to include RequestID if needed
	}

	return ctx.JSON(http.StatusOK, response)
}

// RespondWithErrorModel : 새로운 ErrorResponse 모델을 사용하여 에러 응답을 보냅니다
func RespondWithErrorModel(ctx echo.Context, statusCode int, code, message, details string) error {
	response := &models.ErrorResponse{
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
	response := &models.SuccessResponse{
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

// RespondWithHealthModel : 새로운 HealthResponse 모델을 사용하여 헬스체크 응답을 보냄
func RespondWithHealthModel(ctx echo.Context, service, status, message string) error {
	response := &models.HealthResponse{
		Status:    status,
		Service:   service,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}

	// 로그 기록
	zap.L().Info("헬스체크 응답",
		zap.String("service", service),
		zap.String("status", status),
		zap.String("message", message),
	)

	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return ctx.JSON(statusCode, response)
}

// RespondWithValidationError : 유효성 검사 에러 응답
func RespondWithValidationError(ctx echo.Context, errors []models.ValidationError) error {
	response := &models.ValidationErrorResponse{
		ErrorResponse: models.ErrorResponse{
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
