# KubeMigrate

KubeMigrate는 Kubernetes 클러스터 간 백업 및 복구 검증을 위한 종합적인 API 서버입니다.  
멀티 클러스터 환경에서 Velero 기반 백업/복원 관리, Helm 및 MinIO 연동을 지원하며, 스토리지 클래스 비교 검증, 백업 무결성 확인, 복구 검증 등의 기능을 제공합니다.

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
├── pkg/                    # 패키지 라이브러리
│   ├── cache/             # 캐시 시스템
│   ├── client/            # 클라이언트 인터페이스
│   ├── config/            # 설정 관리
│   ├── handlers/          # HTTP 핸들러
│   ├── health/            # 헬스체크
│   ├── interfaces/        # 인터페이스 정의
│   ├── middleware/        # 미들웨어
│   ├── models/            # 데이터 모델
│   ├── response/          # 응답 처리
│   ├── router/            # 라우팅
│   ├── services/          # 비즈니스 로직
│   ├── utils/             # 유틸리티
│   └── validator/         # 검증 로직
├── docs/                  # 문서
├── docker/                # Docker 설정
└── nginx/                 # Nginx 설정
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

- **`/api/v1/health`** : 서버 상태 및 기능 확인
  ```json
  {
    "status": "healthy",
    "version": "v1",
    "message": "server, kubernetes, velero all reachable",
    "features": [
      "single-cluster",
      "multi-cluster",
      "backup-migration",
      "storage-validation"
    ]
  }
  ```

### 🔧 Kubernetes API

- **`/api/v1/kube/health`** : Kubernetes 클러스터 연결 확인
- **`/api/v1/kube/pods`** : Pod 목록 조회
- **`/api/v1/kube/storage-classes`** : StorageClass 조회 및 비교

### 📦 Velero API

- **`/api/v1/velero/health`** : Velero 연결 확인
- **`/api/v1/velero/backups`** : Backup 목록 조회
- **`/api/v1/velero/restores`** : Restore 목록 조회
- **`/api/v1/velero/backup-repositories`** : BackupRepository 조회
- **`/api/v1/velero/backup-storage-locations`** : BackupStorageLocation 조회
- **`/api/v1/velero/volume-snapshot-locations`** : VolumeSnapshotLocation 조회

### ⚙️ Helm API

- **`/api/v1/helm/health`** : Helm 연결 확인
- **`/api/v1/helm/chart_check`** : 특정 Helm 차트 설치 여부 확인
- **`/api/v1/helm/chart_install`** : Helm 차트 설치

### 🗄️ MinIO API

- **`/api/v1/minio/health`** : MinIO 연결 확인
- **`/api/v1/minio/bucket_check`** : 버킷 존재 여부 확인 및 생성