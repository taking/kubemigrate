package errors

import "errors"

// IsErrorCode : 에러 코드 확인
func IsErrorCode(err error, code string) bool {
	var wrapper *ErrorWrapper
	if errors.As(err, &wrapper) {
		return wrapper.Code == code
	}
	return false
}

// GetErrorCode : 에러 코드 반환
func GetErrorCode(err error) string {
	var wrapper *ErrorWrapper
	if errors.As(err, &wrapper) {
		return wrapper.Code
	}
	return ""
}
