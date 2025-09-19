# KubeMigrate

Kubernetes 리소스 마이그레이션을 위한 통합 관리 도구

## 개요

KubeMigrate는 Kubernetes 클러스터 간 리소스 마이그레이션을 위한 통합 관리 도구입니다. Velero, MinIO, Helm을 활용하여 백업, 복원, 차트 관리를 제공합니다.

## 주요 기능

### 1. Kubernetes 리소스 관리
- Pod, ConfigMap, Secret, Namespace, StorageClass 조회
- 리소스별 상세 정보 및 목록 조회
- 네임스페이스별 필터링 지원

### 2. Velero 백업/복원
- Velero 설치 및 MinIO 연동 설정
- 백업 생성 및 관리
- 복원 작업 수행
- BackupStorageLocation 관리
- VolumeSnapshotLocation 관리

### 3. MinIO 객체 저장소
- 버킷 생성 및 관리
- 객체 업로드/다운로드
- Presigned URL 생성
- 객체 메타데이터 조회

### 4. Helm 차트 관리
- 차트 설치/업그레이드/제거
- 차트 목록 조회
- 차트 Values 조회
- 차트 상태 확인

## API 엔드포인트

### Kubernetes API
- `POST /v1/kubernetes/health` - Kubernetes 연결 테스트
- `GET /v1/kubernetes/resources` - Kubernetes 리소스 조회

### Velero API
- `POST /v1/velero/health` - Velero 연결 테스트
- `POST /v1/velero/install` - Velero 설치 및 MinIO 연동 (비동기)
- `GET /v1/velero/backups` - 백업 목록 조회
- `GET /v1/velero/restores` - 복원 목록 조회
- `GET /v1/velero/storage-locations` - 저장소 위치 조회
- `GET /v1/velero/status/{jobId}` - 작업 상태 조회
- `GET /v1/velero/status/{jobId}/logs` - 작업 로그 조회

### MinIO API
- `POST /v1/minio/health` - MinIO 연결 테스트
- `POST /v1/minio/buckets` - 버킷 목록 조회
- `POST /v1/minio/buckets/{bucket}/objects` - 객체 목록 조회
- `PUT /v1/minio/buckets/{bucket}/objects/{object}` - 객체 업로드

### Helm API
- `POST /v1/helm/health` - Helm 연결 테스트
- `GET /v1/helm/charts` - 차트 목록 조회
- `POST /v1/helm/charts` - 차트 설치 (비동기)
- `PUT /v1/helm/charts/{name}` - 차트 업그레이드 (비동기)
- `DELETE /v1/helm/charts/{name}` - 차트 제거 (비동기)
- `GET /v1/helm/charts/{name}/status` - 차트 상태 조회
- `GET /v1/helm/charts/{name}/values` - 차트 Values 조회
- `GET /v1/helm/jobs/{jobId}` - 작업 상태 조회

## 설정

### 환경변수
- `SERVER_HOST` - 서버 호스트 (기본값: localhost)
- `SERVER_PORT` - 서버 포트 (기본값: 8080)
- `LOG_LEVEL` - 로그 레벨 (기본값: info)
- `LOG_FORMAT` - 로그 포맷 (기본값: pretty)

### 설정 파일
`.env` 파일을 통해 환경변수를 설정할 수 있습니다.

## 설치 및 실행

### 빌드
```bash
go build -o bin/kubemigrate-cli cmd/main.go
```

### 실행
```bash
./bin/kubemigrate-cli
```

## 개발

### 프로젝트 구조
```
├── cmd/                    # 메인 애플리케이션
├── internal/               # 내부 패키지
│   ├── api/               # API 핸들러 (kubernetes, minio, helm, velero)
│   ├── handler/           # 공통 핸들러 (BaseHandler)
│   ├── validator/         # 검증 로직 (ValidationManager)
│   ├── response/          # 응답 처리 (ResponseManager)
│   ├── job/               # 작업 관리 (JobManager, WorkerPool)
│   ├── installer/         # 설치 로직 (VeleroInstaller)
│   ├── cache/             # 캐시 관리 (LRU Cache)
│   └── mocks/            # Mock 클라이언트
├── pkg/                   # 공용 패키지
│   ├── client/           # 클라이언트 인터페이스 (kubernetes, minio, helm, velero)
│   ├── config/           # 설정 관리 (ConfigManager)
│   ├── types/            # 타입 정의 (kubernetes, minio, helm, velero)
│   └── utils/            # 유틸리티 함수
└── docs/                 # 문서
```

### Layered Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   Kubernetes│ │    MinIO    │ │    Helm     │ │  Velero │ │
│  │   Handler   │ │   Handler   │ │   Handler   │ │ Handler │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
└──────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────┐
│                     Business Layer                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   Kubernetes│ │    MinIO    │ │    Helm     │ │  Velero │ │
│  │   Service   │ │   Service   │ │   Service   │ │ Service │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Common Services                            │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌────────────────────┐ │ │
│  │  │   Job       │ │  Installer  │ │    Response        │ │ │
│  │  │  Manager    │ │   Service   │ │    Manager         │ │ │
│  │  └─────────────┘ └─────────────┘ └────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
┌──────────────────────────────────────────────────────────────┐
│                      Data Layer                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   Kubernetes│ │    MinIO    │ │    Helm     │ │  Velero │ │
│  │   Client    │ │   Client    │ │   Client    │ │  Client │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Infrastructure                             │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌────────────────────┐ │ │
│  │  │   Config    │ │   Cache     │ │    Validation      │ │ │
│  │  │  Manager    │ │  Manager    │ │    Manager         │ │ │
│  │  └─────────────┘ └─────────────┘ └────────────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### 테스트
```bash
go test ./...
```

### Swagger 문서
Swagger 문서는 `/docs/swagger-ui.html`에서 확인할 수 있습니다.

## 라이선스

MIT License
