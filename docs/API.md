# KubeMigrate API 문서

## 개요

KubeMigrate는 Kubernetes 클러스터 마이그레이션 및 백업 검증을 위한 통합 API 서버입니다. Velero, Helm, Kubernetes, MinIO와의 통합을 통해 클러스터 관리 작업을 자동화합니다.

## 기본 정보

- **Base URL**: `http://localhost:9091/api/v1`
- **Content-Type**: `application/json`
- **API Version**: v1

## 인증

현재 API는 인증을 요구하지 않습니다. 프로덕션 환경에서는 적절한 인증 메커니즘을 구현하는 것을 권장합니다.

## 응답 형식

### 성공 응답

```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": { ... },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req-123456"
}
```

### 에러 응답

```json
{
  "status": "error",
  "code": "VALIDATION_ERROR",
  "message": "Invalid input parameters",
  "details": "kubeconfig is required",
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "req-123456"
}
```

## API 엔드포인트

### 1. 헬스체크 및 시스템 관리

#### 서버 상태 확인
```http
GET /api/v1/health
```

#### 캐시 통계 조회
```http
GET /api/v1/cache/stats
```

#### 상세 캐시 통계 조회
```http
GET /api/v1/cache/detailed
```

**응답 예시:**
```json
{
  "status": "success",
  "data": {
    "summary": {
      "active_clients": 2,
      "total_clients": 2,
      "capacity": 100
    },
    "active_clients": [
      {
        "api_type": "kubernetes",
        "cache_key": "a1b2c3d4e5f6...",
        "readable_key": "kubernetes:a1b2c3...f6e5d4",
        "created_at": "2024-01-01T10:30:45Z",
        "age_seconds": 120,
        "config": {
          "kubeconfig": "eyJh...abcd",
          "has_config": true
        }
      }
    ],
    "performance": {
      "hit_rate": 85.5,
      "miss_rate": 14.5,
      "total_hits": 42,
      "total_misses": 7,
      "average_age_seconds": 150.5,
      "oldest_client_seconds": 300,
      "newest_client_seconds": 45
    }
  }
}
```

#### 캐시 정리
```http
# 전체 캐시 정리 (GET 또는 POST)
GET /api/v1/cache/cleanup
POST /api/v1/cache/cleanup

# 특정 캐시 키 정리
DELETE /api/v1/cache/clean/:cache_key
```

**응답 예시:**
```json
{
  "status": "success",
  "message": "Cache cleanup completed"
}
```

**특정 키 정리 응답:**
```json
{
  "status": "success",
  "message": "Cache item removed successfully",
  "data": {
    "cache_key": "abc123...",
    "removed": true
  }
}
```

#### Velero API 설정 예시
Velero API는 Kubernetes와 MinIO 설정을 모두 사용합니다:

**요청:**
```json
{
  "kubeconfig": {
    "kubeconfig": "{{base64_local_kubeconfig}}"
  },
  "minio": {
    "endpoint": "{{minio_url}}",
    "accessKey": "{{minio_accesskey}}",
    "secretKey": "{{minio_secretkey}}",
    "useSSL": false
  }
}
```

**캐시 통계 응답:**
```json
{
  "status": "success",
  "data": {
    "summary": {
      "active_clients": 1,
      "total_clients": 1,
      "capacity": 100
    },
    "active_clients": [
      {
        "api_type": "velero",
        "cache_key": "abc123...",
        "readable_key": "velero:abc123...",
        "created_at": "2024-01-01T10:30:45Z",
        "age_seconds": 120,
        "config": {
          "kubernetes": {
            "kubeconfig": "eyJh...abcd",
            "has_config": true
          },
          "minio": {
            "endpoint": "minio.example.com",
            "access_key": "min****...****admin",
            "secret_key": "min****...****123",
            "use_ssl": false,
            "has_config": true
          },
          "has_config": true
        }
      }
    ]
  }
}
```

#### 메모리 통계 조회
```http
GET /api/v1/memory/stats
```

#### 메모리 최적화
```http
POST /api/v1/memory/optimize
```

#### 메모리 사용률 조회
```http
GET /api/v1/memory/usage
```

### 2. Velero API

#### Velero 연결 상태 확인
```http
POST /api/v1/velero/health
Content-Type: application/json

{
  "kubeconfig": {
    "kubeconfig": "base64-encoded-kubeconfig"
  },
  "minio": {
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin123",
    "useSSL": false
  }
}
```

#### 백업 목록 조회
```http
POST /api/v1/velero/backups
Content-Type: application/json

{
  "kubeconfig": { ... },
  "minio": { ... }
}
```

#### 복원 목록 조회
```http
POST /api/v1/velero/restores
Content-Type: application/json

{
  "kubeconfig": { ... },
  "minio": { ... }
}
```

