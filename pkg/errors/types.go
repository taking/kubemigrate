package errors

import (
	"fmt"
	"time"
)

// ErrorType 에러 타입 정의
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	ErrorTypeAuthorization  ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeExternal       ErrorType = "EXTERNAL_ERROR"
	ErrorTypeTimeout        ErrorType = "TIMEOUT_ERROR"
	ErrorTypeRateLimit      ErrorType = "RATE_LIMIT_ERROR"
)

// ErrorCode 에러 코드 정의
type ErrorCode string

const (
	// 클라이언트 관련 에러
	CodeInvalidRequest       ErrorCode = "INVALID_REQUEST"
	CodeMissingParameter     ErrorCode = "MISSING_PARAMETER"
	CodeInvalidParameter     ErrorCode = "INVALID_PARAMETER"
	CodeInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
	CodeUnsupportedPath      ErrorCode = "UNSUPPORTED_PATH"
	CodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceConflict     ErrorCode = "RESOURCE_CONFLICT"

	// 인증/인가 관련 에러
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeTokenExpired ErrorCode = "TOKEN_EXPIRED"

	// 서버 관련 에러
	CodeInternalError      ErrorCode = "INTERNAL_ERROR"
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	CodeTimeout            ErrorCode = "TIMEOUT"
	CodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"

	// 외부 서비스 관련 에러
	CodeKubernetesError ErrorCode = "KUBERNETES_ERROR"
	CodeHelmError       ErrorCode = "HELM_ERROR"
	CodeVeleroError     ErrorCode = "VELERO_ERROR"
	CodeMinioError      ErrorCode = "MINIO_ERROR"
	CodeNetworkError    ErrorCode = "NETWORK_ERROR"
)

// ErrorResponse 표준 에러 응답 구조체
type ErrorResponse struct {
	Type       ErrorType              `json:"type"`
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Cause      error                  `json:"-"`
}

// Error error 인터페이스 구현
func (e *ErrorResponse) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 원본 에러 반환
func (e *ErrorResponse) Unwrap() error {
	return e.Cause
}
