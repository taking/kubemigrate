package helpers

import (
	"net/http"
	"taking.kr/velero/models"
	"taking.kr/velero/validation"

	"github.com/labstack/echo/v4"
)

// BindAndValidateKubeConfig : 요청 바인딩 + kubeconfig 유효성 검사
func BindAndValidateKubeConfig(ctx echo.Context, validator *validation.RequestValidator) (models.KubeConfig, error) {
	var req models.KubeConfig
	if err := ctx.Bind(&req); err != nil {
		return req, JSONError(ctx, http.StatusBadRequest, "invalid request body")
	}

	decodeKubeConfig, err := validator.ValidateKubeConfigRequest(&req)
	if err != nil {
		return req, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.KubeConfig = decodeKubeConfig
	return req, nil
}

// ResolveNamespace : namespace 확인 (요청 > query > 기본값)
func ResolveNamespace(req *models.KubeConfig, ctx echo.Context, defaultNS string) string {
	if req.Namespace != "" {
		return req.Namespace
	}
	if ns := ctx.QueryParam("namespace"); ns != "" {
		return ns
	}
	return defaultNS
}

// JSONError : 에러 응답 반환
func JSONError(ctx echo.Context, status int, message string) error {
	return ctx.JSON(status, map[string]string{"error": message})
}

// JSONStatus : 상태 메시지 응답
func JSONStatus(ctx echo.Context, status, message string) error {
	return ctx.JSON(http.StatusOK, map[string]string{"status": status, "message": message})
}

// JSONSuccess : 데이터 성공 응답
func JSONSuccess(ctx echo.Context, data interface{}) error {
	return ctx.JSON(http.StatusOK, map[string]interface{}{"status": "success", "data": data})
}
