package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/config"
)

// TestMinioConfigParser_Parse : MinIO 설정 파서 테스트
func TestMinioConfigParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedError  bool
		expectedConfig config.MinioConfig
	}{
		{
			name: "Valid MinIO config",
			requestBody: map[string]interface{}{
				"endpoint":  "localhost:9000",
				"accessKey": "minioadmin",
				"secretKey": "minioadmin123",
				"useSSL":    false,
			},
			expectedError: false,
			expectedConfig: config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
		},
		{
			name: "Invalid JSON",
			requestBody: map[string]interface{}{
				"endpoint": "localhost:9000",
				// missing required fields
			},
			expectedError: false, // JSON 파싱은 성공하지만 검증에서 실패
			expectedConfig: config.MinioConfig{
				Endpoint: "localhost:9000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			e := echo.New()
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var minioConfig config.MinioConfig
			parser := NewMinioConfigParser(&minioConfig)

			// When
			err := parser.Parse(c)

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestMinioConfigParser_Validate : MinIO 설정 검증 테스트
func TestMinioConfigParser_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        config.MinioConfig
		expectedError bool
	}{
		{
			name: "Valid config",
			config: config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: false,
		},
		{
			name: "Empty endpoint",
			config: config.MinioConfig{
				Endpoint:  "",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: true,
		},
		{
			name: "Template variables in endpoint",
			config: config.MinioConfig{
				Endpoint:  "{{minio_url}}",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			parser := NewMinioConfigParser(&tt.config)

			// When
			err := parser.Validate()

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestKubeConfigParser_Parse : Kubernetes 설정 파서 테스트
func TestKubeConfigParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedError  bool
		expectedConfig config.KubeConfig
	}{
		{
			name: "Valid kubeconfig",
			requestBody: map[string]interface{}{
				"kubeconfig": "apiVersion: v1\nkind: Config",
			},
			expectedError: false,
			expectedConfig: config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
			},
		},
		{
			name: "Empty kubeconfig",
			requestBody: map[string]interface{}{
				"kubeconfig": "",
			},
			expectedError: false, // 파싱은 성공하지만 검증에서 실패
			expectedConfig: config.KubeConfig{
				KubeConfig: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			e := echo.New()
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var kubeConfig config.KubeConfig
			parser := NewKubeConfigParser(&kubeConfig)

			// When
			err := parser.Parse(c)

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestKubeConfigParser_Validate : Kubernetes 설정 검증 테스트
func TestKubeConfigParser_Validate(t *testing.T) {
	tests := []struct {
		name          string
		config        config.KubeConfig
		expectedError bool
	}{
		{
			name: "Valid config",
			config: config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
			},
			expectedError: false,
		},
		{
			name: "Empty kubeconfig",
			config: config.KubeConfig{
				KubeConfig: "",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			parser := NewKubeConfigParser(&tt.config)

			// When
			err := parser.Validate()

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestVeleroConfigParser_Parse : Velero 설정 파서 테스트
func TestVeleroConfigParser_Parse(t *testing.T) {
	tests := []struct {
		name          string
		requestBody   map[string]interface{}
		expectedError bool
	}{
		{
			name: "Valid Velero config",
			requestBody: map[string]interface{}{
				"kubeconfig": map[string]interface{}{
					"kubeconfig": "apiVersion: v1\nkind: Config",
				},
				"minio": map[string]interface{}{
					"endpoint":  "localhost:9000",
					"accessKey": "minioadmin",
					"secretKey": "minioadmin123",
					"useSSL":    false,
				},
			},
			expectedError: false,
		},
		{
			name: "Invalid JSON structure",
			requestBody: map[string]interface{}{
				"kubeconfig": map[string]interface{}{
					"kubeconfig": "apiVersion: v1\nkind: Config",
				},
				"minio": "invalid", // minio should be an object
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			e := echo.New()
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			var kubeConfig config.KubeConfig
			var veleroConfig config.VeleroConfig
			var minioConfig config.MinioConfig
			parser := NewVeleroConfigParser(&kubeConfig, &veleroConfig, &minioConfig)

			// When
			err := parser.Parse(c)

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestVeleroConfigParser_Validate : Velero 설정 검증 테스트
func TestVeleroConfigParser_Validate(t *testing.T) {
	tests := []struct {
		name          string
		kubeConfig    config.KubeConfig
		veleroConfig  config.VeleroConfig
		minioConfig   config.MinioConfig
		expectedError bool
	}{
		{
			name: "Valid config",
			kubeConfig: config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
			},
			veleroConfig: config.VeleroConfig{},
			minioConfig: config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: false,
		},
		{
			name: "Empty kubeconfig",
			kubeConfig: config.KubeConfig{
				KubeConfig: "",
			},
			veleroConfig: config.VeleroConfig{},
			minioConfig: config.MinioConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: true,
		},
		{
			name: "Empty minio endpoint",
			kubeConfig: config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
			},
			veleroConfig: config.VeleroConfig{},
			minioConfig: config.MinioConfig{
				Endpoint:  "",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: true,
		},
		{
			name: "Template variables in minio endpoint",
			kubeConfig: config.KubeConfig{
				KubeConfig: "apiVersion: v1\nkind: Config",
			},
			veleroConfig: config.VeleroConfig{},
			minioConfig: config.MinioConfig{
				Endpoint:  "{{minio_url}}",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin123",
				UseSSL:    false,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			parser := NewVeleroConfigParser(&tt.kubeConfig, &tt.veleroConfig, &tt.minioConfig)

			// When
			err := parser.Validate()

			// Then
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestBaseHandler_handleConfigError : 공통 에러 처리 함수 테스트
func TestBaseHandler_handleConfigError(t *testing.T) {
	baseHandler := &BaseHandler{}

	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Template variables error",
			error:          fmt.Errorf("minio endpoint contains template variables"),
			expectedStatus: 400,
			expectedCode:   "INVALID_MINIO_CONFIG",
		},
		{
			name:           "Endpoint required error",
			error:          fmt.Errorf("minio endpoint is required"),
			expectedStatus: 400,
			expectedCode:   "INVALID_MINIO_CONFIG",
		},
		{
			name:           "Invalid request body error",
			error:          fmt.Errorf("invalid request body"),
			expectedStatus: 400,
			expectedCode:   "INVALID_REQUEST_BODY",
		},
		{
			name:           "Invalid kubernetes config error",
			error:          fmt.Errorf("invalid kubernetes configuration"),
			expectedStatus: 400,
			expectedCode:   "INVALID_KUBERNETES_CONFIG",
		},
		{
			name:           "Unsupported API path error",
			error:          fmt.Errorf("unsupported API path"),
			expectedStatus: 400,
			expectedCode:   "UNSUPPORTED_API_PATH",
		},
		{
			name:           "Generic error",
			error:          fmt.Errorf("some other error"),
			expectedStatus: 400,
			expectedCode:   "CONFIG_PARSE_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given - 각 테스트마다 새로운 요청과 응답 생성
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// When
			err := baseHandler.handleConfigError(c, tt.error)

			// Then
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// 응답 상태 코드 확인
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// 응답 본문 확인
			var response map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}

			if code, ok := response["code"].(string); !ok || code != tt.expectedCode {
				t.Errorf("Expected code %s, got %v", tt.expectedCode, response["code"])
			}
		})
	}
}
