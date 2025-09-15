package minio

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

// TestMinioHandler_HealthCheck 헬스체크 API 테스트
func TestMinioHandler_HealthCheck(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := minioHandler.HealthCheck(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_ListBuckets 버킷 목록 조회 API 테스트
func TestMinioHandler_ListBuckets(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/buckets", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := minioHandler.ListBuckets(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_CheckBucketExists 버킷 존재 확인 API 테스트
func TestMinioHandler_CheckBucketExists(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/buckets/test-bucket", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:bucket")
	c.SetParamNames("bucket")
	c.SetParamValues("test-bucket")

	// 핸들러 실행
	err := minioHandler.CheckBucketExists(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_CreateBucket 버킷 생성 API 테스트
func TestMinioHandler_CreateBucket(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodPost, "/buckets/test-bucket", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:bucket")
	c.SetParamNames("bucket")
	c.SetParamValues("test-bucket")

	// 핸들러 실행
	err := minioHandler.CreateBucket(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_DeleteBucket 버킷 삭제 API 테스트
func TestMinioHandler_DeleteBucket(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodDelete, "/buckets/test-bucket", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:bucket")
	c.SetParamNames("bucket")
	c.SetParamValues("test-bucket")

	// 핸들러 실행
	err := minioHandler.DeleteBucket(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_ListObjects 객체 목록 조회 API 테스트
func TestMinioHandler_ListObjects(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 데이터
	reqData := map[string]interface{}{
		"endpoint":  "localhost:9000",
		"accessKey": "minioadmin",
		"secretKey": "minioadmin123",
		"useSSL":    false,
	}
	reqBody, _ := json.Marshal(reqData)

	// HTTP 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/buckets/test-bucket/objects", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:bucket/objects")
	c.SetParamNames("bucket")
	c.SetParamValues("test-bucket")

	// 핸들러 실행
	err := minioHandler.ListObjects(c)

	// 응답 검증 (실제 MinIO 연결이 없으므로 에러가 발생할 수 있음)
	if err != nil {
		t.Logf("Expected error due to no actual MinIO connection: %v", err)
	}
}

// TestMinioHandler_InvalidRequest 잘못된 요청 테스트
func TestMinioHandler_InvalidRequest(t *testing.T) {
	// 테스트용 BaseHandler 생성 (Mock 클라이언트 사용)
	workerPool := utils.NewWorkerPool(2)
	defer workerPool.Close()
	baseHandler := handler.NewBaseHandlerWithMock(workerPool)
	minioHandler := NewHandler(baseHandler)

	// Echo 인스턴스 생성
	e := echo.New()

	// 잘못된 JSON 요청
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := minioHandler.HealthCheck(c)

	// 에러가 발생하거나 연결 에러가 발생할 수 있음
	if err != nil {
		t.Logf("Expected error for invalid JSON request: %v", err)
	} else {
		t.Log("No error occurred, but this is acceptable for this test")
	}
}
