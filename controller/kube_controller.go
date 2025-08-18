package controller

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"taking.kr/velero/clients"
	"taking.kr/velero/interfaces"
	"taking.kr/velero/utils"
)

type KubeController struct {
	service interfaces.KubeService
}

func NewKubeController(s interfaces.KubeService) *KubeController {
	return &KubeController{service: s}
}

// 공통 처리 함수
func (bc *KubeController) handleWithNamespace(
	c echo.Context,
	handler func(ctx context.Context, ns string) (interface{}, error),
) error {
	ns := c.QueryParam("namespace")
	if ns == "" {
		ns = "default"
	}
	result, err := handler(context.Background(), ns)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}

func (bc *KubeController) GetResources(c echo.Context) error {
	return bc.handleWithNamespace(c, func(ctx context.Context, ns string) (interface{}, error) {
		raw := c.Request().Header.Get("X-Kubeconfig")
		if raw == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "missing X-Kubeconfig header")
		}
		cfg, err := utils.ParseRestConfigFromRaw(raw)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		svc := bc.service
		if svc == nil {
			s, err := clients.NewKubeClientFromRestConfig(cfg)
			if err != nil {
				return nil, err
			}
			svc = s
		}

		resourceReq := schema.GroupVersionResource{
			Group:    c.Param("group"),
			Version:  c.Param("version"),
			Resource: c.Param("resource"),
		}

		return svc.GetResources(ctx, resourceReq, ns, c.Param("name"))
	})
}
