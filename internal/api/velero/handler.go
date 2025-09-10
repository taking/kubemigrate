package velero

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client"
)

// Handler : Velero 관련 HTTP 핸들러
type Handler struct {
	*handler.BaseHandler
}

// NewHandler : 새로운 Velero 핸들러 생성
func NewHandler(base *handler.BaseHandler) *Handler {
	return &Handler{
		BaseHandler: base,
	}
}

// HealthCheck : Velero 연결 테스트
// @Summary Velero Connection Test
// @Description Test Velero connection with provided configuration
// @Tags velero
// @Accept json
// @Produce json
// @Param request body config.VeleroConfig true "Velero configuration"
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/velero/health [post]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.HandleResourceClient(c, "velero-health", func(client client.Client, ctx context.Context) (interface{}, error) {
		// Velero 연결 테스트
		_, err := client.Velero().GetBackups(ctx, "velero")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "velero",
			"message": "Velero connection is working",
		}, nil
	})
}
