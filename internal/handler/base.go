// Package handler HTTP 요청을 처리하는 핸들러들을 관리합니다.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/cache"
	"github.com/taking/kubemigrate/internal/job"
	"github.com/taking/kubemigrate/internal/logger"
	"github.com/taking/kubemigrate/internal/mocks"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/config"
	"github.com/taking/kubemigrate/pkg/constants"
	pkgutils "github.com/taking/kubemigrate/pkg/utils"
)

// BaseHandler : 모든 핸들러의 기본 구조
type BaseHandler struct {
	KubernetesValidator *validator.KubernetesValidator
	MinioValidator      *validator.MinioValidator
	ConfigManager       *config.ConfigManager
	ValidationManager   *validator.ValidationManager
	ConfigBinder        *pkgutils.ConfigBinder
	workerPool          *job.WorkerPool
	clientCache         *cache.LRUCache
	useMockClient       bool // 테스트용 Mock 클라이언트 사용 여부
}

// NewBaseHandler : 기본 핸들러 생성
func NewBaseHandler(workerPool *job.WorkerPool) *BaseHandler {
	baseHandler := &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		ConfigManager:       config.NewConfigManager(),
		ValidationManager:   validator.NewValidationManager(),
		ConfigBinder:        pkgutils.NewConfigBinder(),
		workerPool:          workerPool,
		clientCache:         cache.NewLRUCache(100), // 최대 100개 항목
		useMockClient:       false,
	}

	// 설정 검증
	if err := baseHandler.ValidateConfiguration(); err != nil {
		// 로그만 출력하고 계속 진행 (개발 환경에서는 유연하게)
		fmt.Printf("Warning: Configuration validation failed: %v\n", err)
	}

	return baseHandler
}

// NewBaseHandlerWithMock : Mock 클라이언트를 사용하는 핸들러 생성 (테스트용)
func NewBaseHandlerWithMock(workerPool *job.WorkerPool) *BaseHandler {
	return &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		ConfigManager:       config.NewConfigManager(),
		ValidationManager:   validator.NewValidationManager(),
		ConfigBinder:        pkgutils.NewConfigBinder(),
		workerPool:          workerPool,
		clientCache:         cache.NewLRUCache(100), // 최대 100개 항목
		useMockClient:       true,
	}
}

// HealthCheckConfig : HealthCheck 설정
type HealthCheckConfig struct {
	ServiceName string
	DefaultNS   string
	HealthFunc  func(client.Client, context.Context) error
}

// HealthCheck : 공통 HealthCheck 함수
func (h *BaseHandler) HealthCheck(c echo.Context, config HealthCheckConfig) error {
	return h.HandleResourceClient(c, config.ServiceName+"-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// 네임스페이스 결정
		namespace := h.ResolveNamespace(c, config.DefaultNS)

		// HealthCheck 함수 실행
		if err := config.HealthFunc(client, ctx); err != nil {
			return nil, err
		}

		// 응답 데이터 구성
		response := map[string]interface{}{
			"service": config.ServiceName,
			"message": config.ServiceName + " connection is working",
		}

		// 네임스페이스가 있는 경우에만 추가
		if namespace != "" {
			response["namespace"] = namespace
		}

		return response, nil
	})
}

// HandleError : 공통 에러 처리 함수
func (h *BaseHandler) HandleError(c echo.Context, config response.ErrorHandlerConfig, err error) error {
	return response.HandleError(c, config, err)
}

// HandleValidationError : 공통 검증 에러 처리 함수
func (h *BaseHandler) HandleValidationError(c echo.Context, serviceName, operation string, err error) error {
	return response.HandleValidationError(c, serviceName, operation, err)
}

// HandleConnectionError : 공통 연결 에러 처리 함수
func (h *BaseHandler) HandleConnectionError(c echo.Context, serviceName, operation string, err error) error {
	return response.HandleConnectionError(c, serviceName, operation, err)
}

// HandleInternalError : 공통 내부 에러 처리 함수
func (h *BaseHandler) HandleInternalError(c echo.Context, serviceName, operation string, err error) error {
	return response.HandleInternalError(c, serviceName, operation, err)
}

// ValidationConfig : 검증 설정
type ValidationConfig struct {
	ServiceName string
	ConfigType  string
}

// ValidateKubeConfig : Kubernetes 설정 검증 및 바인딩
func (h *BaseHandler) ValidateKubeConfig(c echo.Context, serviceName string) (config.KubeConfig, error) {
	return h.ValidationManager.ValidateKubeConfigInternal(c, serviceName)
}

