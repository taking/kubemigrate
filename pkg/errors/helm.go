package errors

import (
	"fmt"
	"strings"
)

// WrapHelmError : Helm 오류 래핑
func WrapHelmError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// 공통 Helm 오류
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
