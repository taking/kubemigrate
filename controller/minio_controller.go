package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"taking.kr/velero/clients"
	"taking.kr/velero/utils"
)

// MinioController : MinIO 관련 API 컨트롤러
type MinioController struct {
	*BaseController
}

func NewMinioController() *MinioController {
	return &MinioController{
		BaseController: NewBaseController(),
	}
}

// CheckMinioConnection : MinIO 서버 연결 상태를 확인하고 결과를 반환
// CheckMinioConnection godoc
// @Summary Minio 연결 확인
// @Description MinioConfig을 사용하여 Minio 연결 검증
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfigRequest true "Minio 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /minio/health [get]
func (c *MinioController) CheckMinioConnection(ctx echo.Context) error {
	// minioConfig 바인딩 및 검증
	req, err := c.BindAndValidateMinioConfig(ctx)
	if err != nil {
		return err
	}

	// Minio 클라이언트 생성
	client, err := clients.NewMinioClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	// Minio 연결 상태 확인
	return c.HandleHealthCheck(ctx, client, "Minio")
}

// CreateBucketIfNotExists : 버킷 존재 여부 확인 후 없으면 생성, 상태 메시지 반환
// CreateBucketIfNotExists godoc
// @Summary Minio 버킷 확인
// @Description MinioConfig을 사용하여 Minio 버킷 존재 여부 확인 후, 없으면 생성
// @Tags minio
// @Accept json
// @Produce json
// @Param request body models.MinioConfigRequest true "Minio 연결에 필요한 값"
// @Success 200 {object} models.SwaggerSuccessResponse "연결 성공"
// @Failure 400 {object} models.SwaggerErrorResponse "잘못된 요청"
// @Failure 500 {object} models.SwaggerErrorResponse "서버 내부 오류"
// @Failure 503 {object} models.SwaggerErrorResponse "서비스 이용 불가"
// @Router /minio/bucket_check [post]
func (c *MinioController) CreateBucketIfNotExists(ctx echo.Context) error {
	// minioConfig 바인딩 및 검증
	req, err := c.BindAndValidateMinioConfig(ctx)
	if err != nil {
		return err
	}

	// Minio 클라이언트 생성
	client, err := clients.NewMinioClient(req)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	req.BucketName = utils.DefaultString(req.BucketName, "velero") // Bucket 이름이 비어 있으면 기본값 사용

	// 버킷이 존재하는지 확인하고, 없으면 새로 생성
	msg, err := client.CreateBucketIfNotExists(ctx.Request().Context(), req.BucketName, req.Region)
	if err != nil {
		return utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
	}

	return utils.RespondStatus(ctx, "success", msg)
}
