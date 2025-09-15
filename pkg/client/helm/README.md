# Helm 클라이언트

Kubernetes 클러스터에서 Helm 차트를 관리하기 위한 통합 클라이언트입니다. 차트 설치, 업그레이드, 제거 및 조회 기능을 제공합니다.

## 📋 개요

Helm 클라이언트는 Kubernetes 패키지 매니저인 Helm과 상호작용하기 위한 고수준 인터페이스를 제공합니다. 이 클라이언트는 복잡한 Helm 명령어를 단순화하고, 일관된 에러 처리를 제공하며, 타입 안전성을 보장합니다.

## ✨ 주요 특징

- **📦 차트 관리**: Helm 차트 설치, 업그레이드, 제거
- **🔍 차트 조회**: 설치된 차트 목록 및 상세 정보 조회
- **✅ 설치 확인**: 차트 설치 상태 확인
- **🏥 Health Check**: Helm 연결 상태 확인
- **🏷️ 네임스페이스 지원**: 다중 네임스페이스 지원
- **⚡ 성능 최적화**: 효율적인 차트 관리 및 캐싱
- **🔧 설정 유연성**: 다양한 설정 옵션 지원

## 빠른 시작

```go
import "github.com/taking/kubemigrate/pkg/client/helm"

// 클라이언트 생성
client := helm.NewClient()

// 설치된 차트 목록 조회
charts, err := client.GetCharts(ctx, "")
if err != nil {
    return err
}

// 차트 목록 출력
for _, chart := range charts {
    fmt.Printf("Chart: %s (Namespace: %s)\n", chart.Name, chart.Namespace)
}
```

## API 참조

### GetCharts

설치된 모든 Helm 차트 목록을 조회합니다.

