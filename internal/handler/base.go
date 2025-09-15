// Package handler HTTP 요청을 처리하는 핸들러들을 관리합니다.
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/cache"
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
	workerPool          *pkgutils.WorkerPool
	clientCache         *cache.LRUCache
	useMockClient       bool // 테스트용 Mock 클라이언트 사용 여부
}

// NewBaseHandler : 기본 핸들러 생성
func NewBaseHandler(workerPool *pkgutils.WorkerPool) *BaseHandler {
	return &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		workerPool:          workerPool,
		clientCache:         cache.NewLRUCache(100), // 최대 100개 항목
		useMockClient:       false,
	}
}

// NewBaseHandlerWithMock : Mock 클라이언트를 사용하는 핸들러 생성 (테스트용)
func NewBaseHandlerWithMock(workerPool *pkgutils.WorkerPool) *BaseHandler {
	return &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		workerPool:          workerPool,
		clientCache:         cache.NewLRUCache(100), // 최대 100개 항목
		useMockClient:       true,
	}
}

// HandleResourceClient : 통합 클라이언트를 사용한 리소스 처리
func (h *BaseHandler) HandleResourceClient(c echo.Context, cacheKey string,
	getResource func(client.Client, context.Context) (interface{}, error)) error {

	// API 타입별 설정 파싱 및 검증
	kubeConfig, veleroConfig, minioConfig, err := h.parseAndValidateConfig(c, cacheKey)
	if err != nil {
		return err
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
					return client.NewClientWithConfig(nil, nil, nil, minioConfig)
				}

				// Velero API인 경우 Kubernetes + MinIO 조합 클라이언트 생성
				if strings.Contains(c.Request().URL.Path, "/velero/") {
					// Velero는 Kubernetes 클라이언트를 사용하지만, MinIO 설정도 함께 전달
					return client.NewClientWithConfig(kubeConfig, kubeConfig, veleroConfig, minioConfig)
				}

				// 기본 Kubernetes/Helm API
				return client.NewClientWithConfig(kubeConfig, kubeConfig, veleroConfig, minioConfig)
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

// parseAndValidateConfig : API 타입별 설정 파싱 및 검증
func (h *BaseHandler) parseAndValidateConfig(c echo.Context, cacheKey string) (
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

	return kubeConfig, veleroConfig, minioConfig, response.RespondWithErrorModel(c, http.StatusBadRequest,
		"UNSUPPORTED_API_PATH",
		"Unsupported API path",
		fmt.Sprintf("API path not supported: %s", c.Request().URL.Path))
}

// parseKubeConfig : Kubernetes 설정 파싱 및 검증
func (h *BaseHandler) parseKubeConfig(c echo.Context, kubeConfig *config.KubeConfig) error {
	var req struct {
		KubeConfig string `json:"kubeconfig"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid request body",
			err.Error())
	}

	kubeConfig.KubeConfig = req.KubeConfig
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_KUBERNETES_CONFIG",
			"Invalid Kubernetes configuration",
			err.Error())
	}

	return nil
}

// parseMinioConfig : MinIO 설정 파싱 및 검증
func (h *BaseHandler) parseMinioConfig(c echo.Context, minioConfig *config.MinioConfig) error {
	var req struct {
		Endpoint  string `json:"endpoint"`
		AccessKey string `json:"accessKey"`
		SecretKey string `json:"secretKey"`
		UseSSL    bool   `json:"useSSL"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid request body",
			err.Error())
	}

	// MinIO 설정 매핑
	minioConfig.Endpoint = req.Endpoint
	minioConfig.AccessKey = req.AccessKey
	minioConfig.SecretKey = req.SecretKey
	minioConfig.UseSSL = req.UseSSL

	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			err.Error())
	}

	return nil
}

// parseVeleroConfig : Velero 설정 파싱 및 검증
func (h *BaseHandler) parseVeleroConfig(c echo.Context, kubeConfig *config.KubeConfig,
	veleroConfig *config.VeleroConfig, minioConfig *config.MinioConfig) error {

	var req struct {
		KubeConfig struct {
			KubeConfig string `json:"kubeconfig"`
		} `json:"kubeconfig"`
		Minio config.MinioConfig `json:"minio"`
	}

	if err := c.Bind(&req); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_REQUEST_BODY",
			"Invalid request body",
			err.Error())
	}

	// Kubernetes 설정
	kubeConfig.KubeConfig = req.KubeConfig.KubeConfig
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_KUBERNETES_CONFIG",
			"Invalid Kubernetes configuration",
			err.Error())
	}

	// Velero 설정 (Kubernetes 설정과 동일)
	veleroConfig.KubeConfig = *kubeConfig

	// MinIO 설정
	*minioConfig = req.Minio
	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return response.RespondWithErrorModel(c, http.StatusBadRequest,
			"INVALID_MINIO_CONFIG",
			"Invalid MinIO configuration",
			err.Error())
	}

	return nil
}

// GetCacheStats : 캐시 통계 정보 조회
func (h *BaseHandler) GetCacheStats() map[string]interface{} {
	return h.clientCache.Stats()
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