// ValidateMinioConfig : MinIO 설정 검증 및 바인딩
func (h *BaseHandler) ValidateMinioConfig(c echo.Context, serviceName string) (config.MinioConfig, error) {
	return h.ValidationManager.ValidateMinioConfigInternal(c, serviceName)
}

// ValidateVeleroConfig : Velero 설정 검증 및 바인딩
func (h *BaseHandler) ValidateVeleroConfig(c echo.Context, serviceName string) (config.VeleroConfig, error) {
	return h.ValidationManager.ValidateVeleroConfigInternal(c, serviceName)
}

// HandleResourceClient : 통합 클라이언트를 사용한 리소스 처리
func (h *BaseHandler) HandleResourceClient(c echo.Context, cacheKey string,
	getResource func(client.Client, context.Context) (interface{}, error)) error {

	// API 타입별 설정 파싱 및 검증
	kubeConfig, veleroConfig, minioConfig, err := h.parseConfig(c, cacheKey)
	if err != nil {
		// 설정 파싱 실패 시 공통 에러 처리 함수 사용
		return h.handleConfigError(c, err)
	}

	// API 경로를 기반으로 정확한 API 타입 결정
	apiType := h.determineApiTypeFromPath(c.Request().URL.Path)

	// 캐시에서 클라이언트 조회 또는 생성
	var unifiedClient client.Client
	if h.useMockClient {
		// 테스트용 Mock 클라이언트 사용
		unifiedClient = mocks.NewMockClient()
	} else {
		unifiedClient = h.clientCache.GetOrCreateWithApiType(
			kubeConfig,
			kubeConfig,
			veleroConfig,
			minioConfig,
			apiType,
			func() client.Client {
				// MinIO API인 경우 minioConfig만 유효하므로 명시적으로 처리
				if strings.Contains(c.Request().URL.Path, "/minio/") {
					// MinIO 설정이 유효한지 먼저 확인
					if minioConfig.Endpoint == "" {
						logger.Error("MinIO configuration is missing",
							logger.String("path", c.Request().URL.Path),
						)
						return mocks.NewMockClient()
					}

					client, err := client.NewClientWithConfig(nil, nil, nil, minioConfig)
					if err != nil {
						logger.Error("Failed to create MinIO client",
							logger.String("error", err.Error()),
							logger.String("endpoint", minioConfig.Endpoint),
						)
						return mocks.NewMockClient()
					}
					return client
				}

				// Velero API인 경우 Kubernetes + MinIO 조합 클라이언트 생성
				if strings.Contains(c.Request().URL.Path, "/velero/") {
					// Velero는 Kubernetes 클라이언트를 사용하지만, MinIO 설정도 함께 전달
					client, err := client.NewClientWithConfig(kubeConfig, kubeConfig, veleroConfig, minioConfig)
					if err != nil {
						logger.Error("Failed to create Velero client",
							logger.String("error", err.Error()),
						)
						return mocks.NewMockClient()
					}
					return client
				}

				// 기본 Kubernetes/Helm API
				client, err := client.NewClientWithConfig(kubeConfig, kubeConfig, veleroConfig, minioConfig)
				if err != nil {
					logger.Error("Failed to create client",
						logger.String("error", err.Error()),
					)
					return mocks.NewMockClient()
				}
				return client
			},
		)
	}

	// 리소스 조회 (타임아웃 설정)
	ctx, cancel := context.WithTimeout(c.Request().Context(), constants.DefaultRequestTimeout)
	defer cancel()

	resource, err := getResource(unifiedClient, ctx)
	if err != nil {
		// 에러 타입에 따른 상태 코드 결정
		statusCode := http.StatusInternalServerError
		errorCode := "RESOURCE_FETCH_FAILED"
		message := fmt.Sprintf("Failed to get %s", cacheKey)

		// 에러 메시지에 따른 상태 코드 조정
		if strings.Contains(err.Error(), "unsupported resource kind") {
			statusCode = http.StatusBadRequest
			errorCode = "UNSUPPORTED_RESOURCE"
			message = "Unsupported resource kind"
		}

		// 구조화된 로깅 사용
		logger.Error("Resource fetch failed",
			logger.String("cache_key", cacheKey),
			logger.String("error", err.Error()),
			logger.Int("status_code", statusCode),
		)

		return response.RespondWithErrorModel(c, statusCode, errorCode, message, err.Error())
	}

	return response.RespondWithData(c, http.StatusOK, resource)
}

