# KubeMigrate

KubeMigrate는 Kubernetes 클러스터 간 백업 및 복구 검증을 위한 종합적인 API 서버입니다.  
멀티 클러스터 환경에서 Velero 기반 백업/복원 관리, Helm 및 MinIO 연동을 지원하며, 스토리지 클래스 비교 검증, 백업 무결성 확인, 복구 검증 등의 기능을 제공합니다.

---

## 📚 API 문서

- **Swagger UI**: [https://taking.github.io/kubemigrate/](https://taking.github.io/kubemigrate/)
- **로컬 실행**: http://localhost:9091/docs

## 🚀 CI/CD 파이프라인

### 자동화된 워크플로우
- **Pull Request**: 테스트, 린터, 빌드 검증
- **Main 브랜치 Push**: Swagger 문서 생성, Docker 빌드, GitHub Pages 배포
- **태그 릴리스**: 릴리스 노트 자동 생성, GitHub Packages 푸시

### GitHub Packages (GHCR)
- **이미지**: `ghcr.io/taking/kubemigrate:latest`
- **태그**: `latest`, `main`, `v1.0.0` 등
- **지원 아키텍처**: `linux/amd64`, `linux/arm64`, `linux/arm/v7`
- **특징**: UPX 압축으로 89% 크기 감소

## 🚀 주요 기능

> baseUrl : http://localhost:9091/api/v1

### 🔧 Kubernetes 클러스터 관리
- **클러스터 연결 상태 확인** (`/kube/health`)
- **특정 네임스페이스 Pod 목록 조회** (`/kube/pods`)
- **스토리지 클래스 조회 및 비교** (`/kube/storage-classes`)
- **멀티 클러스터 환경 지원**

### 📦 Velero 백업/복구 관리
- **Velero 연결 상태 확인** (`/velero/health`)
- **Backup 목록 조회 및 검증** (`/velero/backups`)
- **Restore 목록 조회 및 검증** (`/velero/restores`)
- **BackupRepository 상태 확인** (`/velero/backup-repositories`)
- **BackupStorageLocation 관리** (`/velero/backup-storage-locations`)
- **VolumeSnapshotLocation 관리** (`/velero/volume-snapshot-locations`)

### ⚙️ Helm 차트 관리
- **Helm 연결 상태 확인** (`/helm/health`)
- **Helm 차트 설치 여부 확인** (`/helm/chart_check`)
- **Helm 차트 자동 설치** (`/helm/chart_install`)
- **Velero Helm 차트 관리**

### 🗄️ MinIO 스토리지 관리
- **MinIO 연결 상태 확인** (`/minio/health`)
- **MinIO 버킷 확인 및 자동 생성** (`/minio/bucket_check`)
- **스토리지 백엔드 검증**

---

## 🛠️ 설치 및 실행

### 사전 요구사항
- Go 1.24.5 이상
- Kubernetes 클러스터 접근 권한
- Velero CLI (선택사항)
- Helm CLI (선택사항)
- MinIO 서버 (선택사항)

### 설치 방법

```bash
# 레포지토리 클론
git clone https://github.com/taking/kubemigrate.git
cd kubemigrate

# 의존성 설치
go mod tidy

# 서버 실행
go run cmd/main.go
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

### GitHub Packages에서 Docker 이미지 사용

```bash
# 멀티 아키텍처 이미지 pull (자동으로 플랫폼 감지)
docker pull ghcr.io/taking/kubemigrate:latest

# 컨테이너 실행
docker run --rm -p 9091:9091 ghcr.io/taking/kubemigrate:latest

# 아키텍처 확인
docker run --rm ghcr.io/taking/kubemigrate:latest uname -m

# 특정 아키텍처 pull (선택사항)
docker pull --platform linux/arm64 ghcr.io/taking/kubemigrate:latest
```

> **참고**: 
> - GitHub Container Registry (GHCR) 사용: `ghcr.io/taking/kubemigrate:latest`
> - 지원 플랫폼: `linux/amd64`, `linux/arm64`, `linux/arm/v7`
> - UPX 압축으로 이미지 크기 최적화 (89% 감소)

---

## ⚙️ 환경 변수 설정

| 환경 변수 | 설명 | 기본값 |
|-----------|------|--------|
| `PORT` | 서버 포트 | `9091` |
| `READ_TIMEOUT` | 요청 읽기 타임아웃 | `30s` |
| `WRITE_TIMEOUT` | 응답 쓰기 타임아웃 | `30s` |
| `IDLE_TIMEOUT` | 연결 유지 타임아웃 | `120s` |
| `HEALTH_CHECK_TIMEOUT` | 헬스체크 요청 타임아웃 | `5s` |
| `REQUEST_TIMEOUT` | 일반 API 요청 타임아웃 | `30s` |
| `LOG_LEVEL` | 로그 레벨 | `info` |
| `LOG_FORMAT` | 로그 포맷 | `json` |

---

## 📚 API 구조

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

---

## 📖 API 문서

### Swagger UI
Swagger UI(`http://localhost:9091/docs`)를 통해 상세한 API 문서를 확인할 수 있습니다.

### 주요 모델
- **`KubeConfigRequest`** : Kubernetes 클러스터 연결 정보
- **`HelmConfigRequest`** : Helm 연결 정보
- **`HelmInstallChartRequest`** : Helm 차트 설치 요청
- **`MinioConfigRequest`** : MinIO 연결 정보
- **`SwaggerSuccessResponse`** : 성공 응답
- **`SwaggerErrorResponse`** : 에러 응답

---

## 🏗️ 아키텍처

### 미들웨어
- **Request ID 생성** : 요청 추적을 위한 고유 ID
- **구조화된 로깅** : JSON 형식의 상세 로그
- **Panic 복구** : 서버 안정성 보장
- **CORS 설정** : 모든 도메인 허용
- **Gzip 압축** : 네트워크 효율성 향상
- **요청 타임아웃** : 리소스 보호
- **레이트 제한** : 초당 100 요청, 최대 버스트 50, 1분간 만료

### 인터페이스 정의

#### HealthChecker
- **`HealthCheck(ctx context.Context) error`** : 리소스/클러스터 연결 상태 확인

#### KubernetesClient
- **`GetPods(ctx context.Context)`** : Pod 목록 조회
- **`GetStorageClasses(ctx context.Context)`** : StorageClass 목록 조회

#### VeleroClient
- Velero 관련 CRUD 메서드 제공
- Backup, Restore, StorageLocation, VolumeSnapshotLocation, PodVolumeRestore 등

#### HelmClient
- Helm 차트 설치, 설치 여부 확인, 캐시 무효화

#### MinioClient
- 버킷 생성 및 확인

---

## 🎯 사용 사례

### 1. 클러스터 마이그레이션 검증
```bash
# Source 클러스터 백업 검증
curl -X GET "http://localhost:9091/api/v1/velero/backups" \
  -H "Content-Type: application/json"

# Destination 클러스터 스토리지 클래스 확인
curl -X GET "http://localhost:9091/api/v1/kube/storage-classes" \
  -H "Content-Type: application/json"
```

### 2. Velero 설치 및 설정
```bash
# Velero Helm 차트 설치 확인
curl -X POST "http://localhost:9091/api/v1/helm/chart_check" \
  -H "Content-Type: application/json" \
  -d '{"chart_name": "velero", "namespace": "velero"}'

# Velero Helm 차트 설치
curl -X POST "http://localhost:9091/api/v1/helm/chart_install" \
  -H "Content-Type: application/json" \
  -d '{"chart_name": "velero", "namespace": "velero"}'
```

### 3. MinIO 스토리지 설정
```bash
# MinIO 버킷 확인 및 생성
curl -X POST "http://localhost:9091/api/v1/minio/bucket_check" \
  -H "Content-Type: application/json" \
  -d '{"bucket_name": "velero-backups"}'
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

### 빌드 및 테스트
```bash
# 의존성 업데이트
go mod tidy

# 빌드
go build -o kubemigrate cmd/main.go

# 테스트 실행
go test ./...

# 린트 검사
golangci-lint run
```

---

## 📄 라이선스

Apache 2.0  
[Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html)

---

## 🤝 기여하기

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📞 지원

- **이슈 리포트**: [GitHub Issues](https://github.com/taking/kubemigrate/issues)
- **문서**: [API Documentation](http://localhost:9091/docs)
- **이메일**: support@taking.kr

---

## 🙏 감사의 말

이 프로젝트는 다음 오픈소스 프로젝트들의 도움을 받았습니다:
- [Velero](https://velero.io/) - Kubernetes 백업 및 복구 도구
- [Echo](https://echo.labstack.com/) - Go 웹 프레임워크
- [Helm](https://helm.sh/) - Kubernetes 패키지 매니저
- [MinIO](https://min.io/) - 오브젝트 스토리지