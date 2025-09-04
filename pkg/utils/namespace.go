package utils

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/pkg/models"
)

// ResolveNamespace : 네임스페이스 결정
func ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return defaultNS
}
