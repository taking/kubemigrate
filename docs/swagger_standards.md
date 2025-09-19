# Swagger 주석 표준

## 기본 구조

모든 API 엔드포인트는 다음 구조의 Swagger 주석을 사용해야 합니다:

```go
// FunctionName : 함수 설명
// @Summary API 요약
// @Description API 상세 설명
// @Tags 태그명
// @Accept json
// @Produce json
// @Param 파라미터 설명
// @Success 응답 코드 {object} 응답 타입
// @Failure 에러 코드 {object} 에러 타입
// @Router 경로 [HTTP메서드]
```

## 표준 응답 타입

### 성공 응답
- `200 {object} response.SuccessResponse` - 성공
- `201 {object} response.SuccessResponse` - 생성 성공

### 에러 응답
- `400 {object} response.ErrorResponse` - 잘못된 요청
- `401 {object} response.ErrorResponse` - 인증 실패
- `403 {object} response.ErrorResponse` - 권한 없음
- `404 {object} response.ErrorResponse` - 리소스 없음
- `500 {object} response.ErrorResponse` - 서버 에러

## 표준 태그

- `kubernetes` - Kubernetes 관련 API
- `helm` - Helm 관련 API
- `minio` - MinIO 관련 API
- `velero` - Velero 관련 API

## 파라미터 표준

### Path 파라미터
```go
// @Param name path string true "파라미터 설명"
```

### Query 파라미터
```go
// @Param name query string false "파라미터 설명 (기본값: 'default')"
```

### Body 파라미터
```go
// @Param request body config.ConfigType true "설정 설명"
```

### Form 파라미터
```go
// @Param field formData string true "필드 설명"
// @Param file formData file true "파일 설명"
```

## 예시

```go
// GetResources : Kubernetes 리소스 조회
// @Summary Get Kubernetes Resources
// @Description Get Kubernetes resources by kind, name (optional) and namespace
// @Tags kubernetes
// @Accept json
// @Produce json
// @Param request body config.KubeConfig true "Kubernetes configuration"
// @Param kind path string true "Resource kind (pods, configmaps, secrets, storage-classes)"
// @Param name path string false "Resource name (empty for list, specific name for single resource)"
// @Param namespace query string false "Namespace name (default: 'default', all namespaces: 'all')"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /v1/kubernetes/{kind}/{name} [get]
func (h *Handler) GetResources(c echo.Context) error {
    // 구현
}
```
