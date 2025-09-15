# 개발 가이드

## 프로젝트 구조

```
kubemigrate/
├── cmd/                    # 애플리케이션 진입점
│   └── main.go
├── internal/               # 내부 패키지 (외부에서 import 불가)
│   ├── api/               # API 핸들러
│   │   ├── helm/
│   │   ├── kubernetes/
│   │   ├── minio/
│   │   └── velero/
│   ├── cache/             # 캐시 구현
│   ├── config/            # 설정 관리
│   ├── constants/         # 상수 정의
│   ├── handler/           # 공통 핸들러
│   ├── logger/            # 로깅
│   ├── middleware/        # HTTP 미들웨어
│   ├── response/          # 응답 유틸리티
│   ├── server/            # 서버 설정
│   │   └── routes/        # 라우트 정의
│   └── validator/         # 검증 로직
├── pkg/                   # 공개 패키지 (외부에서 import 가능)
│   ├── client/            # 클라이언트 구현
│   ├── response/          # 응답 타입
│   └── utils/             # 유틸리티
├── docs/                  # 문서
├── docker/                # Docker 설정
└── example/               # 예제 코드
```

## 개발 환경 설정

### 1. 필수 도구

- Go 1.21 이상
- Docker (선택사항)
- Kubernetes 클러스터 (테스트용)
- MinIO 서버 (테스트용)

### 2. 개발 도구 설치

```bash
# Go 도구 설치
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest

# 또는 Makefile 사용
make install-tools
```

### 3. 프로젝트 설정

```bash
# 저장소 클론
git clone https://github.com/taking/kubemigrate.git
cd kubemigrate

# 의존성 설치
go mod download

# 개발 환경 설정
cp .env.example .env
```

## 코딩 규칙

### 1. Go 코딩 스타일

