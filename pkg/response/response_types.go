package response

import "time"

// SuccessResponse 표준 성공 응답 구조체
type SuccessResponse struct {
	Status    string      `json:"status"`               // 항상 "success"
	Message   string      `json:"message,omitempty"`    // 성공 메시지 (선택적)
	Data      interface{} `json:"data,omitempty"`       // 응답 데이터 (선택적)
	Timestamp time.Time   `json:"timestamp"`            // 응답 생성 시각
	RequestID string      `json:"request_id,omitempty"` // 요청 추적용 ID (선택적)
}