#### 백업 저장소 조회
```http
GET /api/v1/velero/repositories
```

#### 스토리지 위치 조회
```http
GET /api/v1/velero/storage-locations
```

### 3. Helm API

#### Helm 연결 상태 확인
```http
POST /api/v1/helm/health
Content-Type: application/json

{
  "kubeconfig": "base64-encoded-kubeconfig"
}
```

#### 차트 설치
```http
POST /api/v1/helm/charts?releaseName=my-release&chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz&version=15.4.2&namespace=default
Content-Type: application/json

{
  "kubeconfig": "base64-encoded-kubeconfig"
}
```

#### 차트 목록 조회
```http
GET /api/v1/helm/charts
```

#### 특정 차트 조회
```http
GET /api/v1/helm/charts/{name}
```

#### 차트 업그레이드
```http
PUT /api/v1/helm/charts/{name}
Content-Type: application/json

{
  "kubeconfig": "base64-encoded-kubeconfig",
  "chartPath": "path/to/chart",
  "version": "1.0.0",
  "namespace": "default"
}
```

#### 차트 제거
```http
DELETE /api/v1/helm/charts/{name}
Content-Type: application/json

{
  "kubeconfig": "base64-encoded-kubeconfig"
}
```

### 4. Kubernetes API

#### Kubernetes 연결 상태 확인
```http
POST /api/v1/kubernetes/health
Content-Type: application/json

{
  "kubeconfig": "base64-encoded-kubeconfig"
}
```

#### 리소스 목록 조회
```http
GET /api/v1/kubernetes/{kind}?namespace=default
```

#### 특정 리소스 조회
```http
GET /api/v1/kubernetes/{kind}/{name}?namespace=default
```

**지원하는 리소스 종류:**
- `pods`
- `configmaps`
- `secrets`
- `storage-classes`

### 5. MinIO API

#### MinIO 연결 상태 확인
```http
POST /api/v1/minio/health
Content-Type: application/json

{
  "endpoint": "localhost:9000",
  "accessKey": "minioadmin",
  "secretKey": "minioadmin123",
  "useSSL": false
}
```

#### 버킷 목록 조회
```http
GET /api/v1/minio/buckets
```

#### 버킷 존재 확인
```http
GET /api/v1/minio/buckets/{bucket}
```

#### 버킷 생성
```http
POST /api/v1/minio/buckets/{bucket}
```

#### 버킷 삭제
```http
DELETE /api/v1/minio/buckets/{bucket}
```

#### 객체 목록 조회
```http
GET /api/v1/minio/buckets/{bucket}/objects
```

#### 객체 업로드
```http
POST /api/v1/minio/buckets/{bucket}/objects/{object}
Content-Type: application/octet-stream

[binary data]
```

#### 객체 다운로드
```http
GET /api/v1/minio/buckets/{bucket}/objects/{object}
```

## 에러 코드

| 코드 | 설명 |
|------|------|
| `VALIDATION_ERROR` | 입력 데이터 검증 실패 |
| `CONNECTION_ERROR` | 외부 서비스 연결 실패 |
| `NOT_FOUND` | 요청한 리소스를 찾을 수 없음 |
| `INTERNAL_ERROR` | 서버 내부 오류 |
| `TIMEOUT` | 요청 시간 초과 |

## 사용 예제

### 1. Kubernetes Pod 목록 조회

```bash
curl -X POST "http://localhost:9091/api/v1/kubernetes/health" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64-encoded-kubeconfig"
  }'

curl -X GET "http://localhost:9091/api/v1/kubernetes/pods?namespace=default"
```

### 2. Helm 차트 설치

```bash
curl -X POST "http://localhost:9091/api/v1/helm/charts?releaseName=nginx&chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz&version=15.4.2&namespace=default" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64-encoded-kubeconfig"
  }'
```

### 3. MinIO 버킷 생성

```bash
curl -X POST "http://localhost:9091/api/v1/minio/buckets/my-bucket" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin123",
    "useSSL": false
  }'
```

## 제한사항

- **KubeConfig 크기**: 최대 100KB
- **네임스페이스**: Kubernetes 네이밍 규칙 준수 (최대 63자)
- **Rate Limiting**: 초당 100개 요청, 최대 50개 버스트
- **타임아웃**: 요청당 30초

## 모니터링

API 서버는 다음 메트릭을 제공합니다:

- 메모리 사용량 통계
- 캐시 히트율
- 요청 처리 시간
- 에러율

이러한 메트릭은 `/api/v1/memory/stats`와 `/api/v1/cache/stats` 엔드포인트를 통해 조회할 수 있습니다.
