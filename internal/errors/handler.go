package errors

import (
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// ErrorHandler 에러 핸들러 구조체
type ErrorHandler struct {
	logger *slog.Logger
	debug  bool
}

// NewErrorHandler 새로운 에러 핸들러 생성
func NewErrorHandler(logger *slog.Logger, debug bool) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		debug:  debug,
	}
}

// HandleError 에러 처리 및 응답
func (eh *ErrorHandler) HandleError(c echo.Context, err error, operation string) error {
	errorResp := eh.wrapError(err, operation, c)

	// 에러 로깅
	eh.logError(errorResp, c)

	// HTTP 상태 코드 결정
	statusCode := eh.getStatusCode(errorResp.Type)

	// 에러 응답 반환
	return c.JSON(statusCode, errorResp)
}

// wrapError 에러를 ErrorResponse로 래핑
func (eh *ErrorHandler) wrapError(err error, operation string, c echo.Context) *ErrorResponse {
	// 이미 ErrorResponse인 경우
	if errorResp, ok := err.(*ErrorResponse); ok {
		errorResp.RequestID = eh.getRequestID(c)
		errorResp.Timestamp = time.Now()
		return errorResp
	}

	// 일반 에러를 ErrorResponse로 변환
	errorResp := &ErrorResponse{
		Type:    ErrorTypeInternal,
		Code:    CodeInternalError,
		Message: "An unexpected error occurred",
		Details: err.Error(),
		Context: map[string]interface{}{
			"operation": operation,
		},
		Timestamp: time.Now(),
		RequestID: eh.getRequestID(c),
		Cause:     err,
	}

	// 디버그 모드에서 스택 트레이스 추가
	if eh.debug {
		errorResp.StackTrace = eh.getStackTrace()
	}

	return errorResp
}

// logError 에러 로깅
func (eh *ErrorHandler) logError(err *ErrorResponse, c echo.Context) {
	attrs := []slog.Attr{
		slog.String("error_type", string(err.Type)),
		slog.String("error_code", string(err.Code)),
		slog.String("request_id", err.RequestID),
		slog.String("method", c.Request().Method),
		slog.String("path", c.Request().URL.Path),
		slog.String("user_agent", c.Request().UserAgent()),
		slog.String("remote_ip", c.RealIP()),
	}

	// 컨텍스트 정보 추가
	for k, v := range err.Context {
		attrs = append(attrs, slog.Any(k, v))
	}

	// 에러 레벨에 따른 로깅
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}

	switch err.Type {
	case ErrorTypeValidation, ErrorTypeNotFound:
		eh.logger.Warn("Request error", args...)
	case ErrorTypeAuthentication, ErrorTypeAuthorization:
		eh.logger.Warn("Authentication/Authorization error", args...)
	case ErrorTypeInternal, ErrorTypeExternal:
		eh.logger.Error("Internal/External error", args...)
	default:
		eh.logger.Error("Unknown error", args...)
	}
}

// getStatusCode 에러 타입에 따른 HTTP 상태 코드 반환
func (eh *ErrorHandler) getStatusCode(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID 요청 ID 추출
func (eh *ErrorHandler) getRequestID(c echo.Context) string {
	if requestID := c.Response().Header().Get(echo.HeaderXRequestID); requestID != "" {
		return requestID
	}
	return "unknown"
}

// getStackTrace 스택 트레이스 생성
func (eh *ErrorHandler) getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
