package velero

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client/velero"
	"github.com/taking/kubemigrate/pkg/errors"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VeleroPlugin Velero 플러그인
type VeleroPlugin struct {
	client    velero.Client
	config    map[string]interface{}
	validator *validator.KubernetesValidator
}

// NewPlugin 새로운 Velero 플러그인 생성
func NewPlugin() *VeleroPlugin {
	return &VeleroPlugin{
		validator: validator.NewKubernetesValidator(),
	}
}

// Name 플러그인 이름
func (p *VeleroPlugin) Name() string {
	return "velero"
}

// Version 플러그인 버전
func (p *VeleroPlugin) Version() string {
	return "1.0.0"
}

// Description 플러그인 설명
func (p *VeleroPlugin) Description() string {
	return "Velero backup and restore management plugin"
}

// Initialize 플러그인 초기화
func (p *VeleroPlugin) Initialize(config map[string]interface{}) error {
	p.config = config

	// Velero 클라이언트 초기화 (기본 설정 사용)
	// TODO: 설정 기반 클라이언트 초기화 구현 필요
	p.client = velero.NewClient()

	return nil
}

// Shutdown 플러그인 종료
func (p *VeleroPlugin) Shutdown() error {
	// 정리 작업이 필요한 경우 여기에 구현
	return nil
}

// RegisterRoutes 라우트 등록
func (p *VeleroPlugin) RegisterRoutes(router *echo.Group) error {
	// Velero 관련 라우트 등록 (기존과 동일한 구조)
	veleroGroup := router.Group("/velero")

	// 헬스체크
	veleroGroup.POST("/health", p.HealthCheckHandler)

	// 백업 관련 (기존과 동일)
	veleroGroup.POST("/backups", p.GetBackupsHandler)
	veleroGroup.GET("/repositories", p.GetBackupRepositoriesHandler)
	veleroGroup.GET("/storage-locations", p.GetBackupStorageLocationsHandler)

	// 복구 관련 (기존과 동일)
	veleroGroup.POST("/restores", p.GetRestoresHandler)
	veleroGroup.GET("/volume-snapshot-locations", p.GetVolumeSnapshotLocationsHandler)
	veleroGroup.GET("/pod-volume-restores", p.GetPodVolumeRestoresHandler)

	return nil
}

// HealthCheck 헬스체크
func (p *VeleroPlugin) HealthCheck(ctx context.Context) error {
	_, err := p.client.GetBackups(ctx, "velero")
	return err
}

// GetServiceType 서비스 타입
func (p *VeleroPlugin) GetServiceType() string {
	return "velero"
}

// GetClient 클라이언트 반환
func (p *VeleroPlugin) GetClient() interface{} {
	return p.client
}

// SetPluginManager 플러그인 매니저 설정
func (p *VeleroPlugin) SetPluginManager(manager interface{}) {
	// Velero 플러그인에서는 현재 사용하지 않음
}

// HealthCheckHandler 헬스체크 핸들러
func (p *VeleroPlugin) HealthCheckHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// Velero 연결 테스트
	_, err := p.client.GetBackups(c.Request().Context(), "velero")
	if err != nil {
		return errors.NewExternalError("velero", "GetBackups", err)
	}

	return response.RespondWithSuccessModel(c, 200, "Velero connection is working", map[string]interface{}{
		"service": "velero",
		"message": "Velero connection is working",
	})
}

// GetBackupsHandler 백업 목록 조회 핸들러
func (p *VeleroPlugin) GetBackupsHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	backups, err := p.client.GetBackups(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetBackups", err)
	}

	return response.RespondWithData(c, 200, backups)
}

// GetBackupHandler 특정 백업 조회 핸들러
func (p *VeleroPlugin) GetBackupHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	backupName := c.Param("name")
	if backupName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing backup name", "backup name parameter is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	backup, err := p.client.GetBackup(c.Request().Context(), namespace, backupName)
	if err != nil {
		return errors.NewExternalError("velero", "GetBackup", err)
	}

	return response.RespondWithData(c, 200, backup)
}