// parseConfig : API 타입별 설정 파싱
func (h *BaseHandler) parseConfig(c echo.Context, cacheKey string) (
	config.KubeConfig, config.VeleroConfig, config.MinioConfig, error) {

	var kubeConfig config.KubeConfig
	var veleroConfig config.VeleroConfig
	var minioConfig config.MinioConfig

	// MinIO API인지 확인
	isMinioAPI := strings.Contains(c.Request().URL.Path, "/minio/")

	if isMinioAPI {
		// MinIO API: minio 설정만 필요
		return kubeConfig, veleroConfig, minioConfig, h.parseMinioConfig(c, &minioConfig)
	}

	// Kubernetes/Helm API: kubeconfig만 필요
	if strings.Contains(c.Request().URL.Path, "/kubernetes/") || strings.Contains(c.Request().URL.Path, "/helm/") {
		return kubeConfig, veleroConfig, minioConfig, h.parseKubeConfig(c, &kubeConfig)
	}

	// Velero API: kubeconfig와 minio 설정 필요
	if strings.Contains(c.Request().URL.Path, "/velero/") {
		return kubeConfig, veleroConfig, minioConfig, h.parseVeleroConfig(c, &kubeConfig, &veleroConfig, &minioConfig)
	}

	return kubeConfig, veleroConfig, minioConfig, fmt.Errorf("unsupported API path: %s", c.Request().URL.Path)
}

// parseKubeConfig : Kubernetes 설정 파싱 및 검증 (통합 파서 사용)
func (h *BaseHandler) parseKubeConfig(c echo.Context, kubeConfig *config.KubeConfig) error {
	parser := NewKubeConfigParser(kubeConfig)

	// 파싱
	if err := parser.Parse(c); err != nil {
		return err
	}

	// 검증
	if err := parser.Validate(); err != nil {
		return err
	}

	// 추가 검증 (KubernetesValidator 사용)
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return fmt.Errorf("invalid kubernetes configuration: %w", err)
	}

	return nil
}

// parseMinioConfig : MinIO 설정 파싱 및 검증 (통합 파서 사용)
func (h *BaseHandler) parseMinioConfig(c echo.Context, minioConfig *config.MinioConfig) error {
	parser := NewMinioConfigParser(minioConfig)

	// 파싱
	if err := parser.Parse(c); err != nil {
		return err
	}

	// 검증
	if err := parser.Validate(); err != nil {
		return err
	}

	// 추가 검증 (MinioValidator 사용)
	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return fmt.Errorf("invalid minio configuration: %w", err)
	}

	return nil
}

// parseVeleroConfig : Velero 설정 파싱 및 검증 (통합 파서 사용)
func (h *BaseHandler) parseVeleroConfig(c echo.Context, kubeConfig *config.KubeConfig,
	veleroConfig *config.VeleroConfig, minioConfig *config.MinioConfig) error {

	parser := NewVeleroConfigParser(kubeConfig, veleroConfig, minioConfig)

	// 파싱
	if err := parser.Parse(c); err != nil {
		return err
	}

	// 검증
	if err := parser.Validate(); err != nil {
		return err
	}

	// 추가 검증 (KubernetesValidator 사용)
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return fmt.Errorf("invalid kubernetes configuration: %w", err)
	}

	// 추가 검증 (MinioValidator 사용)
	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return fmt.Errorf("invalid minio configuration: %w", err)
	}

	return nil
}

// GetCacheStats : 캐시 통계 정보 조회
func (h *BaseHandler) GetCacheStats() map[string]interface{} {
	return h.clientCache.Stats()
}

// ===== 공통 설정 관리 함수들 =====

// ConfigParser : 설정 파서 인터페이스
type ConfigParser interface {
	Parse(c echo.Context) error
	Validate() error
}

// MinioConfigParser : MinIO 설정 파서
type MinioConfigParser struct {
	config *config.MinioConfig
}

// NewMinioConfigParser : MinIO 설정 파서 생성
func NewMinioConfigParser(config *config.MinioConfig) *MinioConfigParser {
	return &MinioConfigParser{config: config}
}

// Parse : MinIO 설정 파싱
func (p *MinioConfigParser) Parse(c echo.Context) error {
	var req struct {
		Endpoint  string `json:"endpoint"`
		AccessKey string `json:"accessKey"`
		SecretKey string `json:"secretKey"`
		UseSSL    bool   `json:"useSSL"`
	}

	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// MinIO 설정 매핑
	p.config.Endpoint = req.Endpoint
	p.config.AccessKey = req.AccessKey
	p.config.SecretKey = req.SecretKey
	p.config.UseSSL = req.UseSSL

	return nil
}

