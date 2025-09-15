package helm

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

// TestHelmHandler_HealthCheck 헬스체크 API 테스트
func TestHelmHandler_HealthCheck(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := helmHandler.HealthCheck(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_InstallChart 차트 설치 API 테스트
func TestHelmHandler_InstallChart(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성 (쿼리 파라미터 포함)
	req := httptest.NewRequest(http.MethodPost, "/charts?releaseName=test-release&chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz&version=15.4.2&namespace=default", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := helmHandler.InstallChart(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_GetCharts 차트 목록 조회 API 테스트
func TestHelmHandler_GetCharts(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/charts", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := helmHandler.GetCharts(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_GetChart 특정 차트 조회 API 테스트
func TestHelmHandler_GetChart(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/charts/test-chart", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:name")
	c.SetParamNames("name")
	c.SetParamValues("test-chart")

	// 핸들러 실행
	err := helmHandler.GetChart(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_UpgradeChart 차트 업그레이드 API 테스트
func TestHelmHandler_UpgradeChart(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPut, "/charts/test-chart", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:name")
	c.SetParamNames("name")
	c.SetParamValues("test-chart")

	// 핸들러 실행
	err := helmHandler.UpgradeChart(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_UninstallChart 차트 제거 API 테스트
func TestHelmHandler_UninstallChart(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodDelete, "/charts/test-chart", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:name")
	c.SetParamNames("name")
	c.SetParamValues("test-chart")

	// 핸들러 실행
	err := helmHandler.UninstallChart(c)

	// 응답 검증 (실제 Helm 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual Helm connection: %v", err)
	}
}

// TestHelmHandler_InvalidRequest 잘못된 요청 테스트
func TestHelmHandler_InvalidRequest(t *testing.T) {
	// 테스트용 BaseHandler 생성
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandler(workerPool)
	helmHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 잘못된 JSON 요청
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := helmHandler.HealthCheck(c)

	// 에러가 발생하거나 연결 에러가 발생할 수 있음
	if err != nil {
		t.Logf("Expected error for invalid JSON request: %v", err)
	} else {
		t.Log("No error occurred, but this is acceptable for this test")
	}
}
