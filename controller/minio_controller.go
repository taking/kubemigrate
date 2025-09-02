package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/helpers"
	"taking.kr/velero/models"
)

// MinioController : MinIO 관련 API 컨트롤러
type MinioController struct{}

func NewMinioController() *MinioController {
	return &MinioController{}
}

// CheckMinioConnection : MinIO 서버 연결 상태를 확인하고 결과를 반환
func (c *MinioController) CheckMinioConnection(ctx echo.Context) error {
	var cfg models.MinioConfig
	if err := ctx.Bind(&cfg); err != nil {
		return helpers.JSONError(ctx, http.StatusBadRequest, "invalid request body")
	}

	client, err := clients.NewMinioClient(cfg)
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	if err := client.HealthCheck(ctx.Request().Context()); err != nil {
		return helpers.JSONError(ctx, http.StatusServiceUnavailable, "Minio cluster unhealthy: "+err.Error())
	}

	return helpers.JSONStatus(ctx, "healthy", "MinIO connection successful")
}

// CreateBucketIfNotExists : 버킷 존재 여부 확인 후 없으면 생성, 상태 메시지 반환
func (c *MinioController) CreateBucketIfNotExists(ctx echo.Context) error {
	var req struct {
		models.MinioConfig
		BucketName string `json:"bucketName"`
		Region     string `json:"region"`
	}

	if err := ctx.Bind(&req); err != nil {
		return helpers.JSONError(ctx, http.StatusBadRequest, "invalid request body")
	}

	client, err := clients.NewMinioClient(req.MinioConfig)
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	if req.Region == "" {
		req.Region = "us-east-1"
	}

	msg, err := client.CreateBucketIfNotExists(ctx.Request().Context(), req.BucketName, req.Region)
	if err != nil {
		return helpers.JSONError(ctx, http.StatusInternalServerError, err.Error())
	}

	return helpers.JSONStatus(ctx, "success", msg)
}
