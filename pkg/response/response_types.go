package response

import "time"

// ErrorResponse 표준 에러 응답 구조체
type ErrorResponse struct {
	Status    string    `json:"status"`               // 항상 "error"
	Code      string    `json:"code"`                 // 에러 코드 (예: "VALIDATION_FAILED", "CONNECTION_FAILED")
	Message   string    `json:"message"`              // 사용자 친화적인 에러 메시지
	Details   string    `json:"details,omitempty"`    // 상세 에러 정보 (선택적)
	Timestamp time.Time `json:"timestamp"`            // 에러 발생 시각
	RequestID string    `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}

// SuccessResponse 표준 성공 응답 구조체
type SuccessResponse struct {
	Status    string      `json:"status"`               // 항상 "success"
	Message   string      `json:"message,omitempty"`    // 성공 메시지 (선택적)
	Data      interface{} `json:"data,omitempty"`       // 응답 데이터 (선택적)
	Timestamp time.Time   `json:"timestamp"`            // 응답 생성 시각
	RequestID string      `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}

// ValidationError 유효성 검사 에러 구조체
type ValidationError struct {
	Field   string `json:"field"`           // 에러가 발생한 필드명
	Message string `json:"message"`         // 에러 메시지
	Value   string `json:"value,omitempty"` // 잘못된 값 (선택적)
}

// ValidationErrorResponse 유효성 검사 에러 응답 구조체
type ValidationErrorResponse struct {
	ErrorResponse
	Errors []ValidationError `json:"errors"` // 유효성 검사 에러 목록
}