- [Effective Go](https://golang.org/doc/effective_go.html) 가이드라인 준수
- `gofmt` 사용하여 코드 포맷팅
- `golangci-lint` 사용하여 린팅

### 2. 패키지 구조

- `internal/`: 애플리케이션 내부 로직
- `pkg/`: 외부에서 사용할 수 있는 공개 패키지
- 각 패키지는 단일 책임 원칙 준수

### 3. 네이밍 규칙

- 패키지명: 소문자, 간결하게
- 함수/메서드: PascalCase (공개), camelCase (비공개)
- 상수: PascalCase
- 변수: camelCase

### 4. 주석 규칙

- 모든 공개 함수/타입에 주석 작성
- 한글로 작성 (일관성 유지)
- 예시 코드 포함 권장

```go
// GetPods Pod 목록을 조회합니다.
// namespace가 비어있으면 모든 네임스페이스에서 조회합니다.
// name이 제공되면 특정 Pod를 조회합니다.
func GetPods(ctx context.Context, namespace, name string) (interface{}, error) {
    // 구현...
}
```

## 테스트 작성

### 1. 테스트 파일 규칙

- 테스트 파일명: `*_test.go`
- 테스트 함수명: `TestXxx` 또는 `TestXxx_Yyy`
- 벤치마크: `BenchmarkXxx`

### 2. 테스트 구조

```go
func TestFunctionName(t *testing.T) {
    // Given
    input := "test input"
    expected := "expected output"
    
    // When
    result, err := FunctionName(input)
    
    // Then
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### 3. 테스트 실행

```bash
# 모든 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./internal/api/kubernetes

# 커버리지 포함
go test -cover ./...

# 벤치마크 실행
go test -bench=. ./...
```

## API 개발

### 1. 새로운 API 추가

1. **핸들러 생성** (`internal/api/{service}/handler.go`)
2. **라우트 추가** (`internal/server/routes/{service}.go`)
3. **테스트 작성** (`internal/api/{service}/handler_test.go`)
4. **문서 업데이트** (`docs/API.md`)

### 2. 핸들러 작성 예시

```go
// GetResources 리소스를 조회합니다.
func (h *Handler) GetResources(c echo.Context) error {
    return h.HandleResourceClient(c, "resources", func(client client.Client, ctx context.Context) (interface{}, error) {
        // 비즈니스 로직 구현
        return client.Kubernetes().GetPods(ctx, namespace, name)
    })
}
```

### 3. 라우트 추가 예시

```go
// SetupKubernetesRoutes Kubernetes 관련 라우트를 설정합니다.
func SetupKubernetesRoutes(e *echo.Echo, kubernetesHandler *kubernetes.Handler) {
    api := e.Group("/api/v1")
    k8sGroup := api.Group("/kubernetes")
    
    k8sGroup.GET("/:kind", kubernetesHandler.GetResources)
}
```

## 로깅

### 1. 로그 레벨

- `DEBUG`: 상세한 디버깅 정보
- `INFO`: 일반적인 정보
- `WARN`: 경고 메시지
- `ERROR`: 에러 메시지
- `FATAL`: 치명적 에러

### 2. 로그 사용법

```go
import "github.com/taking/kubemigrate/internal/logger"

// 기본 로그
logger.Info("Operation completed successfully")

// 구조화된 로그
logger.Info("User login",
    logger.String("user_id", "12345"),
    logger.String("ip", "192.168.1.1"),
)

// 에러 로그
logger.Error("Failed to process request",
    logger.ErrorAttr(err),
    logger.String("request_id", "req-123"),
)
```

## 캐시 관리

### 1. LRU 캐시 사용

```go
// 캐시 생성
cache := cache.NewLRUCache(100, 5*time.Minute)

// 값 저장
cache.Set("key", value)

// 값 조회
if value, exists := cache.Get("key"); exists {
    // 사용
}

// 캐시 정리
cache.Cleanup()
```

### 2. 캐시 키 생성

```go
import "github.com/taking/kubemigrate/pkg/utils"

// SHA256 해시로 키 생성
key := utils.GenerateCacheKey("input string")
```

## 에러 처리

### 1. 에러 응답 형식

```go
// 유효성 검사 에러
return response.RespondWithValidationError(c, []responseTypes.ValidationError{
    {Field: "kubeconfig", Message: "kubeconfig is required"},
})

// 일반 에러
return response.RespondWithErrorModel(c, http.StatusBadRequest, 
    "VALIDATION_ERROR", "Invalid input", "kubeconfig is required")
```

### 2. 에러 로깅

```go
logger.Error("Failed to process request",
    logger.ErrorAttr(err),
    logger.String("operation", "GetPods"),
    logger.String("namespace", namespace),
)
```

## 성능 최적화

### 1. 메모리 모니터링

```go
// 메모리 통계 조회
stats := utils.GetMemoryStats()

// 메모리 사용률 확인
if utils.IsMemoryHigh(80.0) {
    // 최적화 수행
    utils.OptimizeMemory()
}
```

### 2. 워커 풀 사용

```go
// 워커 풀 생성
pool := utils.NewWorkerPool(10)
defer pool.Close()

// 작업 제출
pool.Submit(func() {
    // 비동기 작업
})
```

## 배포

### 1. 빌드

```bash
# 릴리스 빌드
make build

# Docker 이미지 빌드
make docker-build
```

### 2. 환경 설정

```bash
# 프로덕션 환경 변수
export LOG_LEVEL=info
export LOG_FORMAT=json
export SERVER_PORT=9091
```

## 디버깅

### 1. 로그 레벨 설정

```bash
# 디버그 모드
LOG_LEVEL=debug ./kubemigrate

# JSON 로그
LOG_FORMAT=json ./kubemigrate
```

### 2. 프로파일링

```go
import _ "net/http/pprof"

// 프로파일링 서버 시작
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

## 기여 가이드

### 1. Pull Request 프로세스

1. 이슈 생성 또는 기존 이슈 확인
2. 기능 브랜치 생성
3. 코드 작성 및 테스트
4. Pull Request 생성
5. 코드 리뷰 및 수정
6. 머지

### 2. 커밋 메시지 규칙

```
type(scope): description

- feat: 새로운 기능
- fix: 버그 수정
- docs: 문서 수정
- style: 코드 포맷팅
- refactor: 리팩토링
- test: 테스트 추가/수정
- chore: 빌드/설정 변경
```

### 3. 코드 리뷰 체크리스트

- [ ] 코드가 Go 스타일 가이드라인을 준수하는가?
- [ ] 적절한 테스트가 작성되었는가?
- [ ] 문서가 업데이트되었는가?
- [ ] 에러 처리가 적절한가?
- [ ] 로깅이 적절한가?
- [ ] 성능에 영향을 주는가?
