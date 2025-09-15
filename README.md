# KubeMigrate

KubeMigrate는 Kubernetes 클러스터 간 백업 및 복구 검증을 위한 종합적인 API 서버입니다.  
멀티 클러스터 환경에서 Velero 기반 백업/복원 관리, Helm 및 MinIO 연동을 지원하며, 스토리지 클래스 비교 검증, 백업 무결성 확인, 복구 검증 등의 기능을 제공합니다.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

## ✨ 주요 기능

- **🔧 Kubernetes 관리**: 클러스터 리소스 조회 및 관리
- **💾 Velero 통합**: 백업/복원 작업 모니터링 및 관리
- **📦 Helm 지원**: URL 기반 차트 설치, 업그레이드, 관리
- **🗄️ MinIO 연동**: 객체 스토리지 버킷 및 파일 관리
- **🚀 RESTful API**: 일관된 API 디자인으로 쉬운 통합
- **📚 Swagger 문서**: 자동 생성된 API 문서
- **🧪 Bruno 컬렉션**: 포함된 API 테스트 도구
- **⚡ 고성능**: LRU 캐시 및 워커 풀을 통한 최적화
- **📊 모니터링**: 메모리 사용량 및 성능 모니터링

---

## 📖 문서

- [🚀 빠른 시작 가이드](docs/QUICKSTART.md) - 설치부터 첫 API 호출까지
- [📚 API 문서](docs/API.md) - 모든 엔드포인트 상세 설명
- [🛠️ 개발 가이드](docs/DEVELOPMENT.md) - 개발 환경 설정 및 기여 방법
- [📋 코드 리뷰 체크리스트](docs/CODE_REVIEW_CHECKLIST.md) - 코드 품질 관리 가이드

## 🎯 클라이언트 라이브러리

### Go SDK
- [Kubernetes 클라이언트](pkg/client/kubernetes/README.md) - Kubernetes 리소스 관리
- [Helm 클라이언트](pkg/client/helm/README.md) - Helm 차트 관리
- [MinIO 클라이언트](pkg/client/minio/README.md) - 객체 스토리지 관리
- [Velero 클라이언트](pkg/client/velero/README.md) - 백업/복원 관리
- [예제 코드](example/README.md) - 실제 사용 예제들

---

## 🛠️ 설치 및 실행

### 사전 요구사항
- Go 1.24.5 이상
- Kubernetes 클러스터 접근 권한
- Helm
- MinIO
- Velero CRD

### 설치 방법

```bash
# 레포지토리 클론
git clone https://github.com/taking/kubemigrate.git
```

### 실행 방법

```bash
# kubemigrate 폴더 이동
cd kubemigrate

# 서버 실행
make runWithSwagger
```

### 빌드 및 기타
```bash
# 의존성 업데이트
make deps

# 빌드
make build-compressed

# 코드 포맷팅
make format

# 린트 검사
make lint

# Swagger 업데이트
make swagger
```

### Docker를 이용한 실행

```bash
# Docker 이미지 빌드
make docker-build

# 컨테이너 실행
make docker-run

# 또는 docker-compose 사용
make compose-up
```

---

## 🔧 개발

### 프로젝트 구조
```
kubemigrate/
├── cmd/                    # 메인 애플리케이션
├── internal/               # 내부 패키지
│   ├── api/               # API 핸들러
│   │   ├── helm/          # Helm API 핸들러
│   │   ├── kubernetes/    # Kubernetes API 핸들러
│   │   ├── minio/         # MinIO API 핸들러
│   │   └── velero/        # Velero API 핸들러
│   ├── config/            # 설정 관리
│   ├── errors/            # 에러 정의
│   ├── handler/           # 기본 핸들러
│   ├── logger/            # 로깅
│   ├── middleware/        # 미들웨어
│   ├── response/          # 응답 처리
│   ├── server/            # 서버 설정
│   └── validator/         # 검증 로직
├── pkg/                    # 공개 패키지
│   ├── client/            # 클라이언트 인터페이스
│   │   ├── helm/          # Helm 클라이언트
│   │   ├── kubernetes/    # Kubernetes 클라이언트
│   │   ├── minio/         # MinIO 클라이언트
│   │   └── velero/        # Velero 클라이언트
│   ├── response/          # 응답 타입
│   └── utils/             # 유틸리티
├── docs/                  # 문서 (Swagger)
├── docker/                # Docker 설정
├── .bruno/                # Bruno API 컬렉션
└── example/               # 사용 예제
```

---

## ⚙️ 환경 변수 설정

| 환경 변수 | 설명 | 기본값 |
|-----------|------|--------|
| `SERVER_HOST` | 서버 주소 | `localhost` |
| `SERVER_PORT` | 서버 포트 | `9091` |
| `READ_TIMEOUT` | 요청 읽기 타임아웃 | `30s` |
| `WRITE_TIMEOUT` | 응답 쓰기 타임아웃 | `30s` |
| `IDLE_TIMEOUT` | 연결 유지 타임아웃 | `120s` |
| `HEALTH_CHECK_TIMEOUT` | 헬스체크 요청 타임아웃 | `5s` |
| `REQUEST_TIMEOUT` | 일반 API 요청 타임아웃 | `30s` |
| `LOG_LEVEL` | 로그 레벨 | `info` |
| `LOG_FORMAT` | 로그 포맷 | `json` |

