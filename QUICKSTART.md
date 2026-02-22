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

### 4. Make 명령어 사용

```bash
# 개발 서버 실행 (Swagger 포함)
make runWithSwagger

# 프로덕션 빌드
make build-compressed

# Docker 빌드 및 실행
make docker-build
make docker-run
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
export LOG_FORMAT=json
```

### 고급 설정

```bash
# 타임아웃 설정
export HEALTH_CHECK_TIMEOUT=5s
export REQUEST_TIMEOUT=30s

# 캐시 설정 (TTL 기반)
export CACHE_TTL=30m
export CACHE_CAPACITY=100
```

## 첫 번째 API 호출

### 1. 서버 상태 확인

```bash
curl http://localhost:9091/api/v1/health
```

### 2. Kubernetes 클러스터 연결 확인

```bash
curl -X POST http://localhost:9091/api/v1/kubernetes/health \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }'
```

### 3. MinIO 연결 확인

```bash
curl -X POST http://localhost:9091/api/v1/minio/health \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin123",
    "useSSL": false
  }'
```

### 4. Velero 연결 확인

```bash
curl -X POST http://localhost:9091/api/v1/velero/health \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig",
    "minio": {
      "endpoint": "localhost:9000",
      "accessKey": "minioadmin",
      "secretKey": "minioadmin123",
      "useSSL": false
    }
  }'
```

## API 문서 확인

### Swagger UI 접근

- **로컬**: http://localhost:9091/docs
- **온라인**: https://taking.github.io/kubemigrate/

### Bruno 컬렉션 사용

1. [Bruno](https://www.usebruno.com/) 설치
2. `.bruno/` 폴더를 Bruno에서 열기
3. 환경 변수 설정:
   - `{{base_url}}`: http://localhost:9091
   - `{{base64_local_kubeconfig}}`: base64 인코딩된 kubeconfig
4. API 테스트 실행

## 일반적인 사용 사례

### 1. Kubernetes 리소스 조회

```bash
# Pod 목록 조회
curl -X GET "http://localhost:9091/api/v1/kubernetes/pods" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig",
    "namespace": "default"
  }'

# 특정 Pod 조회
curl -X GET "http://localhost:9091/api/v1/kubernetes/pods/my-pod" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }'
```

### 2. MinIO 버킷 관리

```bash
# 버킷 목록 조회
curl -X GET "http://localhost:9091/api/v1/minio/buckets" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin123",
    "useSSL": false
  }'

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

### 3. Helm 차트 설치

```bash
# Helm 차트 설치 (비동기)
curl -X POST "http://localhost:9091/api/v1/helm/charts" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }' \
  -G \
  --data-urlencode "releaseName=nginx" \
  --data-urlencode "chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz" \
  --data-urlencode "version=15.4.2" \
  --data-urlencode "namespace=default"
```

### 4. Velero 백업 관리

```bash
# Velero 설치 (비동기)
curl -X POST "http://localhost:9091/api/v1/velero/install?namespace=velero&force=false" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig",
    "minio": {
      "endpoint": "localhost:9000",
      "accessKey": "minioadmin",
      "secretKey": "minioadmin123",
      "useSSL": false
    }
  }'

# 백업 목록 조회
curl -X GET "http://localhost:9091/api/v1/velero/backups" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }'
```

## 문제 해결

### 일반적인 문제

1. **연결 실패**: kubeconfig가 올바른지 확인
2. **인증 실패**: MinIO 자격 증명 확인
3. **타임아웃**: 네트워크 연결 및 클러스터 상태 확인

### 로그 확인

```bash
# 서버 로그 확인
tail -f logs/kubemigrate.log

# 디버그 모드 실행
LOG_LEVEL=debug ./kubemigrate
```

### 성능 모니터링

```bash
# 캐시 통계 확인
curl http://localhost:9091/api/v1/health

# 메모리 사용량 확인
curl http://localhost:9091/api/v1/health | jq '.memory'
```

## 다음 단계

- [API 문서](API.md) - 모든 엔드포인트 상세 설명
- [개발 가이드](DEVELOPMENT.md) - 개발 환경 설정
- [예제 코드](../example/) - 실제 사용 예제들