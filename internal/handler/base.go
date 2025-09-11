package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/errors"
	"github.com/taking/kubemigrate/pkg/plugin/cache"
	pkgutils "github.com/taking/kubemigrate/pkg/utils"
)

// BaseHandler : 모든 핸들러의 기본 구조
type BaseHandler struct {
	KubernetesValidator *validator.KubernetesValidator
	MinioValidator      *validator.MinioValidator
	workerPool          *pkgutils.WorkerPool
	cacheManager        *cache.Manager
}

// NewBaseHandler : 기본 핸들러 생성
func NewBaseHandler(workerPool *pkgutils.WorkerPool) *BaseHandler {
	return &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		workerPool:          workerPool,
		cacheManager:        cache.NewManager(5 * time.Minute), // 5분 TTL
	}
}

// HandleResourceClient : 통합 클라이언트를 사용한 리소스 처리
func (h *BaseHandler) HandleResourceClient(c echo.Context, cacheKey string,
	getResource func(client.Client, context.Context) (interface{}, error)) error {

	// API 타입별 설정 파싱 및 검증
	kubeConfig, veleroConfig, minioConfig, err := h.parseAndValidateConfig(c)
	if err != nil {
		return err
	}

	// API 타입 감지
	apiType := h.detectApiType(c.Request().URL.Path)

	// 캐시에서 클라이언트 조회 또는 생성
	configMap := map[string]interface{}{
		"kubeconfig":       kubeConfig.Config,
		"minio_endpoint":   minioConfig.Endpoint,
		"minio_access_key": minioConfig.AccessKey,
		"minio_secret_key": minioConfig.SecretKey,
		"minio_use_ssl":    minioConfig.UseSSL,
	}

	unifiedClient, err := h.cacheManager.GetCachedClient(apiType, configMap)
	if err != nil {
		// 캐시 오류 시 새 클라이언트 생성
		if strings.Contains(c.Request().URL.Path, "/minio/") {
			unifiedClient = client.NewClientWithConfig(nil, nil, nil, minioConfig)
		} else {
			unifiedClient = client.NewClientWithConfig(kubeConfig, kubeConfig, veleroConfig, minioConfig)
		}
	}

	// 리소스 조회 (타임아웃 설정)
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	resource, err := getResource(unifiedClient, ctx)
	if err != nil {
		// 디버깅을 위한 로그 추가
		fmt.Printf("DEBUG: Resource fetch failed for %s: %v\n", cacheKey, err)
		fmt.Printf("DEBUG: Unified client: %+v\n", unifiedClient)

		// 공통 에러 처리 패키지 사용
		return errors.NewInternalError(cacheKey, err)
	}

	return response.RespondWithData(c, http.StatusOK, resource)
}

// parseAndValidateConfig : API 타입별 설정 파싱 및 검증
func (h *BaseHandler) parseAndValidateConfig(c echo.Context) (
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

	return kubeConfig, veleroConfig, minioConfig, errors.NewValidationError(errors.CodeUnsupportedPath, "Unsupported API path", fmt.Sprintf("API path not supported: %s", c.Request().URL.Path))
}

// parseKubeConfig : Kubernetes 설정 파싱 및 검증
func (h *BaseHandler) parseKubeConfig(c echo.Context, kubeConfig *config.KubeConfig) error {
	var req struct {
		KubeConfig string `json:"kubeconfig"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	kubeConfig.Config = req.KubeConfig
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return errors.NewValidationError(errors.CodeInvalidConfiguration, "Invalid Kubernetes configuration", err.Error())
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
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// MinIO 설정 매핑
	minioConfig.Endpoint = req.Endpoint
	minioConfig.AccessKey = req.AccessKey
	minioConfig.SecretKey = req.SecretKey
	minioConfig.UseSSL = req.UseSSL

	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return errors.NewValidationError(errors.CodeInvalidConfiguration, "Invalid MinIO configuration", err.Error())
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
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// Kubernetes 설정
	kubeConfig.Config = req.KubeConfig.KubeConfig
	if _, err := h.KubernetesValidator.ValidateKubernetesConfig(kubeConfig); err != nil {
		return errors.NewValidationError(errors.CodeInvalidConfiguration, "Invalid Kubernetes configuration", err.Error())
	}

	// Velero 설정 (Kubernetes 설정과 동일)
	veleroConfig.KubeConfig = *kubeConfig

	// MinIO 설정
	*minioConfig = req.Minio
	if err := h.MinioValidator.ValidateMinioConfig(minioConfig); err != nil {
		return errors.NewValidationError(errors.CodeInvalidConfiguration, "Invalid MinIO configuration", err.Error())
	}

	return nil
}

// GetCacheStats : 캐시 통계 정보 조회
func (h *BaseHandler) GetCacheStats() map[string]interface{} {
	return h.cacheManager.GetStats()
}

// CleanupCache : 만료된 캐시 정리
func (h *BaseHandler) CleanupCache() {
	h.cacheManager.Cleanup()
}

// InvalidateCache : 특정 설정의 캐시 무효화
func (h *BaseHandler) InvalidateCache(apiType string, kubeConfig config.KubeConfig, helmConfig config.KubeConfig, veleroConfig config.VeleroConfig, minioConfig config.MinioConfig) {
	configMap := map[string]interface{}{
		"kubeconfig":       kubeConfig.Config,
		"minio_endpoint":   minioConfig.Endpoint,
		"minio_access_key": minioConfig.AccessKey,
		"minio_secret_key": minioConfig.SecretKey,
		"minio_use_ssl":    minioConfig.UseSSL,
	}
	h.cacheManager.Invalidate(apiType, configMap)
}

// InvalidateAllCache : 모든 캐시 무효화
func (h *BaseHandler) InvalidateAllCache() {
	h.cacheManager.InvalidateAll()
}

// detectApiType : API 타입 감지
func (h *BaseHandler) detectApiType(path string) string {
	if strings.Contains(path, "/minio/") {
		return "minio"
	} else if strings.Contains(path, "/helm/") {
		return "helm"
	} else if strings.Contains(path, "/velero/") {
		return "velero"
	} else if strings.Contains(path, "/kubernetes/") {
		return "kubernetes"
	}
	return "unknown"
}
