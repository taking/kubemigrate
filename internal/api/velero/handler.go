package velero

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/handler"
	"github.com/taking/kubemigrate/pkg/client/velero"
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

// HealthCheck : Velero 연결 상태 확인
// @Summary Velero Health Check
// @Description Check Velero connection status
// @Tags velero
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/velero/health [get]
func (h *Handler) HealthCheck(c echo.Context) error {
	return h.BaseHandler.HandleVeleroResource(c, "velero-health", func(veleroClient velero.Client, ctx context.Context) (interface{}, error) {
		// 간단한 Velero 연결 테스트 (백업 목록 조회)
		_, err := veleroClient.GetBackups(ctx, "velero")
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"service": "velero",
			"status":  "healthy",
			"message": "Velero connection is working",
		}, nil
	})
}
