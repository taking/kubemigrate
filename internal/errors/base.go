package errors

import (
	"errors"
	"fmt"
)

// 커스텀 에러 타입
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

// ErrorWrapper : 에러 래핑
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

// Wrap : 에러 래핑
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &ErrorWrapper{
		Err:     err,
		Message: message,
	}
}

// WrapWithCode : 에러 래핑 및 에러 코드 추가
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