// Validate : MinIO 설정 검증
func (p *MinioConfigParser) Validate() error {
	// 설정 검증 전에 기본값 체크
	if p.config.Endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}

	// 템플릿 변수가 포함된 경우 명확한 에러 메시지 제공
	if strings.Contains(p.config.Endpoint, "{{") || strings.Contains(p.config.Endpoint, "}}") {
		return fmt.Errorf("minio endpoint contains template variables. Please provide actual endpoint URL (e.g., 'localhost:9000' or 'minio.example.com:9000')")
	}

	return nil
}

// KubeConfigParser : Kubernetes 설정 파서
type KubeConfigParser struct {
	config *config.KubeConfig
}

// NewKubeConfigParser : Kubernetes 설정 파서 생성
func NewKubeConfigParser(config *config.KubeConfig) *KubeConfigParser {
	return &KubeConfigParser{config: config}
}

// Parse : Kubernetes 설정 파싱
func (p *KubeConfigParser) Parse(c echo.Context) error {
	var req struct {
		KubeConfig string `json:"kubeconfig"`
	}

	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	p.config.KubeConfig = req.KubeConfig
	return nil
}

// Validate : Kubernetes 설정 검증
func (p *KubeConfigParser) Validate() error {
	if p.config.KubeConfig == "" {
		return fmt.Errorf("kubeconfig is required")
	}
	return nil
}

// VeleroConfigParser : Velero 설정 파서
type VeleroConfigParser struct {
	kubeConfig   *config.KubeConfig
	veleroConfig *config.VeleroConfig
	minioConfig  *config.MinioConfig
}

// NewVeleroConfigParser : Velero 설정 파서 생성
func NewVeleroConfigParser(kubeConfig *config.KubeConfig, veleroConfig *config.VeleroConfig, minioConfig *config.MinioConfig) *VeleroConfigParser {
	return &VeleroConfigParser{
		kubeConfig:   kubeConfig,
		veleroConfig: veleroConfig,
		minioConfig:  minioConfig,
	}
}

// Parse : Velero 설정 파싱
func (p *VeleroConfigParser) Parse(c echo.Context) error {
	var req struct {
		KubeConfig struct {
			KubeConfig string `json:"kubeconfig"`
		} `json:"kubeconfig"`
		Minio config.MinioConfig `json:"minio"`
	}

	if err := c.Bind(&req); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// Kubernetes 설정
	p.kubeConfig.KubeConfig = req.KubeConfig.KubeConfig
	p.veleroConfig.KubeConfig = *p.kubeConfig

	// MinIO 설정
	*p.minioConfig = req.Minio

	return nil
}

// Validate : Velero 설정 검증
func (p *VeleroConfigParser) Validate() error {
	// Kubernetes 설정 검증
	if p.kubeConfig.KubeConfig == "" {
		return fmt.Errorf("kubeconfig is required")
	}

	// MinIO 설정 검증
	if p.minioConfig.Endpoint == "" {
		return fmt.Errorf("minio endpoint is required")
	}

	// 템플릿 변수가 포함된 경우 명확한 에러 메시지 제공
	if strings.Contains(p.minioConfig.Endpoint, "{{") || strings.Contains(p.minioConfig.Endpoint, "}}") {
		return fmt.Errorf("minio endpoint contains template variables. Please provide actual endpoint URL (e.g., 'localhost:9000' or 'minio.example.com:9000')")
	}

	return nil
}

// ===== 공통 에러 처리 함수들 =====

// handleConfigError : 설정 파싱 에러를 적절한 HTTP 응답으로 변환
func (h *BaseHandler) handleConfigError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	errorMsg := err.Error()

	// 에러 타입별 처리
	if strings.Contains(errorMsg, "template variables") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			errorMsg)
	}
	if strings.Contains(errorMsg, "endpoint is required") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			errorMsg)
	}
	if strings.Contains(errorMsg, "invalid request body") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid request body",
			errorMsg)
	}
	if strings.Contains(errorMsg, "invalid minio configuration") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			errorMsg)
	}
	if strings.Contains(errorMsg, "invalid kubernetes configuration") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_KUBERNETES_CONFIG",
			"Invalid Kubernetes configuration",
			errorMsg)
	}
	if strings.Contains(errorMsg, "unsupported API path") {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"UNSUPPORTED_API_PATH",
			"Unsupported API path",
			errorMsg)
	}

	// 기타 에러
	return response.RespondWithErrorModel(c, http.StatusBadRequest,
		"CONFIG_PARSE_ERROR",
		"Configuration parsing failed",
		errorMsg)
}

// ===== ConfigManager 활용 메서드들 =====

