package kubernetes

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/config"
	"github.com/taking/kubemigrate/internal/response"
	"github.com/taking/kubemigrate/internal/validator"
	"github.com/taking/kubemigrate/pkg/client/kubernetes"
	"github.com/taking/kubemigrate/pkg/errors"
	"github.com/taking/kubemigrate/pkg/utils"
	// v1 "k8s.io/api/core/v1" // 사용하지 않음
)

// KubernetesPlugin Kubernetes 플러그인
type KubernetesPlugin struct {
	client        kubernetes.Client
	config        map[string]interface{}
	validator     *validator.KubernetesValidator
	pluginManager interface{} // 플러그인 매니저 참조 (캐시 사용을 위해)
}

// NewPlugin 새로운 Kubernetes 플러그인 생성
func NewPlugin() *KubernetesPlugin {
	return &KubernetesPlugin{
		validator: validator.NewKubernetesValidator(),
	}
}

// Name 플러그인 이름
func (p *KubernetesPlugin) Name() string {
	return "kubernetes"
}

// Version 플러그인 버전
func (p *KubernetesPlugin) Version() string {
	return "1.0.0"
}

// Description 플러그인 설명
func (p *KubernetesPlugin) Description() string {
	return "Kubernetes resource management plugin"
}

// Initialize 플러그인 초기화
func (p *KubernetesPlugin) Initialize(config map[string]interface{}) error {
	p.config = config

	// 기본 클라이언트 초기화 (캐시는 필요시 생성)
	p.client = kubernetes.NewClient()

	return nil
}

// Shutdown 플러그인 종료
func (p *KubernetesPlugin) Shutdown() error {
	// 정리 작업이 필요한 경우 여기에 구현
	return nil
}

// RegisterRoutes 라우트 등록
func (p *KubernetesPlugin) RegisterRoutes(router *echo.Group) error {
	// Kubernetes 관련 라우트 등록
	k8sGroup := router.Group("/kubernetes")

	// 헬스체크
	k8sGroup.POST("/health", p.HealthCheckHandler)

	// 통합 API (조회) - 기존과 동일
	k8sGroup.GET("/:kind", p.GetResourcesHandler)      // 리스트 조회
	k8sGroup.GET("/:kind/:name", p.GetResourceHandler) // 단일 조회

	return nil
}

// HealthCheck 헬스체크
func (p *KubernetesPlugin) HealthCheck(ctx context.Context) error {
	_, err := p.client.GetPods(ctx, "default", "")
	return err
}

// GetServiceType 서비스 타입
func (p *KubernetesPlugin) GetServiceType() string {
	return "kubernetes"
}

// GetClient 클라이언트 반환
func (p *KubernetesPlugin) GetClient() interface{} {
	return p.client
}

// SetPluginManager 플러그인 매니저 설정
func (p *KubernetesPlugin) SetPluginManager(manager interface{}) {
	p.pluginManager = manager
}

// HealthCheckHandler 헬스체크 핸들러
func (p *KubernetesPlugin) HealthCheckHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	// 미들웨어에서 캐시된 클라이언트 사용
	var clientToUse kubernetes.Client
	if cachedClient := c.Get("cached_client"); cachedClient != nil {
		if k8sClient, ok := cachedClient.(kubernetes.Client); ok {
			clientToUse = k8sClient
		}
	}

	// 캐시된 클라이언트가 없으면 기본 클라이언트 사용
	if clientToUse == nil {
		clientToUse = p.client
	}

	// Kubernetes 연결 테스트
	_, err := clientToUse.GetPods(c.Request().Context(), "default", "")
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetPods", err)
	}

	return response.RespondWithSuccessModel(c, 200, "Kubernetes connection is working", map[string]interface{}{
		"status": "UP",
	})
}

// GetPodsHandler 파드 목록 조회 핸들러
func (p *KubernetesPlugin) GetPodsHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	pods, err := p.client.GetPods(c.Request().Context(), namespace, "")
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetPods", err)
	}

	return response.RespondWithData(c, 200, pods)
}

// GetPodHandler 특정 파드 조회 핸들러
func (p *KubernetesPlugin) GetPodHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	podName := c.Param("name")

	pod, err := p.client.GetPods(c.Request().Context(), namespace, podName)
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetPods", err)
	}

	return response.RespondWithData(c, 200, pod)
}

// GetConfigMapsHandler ConfigMap 목록 조회 핸들러
func (p *KubernetesPlugin) GetConfigMapsHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	configMaps, err := p.client.GetConfigMaps(c.Request().Context(), namespace, "")
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetConfigMaps", err)
	}

	return response.RespondWithData(c, 200, configMaps)
}

// GetConfigMapHandler 특정 ConfigMap 조회 핸들러
func (p *KubernetesPlugin) GetConfigMapHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	configMapName := c.Param("name")

	configMapResult, err := p.client.GetConfigMaps(c.Request().Context(), namespace, configMapName)
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetConfigMaps", err)
	}

	return response.RespondWithData(c, 200, configMapResult)
}

