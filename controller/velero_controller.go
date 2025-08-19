package controller

import (
	"context"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/interfaces"
	"taking.kr/velero/models"
	"taking.kr/velero/utils"
)

type VeleroBodyController struct{}

func NewVeleroBodyController() *VeleroBodyController {
	return &VeleroBodyController{}
}

// getVeleroServiceFromRequest는 요청에서 kubeconfig를 추출하여 서비스를 생성합니다
func (c *VeleroBodyController) getVeleroServiceFromRequest(ctx echo.Context) (interfaces.VeleroService, string, error) {
	var req models.VeleroRequest

	if err := ctx.Bind(&req); err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	// kubeconfig 유효성 검사
	if req.SourceKubeconfig == "" {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "sourceKubeconfig is required")
	}

	// namespace 설정
	namespace := req.Namespace
	if namespace == "" {
		namespace = ctx.QueryParam("namespace")
		if namespace == "" {
			namespace = "velero"
		}
	}

	decodeKubeconfig, err := utils.DecodeIfBase64(req.SourceKubeconfig)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Velero 클라이언트 생성
	service, err := clients.NewVeleroClientFromRawConfig(decodeKubeconfig)
	if err != nil {
		return nil, "", echo.NewHTTPError(http.StatusBadRequest, "Invalid kubeconfig: "+err.Error())
	}

	return service, namespace, nil
}

// 공통 요청 처리 함수
func (c *VeleroBodyController) handleRequest(
	ctx echo.Context,
	handler func(context.Context, interfaces.VeleroService, string) (interface{}, error),
) error {
	service, namespace, err := c.getVeleroServiceFromRequest(ctx)
	if err != nil {
		return err
	}

	result, err := handler(ctx.Request().Context(), service, namespace)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"data":    result,
		"success": true,
	})
}

func (c *VeleroBodyController) GetBackups(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetBackups(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetRestores(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetRestores(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetBackupRepositories(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetBackupRepositories(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetBackupStorageLocations(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetBackupStorageLocations(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetVolumeSnapshotLocations(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetVolumeSnapshotLocations(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetPodVolumeRestores(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetPodVolumeRestores(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDownloadRequests(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDownloadRequests(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDataUploads(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDataUploads(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetDataDownloads(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetDataDownloads(reqCtx, namespace)
	})
}

func (c *VeleroBodyController) GetServerStatusRequests(ctx echo.Context) error {
	return c.handleRequest(ctx, func(reqCtx context.Context, service interfaces.VeleroService, namespace string) (interface{}, error) {
		return service.GetServerStatusRequests(reqCtx, namespace)
	})
}

// 대상 클러스터 검증
func (c *VeleroBodyController) ValidateDestination(ctx echo.Context) error {
	var req models.VeleroRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.DestinationKubeconfig == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "destinationKubeconfig is required")
	}

	// 대상 클러스터 서비스 생성
	destService, err := clients.NewVeleroClientFromRawConfig(req.DestinationKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid destination kubeconfig: "+err.Error())
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "velero"
	}

	// Velero 설치 상태 확인
	_, err = destService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"valid":   false,
			"error":   err.Error(),
			"message": "Destination cluster validation failed",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"valid":   true,
		"message": "Destination cluster is ready for Velero operations",
	})
}

// 스토리지 클래스 비교
func (c *VeleroBodyController) CompareStorageClasses(ctx echo.Context) error {
	var req models.VeleroRequest
	if err := ctx.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if req.SourceKubeconfig == "" || req.DestinationKubeconfig == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Both source and destination kubeconfigs are required")
	}

	// 소스와 대상 클러스터의 스토리지 클래스를 가져오는 로직
	// 여기서는 간단한 예시로 구현

	sourceService, err := clients.NewVeleroClientFromRawConfig(req.SourceKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid source kubeconfig: "+err.Error())
	}

	destService, err := clients.NewVeleroClientFromRawConfig(req.DestinationKubeconfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid destination kubeconfig: "+err.Error())
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "velero"
	}

	// 간단한 비교 로직 (실제로는 더 복잡한 구현 필요)
	sourceLocations, err := sourceService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get source storage locations: "+err.Error())
	}

	destLocations, err := destService.GetBackupStorageLocations(ctx.Request().Context(), namespace)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get destination storage locations: "+err.Error())
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"sourceStorageLocations":      len(sourceLocations),
		"destinationStorageLocations": len(destLocations),
		"compatible":                  len(sourceLocations) > 0 && len(destLocations) > 0,
		"message":                     "Storage location comparison completed",
	})
}
