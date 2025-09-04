package errors

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Custom error types
var (
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrConnectionFailed = errors.New("connection failed")
	ErrAuthentication   = errors.New("authentication failed")
	ErrAuthorization    = errors.New("authorization failed")
	ErrResourceNotFound = errors.New("resource not found")
	ErrTimeout          = errors.New("operation timeout")
	ErrValidation       = errors.New("validation failed")
	ErrInternal         = errors.New("internal server error")
)

// ErrorWrapper wraps errors with additional context
type ErrorWrapper struct {
	Err     error
	Message string
	Code    string
}

func (e *ErrorWrapper) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

func (e *ErrorWrapper) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &ErrorWrapper{
		Err:     err,
		Message: message,
	}
}

// WrapWithCode wraps an error with additional context and error code
func WrapWithCode(err error, message, code string) error {
	if err == nil {
		return nil
	}
	return &ErrorWrapper{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// WrapK8sError wraps Kubernetes API errors with meaningful messages
func WrapK8sError(namespace string, err error, resource string) error {
	if err == nil {
		return nil
	}

	switch {
	case apierrors.IsUnauthorized(err):
		return WrapWithCode(err, "unauthorized: check service account or kubeconfig permissions", "UNAUTHORIZED")
	case apierrors.IsForbidden(err):
		return WrapWithCode(err, fmt.Sprintf("forbidden: access denied to list %s in namespace %q", resource, namespace), "FORBIDDEN")
	case apierrors.IsNotFound(err):
		return WrapWithCode(err, fmt.Sprintf("not found: namespace %q or resource %q might not exist", namespace, resource), "NOT_FOUND")
	case apierrors.IsTimeout(err):
		return WrapWithCode(err, fmt.Sprintf("timeout while communicating with Kubernetes API server for %s", resource), "TIMEOUT")
	default:
		return WrapWithCode(err, fmt.Sprintf("failed to list %s in namespace %q", resource, namespace), "K8S_ERROR")
	}
}

// WrapMinioError wraps MinIO errors with meaningful messages
func WrapMinioError(err error) error {
	if err == nil {
		return nil
	}

	// Context related errors
	if errors.Is(err, context.DeadlineExceeded) {
		return WrapWithCode(err, "minio connection timed out", "TIMEOUT")
	}
	if errors.Is(err, context.Canceled) {
		return WrapWithCode(err, "minio request was canceled", "CANCELED")
	}

	errStr := err.Error()

	// Authentication/Authorization
	if strings.Contains(errStr, "invalid Access Key") || strings.Contains(errStr, "signature") {
		return WrapWithCode(err, "minio authentication failed: invalid access key/secret key", "AUTH_FAILED")
	}
	if strings.Contains(errStr, "authorization") || strings.Contains(errStr, "AccessDenied") {
		return WrapWithCode(err, "minio authorization failed: access denied", "AUTH_FAILED")
	}

	// Network/SSL issues
	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "x509") {
		return WrapWithCode(err, "minio SSL/TLS handshake failed", "SSL_ERROR")
	}
	if strings.Contains(errStr, "connection refused") {
		return WrapWithCode(err, "minio connection refused: check endpoint or network", "CONNECTION_REFUSED")
	}
	if strings.Contains(errStr, "no such host") {
		return WrapWithCode(err, "minio endpoint not reachable: DNS resolution failed", "DNS_ERROR")
	}

	// Default
	return WrapWithCode(err, "minio operation failed", "MINIO_ERROR")
}

// WrapHelmError wraps Helm errors with meaningful messages
func WrapHelmError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Common Helm errors
	if strings.Contains(errStr, "chart not found") {
		return WrapWithCode(err, fmt.Sprintf("helm chart not found during %s", operation), "CHART_NOT_FOUND")
	}
	if strings.Contains(errStr, "release not found") {
		return WrapWithCode(err, fmt.Sprintf("helm release not found during %s", operation), "RELEASE_NOT_FOUND")
	}
	if strings.Contains(errStr, "already exists") {
		return WrapWithCode(err, fmt.Sprintf("helm release already exists during %s", operation), "RELEASE_EXISTS")
	}

	return WrapWithCode(err, fmt.Sprintf("helm %s operation failed", operation), "HELM_ERROR")
}

// WrapVeleroError wraps Velero errors with meaningful messages
func WrapVeleroError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Common Velero errors
	if strings.Contains(errStr, "backup not found") {
		return WrapWithCode(err, fmt.Sprintf("velero backup not found during %s", operation), "BACKUP_NOT_FOUND")
	}
	if strings.Contains(errStr, "restore not found") {
		return WrapWithCode(err, fmt.Sprintf("velero restore not found during %s", operation), "RESTORE_NOT_FOUND")
	}

	return WrapWithCode(err, fmt.Sprintf("velero %s operation failed", operation), "VELERO_ERROR")
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code string) bool {
	var wrapper *ErrorWrapper
	if errors.As(err, &wrapper) {
		return wrapper.Code == code
	}
	return false
}

// GetErrorCode returns the error code if available
func GetErrorCode(err error) string {
	var wrapper *ErrorWrapper
	if errors.As(err, &wrapper) {
		return wrapper.Code
	}
	return ""
}
