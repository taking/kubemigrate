# Velero API Server

Velero API Server는 Kubernetes 및 Velero 기반 백업/복원 관리, Helm 및 MinIO 연동을 지원하는 REST API 서버입니다.  
멀티 클러스터 환경에서도 사용 가능하며, 백업/복원 상태 확인, 스토리지 검증, Helm 차트 설치 여부 확인 등의 기능을 제공합니다.

---

## 주요 기능

### Kubernetes
- 클러스터 연결 상태 확인 (`/kube/health`)
- 특정 네임스페이스 Pod 목록 조회 (`/kube/pods`)
- 스토리지 클래스 조회 (`/kube/storage-classes`)

### Velero
- Velero 연결 상태 확인 (`/velero/health`)
- Backup 목록 조회 (`/velero/backups`)
- Restore 목록 조회 (`/velero/restores`)
- BackupRepository 조회 (`/velero/backup-repositories`)
- BackupStorageLocation 조회 (`/velero/backup-storage-locations`)
- VolumeSnapshotLocation 조회 (`/velero/volume-snapshot-locations`)

### Helm
- Helm 연결 상태 확인 (`/helm/health`)
- Helm 차트 설치 여부 확인 (`/helm/chart_check`)
- Helm 차트 설치 (`/helm/install`)

### MinIO
- MinIO 연결 상태 확인 (`/minio/health`)
- MinIO 버킷 확인 및 없으면 생성 (`/minio/bucket_check`)

---

## 설치 및 실행

```bash
# 레포지토리 클론
git clone <repository_url>
cd velero-api-server

# 의존성 설치
go mod tidy

# 서버 실행
go run main.go
```

### 환경 변수 설정 (선택)

| 환경 변수 | 설명 | 기본값 |
|-----------|------|--------|
| PORT | 서버 포트 | 9091 |
| READ_TIMEOUT | 요청 읽기 타임아웃 | 30s |
| WRITE_TIMEOUT | 응답 쓰기 타임아웃 | 30s |
| IDLE_TIMEOUT | 연결 유지 타임아웃 | 120s |
| HEALTH_CHECK_TIMEOUT | 헬스체크 요청 타임아웃 | 5s |
| REQUEST_TIMEOUT | 일반 API 요청 타임아웃 | 30s |
| LOG_LEVEL | 로그 레벨 | info |
| LOG_FORMAT | 로그 포맷 | json |

---

## API 구조

### 공통

- `/api/v1/health` : 서버 상태 및 기능 확인
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

### Kubernetes
- `/api/v1/kube/health` : Kubernetes 연결 확인
- `/api/v1/kube/pods` : Pod 목록 조회
- `/api/v1/kube/storage-classes` : StorageClass 조회

### Velero
- `/api/v1/velero/health` : Velero 연결 확인
- `/api/v1/velero/backups` : Backup 목록 조회
- `/api/v1/velero/restores` : Restore 목록 조회
- `/api/v1/velero/backup-repositories` : BackupRepository 조회
- `/api/v1/velero/backup-storage-locations` : BackupStorageLocation 조회
- `/api/v1/velero/volume-snapshot-locations` : VolumeSnapshotLocation 조회

### Helm
- `/api/v1/helm/health` : Helm 연결 확인
- `/api/v1/helm/chart_check` : 특정 Helm 차트 설치 여부 확인
- `/api/v1/helm/install` : Helm 차트 설치

### MinIO
- `/api/v1/minio/health` : MinIO 연결 확인
- `/api/v1/minio/bucket_check` : 버킷 존재 여부 확인 및 생성

---

## Swagger 문서

Swagger UI(`http://localhost:9091/docs`)를 통해 API 문서를 확인할 수 있습니다.

`docs/swagger.json` 파일을 기반으로 API 요청/응답 구조와 모델 정의를 제공합니다.

#### 주요 모델
- `KubeConfigRequest` : Kubernetes 연결 정보
- `HelmConfigRequest` : Helm 연결 정보
- `HelmInstallChartRequest` : Helm 차트 설치 요청
- `MinioConfigRequest` : MinIO 연결 정보
- `SwaggerSuccessResponse` : 성공 응답
- `SwaggerErrorResponse` : 에러 응답

---

## 미들웨어

- Request ID 생성
- 로깅 (JSON 형식)
- Panic 복구
- CORS 설정 (모든 도메인 허용)
- Gzip 압축
- 요청 타임아웃
- 레이트 제한 (초당 100 요청, 최대 버스트 50, 1분간 만료)

---

## 인터페이스 정의

### HealthChecker
- `HealthCheck(ctx context.Context) error` : 리소스/클러스터 연결 상태 확인

### KubernetesClient
- `GetPods(ctx context.Context)` : Pod 목록 조회
- `GetStorageClasses(ctx context.Context)` : StorageClass 목록 조회

### VeleroClient
- Velero 관련 CRUD 메서드 제공
- Backup, Restore, StorageLocation, VolumeSnapshotLocation, PodVolumeRestore 등

### HelmClient
- Helm 차트 설치, 설치 여부 확인, 캐시 무효화

### MinioClient
- 버킷 생성 및 확인

---

## 라이선스

Apache 2.0  
[Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html)

