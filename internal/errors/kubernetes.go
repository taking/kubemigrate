package errors

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// WrapK8sError : Kubernetes API 오류 래핑
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
