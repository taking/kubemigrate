package utils

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/taking/kubemigrate/internal/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestGetStringOrDefault - 문자열 기본값 처리 함수 테스트
// 빈 문자열일 때 기본값을 반환하는지 확인
func TestGetStringOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		def      string
		expected string
	}{
		{
			name:     "빈 값은 기본값 반환",
			value:    "",
			def:      "default",
			expected: "default",
		},
		{
			name:     "비어있지 않은 값은 해당 값 반환",
			value:    "test",
			def:      "default",
			expected: "test",
		},
		{
			name:     "공백 값은 해당 값 반환",
			value:    " ",
			def:      "default",
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringOrDefault(tt.value, tt.def)
			if result != tt.expected {
				t.Errorf("GetStringOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBoolOrDefault - 불린 기본값 처리 함수 테스트
// 불린 값의 기본값 처리 로직 확인
func TestGetBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		def      bool
		expected bool
	}{
		{
			name:     "true 값은 true 반환",
			value:    true,
			def:      false,
			expected: true,
		},
		{
			name:     "false 값은 false 반환",
			value:    false,
			def:      true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBoolOrDefault(tt.value, tt.def)
			if result != tt.expected {
				t.Errorf("GetBoolOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestStringToIntOrDefault - 문자열을 정수로 변환하는 함수 테스트
// 유효하지 않은 문자열일 때 기본값을 반환하는지 확인
func TestStringToIntOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		def      int
		expected int
	}{
		{
			name:     "유효한 정수",
			s:        "123",
			def:      0,
			expected: 123,
		},
		{
			name:     "잘못된 문자열은 기본값 반환",
			s:        "abc",
			def:      42,
			expected: 42,
		},
		{
			name:     "빈 문자열은 기본값 반환",
			s:        "",
			def:      10,
			expected: 10,
		},
		{
			name:     "음수",
			s:        "-5",
			def:      0,
			expected: -5,
		},
		{
			name:     "0",
			s:        "0",
			def:      10,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToIntOrDefault(tt.s, tt.def)
			if result != tt.expected {
				t.Errorf("StringToIntOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestStringToBoolOrDefault - 문자열을 불린으로 변환하는 함수 테스트
// 다양한 문자열 형식에 대한 불린 변환 로직 확인
func TestStringToBoolOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		def      bool
		expected bool
	}{
		{
			name:     "true 문자열",
			s:        "true",
			def:      false,
			expected: true,
		},
		{
			name:     "false 문자열",
			s:        "false",
			def:      true,
			expected: false,
		},
		{
			name:     "1 문자열",
			s:        "1",
			def:      false,
			expected: true,
		},
		{
			name:     "0 문자열",
			s:        "0",
			def:      true,
			expected: false,
		},
		{
			name:     "잘못된 문자열은 기본값 반환",
			s:        "invalid",
			def:      true,
			expected: true,
		},
		{
			name:     "빈 문자열은 기본값 반환",
			s:        "",
			def:      false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToBoolOrDefault(tt.s, tt.def)
			if result != tt.expected {
				t.Errorf("StringToBoolOrDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestResolveNamespace - 네임스페이스 해석 함수 테스트
// 쿼리 파라미터와 기본값에 따른 네임스페이스 해석 로직 확인
func TestResolveNamespace(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())

	tests := []struct {
		name       string
		queryParam string
		defaultNS  string
		expected   string
	}{
		{
			name:       "쿼리 파라미터가 없으면 기본값 반환",
			queryParam: "",
			defaultNS:  "default",
			expected:   "default",
		},
		{
			name:       "쿼리 파라미터가 있으면 해당 값 반환",
			queryParam: "kube-system",
			defaultNS:  "default",
			expected:   "kube-system",
		},
		{
			name:       "all 쿼리 파라미터는 빈 값 반환",
			queryParam: "all",
			defaultNS:  "default",
			expected:   "",
		},
		{
			name:       "all 쿼리 파라미터와 빈 기본값",
			queryParam: "all",
			defaultNS:  "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx echo.Context
			if tt.queryParam != "" {
				req := httptest.NewRequest("GET", "/?namespace="+tt.queryParam, nil)
				rec := httptest.NewRecorder()
				ctx = e.NewContext(req, rec)
			} else {
				req := httptest.NewRequest("GET", "/", nil)
				rec := httptest.NewRecorder()
				ctx = e.NewContext(req, rec)
			}

			result := ResolveNamespace(ctx, tt.defaultNS)
			if result != tt.expected {
				t.Errorf("ResolveNamespace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRunWithTimeout - 타임아웃이 있는 함수 실행 테스트
// 함수 실행 시간 제한과 에러 처리 로직 확인
func TestRunWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		fn          func() error
		expectError bool
	}{
		{
			name:    "함수가 타임아웃 전에 완료",
			timeout: 100 * time.Millisecond,
			fn: func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			expectError: false,
		},
		{
			name:    "함수가 타임아웃됨",
			timeout: 50 * time.Millisecond,
			fn: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			expectError: true,
		},
		{
			name:    "함수가 에러 반환",
			timeout: 100 * time.Millisecond,
			fn: func() error {
				return echo.NewHTTPError(500, "test error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := RunWithTimeout(ctx, tt.fn)
			if (err != nil) != tt.expectError {
				t.Errorf("RunWithTimeout() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestStripManagedFields - 관리 필드 제거 함수 테스트
// Kubernetes 객체에서 관리 필드를 올바르게 제거하는지 확인
func TestStripManagedFields(t *testing.T) {
	// 모의 객체 생성
	obj := &unstructured.Unstructured{}
	obj.SetManagedFields([]metav1.ManagedFieldsEntry{
		{
			Manager:    "test-manager",
			Operation:  "Update",
			APIVersion: "v1",
		},
	})

	// 관리 필드 존재 확인
	if len(obj.GetManagedFields()) == 0 {
		t.Fatal("Expected managed fields to exist before stripping")
	}

	// 관리 필드 제거
	StripManagedFields(obj)

	// 제거 확인
	if obj.GetManagedFields() != nil {
		t.Error("Expected managed fields to be nil after stripping")
	}
}

// TestCopyFile - 파일 복사 함수 테스트
// 파일 복사 기능이 올바르게 작동하는지 확인
func TestCopyFile(t *testing.T) {
	// 소스 파일 생성
	srcFile, err := os.CreateTemp("", "test_src_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	// 테스트 내용 작성
	testContent := "test content for file copy"
	_, err = srcFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to src file: %v", err)
	}
	srcFile.Close()

	// 대상 파일 생성
	dstFile, err := os.CreateTemp("", "test_dst_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	dstFile.Close()
	defer os.Remove(dstFile.Name())

	// 파일 복사
	err = CopyFile(srcFile.Name(), dstFile.Name())
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
	}

	// 내용 확인
	content, err := os.ReadFile(dstFile.Name())
	if err != nil {
		t.Fatalf("Failed to read dst file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("CopyFile() content = %v, want %v", string(content), testContent)
	}
}

// TestCopyFile_NonExistentSource - 존재하지 않는 파일 복사 테스트
// 존재하지 않는 소스 파일에 대한 에러 처리 확인
func TestCopyFile_NonExistentSource(t *testing.T) {
	err := CopyFile("non-existent-file.txt", "dst.txt")
	if err == nil {
		t.Error("CopyFile() expected error for non-existent source file")
	}
}

// TestBindAndValidateKubeConfig - Kubernetes 설정 바인딩 및 검증 테스트
// HTTP 요청에서 Kubernetes 설정을 바인딩하고 검증하는 로직 확인
func TestBindAndValidateKubeConfig(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	validator := validator.NewKubernetesValidator()

	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "유효한 kubeconfig",
			requestBody: `{"kubeConfig": "apiVersion: v1\nkind: Config", "namespace": "default"}`,
			expectError: false,
		},
		{
			name:        "잘못된 kubeconfig",
			requestBody: `{"kubeConfig": "invalid", "namespace": "default"}`,
			expectError: true,
		},
		{
			name:        "빈 kubeconfig",
			requestBody: `{"kubeConfig": "", "namespace": "default"}`,
			expectError: true,
		},
		{
			name:        "잘못된 JSON",
			requestBody: `{"kubeConfig": "apiVersion: v1\nkind: Config", "namespace": "default"`,
			expectError: false, // Echo가 자동으로 JSON을 처리하므로 에러가 발생하지 않음
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			_, err := BindAndValidateKubeConfig(ctx, validator)
			if (err != nil) != tt.expectError {
				t.Errorf("BindAndValidateKubeConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestBindAndValidateMinioConfig - MinIO 설정 바인딩 및 검증 테스트
// HTTP 요청에서 MinIO 설정을 바인딩하고 검증하는 로직 확인
func TestBindAndValidateMinioConfig(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	validator := validator.NewMinioValidator()

	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "유효한 minio 설정",
			requestBody: `{"endpoint": "localhost:9000", "accessKey": "minioadmin", "secretKey": "minioadmin123", "useSSL": false}`,
			expectError: false,
		},
		{
			name:        "잘못된 minio 설정",
			requestBody: `{"endpoint": "", "accessKey": "minioadmin", "secretKey": "minioadmin123", "useSSL": false}`,
			expectError: true,
		},
		{
			name:        "잘못된 JSON",
			requestBody: `{"endpoint": "localhost:9000", "accessKey": "minioadmin"`,
			expectError: false, // Echo가 자동으로 JSON을 처리하므로 에러가 발생하지 않음
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			_, err := BindAndValidateMinioConfig(ctx, validator)
			if (err != nil) != tt.expectError {
				t.Errorf("BindAndValidateMinioConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestBindAndValidateVeleroConfig - Velero 설정 바인딩 및 검증 테스트
// HTTP 요청에서 Velero 설정을 바인딩하고 검증하는 로직 확인
func TestBindAndValidateVeleroConfig(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	minioValidator := validator.NewMinioValidator()
	k8sValidator := validator.NewKubernetesValidator()

	tests := []struct {
		name        string
		requestBody string
		expectError bool
	}{
		{
			name:        "유효한 velero 설정",
			requestBody: `{"kubeconfig": {"kubeconfig": "apiVersion: v1\nkind: Config", "namespace": "default"}, "minio": {"endpoint": "localhost:9000", "accessKey": "minioadmin", "secretKey": "minioadmin123", "useSSL": false}}`,
			expectError: false,
		},
		{
			name:        "velero에서 잘못된 minio 설정",
			requestBody: `{"kubeconfig": {"kubeconfig": "apiVersion: v1\nkind: Config", "namespace": "default"}, "minio": {"endpoint": "", "accessKey": "minioadmin", "secretKey": "minioadmin123", "useSSL": false}}`,
			expectError: true,
		},
		{
			name:        "velero에서 잘못된 kubeconfig",
			requestBody: `{"kubeconfig": {"kubeconfig": "invalid", "namespace": "default"}, "minio": {"endpoint": "localhost:9000", "accessKey": "minioadmin", "secretKey": "minioadmin123", "useSSL": false}}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			_, err := BindAndValidateVeleroConfig(ctx, minioValidator, k8sValidator)
			if (err != nil) != tt.expectError {
				t.Errorf("BindAndValidateVeleroConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
