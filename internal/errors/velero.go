package errors

import (
	"fmt"
	"strings"
)

// WrapVeleroError : Velero 오류 래핑
func WrapVeleroError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// 공통 Velero 오류
	if strings.Contains(errStr, "backup not found") {
		return WrapWithCode(err, fmt.Sprintf("velero backup not found during %s", operation), "BACKUP_NOT_FOUND")
	}
	if strings.Contains(errStr, "restore not found") {
		return WrapWithCode(err, fmt.Sprintf("velero restore not found during %s", operation), "RESTORE_NOT_FOUND")
	}

	return WrapWithCode(err, fmt.Sprintf("velero %s operation failed", operation), "VELERO_ERROR")
}
