package utils

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

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
