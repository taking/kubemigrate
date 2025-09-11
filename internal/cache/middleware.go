package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

// CacheMiddleware 캐시 미들웨어
func CacheMiddleware(cacheManager *Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 요청에서 설정 추출
			apiType, config := extractConfigFromRequest(c)
			if apiType == "" {
				// 설정이 없으면 캐시 없이 진행
				return next(c)
			}

			// 캐시에서 클라이언트 조회/생성
			client, err := cacheManager.GetCachedClient(apiType, config)
			if err != nil {
				// 캐시 오류 시 로그만 남기고 계속 진행
				c.Logger().Warnf("Cache error: %v", err)
				return next(c)
			}

			// 컨텍스트에 캐시된 클라이언트 저장
			c.Set("cached_client", client)
			c.Set("api_type", apiType)
			c.Set("config_hash", generateConfigHash(config))

			return next(c)
		}
	}
}

// extractConfigFromRequest 요청에서 설정 추출
func extractConfigFromRequest(c echo.Context) (string, map[string]interface{}) {
	apiType := detectApiType(c.Request().URL.Path)
	if apiType == "" {
		return "", nil
	}

	config := extractConfigByApiType(c, apiType)
	return apiType, config
}

// detectApiType 경로에서 API 타입 감지
func detectApiType(path string) string {
	if strings.Contains(path, "/kubernetes/") {
		return "kubernetes"
	} else if strings.Contains(path, "/minio/") {
		return "minio"
	} else if strings.Contains(path, "/helm/") {
		return "helm"
	} else if strings.Contains(path, "/velero/") {
		return "velero"
	}
	return ""
}

// extractConfigByApiType API 타입별 설정 추출
func extractConfigByApiType(c echo.Context, apiType string) map[string]interface{} {
	config := make(map[string]interface{})

	if c.Request().Method == "POST" {
		// POST Body에서 설정 추출
		body, err := io.ReadAll(c.Request().Body)
		if err == nil {
			// Body 복원
			c.Request().Body = io.NopCloser(strings.NewReader(string(body)))

			var requestData map[string]interface{}
			if json.Unmarshal(body, &requestData) == nil {
				config = extractConfigFromBody(requestData, apiType)
			}
		}
	} else {
		// GET Query Parameter에서 설정 추출
		config = extractConfigFromQuery(c, apiType)
	}

	return config
}

// extractConfigFromBody POST Body에서 설정 추출
func extractConfigFromBody(requestData map[string]interface{}, apiType string) map[string]interface{} {
	config := make(map[string]interface{})

	switch apiType {
	case "kubernetes", "helm":
		if kubeconfig, ok := requestData["kubeconfig"].(string); ok && kubeconfig != "" {
			config["kubeconfig"] = kubeconfig
		}
	case "minio":
		if endpoint, ok := requestData["endpoint"].(string); ok {
			config["minio_endpoint"] = endpoint
		}
		if accessKey, ok := requestData["accessKey"].(string); ok {
			config["minio_access_key"] = accessKey
		}
		if secretKey, ok := requestData["secretKey"].(string); ok {
			config["minio_secret_key"] = secretKey
		}
		if useSSL, ok := requestData["useSSL"].(bool); ok {
			config["minio_use_ssl"] = useSSL
		}
	case "velero":
		if kubeconfigData, ok := requestData["kubeconfig"].(map[string]interface{}); ok {
			if kubeconfig, ok := kubeconfigData["kubeconfig"].(string); ok {
				config["kubeconfig"] = kubeconfig
			}
		}
		if minioData, ok := requestData["minio"].(map[string]interface{}); ok {
			if endpoint, ok := minioData["endpoint"].(string); ok {
				config["minio_endpoint"] = endpoint
			}
			if accessKey, ok := minioData["accessKey"].(string); ok {
				config["minio_access_key"] = accessKey
			}
			if secretKey, ok := minioData["secretKey"].(string); ok {
				config["minio_secret_key"] = secretKey
			}
			if useSSL, ok := minioData["useSSL"].(bool); ok {
				config["minio_use_ssl"] = useSSL
			}
		}
	}

	return config
}

// extractConfigFromQuery GET Query Parameter에서 설정 추출
func extractConfigFromQuery(c echo.Context, apiType string) map[string]interface{} {
	config := make(map[string]interface{})

	switch apiType {
	case "kubernetes", "helm", "velero":
		if kubeConfig := c.QueryParam("kubeconfig"); kubeConfig != "" {
			config["kubeconfig"] = kubeConfig
		}
	}

	// MinIO와 Velero는 추가 설정 필요
	if apiType == "minio" || apiType == "velero" {
		if endpoint := c.QueryParam("endpoint"); endpoint != "" {
			config["minio_endpoint"] = endpoint
		}
		if accessKey := c.QueryParam("access_key"); accessKey != "" {
			config["minio_access_key"] = accessKey
		}
		if secretKey := c.QueryParam("secret_key"); secretKey != "" {
			config["minio_secret_key"] = secretKey
		}
		if useSSL := c.QueryParam("use_ssl"); useSSL != "" {
			config["minio_use_ssl"] = useSSL == "true"
		}
	}

	return config
}

// generateConfigHash 설정 기반 해시 생성
func generateConfigHash(config map[string]interface{}) string {
	// 설정을 정렬된 JSON으로 변환하여 일관된 해시 생성
	configBytes, err := json.Marshal(config)
	if err != nil {
		// JSON 변환 실패 시 기본 해시 생성
		configStr := fmt.Sprintf("%v", config)
		hash := sha256.Sum256([]byte(configStr))
		return hex.EncodeToString(hash[:])
	}

	hash := sha256.Sum256(configBytes)
	return hex.EncodeToString(hash[:])
}

// GetCachedClientFromContext 컨텍스트에서 캐시된 클라이언트 조회
func GetCachedClientFromContext(c echo.Context) interface{} {
	return c.Get("cached_client")
}

// GetApiTypeFromContext 컨텍스트에서 API 타입 조회
func GetApiTypeFromContext(c echo.Context) string {
	return c.Get("api_type").(string)
}

// GetConfigHashFromContext 컨텍스트에서 설정 해시 조회
func GetConfigHashFromContext(c echo.Context) string {
	return c.Get("config_hash").(string)
}

// InvalidateCacheByConfig 설정 변경 시 캐시 무효화
func InvalidateCacheByConfig(cacheManager *Manager, apiType string, oldConfig, newConfig map[string]interface{}) {
	// 설정이 변경되었는지 확인
	if !reflect.DeepEqual(oldConfig, newConfig) {
		cacheManager.Invalidate(apiType, oldConfig)
	}
}
