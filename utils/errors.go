package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// WrapK8sError : Kubernetes 관련 에러 래핑
func WrapK8sError(namespace string, err error, resource string) error {
	if err == nil {
		return nil
	}
	switch {
	case apierrors.IsUnauthorized(err):
		return fmt.Errorf("unauthorized: check service account or kubeconfig permissions: %w", err)
	case apierrors.IsForbidden(err):
		return fmt.Errorf("forbidden: access denied to list %s in namespace %q: %w", resource, namespace, err)
	case apierrors.IsNotFound(err):
		return fmt.Errorf("not found: namespace %q or resource %q might not exist: %w", namespace, resource, err)
	case apierrors.IsTimeout(err):
		return fmt.Errorf("timeout while communicating with Kubernetes API server for %s: %w", resource, err)
	default:
		return fmt.Errorf("failed to list %s in namespace %q: %w", resource, namespace, err)
	}
}

// WrapMinioError : MinIO 관련 에러 래핑
func WrapMinioError(err error) error {
	if err == nil {
		return nil
	}

	// Context 관련 에러
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("minio connection timed out: %w", err)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("minio request was canceled: %w", err)
	}

	// 인증/인가 관련
	if strings.Contains(err.Error(), "invalid Access Key") ||
		strings.Contains(err.Error(), "signature") {
		return fmt.Errorf("minio authentication failed: invalid access key/secret key: %w", err)
	}
	if strings.Contains(err.Error(), "authorization") ||
		strings.Contains(err.Error(), "AccessDenied") {
		return fmt.Errorf("minio authorization failed: access denied: %w", err)
	}

	// 네트워크 / SSL 문제
	if strings.Contains(err.Error(), "certificate") ||
		strings.Contains(err.Error(), "x509") {
		return fmt.Errorf("minio SSL/TLS handshake failed: %w", err)
	}
	if strings.Contains(err.Error(), "connection refused") {
		return fmt.Errorf("minio connection refused: check endpoint or network: %w", err)
	}
	if strings.Contains(err.Error(), "no such host") {
		return fmt.Errorf("minio endpoint not reachable: DNS resolution failed: %w", err)
	}

	// 기본
	return fmt.Errorf("minio operation failed: %w", err)
}
