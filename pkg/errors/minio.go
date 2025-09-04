package errors

import (
	"context"
	"errors"
	"strings"
)

// WrapMinioError : MinIO 오류 래핑
func WrapMinioError(err error) error {
	if err == nil {
		return nil
	}

	// 컨텍스트 관련 오류
	if errors.Is(err, context.DeadlineExceeded) {
		return WrapWithCode(err, "minio connection timed out", "TIMEOUT")
	}
	if errors.Is(err, context.Canceled) {
		return WrapWithCode(err, "minio request was canceled", "CANCELED")
	}

	errStr := err.Error()

	// 인증/권한 오류
	if strings.Contains(errStr, "invalid Access Key") || strings.Contains(errStr, "signature") {
		return WrapWithCode(err, "minio authentication failed: invalid access key/secret key", "AUTH_FAILED")
	}
	if strings.Contains(errStr, "authorization") || strings.Contains(errStr, "AccessDenied") {
		return WrapWithCode(err, "minio authorization failed: access denied", "AUTH_FAILED")
	}

	// 네트워크/SSL 오류
	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "x509") {
		return WrapWithCode(err, "minio SSL/TLS handshake failed", "SSL_ERROR")
	}
	if strings.Contains(errStr, "connection refused") {
		return WrapWithCode(err, "minio connection refused: check endpoint or network", "CONNECTION_REFUSED")
	}
	if strings.Contains(errStr, "no such host") {
		return WrapWithCode(err, "minio endpoint not reachable: DNS resolution failed", "DNS_ERROR")
	}

	// 기본 오류
	return WrapWithCode(err, "minio operation failed", "MINIO_ERROR")
}
