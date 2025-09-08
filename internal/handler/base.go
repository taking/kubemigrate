package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client"
	"github.com/taking/kubemigrate/pkg/client/helm"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/client/minio"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/utils"
)

// BaseHandler : 모든 핸들러의 기본 구조
type BaseHandler struct {
	KubernetesValidator *validator.KubernetesValidator
	MinioValidator      *validator.MinioValidator
	workerPool          *utils.WorkerPool
}

// NewBaseHandler : 기본 핸들러 생성
func NewBaseHandler(workerPool *utils.WorkerPool) *BaseHandler {
	return &BaseHandler{
		KubernetesValidator: validator.NewKubernetesValidator(),
		MinioValidator:      validator.NewMinioValidator(),
		workerPool:          workerPool,
	}
}

// HandleKubernetesResource : Kubernetes 리소스 처리 공통 로직
func (h *BaseHandler) HandleKubernetesResource(c echo.Context, cacheKey string,
	getResource func(kubernetes.Client, context.Context) (interface{}, error)) error {

	// Kubernetes 클라이언트 생성
	k8sClient := client.NewClient().Kubernetes()

	// 리소스 조회
	resource, err := getResource(k8sClient, c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get %s: %v", cacheKey, err),
		})
	}

	return c.JSON(http.StatusOK, resource)
}

// HandleVeleroResource : Velero 리소스 처리 공통 로직
func (h *BaseHandler) HandleVeleroResource(c echo.Context, cacheKey string,
	getResource func(velero.Client, context.Context) (interface{}, error)) error {

	// Velero 클라이언트 생성
	veleroClient := client.NewClient().Velero()

	// 리소스 조회
	resource, err := getResource(veleroClient, c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get %s: %v", cacheKey, err),
		})
	}

	return c.JSON(http.StatusOK, resource)
}

// HandleHelmResource : Helm 리소스 처리 공통 로직
func (h *BaseHandler) HandleHelmResource(c echo.Context, cacheKey string,
	getResource func(helm.Client, context.Context) (interface{}, error)) error {

	// Helm 클라이언트 생성
	helmClient := client.NewClient().Helm()

	// 리소스 조회
	resource, err := getResource(helmClient, c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get %s: %v", cacheKey, err),
		})
	}

	return c.JSON(http.StatusOK, resource)
}

// HandleMinioResource : Minio 리소스 처리 공통 로직
func (h *BaseHandler) HandleMinioResource(c echo.Context, cacheKey string,
	getResource func(minio.Client, context.Context) (interface{}, error)) error {

	// MinIO 클라이언트 생성
	minioClient := client.NewClient().Minio()

	// 리소스 조회
	resource, err := getResource(minioClient, c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get %s: %v", cacheKey, err),
		})
	}

	return c.JSON(http.StatusOK, resource)
}