---

## 📚 API 구조

- **Swagger UI**: [https://taking.github.io/kubemigrate/](https://taking.github.io/kubemigrate/)
- **로컬 실행**: http://localhost:9091/docs

### 🔍 공통 엔드포인트

- **`GET /`** : 서버 기본 정보
- **`GET /api/v1/health`** : API 서버 상태 확인

### 🔧 Kubernetes API (`/api/v1/kubernetes`)

- **`POST /health`** : Kubernetes 클러스터 연결 확인
- **`GET /:kind`** : 리소스 목록 조회 (pods, services, deployments 등)
- **`GET /:kind/:name`** : 특정 리소스 조회

### 📦 Velero API (`/api/v1/velero`)

- **`POST /health`** : Velero 연결 확인
- **`POST /backups`** : Backup 목록 조회
- **`POST /restores`** : Restore 목록 조회
- **`GET /repositories`** : BackupRepository 조회
- **`GET /storage-locations`** : BackupStorageLocation 조회
- **`GET /volume-snapshot-locations`** : VolumeSnapshotLocation 조회
- **`GET /pod-volume-restores`** : PodVolumeRestore 조회

### ⚙️ Helm API (`/api/v1/helm`)

- **`POST /health`** : Helm 연결 확인
- **`POST /charts`** : Helm 차트 설치 (URL 기반)
- **`GET /charts`** : 차트 목록 조회
- **`GET /charts/:name`** : 특정 차트 상세 조회
- **`GET /charts/:name/status`** : 차트 설치 상태 확인
- **`PUT /charts/:name`** : 차트 업그레이드
- **`GET /charts/:name/history`** : 차트 히스토리 조회
- **`GET /charts/:name/values`** : 차트 값 조회
- **`DELETE /charts/:name`** : 차트 제거

### 🗄️ MinIO API (`/api/v1/minio`)

- **`POST /health`** : MinIO 연결 확인

#### 버킷 관리
- **`GET /buckets`** : 버킷 목록 조회
- **`GET /buckets/:bucket`** : 버킷 존재 확인
- **`POST /buckets/:bucket`** : 버킷 생성
- **`DELETE /buckets/:bucket`** : 버킷 삭제

#### 객체 관리
- **`GET /buckets/:bucket/objects`** : 객체 목록 조회
- **`POST /buckets/:bucket/objects/:objectName`** : 객체 업로드
- **`GET /buckets/:bucket/objects/:objectName`** : 객체 다운로드
- **`GET /buckets/:bucket/objects/:objectName`** : 객체 정보 조회
- **`POST /buckets/:srcBucket/objects/:srcObject/copy/:dstBucket/:dstObject`** : 객체 복사
- **`DELETE /buckets/:bucket/objects/:objectName`** : 객체 삭제

#### Presigned URL
- **`GET /buckets/:bucket/objects/:objectName/presigned-get`** : Presigned GET URL 생성
- **`PUT /buckets/:bucket/objects/:objectName/presigned-put`** : Presigned PUT URL 생성

---

## 🚀 사용 예제

### Helm 차트 설치 (URL 기반)
```bash
curl -X POST "http://localhost:9091/api/v1/helm/charts" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }' \
  -G \
  --data-urlencode "releaseName=my-nginx" \
  --data-urlencode "chartURL=https://charts.bitnami.com/bitnami/nginx-15.4.2.tgz" \
  --data-urlencode "version=15.4.2" \
  --data-urlencode "namespace=default"
```

### MinIO 객체 업로드
```bash
curl -X POST "http://localhost:9091/api/v1/minio/buckets/my-bucket/objects/test-file.txt" \
  -F "file=@/path/to/local/file.txt" \
  -F 'config={"endpoint":"192.168.1.100:9000","accessKey":"admin","secretKey":"password","useSSL":false}'
```

### Kubernetes 리소스 조회
```bash
curl -X GET "http://localhost:9091/api/v1/kubernetes/pods" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig",
    "namespace": "default"
  }'
```

### Velero 백업 목록 조회
```bash
curl -X POST "http://localhost:9091/api/v1/velero/backups" \
  -H "Content-Type: application/json" \
  -d '{
    "kubeconfig": "base64_encoded_kubeconfig"
  }'
```

---

## 🧪 API 테스트

프로젝트에는 Bruno API 컬렉션이 포함되어 있어 쉽게 API를 테스트할 수 있습니다:

1. **Bruno 설치**: [Bruno 공식 사이트](https://www.usebruno.com/)에서 다운로드
2. **컬렉션 열기**: `.bruno/` 폴더를 Bruno에서 열기
3. **환경 변수 설정**: `{{base_url}}`, `{{base64_local_kubeconfig}}` 등 설정
4. **API 테스트**: 각 서비스별로 분류된 요청들을 실행

### Bruno 컬렉션 구조
```
.bruno/
├── 1_kube/          # Kubernetes API 테스트
├── 2_minio/         # MinIO API 테스트  
├── 3_helm/          # Helm API 테스트
└── velero/          # Velero API 테스트
```