// GetServerConfig : 서버 설정 조회
func (h *BaseHandler) GetServerConfig() config.ServerConfig {
	return h.ConfigManager.GetServerConfig()
}

// GetTimeoutConfig : 타임아웃 설정 조회
func (h *BaseHandler) GetTimeoutConfig() config.TimeoutConfig {
	return h.ConfigManager.GetTimeoutConfig()
}

// GetLoggingConfig : 로깅 설정 조회
func (h *BaseHandler) GetLoggingConfig() config.LoggingConfig {
	return h.ConfigManager.GetLoggingConfig()
}

// ValidateConfiguration : 설정 검증
func (h *BaseHandler) ValidateConfiguration() error {
	return h.ConfigManager.ValidateConfig()
}

// ReloadConfiguration : 설정 재로드
func (h *BaseHandler) ReloadConfiguration() error {
	return h.ConfigManager.Reload()
}

// GetConfigValue : 환경변수 기반 설정 값 조회
func (h *BaseHandler) GetConfigValue(key, defaultValue string) string {
	return config.GetEnvOrDefault(key, defaultValue)
}

// GetConfigDuration : 환경변수 기반 Duration 설정 값 조회
func (h *BaseHandler) GetConfigDuration(key string, defaultValue time.Duration) time.Duration {
	return config.GetDurationOrDefault(key, defaultValue)
}

// GetConfigInt : 환경변수 기반 int 설정 값 조회
func (h *BaseHandler) GetConfigInt(key string, defaultValue int) int {
	return config.GetIntOrDefault(key, defaultValue)
}

// GetConfigBool : 환경변수 기반 bool 설정 값 조회
func (h *BaseHandler) GetConfigBool(key string, defaultValue bool) bool {
	return config.GetBoolOrDefault(key, defaultValue)
}

// ===== Query Parameter 처리 함수들 =====

// ResolveNamespace : 네임스페이스 쿼리 파라미터 결정
func (h *BaseHandler) ResolveNamespace(ctx echo.Context, defaultNS string) string {
	var namespace string

	if ns := ctx.QueryParam("namespace"); ns != "" {
		namespace = ns
	} else {
		return defaultNS
	}

	// "all"을 빈 문자열로 변환 (모든 namespace 조회)
	if namespace == "all" {
		return ""
	}

	return namespace
}

// ResolveBool : boolean 쿼리 파라미터 결정
func (h *BaseHandler) ResolveBool(c echo.Context, param string, defaultValue bool) bool {
	value := c.QueryParam(param)
	if value == "" {
		return defaultValue
	}
	return h.StringToBoolOrDefault(value, defaultValue)
}

// ResolveInt : integer 쿼리 파라미터 결정
func (h *BaseHandler) ResolveInt(c echo.Context, param string, defaultValue int) int {
	value := c.QueryParam(param)
	if value == "" {
		return defaultValue
	}
	return h.StringToIntOrDefault(value, defaultValue)
}

// StringToBoolOrDefault : string을 bool로 변환, 실패하면 기본값 반환
func (h *BaseHandler) StringToBoolOrDefault(s string, def bool) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return def
	}
	return b
}

// StringToIntOrDefault : string을 int로 변환, 실패하면 기본값 반환
func (h *BaseHandler) StringToIntOrDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// GetDetailedCacheStats : 상세한 캐시 통계 정보 조회
func (h *BaseHandler) GetDetailedCacheStats() interface{} {
	return h.clientCache.GetDetailedStats()
}

// CleanupCache : 전체 캐시 정리
func (h *BaseHandler) CleanupCache() {
	h.clientCache.Cleanup()
}

// CleanCacheByKey : 특정 키의 캐시 정리
func (h *BaseHandler) CleanCacheByKey(key string) bool {
	return h.clientCache.CleanByKey(key)
}

// CleanCacheByPattern : 패턴에 맞는 캐시 정리
func (h *BaseHandler) CleanCacheByPattern(pattern string) int {
	return h.clientCache.CleanByPattern(pattern)
}

// determineApiTypeFromPath API 경로를 기반으로 정확한 API 타입을 결정합니다.
func (h *BaseHandler) determineApiTypeFromPath(path string) string {
	if strings.Contains(path, "/minio/") {
		return "minio"
	} else if strings.Contains(path, "/velero/") {
		return "velero"
	} else if strings.Contains(path, "/helm/") {
		return "helm"
	} else if strings.Contains(path, "/kubernetes/") {
		return "kubernetes"
	}

	// 기본값 (대부분의 경우 Kubernetes)
	return "kubernetes"
}
