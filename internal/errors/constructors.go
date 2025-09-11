package errors

import (
	"fmt"
	"time"
)

// NewValidationError 유효성 검증 에러 생성
func NewValidationError(code ErrorCode, message, details string) *ErrorResponse {
	return &ErrorResponse{
		Type:      ErrorTypeValidation,
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// NewNotFoundError 리소스 없음 에러 생성
func NewNotFoundError(resource string) *ErrorResponse {
	return &ErrorResponse{
		Type:      ErrorTypeNotFound,
		Code:      CodeResourceNotFound,
		Message:   fmt.Sprintf("Resource not found: %s", resource),
		Timestamp: time.Now(),
	}
}

// NewInternalError 내부 에러 생성
func NewInternalError(operation string, cause error) *ErrorResponse {
	return &ErrorResponse{
		Type:      ErrorTypeInternal,
		Code:      CodeInternalError,
		Message:   fmt.Sprintf("Internal error in %s", operation),
		Details:   cause.Error(),
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

// NewExternalError 외부 서비스 에러 생성
func NewExternalError(service string, operation string, cause error) *ErrorResponse {
	return &ErrorResponse{
		Type:    ErrorTypeExternal,
		Code:    getServiceErrorCode(service),
		Message: fmt.Sprintf("External service error: %s", service),
		Details: fmt.Sprintf("Operation '%s' failed: %s", operation, cause.Error()),
		Context: map[string]interface{}{
			"service":   service,
			"operation": operation,
		},
		Timestamp: time.Now(),
		Cause:     cause,
	}
}

// NewTimeoutError 타임아웃 에러 생성
func NewTimeoutError(operation string, timeout time.Duration) *ErrorResponse {
	return &ErrorResponse{
		Type:    ErrorTypeTimeout,
		Code:    CodeTimeout,
		Message: fmt.Sprintf("Operation timed out: %s", operation),
		Details: fmt.Sprintf("Timeout after %v", timeout),
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout.String(),
		},
		Timestamp: time.Now(),
	}
}

// NewRateLimitError 레이트 리미트 에러 생성
func NewRateLimitError(limit int, window time.Duration) *ErrorResponse {
	return &ErrorResponse{
		Type:    ErrorTypeRateLimit,
		Code:    CodeRateLimitExceeded,
		Message: "Rate limit exceeded",
		Details: fmt.Sprintf("Limit: %d requests per %v", limit, window),
		Context: map[string]interface{}{
			"limit":  limit,
			"window": window.String(),
		},
		Timestamp: time.Now(),
	}
}

// getServiceErrorCode 서비스별 에러 코드 반환
func getServiceErrorCode(service string) ErrorCode {
	switch service {
	case "kubernetes":
		return CodeKubernetesError
	case "helm":
		return CodeHelmError
	case "velero":
		return CodeVeleroError
	case "minio":
		return CodeMinioError
	default:
		return CodeNetworkError
	}
}