// GetSecretsHandler Secret 목록 조회 핸들러
func (p *KubernetesPlugin) GetSecretsHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	secrets, err := p.client.GetSecrets(c.Request().Context(), namespace, "")
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetSecrets", err)
	}

	return response.RespondWithData(c, 200, secrets)
}

// GetSecretHandler 특정 Secret 조회 핸들러
func (p *KubernetesPlugin) GetSecretHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespace := utils.ResolveNamespace(c, "default")
	secretName := c.Param("name")

	secret, err := p.client.GetSecrets(c.Request().Context(), namespace, secretName)
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetSecrets", err)
	}

	return response.RespondWithData(c, 200, secret)
}

// GetStorageClassesHandler StorageClass 목록 조회 핸들러
func (p *KubernetesPlugin) GetStorageClassesHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	storageClasses, err := p.client.GetStorageClasses(c.Request().Context(), "")
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetStorageClasses", err)
	}

	return response.RespondWithData(c, 200, storageClasses)
}

// GetStorageClassHandler 특정 StorageClass 조회 핸들러
func (p *KubernetesPlugin) GetStorageClassHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	storageClassName := c.Param("name")

	storageClass, err := p.client.GetStorageClasses(c.Request().Context(), storageClassName)
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetStorageClasses", err)
	}

	return response.RespondWithData(c, 200, storageClass)
}

// GetNamespacesHandler Namespace 목록 조회 핸들러
func (p *KubernetesPlugin) GetNamespacesHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespaces, err := p.client.GetNamespaces(c.Request().Context())
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetNamespaces", err)
	}

	return response.RespondWithData(c, 200, namespaces)
}

// GetNamespaceHandler 특정 Namespace 조회 핸들러
func (p *KubernetesPlugin) GetNamespaceHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	namespaceName := c.Param("name")

	namespace, err := p.client.GetNamespace(c.Request().Context(), namespaceName)
	if err != nil {
		return errors.NewExternalError("kubernetes", "GetNamespace", err)
	}

	return response.RespondWithData(c, 200, namespace)
}

// GetResourcesHandler 통합 리소스 목록 조회 핸들러 (기존과 동일)
func (p *KubernetesPlugin) GetResourcesHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	kind := c.Param("kind")
	namespace := utils.ResolveNamespace(c, "default")

	switch kind {
	case "pods":
		resources, err := p.client.GetPods(c.Request().Context(), namespace, "")
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetPods", err)
		}
		return response.RespondWithData(c, 200, resources)
	case "configmaps":
		resources, err := p.client.GetConfigMaps(c.Request().Context(), namespace, "")
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetConfigMaps", err)
		}
		return response.RespondWithData(c, 200, resources)
	case "secrets":
		resources, err := p.client.GetSecrets(c.Request().Context(), namespace, "")
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetSecrets", err)
		}
		return response.RespondWithData(c, 200, resources)
	case "storage-classes":
		resources, err := p.client.GetStorageClasses(c.Request().Context(), "")
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetStorageClasses", err)
		}
		return response.RespondWithData(c, 200, resources)
	case "namespaces":
		resources, err := p.client.GetNamespaces(c.Request().Context())
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetNamespaces", err)
		}
		return response.RespondWithData(c, 200, resources)
	default:
		return errors.NewValidationError("UNSUPPORTED_RESOURCE", "Unsupported resource type", "Resource type '"+kind+"' is not supported")
	}
}

// GetResourceHandler 통합 리소스 단일 조회 핸들러 (기존과 동일)
func (p *KubernetesPlugin) GetResourceHandler(c echo.Context) error {
	var req config.KubeConfig
	if err := c.Bind(&req); err != nil {
		return errors.NewValidationError(errors.CodeInvalidRequest, "Invalid request body", err.Error())
	}

	kind := c.Param("kind")
	name := c.Param("name")
	namespace := utils.ResolveNamespace(c, "default")

	switch kind {
	case "pods":
		resource, err := p.client.GetPods(c.Request().Context(), namespace, name)
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetPods", err)
		}
		return response.RespondWithData(c, 200, resource)
	case "configmaps":
		resource, err := p.client.GetConfigMaps(c.Request().Context(), namespace, name)
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetConfigMaps", err)
		}
		return response.RespondWithData(c, 200, resource)
	case "secrets":
		resource, err := p.client.GetSecrets(c.Request().Context(), namespace, name)
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetSecrets", err)
		}
		return response.RespondWithData(c, 200, resource)
	case "storage-classes":
		resource, err := p.client.GetStorageClasses(c.Request().Context(), name)
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetStorageClasses", err)
		}
		return response.RespondWithData(c, 200, resource)
	case "namespaces":
		resource, err := p.client.GetNamespaces(c.Request().Context())
		if err != nil {
			return errors.NewExternalError("kubernetes", "GetNamespaces", err)
		}
		// 이름으로 필터링
		for _, ns := range resource.Items {
			if ns.Name == name {
				return response.RespondWithData(c, 200, ns)
			}
		}
		return errors.NewValidationError("RESOURCE_NOT_FOUND", "Resource not found", "Namespace '"+name+"' not found")
	default:
		return errors.NewValidationError("UNSUPPORTED_RESOURCE", "Unsupported resource type", "Resource type '"+kind+"' is not supported")
	}
}
