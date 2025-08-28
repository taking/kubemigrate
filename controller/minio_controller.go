package controller

import (
	"context"
	"net/http"
	"taking.kr/velero/models"

	"github.com/labstack/echo/v4"
	"taking.kr/velero/clients"
)

type MinioController struct{}

func NewMinioController() *MinioController {
	return &MinioController{}
}

func (c *MinioController) CheckMinioConnection(ctx echo.Context) error {
	var cfg models.MinioConfig
	if err := ctx.Bind(&cfg); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	client, err := clients.NewMinioClient(cfg)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
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
		"message": "minio connection successful",
	})
}

func (c *MinioController) CreateBucketIfNotExists(ctx echo.Context) error {
	var req struct {
		models.MinioConfig
		BucketName string `json:"bucketName"`
		Region     string `json:"region"` // 기본값: "us-east-1"
	}

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	client, err := clients.NewMinioClient(req.MinioConfig)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if req.Region == "" {
		req.Region = "us-east-1"
	}

	statusMsg, err := client.CreateBucketIfNotExists(context.Background(), req.BucketName, req.Region)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{
			"status": "failed",
			"error":  err.Error(),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"status":  "success",
		"bucket":  req.BucketName,
		"message": statusMsg, // "already exists" 또는 "created successfully"
	})
}
