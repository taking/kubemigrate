# KubeMigrate Go SDK Examples

이 폴더는 KubeMigrate Go SDK의 사용 예제들을 포함하고 있습니다. 각 예제는 특정 기능을 보여주며, 실제 프로젝트에서 어떻게 사용할 수 있는지 설명합니다.

## 📋 목차

- [설치](#설치)
- [기본 사용법](#기본-사용법)
- [예제 파일들](#예제-파일들)
- [API 문서](#api-문서)
- [문제 해결](#문제-해결)

## 🚀 설치

### Go 모듈로 설치

```bash
go get github.com/taking/kubemigrate
```

### go.mod에 추가

```go
module your-project

go 1.21

require (
    github.com/taking/kubemigrate v0.1.0
)
```

## 📖 기본 사용법

### 1. Helm 클라이언트 사용

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/taking/kubemigrate/pkg/client/helm"
)

func main() {
    // Helm 클라이언트 생성
    client := helm.NewClient()
    
    // 차트 목록 조회
    charts, err := client.GetCharts("default")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d charts\n", len(charts))
}
```

### 2. Kubernetes 클라이언트 사용

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/taking/kubemigrate/pkg/client/kubernetes"
)

func main() {
    // Kubernetes 클라이언트 생성
    client := kubernetes.NewClient()
    
    // Pod 목록 조회
    pods, err := client.GetPods("default")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d pods\n", len(pods))
}
```

### 3. MinIO 클라이언트 사용

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/taking/kubemigrate/pkg/client/minio"
)

func main() {
    // MinIO 클라이언트 생성
    client := minio.NewClient()
    
    // 버킷 목록 조회
    buckets, err := client.ListBuckets()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d buckets\n", len(buckets))
}
```

### 4. Velero 클라이언트 사용

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/taking/kubemigrate/pkg/client/velero"
)

func main() {
    // Velero 클라이언트 생성
    client := velero.NewClient()
    
    // 백업 목록 조회
    backups, err := client.GetBackups("default")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d backups\n", len(backups))
}
```

### 5. 유틸리티 함수 사용

```go
package main

import (
    "fmt"
    
    "github.com/taking/kubemigrate/pkg/utils"
)

func main() {
    // 문자열 기본값 처리
    result := utils.GetStringOrDefault("", "default value")
    fmt.Println(result) // "default value"
    
    // 정수 변환
    num := utils.StringToIntOrDefault("123", 0)
    fmt.Println(num) // 123
    
    // 불린 변환
    flag := utils.StringToBoolOrDefault("true", false)
    fmt.Println(flag) // true
}
```

## 📁 예제 파일들

| 디렉토리 | 설명 | 주요 기능 |
|----------|------|-----------|
| `helm/` | Helm 클라이언트 사용 예제 | 차트 설치, 삭제, 목록 조회 |
| `kubernetes/` | Kubernetes 클라이언트 사용 예제 | Pod, Service, StorageClass 조회 |
| `minio/` | MinIO 클라이언트 사용 예제 | 버킷 관리, 객체 업로드/다운로드 |
| `velero/` | Velero 클라이언트 사용 예제 | 백업 생성, 복원, 목록 조회 |
| `utils/` | 유틸리티 함수 사용 예제 | 문자열, 숫자, 불린 변환 |
| `complete/` | 통합 사용 예제 | 모든 클라이언트를 함께 사용 |

### 예제 실행 방법

각 예제는 독립적인 Go 모듈로 구성되어 있습니다:

```bash
# Helm 예제 실행
cd example/helm
go run main.go

# Kubernetes 예제 실행
cd example/kubernetes
go run main.go

# MinIO 예제 실행
cd example/minio
go run main.go

# Velero 예제 실행
cd example/velero
go run main.go

# Utils 예제 실행
cd example/utils
go run main.go

# 통합 예제 실행
cd example/complete
go run main.go
```

## 🔧 설정

### 환경 변수

```bash
# Kubernetes 설정
export KUBECONFIG=/path/to/your/kubeconfig

# MinIO 설정
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin

# Velero 설정
export VELERO_NAMESPACE=velero
```

### 설정 파일

```yaml
# config.yaml
kubernetes:
  kubeconfig: "/path/to/kubeconfig"
  namespace: "default"

minio:
  endpoint: "localhost:9000"
  accessKey: "minioadmin"
  secretKey: "minioadmin"
  useSSL: false

velero:
  namespace: "velero"
```

## 📚 API 문서

### Helm 클라이언트

```go
type HelmClient interface {
    // 차트 목록 조회
    GetCharts(namespace string) ([]*release.Release, error)
    
    // 특정 차트 조회
    GetChart(releaseName, namespace string) (*release.Release, error)
    
    // 차트 설치
    InstallChart(chartName, releaseName, namespace string, values map[string]interface{}) error
    
    // 차트 삭제
    DeleteChart(releaseName, namespace string, dryRun bool) error
    
    // 차트 설치 여부 확인
    IsChartInstalled(releaseName, namespace string) (bool, error)
    
    // 헬스 체크
    HealthCheck() error
}
```

### Kubernetes 클라이언트

```go
type KubernetesClient interface {
    // Pod 목록 조회
    GetPods(namespace string) ([]v1.Pod, error)
    
    // Service 목록 조회
    GetServices(namespace string) ([]v1.Service, error)
    
    // StorageClass 목록 조회
    GetStorageClasses() ([]storagev1.StorageClass, error)
    
    // 헬스 체크
    HealthCheck() error
}
```

### MinIO 클라이언트

```go
type MinioClient interface {
    // 버킷 목록 조회
    ListBuckets() ([]minio.BucketInfo, error)
    
    // 버킷 존재 여부 확인
    BucketExists(bucketName string) (bool, error)
    
    // 객체 업로드
    PutObject(bucketName, objectName string, reader io.Reader, objectSize int64) error
    
    // 객체 다운로드
    GetObject(bucketName, objectName string) (io.Reader, error)
    
    // 헬스 체크
    HealthCheck() error
}
```

### Velero 클라이언트

```go
type VeleroClient interface {
    // 백업 목록 조회
    GetBackups(namespace string) ([]velerov1.Backup, error)
    
    // 백업 저장소 목록 조회
    GetBackupRepositories(namespace string) ([]velerov1.BackupRepository, error)
    
    // 헬스 체크
    HealthCheck() error
}
```

## 🛠️ 문제 해결

### 일반적인 문제들

#### kubernetes 연결 실패

```bash
# kubeconfig 파일 확인
kubectl config current-context

# 클러스터 연결 테스트
kubectl cluster-info
```

#### minio 연결 실패

```bash
# MinIO 서버 상태 확인
curl http://localhost:9000/minio/health/live

# 접근 키/시크릿 키 확인
echo $MINIO_ACCESS_KEY
echo $MINIO_SECRET_KEY
```

#### 3. Velero 연결 실패

```bash
# Velero 네임스페이스 확인
kubectl get pods -n velero

# Velero 서버 상태 확인
kubectl logs -n velero deployment/velero
```

### 로그 활성화

```go
import (
    "log"
    "os"
)

func main() {
    // 디버그 로그 활성화
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetOutput(os.Stdout)
    
    // 클라이언트 사용...
}
```

### 에러 처리

```go
func handleError(err error) {
    if err != nil {
        log.Printf("Error: %v", err)
        // 적절한 에러 처리 로직
    }
}
```

## 🤝 기여하기

버그 리포트나 기능 요청은 [GitHub Issues](https://github.com/taking/kubemigrate/issues)에 등록해주세요.

## 📄 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다.

## 🔗 관련 링크

- [KubeMigrate GitHub](https://github.com/taking/kubemigrate)
- [Kubernetes 공식 문서](https://kubernetes.io/docs/)
- [Helm 공식 문서](https://helm.sh/docs/)
- [MinIO 공식 문서](https://docs.min.io/)
- [Velero 공식 문서](https://velero.io/docs/)
