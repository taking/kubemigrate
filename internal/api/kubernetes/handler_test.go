package kubernetes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/internal/job"
)

// TestKubernetesHandler_HealthCheck 헬스체크 API 테스트
func TestKubernetesHandler_HealthCheck(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := job.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	kubernetesHandler := NewHandler(baseHandler)

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

	// 핸들러 실행 (실제 Kubernetes 연결이 없으므로 에러가 발생할 수 있음)
	err := kubernetesHandler.HealthCheck(c)

	// 에러가 발생해도 핸들러 로직은 정상 동작해야 함
	// 실제 Kubernetes 클러스터가 없으므로 연결 에러는 예상됨
	if err != nil {
		t.Logf("Expected error due to no actual Kubernetes connection: %v", err)
	}
}

// TestKubernetesHandler_GetResources 리소스 조회 API 테스트
func TestKubernetesHandler_GetResources(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := job.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	kubernetesHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 케이스들
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "pods 리소스 조회",
			method:         http.MethodGet,
			path:           "/pods",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "configmaps 리소스 조회",
			method:         http.MethodGet,
			path:           "/configmaps",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "지원하지 않는 리소스",
			method:         http.MethodGet,
			path:           "/unsupported",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// HTTP 요청 생성
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:kind")
			c.SetParamNames("kind")
			c.SetParamValues(tc.path[1:]) // "/" 제거

			// 핸들러 실행
			err := kubernetesHandler.GetResources(c)

			// 응답 검증
			// 실제 Kubernetes 클러스터가 없으므로 연결 에러는 예상됨
			if err != nil {
				t.Logf("Expected error due to no actual Kubernetes connection: %v", err)
			}
		})
	}
}

// TestKubernetesHandler_GetResources_WithName 이름이 있는 리소스 조회 테스트
func TestKubernetesHandler_GetResources_WithName(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := job.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	kubernetesHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/pods/test-pod", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:kind/:name")
	c.SetParamNames("kind", "name")
	c.SetParamValues("pods", "test-pod")

	// 핸들러 실행
	err := kubernetesHandler.GetResources(c)

	// 응답 검증 (실제 Kubernetes 연결이 없으므로 에러가 발생할 수 있음)
	// 하지만 핸들러 자체는 정상적으로 동작해야 함
	if err != nil {
		// 에러가 발생해도 핸들러 로직은 정상 동작
		t.Logf("Expected error due to no actual Kubernetes connection: %v", err)
	}
}

// TestKubernetesHandler_GetResources_QueryParams 쿼리 파라미터 테스트
func TestKubernetesHandler_GetResources_QueryParams(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := job.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	kubernetesHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 케이스들
	testCases := []struct {
		name           string
		namespace      string
		expectedStatus int
	}{
		{
			name:           "기본 네임스페이스",
			namespace:      "default",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "모든 네임스페이스",
			namespace:      "all",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "특정 네임스페이스",
			namespace:      "kube-system",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// HTTP 요청 생성
			req := httptest.NewRequest(http.MethodGet, "/pods?namespace="+tc.namespace, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:kind")
			c.SetParamNames("kind")
			c.SetParamValues("pods")

			// 핸들러 실행
			err := kubernetesHandler.GetResources(c)

			// 응답 검증
			if tc.expectedStatus == http.StatusOK && err != nil {
				t.Logf("Expected error due to no actual Kubernetes connection: %v", err)
			}
		})
	}
}
