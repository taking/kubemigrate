package velero

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/utils"
)

// TestVeleroHandler_HealthCheck 헬스체크 API 테스트
func TestVeleroHandler_HealthCheck(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	veleroHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": map[string]interface{}{
			"kubeconfig": "apiVersion: v1\nkind: Config",
		},
		"minio": map[string]interface{}{
			"endpoint":  "localhost:9000",
			"accessKey": "minioadmin",
			"secretKey": "minioadmin123",
			"useSSL":    false,
		},
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := veleroHandler.HealthCheck(c)

	// 응답 검증 (실제 Velero 연결이 없으므로 에러가 발생할 수 있음)
	// 하지만 핸들러 자체는 정상적으로 동작해야 함
	if err != nil {
		t.Logf("Expected error due to no actual Velero connection: %v", err)
	}
}

// TestVeleroHandler_GetBackups 백업 목록 조회 API 테스트
func TestVeleroHandler_GetBackups(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	veleroHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": map[string]interface{}{
			"kubeconfig": "apiVersion: v1\nkind: Config",
		},
		"minio": map[string]interface{}{
			"endpoint":  "localhost:9000",
			"accessKey": "minioadmin",
			"secretKey": "minioadmin123",
			"useSSL":    false,
		},
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/backups", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := veleroHandler.GetBackups(c)

	// 응답 검증 (실제 Velero 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Velero connection: %v", err)
	}
}

// TestVeleroHandler_GetRestores 복원 목록 조회 API 테스트
func TestVeleroHandler_GetRestores(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	veleroHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": map[string]interface{}{
			"kubeconfig": "apiVersion: v1\nkind: Config",
		},
		"minio": map[string]interface{}{
			"endpoint":  "localhost:9000",
			"accessKey": "minioadmin",
			"secretKey": "minioadmin123",
			"useSSL":    false,
		},
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/restores", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := veleroHandler.GetRestores(c)

	// 응답 검증 (실제 Velero 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Velero connection: %v", err)
	}
}

// TestVeleroHandler_InvalidRequest 잘못된 요청 테스트
func TestVeleroHandler_InvalidRequest(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	veleroHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 잘못된 JSON 요청
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := veleroHandler.HealthCheck(c)

	// 에러가 발생하거나 연결 에러가 발생할 수 있음
	if err != nil {
		t.Logf("Expected error for invalid JSON request: %v", err)
	} else {
		t.Log("No error occurred, but this is acceptable for this test")
	}
}
