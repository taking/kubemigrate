package controller

import (
	"context"
	"net/http"
	"taking.kr/velero/models"
	"taking.kr/velero/validation"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/clients"
)

type KubernetesController struct {
	validator *validation.RequestValidator
}

func NewKubernetesController() *KubernetesController {
	return &KubernetesController{
		validator: validation.NewRequestValidator(),
	}
}

// HealthCheck : Kubernetes 클러스터 연결 확인
func (c *KubernetesController) HealthCheck(ctx echo.Context) error {
	var req models.KubeConfig

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// namespace 설정
	namespace := c.determineNamespace(&req, ctx)

	// kubeconfig 유효성 검사
	decodeKubeConfig, err := c.validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Kubernetes 클라이언트 생성
	client, err := clients.NewKubeClient(models.KubeConfig{
		KubeConfig: decodeKubeConfig,
		Namespace:  namespace,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"Invalid kubeconfig: ": err.Error(),
		})
	}

	if err := client.HealthCheck(context.Background()); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"status":  "healthy",
		"message": "kubernetes cluster connection successful",
	})
}

func (c *KubernetesController) determineNamespace(req *models.KubeConfig, ctx echo.Context) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return "default" // default
}

// GetPods : 네임스페이스 내 모든 Pod 조회
func (c *KubernetesController) GetPods(ctx echo.Context) error {
	var cfg models.KubeConfig
	if err := ctx.Bind(&cfg); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	client, err := clients.NewKubeClient(cfg)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	data, err := client.GetPods(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// GetStorageClasses : 클러스터 내 StorageClass 조회
func (c *KubernetesController) GetStorageClasses(ctx echo.Context) error {
	var cfg models.KubeConfig
	if err := ctx.Bind(&cfg); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	client, err := clients.NewKubeClient(cfg)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	data, err := client.GetStorageClasses(context.Background())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}
