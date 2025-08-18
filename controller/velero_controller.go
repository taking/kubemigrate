package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/clients"
	"taking.kr/velero/interfaces"
	"taking.kr/velero/utils"
)

type VeleroController struct {
	service interfaces.VeleroService
}

func NewVeleroController(s interfaces.VeleroService) *VeleroController {
	return &VeleroController{service: s}
}

// 공통 처리 함수
func (bc *VeleroController) handleWithNamespace(
	c echo.Context,
	handler func(ctx context.Context, ns string) (interface{}, error),
) error {
	ns := c.QueryParam("namespace")
	if ns == "" {
		ns = "velero"
	}
	result, err := handler(context.Background(), ns)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (bc *VeleroController) GetBackups(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		// Prefer existing service if provided, else construct from header per request
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetBackups(ctx, ns)
	})
}

func (bc *VeleroController) GetRestores(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetRestores(ctx, ns)
	})
}

func (bc *VeleroController) GetBackupRepositories(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetBackupRepositories(ctx, ns)
	})
}

func (bc *VeleroController) GetBackupStorageLocations(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetBackupStorageLocations(ctx, ns)
	})
}

func (bc *VeleroController) GetVolumeSnapshotLocations(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetVolumeSnapshotLocations(ctx, ns)
	})
}

func (bc *VeleroController) GetPodVolumeRestores(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetPodVolumeRestores(ctx, ns)
	})
}

func (bc *VeleroController) GetDownloadRequests(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetDownloadRequests(ctx, ns)
	})
}

func (bc *VeleroController) GetDataUploads(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetDataUploads(ctx, ns)
	})
}

func (bc *VeleroController) GetDataDownloads(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetDataDownloads(ctx, ns)
	})
}

func (bc *VeleroController) GetServerStatusRequests(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		svc := bc.service
		if svc == nil {
			raw := c.Request().Header.Get("X-Kubeconfig")
			if raw == "" {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
			}
			cfg, err := utils.ParseRestConfigFromRaw(raw)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			v, err := clients.NewVeleroClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = v
		}
		return svc.GetServerStatusRequests(ctx, ns)
	})
}
