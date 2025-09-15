# KubeMigrate 빠른 시작 가이드

## 설치 및 실행

### 1. 사전 요구사항

- Go 1.21 이상
- Kubernetes 클러스터 (선택사항)
- MinIO 서버 (선택사항)
- Velero (선택사항)

### 2. 소스에서 빌드

```bash
# 저장소 클론
git clone https://github.com/taking/kubemigrate.git
cd kubemigrate

# 의존성 설치
go mod download

# 빌드
go build -o kubemigrate cmd/main.go

# 실행
./kubemigrate
```

### 3. Docker로 실행

```bash
# Docker 이미지 빌드
docker build -t kubemigrate .

# 컨테이너 실행
docker run -p 9091:9091 kubemigrate
```

## 환경 변수 설정

### 기본 설정

```bash
# 서버 설정
export SERVER_HOST=localhost
export SERVER_PORT=9091
export READ_TIMEOUT=30s
export WRITE_TIMEOUT=30s
export IDLE_TIMEOUT=120s

# 로깅 설정
export LOG_LEVEL=info
export LOG_FORMAT=pretty

# 타임아웃 설정
export HEALTH_CHECK_TIMEOUT=5s
export REQUEST_TIMEOUT=30s
```

### Kubernetes 설정

```bash
# KubeConfig 파일 경로 (선택사항)
export KUBECONFIG=/path/to/kubeconfig

# 또는 환경변수로 설정
export KUBE_CONFIG_BASE64="base64-encoded-kubeconfig"
```

### MinIO 설정

```bash
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin123
export MINIO_USE_SSL=false
```

## 기본 사용법

### 1. 서버 시작

```bash
# 기본 설정으로 시작
./kubemigrate

# 환경변수와 함께 시작
SERVER_PORT=8080 LOG_LEVEL=debug ./kubemigrate
```

### 2. 헬스체크

```bash
# 서버 상태 확인
curl http://localhost:9091/api/v1/health

# 응답 예시
{
  "status": "healthy",
  "message": "API server is running",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 3. Kubernetes 리소스 조회

```bash
# Pod 목록 조회
curl -X GET "http://localhost:9091/api/v1/kubernetes/pods?namespace=default"

# 특정 Pod 조회
curl -X GET "http://localhost:9091/api/v1/kubernetes/pods/my-pod?namespace=default"
```

### 4. Helm 차트 관리

```bash
# 차트 목록 조회
curl -X GET "http://localhost:9091/api/v1/helm/charts"

# 차트 설치
curl -X POST "http://localhost:9091/api/v1/helm/charts?releaseName=nginx&chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz&version=15.4.2&namespace=default" \
  -H "Content-Type: application/json" \
  -d '{"kubeconfig": "base64-encoded-kubeconfig"}'
```

### 5. MinIO 버킷 관리

```bash
# 버킷 목록 조회
curl -X GET "http://localhost:9091/api/v1/minio/buckets"

# 버킷 생성
curl -X POST "http://localhost:9091/api/v1/minio/buckets/my-bucket" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin123",
    "useSSL": false
  }'
```

## 개발 환경 설정

### 1. 로컬 개발

```bash
# 개발 모드로 실행 (디버그 로그 활성화)
LOG_LEVEL=debug ./kubemigrate

# 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./internal/api/kubernetes
```

### 2. 코드 포맷팅 및 린팅

```bash
# 코드 포맷팅
go fmt ./...

# 린팅
golangci-lint run

# 또는 Makefile 사용
make lint
make format
```

### 3. 빌드 및 배포

```bash
# 릴리스 빌드
make build

# Docker 이미지 빌드
make docker-build

# 모든 플랫폼용 빌드
make build-all
```

## 문제 해결

### 1. 일반적인 문제

#### 연결 오류
```
Error: Kubernetes cluster unreachable
```
**해결방법**: KubeConfig가 올바른지 확인하고 클러스터에 접근 가능한지 확인

#### MinIO 연결 실패
```
Error: minio client not initialized
```
**해결방법**: MinIO 서버가 실행 중인지 확인하고 엔드포인트 설정 확인

#### 메모리 사용량 높음
```
Warning: High memory usage detected
```
**해결방법**: 메모리 최적화 실행
```bash
curl -X POST http://localhost:9091/api/v1/memory/optimize
```

### 2. 로그 확인

```bash
# 디버그 로그 활성화
LOG_LEVEL=debug ./kubemigrate

# JSON 형식 로그
LOG_FORMAT=json ./kubemigrate
```

### 3. 성능 모니터링

```bash
# 메모리 통계 조회
curl http://localhost:9091/api/v1/memory/stats

# 캐시 통계 조회
curl http://localhost:9091/api/v1/cache/stats

# 메모리 사용률 조회
curl http://localhost:9091/api/v1/memory/usage
```

## 고급 설정

### 1. 커스텀 미들웨어

미들웨어는 `internal/middleware/middleware.go`에서 설정할 수 있습니다.

### 2. 캐시 설정

LRU 캐시 설정은 `internal/constants/constants.go`에서 수정할 수 있습니다.

### 3. 레이트 리미팅

기본 설정:
- 초당 100개 요청
- 최대 50개 버스트
- 1분 만료

## 다음 단계

1. [API 문서](./API.md) 참조
2. [개발 가이드](./DEVELOPMENT.md) 확인
3. [배포 가이드](./DEPLOYMENT.md) 참조
4. [기여 가이드](./CONTRIBUTING.md) 확인