```go
func (c *client) GetCharts(ctx context.Context, namespace string) ([]*release.Release, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `namespace`: 네임스페이스 (현재는 모든 네임스페이스에서 조회)

**반환값:**
- `([]*release.Release, error)`: 설치된 차트 목록

### GetChart

특정 Helm 차트의 상세 정보를 조회합니다.

```go
func (c *client) GetChart(ctx context.Context, releaseName, namespace string, releaseVersion int) (*release.Release, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `releaseName`: 릴리스 이름
- `namespace`: 네임스페이스
- `releaseVersion`: 릴리스 버전 (0이면 최신 버전)

**반환값:**
- `(*release.Release, error)`: 차트 상세 정보

### IsChartInstalled

특정 차트가 설치되어 있는지 확인합니다.

```go
func (c *client) IsChartInstalled(releaseName string) (bool, *release.Release, error)
```

**매개변수:**
- `releaseName`: 릴리스 이름

**반환값:**
- `(bool, *release.Release, error)`: 설치 여부, 릴리스 정보, 에러

### InstallChart

Helm 차트를 설치합니다.

```go
func (c *client) InstallChart(releaseName, chartURL, version string, values map[string]interface{}) error
```

**매개변수:**
- `releaseName`: 릴리스 이름
- `chartURL`: 차트 URL 또는 경로
- `version`: 차트 버전 (빈 문자열이면 최신 버전)
- `values`: 차트 값 (설정)

**예제:**
```go
values := map[string]interface{}{
    "replicaCount": 3,
    "image": map[string]interface{}{
        "repository": "nginx",
        "tag": "1.20",
    },
}

err := client.InstallChart("my-app", "https://charts.example.com/my-app", "1.0.0", values)
if err != nil {
    return err
}
```

### UpgradeChart

설치된 Helm 차트를 업그레이드합니다.

```go
func (c *client) UpgradeChart(releaseName, chartPath string, values map[string]interface{}) error
```

**매개변수:**
- `releaseName`: 릴리스 이름
- `chartPath`: 차트 경로 또는 URL
- `values`: 새로운 차트 값

### UninstallChart

Helm 차트를 제거합니다.

```go
func (c *client) UninstallChart(releaseName, namespace string, dryRun bool) error
```

**매개변수:**
- `releaseName`: 릴리스 이름
- `namespace`: 네임스페이스
- `dryRun`: 시뮬레이션 모드 (true면 실제 제거하지 않음)

### HealthCheck

Helm 연결 상태를 확인합니다.

```go
func (c *client) HealthCheck(ctx context.Context) error
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트

**참고:** HealthCheck 메서드는 현재 인터페이스에 포함되어 있지 않습니다. 기본 클라이언트에서는 사용할 수 없습니다.

## 클라이언트 생성

### 기본 클라이언트

```go
client := helm.NewClient()
```

### 설정을 통한 클라이언트 생성

```go
import "github.com/taking/kubemigrate/internal/config"

cfg := config.HelmConfig{
    KubeConfig: config.KubeConfig{
        KubeConfig: "base64-encoded-kubeconfig",
        Namespace:  "default",
    },
}

client, err := helm.NewClientWithConfig(cfg)
if err != nil {
    return err
}
```

## 에러 처리

모든 메서드는 적절한 에러 처리를 포함합니다:

```go
// 차트 설치
err := client.InstallChart("my-app", "https://charts.example.com/my-app", "1.0.0", values)
if err != nil {
    if strings.Contains(err.Error(), "already installed") {
        // 이미 설치된 경우
        fmt.Println("차트가 이미 설치되어 있습니다.")
    } else {
        return fmt.Errorf("차트 설치 실패: %w", err)
    }
}

// 차트 조회
chart, err := client.GetChart(ctx, "my-app", "default", 0)
if err != nil {
    return fmt.Errorf("차트 조회 실패: %w", err)
}
```

## 모범 사례

1. **설치 전 확인**: `IsChartInstalled`로 중복 설치 방지
2. **에러 처리**: 모든 Helm 작업에 대한 적절한 에러 처리
3. **컨텍스트 사용**: 취소 및 타임아웃을 위해 항상 컨텍스트 전달
4. **Dry Run 활용**: 제거 작업 전에 `dryRun=true`로 시뮬레이션

## 예제

### 차트 설치 및 확인

```go
// 설치 여부 확인
installed, release, err := client.IsChartInstalled("my-app")
if err != nil {
    return err
}

if !installed {
    // 차트 설치
    values := map[string]interface{}{
        "replicaCount": 2,
    }
    
    err = client.InstallChart("my-app", "/path/to/chart", values)
    if err != nil {
        return err
    }
    
    fmt.Println("차트가 성공적으로 설치되었습니다.")
} else {
    fmt.Printf("차트가 이미 설치되어 있습니다. 버전: %d\n", release.Version)
}
```

### 차트 업그레이드

```go
// 설치 여부 확인
installed, _, err := client.IsChartInstalled("my-app")
if err != nil {
    return err
}

if installed {
    // 새로운 설정으로 업그레이드
    newValues := map[string]interface{}{
        "replicaCount": 5,
        "image": map[string]interface{}{
            "tag": "1.21",
        },
    }
    
    err = client.UpgradeChart("my-app", "/path/to/chart", newValues)
    if err != nil {
        return err
    }
    
    fmt.Println("차트가 성공적으로 업그레이드되었습니다.")
}
```

### 차트 제거 (Dry Run)

```go
// Dry Run으로 제거 시뮬레이션
err := client.UninstallChart("my-app", "default", true)
if err != nil {
    if strings.Contains(err.Error(), "[DryRun]") {
        fmt.Println("제거 시뮬레이션이 완료되었습니다.")
    } else {
        return err
    }
}

// 실제 제거
err = client.UninstallChart("my-app", "default", false)
if err != nil {
    return err
}

fmt.Println("차트가 성공적으로 제거되었습니다.")
```

### 모든 차트 목록 조회

```go
charts, err := client.GetCharts(ctx, "")
if err != nil {
    return err
}

fmt.Printf("총 %d개의 차트가 설치되어 있습니다:\n", len(charts))
for _, chart := range charts {
    fmt.Printf("- %s (Namespace: %s, Version: %d, Status: %s)\n", 
        chart.Name, chart.Namespace, chart.Version, chart.Info.Status)
}
```

## 테스트

클라이언트 테스트는 다음과 같이 실행할 수 있습니다:

```bash
go test ./pkg/client/helm/... -v
```

### 테스트 커버리지

현재 테스트는 다음 기능들을 커버합니다:

- ✅ `NewClient()` - 기본 클라이언트 생성
- ✅ `NewClientWithConfig()` - 설정을 통한 클라이언트 생성
- ✅ `GetCharts()` - 차트 목록 조회
- ✅ `GetChart()` - 특정 차트 조회
- ✅ `IsChartInstalled()` - 차트 설치 상태 확인
- ✅ `InstallChart()` - 차트 설치
- ✅ `UninstallChart()` - 차트 제거
- ✅ `UpgradeChart()` - 차트 업그레이드

### 테스트 실행 예제

```go
func TestHelmClient(t *testing.T) {
    // 기본 클라이언트 생성
    client := helm.NewClient()
    if client == nil {
        t.Fatal("NewClient() returned nil")
    }

    // 차트 목록 조회 테스트
    ctx := context.Background()
    charts, err := client.GetCharts(ctx, "default")
    if err != nil {
        t.Logf("GetCharts failed as expected: %v", err)
    } else {
        t.Log("GetCharts succeeded - this might indicate a real cluster is available")
    }
}
```