// CreateBackupHandler 백업 생성 핸들러
func (p *VeleroPlugin) CreateBackupHandler(c echo.Context) error {
	var req struct {
		config.VeleroConfig
		BackupName string            `json:"backupName"`
		Namespaces []string          `json:"namespaces,omitempty"`
		Resources  []string          `json:"resources,omitempty"`
		Labels     map[string]string `json:"labels,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	if req.BackupName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing backup name", "backupName is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	// Velero Backup 객체 생성 (간단한 예시)
	backup := &velerov1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.BackupName,
			Namespace: namespace,
		},
		Spec: velerov1.BackupSpec{
			IncludedNamespaces: req.Namespaces,
			IncludedResources:  req.Resources,
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: req.Labels,
			},
		},
	}

	err := p.client.CreateBackup(c.Request().Context(), namespace, backup)
	if err != nil {
		return errors.NewExternalError("velero", "CreateBackup", err)
	}

	return response.RespondWithData(c, 201, backup)
}

// DeleteBackupHandler 백업 삭제 핸들러
func (p *VeleroPlugin) DeleteBackupHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	backupName := c.Param("name")
	if backupName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing backup name", "backup name parameter is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	err := p.client.DeleteBackup(c.Request().Context(), namespace, backupName)
	if err != nil {
		return errors.NewExternalError("velero", "DeleteBackup", err)
	}

	return response.RespondWithMessage(c, 200, "Backup deleted successfully")
}

// GetRestoresHandler 복구 목록 조회 핸들러
func (p *VeleroPlugin) GetRestoresHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	restores, err := p.client.GetRestores(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetRestores", err)
	}

	return response.RespondWithData(c, 200, restores)
}

// GetRestoreHandler 특정 복구 조회 핸들러
func (p *VeleroPlugin) GetRestoreHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	restoreName := c.Param("name")
	if restoreName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing restore name", "restore name parameter is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	restore, err := p.client.GetRestore(c.Request().Context(), namespace, restoreName)
	if err != nil {
		return errors.NewExternalError("velero", "GetRestore", err)
	}

	return response.RespondWithData(c, 200, restore)
}

// CreateRestoreHandler 복구 생성 핸들러
func (p *VeleroPlugin) CreateRestoreHandler(c echo.Context) error {
	var req struct {
		config.VeleroConfig
		RestoreName string            `json:"restoreName"`
		BackupName  string            `json:"backupName"`
		Namespaces  []string          `json:"namespaces,omitempty"`
		Resources   []string          `json:"resources,omitempty"`
		Labels      map[string]string `json:"labels,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	if req.RestoreName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing restore name", "restoreName is required")
	}
	if req.BackupName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing backup name", "backupName is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	// Velero Restore 객체 생성 (간단한 예시)
	restore := &velerov1.Restore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.RestoreName,
			Namespace: namespace,
		},
		Spec: velerov1.RestoreSpec{
			BackupName:         req.BackupName,
			IncludedNamespaces: req.Namespaces,
			IncludedResources:  req.Resources,
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: req.Labels,
			},
		},
	}

	err := p.client.CreateRestore(c.Request().Context(), namespace, restore)
	if err != nil {
		return errors.NewExternalError("velero", "CreateRestore", err)
	}

	return response.RespondWithData(c, 201, restore)
}

// DeleteRestoreHandler 복구 삭제 핸들러
func (p *VeleroPlugin) DeleteRestoreHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	restoreName := c.Param("name")
	if restoreName == "" {
		return errors.NewValidationError(errors.CodeMissingParameter, "Missing restore name", "restore name parameter is required")
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	err := p.client.DeleteRestore(c.Request().Context(), namespace, restoreName)
	if err != nil {
		return errors.NewExternalError("velero", "DeleteRestore", err)
	}

	return response.RespondWithMessage(c, 200, "Restore deleted successfully")
}

// GetBackupRepositoriesHandler 백업 저장소 목록 조회 핸들러
func (p *VeleroPlugin) GetBackupRepositoriesHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	repositories, err := p.client.GetBackupRepositories(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetBackupRepositories", err)
	}

	return response.RespondWithData(c, 200, repositories)
}

// GetBackupStorageLocationsHandler 백업 스토리지 위치 목록 조회 핸들러
func (p *VeleroPlugin) GetBackupStorageLocationsHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	locations, err := p.client.GetBackupStorageLocations(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetBackupStorageLocations", err)
	}

	return response.RespondWithData(c, 200, locations)
}

// GetVolumeSnapshotLocationsHandler 볼륨 스냅샷 위치 목록 조회 핸들러
func (p *VeleroPlugin) GetVolumeSnapshotLocationsHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	locations, err := p.client.GetVolumeSnapshotLocations(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetVolumeSnapshotLocations", err)
	}

	return response.RespondWithData(c, 200, locations)
}

// GetPodVolumeRestoresHandler Pod 볼륨 복구 목록 조회 핸들러
func (p *VeleroPlugin) GetPodVolumeRestoresHandler(c echo.Context) error {
	var req config.VeleroConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := c.QueryParam("namespace")
	if namespace == "" {
		namespace = "velero"
	}

	restores, err := p.client.GetPodVolumeRestores(c.Request().Context(), namespace)
	if err != nil {
		return errors.NewExternalError("velero", "GetPodVolumeRestores", err)
	}

	return response.RespondWithData(c, 200, restores)
}
