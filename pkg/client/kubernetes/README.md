# Kubernetes 클라이언트

Kubernetes 리소스 작업을 위한 통합 클라이언트로, 리스트 조회와 단일 리소스 조회를 위한 단일 인터페이스를 제공하여 일반적인 작업을 단순화합니다.

## 📋 개요

Kubernetes 클라이언트는 Kubernetes API와 상호작용하기 위한 고수준 인터페이스를 제공합니다. 이 클라이언트는 복잡한 Kubernetes API 호출을 단순화하고, 타입 안전성을 보장하며, 일관된 에러 처리를 제공합니다.

## ✨ 주요 특징

- **🔗 통합 API**: 리스트 및 단일 리소스 작업을 위한 단일 메서드
- **🛡️ 타입 안전성**: 타입 어설션에 대한 명확한 문서화
- **🏷️ 네임스페이스 지원**: "all" 네임스페이스를 포함한 전체 네임스페이스 지원
- **⚠️ 에러 처리**: 상세한 메시지와 함께 포괄적인 에러 처리
- **⚡ 성능 최적화**: 효율적인 리소스 관리 및 캐싱
- **🔧 설정 유연성**: 다양한 설정 옵션 지원

## 기능

- **통합 API**: 리스트 및 단일 리소스 작업을 위한 단일 메서드
- **타입 안전성**: 타입 어설션에 대한 명확한 문서화
- **네임스페이스 지원**: "all" 네임스페이스를 포함한 전체 네임스페이스 지원
- **에러 처리**: 상세한 메시지와 함께 포괄적인 에러 처리

## 빠른 시작

```go
import "github.com/taking/kubemigrate/pkg/client/kubernetes"

// 클라이언트 생성
client := kubernetes.NewClient()

// default 네임스페이스의 모든 파드 목록 조회
response, err := client.GetPods(ctx, "default", "")
if err != nil {
    return err
}

// 리스트 응답에 대한 타입 어설션
podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type")
}

// 파드 반복 처리
for _, pod := range podList.Items {
    fmt.Printf("Pod: %s\n", pod.Name)
}
```

## API 참조

### GetPods

지정된 네임스페이스에서 파드를 조회합니다.

```go
func (c *client) GetPods(ctx context.Context, namespace, name string) (interface{}, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `namespace`: 네임스페이스 이름 (모든 네임스페이스의 경우 "" 사용)
- `name`: 파드 이름 (목록의 경우 "", 특정 파드의 경우 이름 지정)

**반환값:**
- `name`이 비어있을 때: `(*v1.PodList, error)` (모든 파드 목록)
- `name`이 제공될 때: `(*v1.Pod, error)` (단일 파드)

**예제:**
```go
// 모든 파드 목록 조회
response, err := client.GetPods(ctx, "default", "")
podList := response.(*v1.PodList)

// 특정 파드 조회
response, err = client.GetPods(ctx, "default", "my-pod")
pod := response.(*v1.Pod)
```

### GetConfigMaps

지정된 네임스페이스에서 ConfigMap을 조회합니다.

```go
func (c *client) GetConfigMaps(ctx context.Context, namespace, name string) (interface{}, error)
```

**반환값:**
- `name`이 비어있을 때: `(*v1.ConfigMapList, error)` (모든 configmap 목록)
- `name`이 제공될 때: `(*v1.ConfigMap, error)` (단일 configmap)

### GetSecrets

지정된 네임스페이스에서 Secret을 조회합니다.

```go
func (c *client) GetSecrets(ctx context.Context, namespace, name string) (interface{}, error)
```

**반환값:**
- `name`이 비어있을 때: `(*v1.SecretList, error)` (모든 secret 목록)
- `name`이 제공될 때: `(*v1.Secret, error)` (단일 secret)

### GetStorageClasses

StorageClass를 조회합니다 (클러스터 범위 리소스).

```go
func (c *client) GetStorageClasses(ctx context.Context, name string) (interface{}, error)
```

**반환값:**
- `name`이 비어있을 때: `(*storagev1.StorageClassList, error)` (모든 storage class 목록)
- `name`이 제공될 때: `(*storagev1.StorageClass, error)` (단일 storage class)

### GetNamespaces

모든 네임스페이스 목록을 조회합니다.

```go
func (c *client) GetNamespaces(ctx context.Context) (*v1.NamespaceList, error)
```

**반환값:**
- `(*v1.NamespaceList, error)`: 네임스페이스 목록, 에러

### GetNamespace

특정 네임스페이스를 조회합니다.

```go
func (c *client) GetNamespace(ctx context.Context, name string) (*v1.Namespace, error)
```

**반환값:**
- `(*v1.Namespace, error)`: 네임스페이스 정보, 에러

## 타입 어설션 가이드

모든 메서드가 `interface{}`를 반환하므로, 매개변수에 따라 타입 어설션을 수행해야 합니다:

### 목록 작업 (name이 비어있음)

| 메서드 | 예상 타입 |
|--------|-----------|
| `GetPods(ctx, namespace, "")` | `*v1.PodList` |
| `GetConfigMaps(ctx, namespace, "")` | `*v1.ConfigMapList` |
| `GetSecrets(ctx, namespace, "")` | `*v1.SecretList` |
| `GetStorageClasses(ctx, "")` | `*storagev1.StorageClassList` |

### 단일 리소스 작업 (name이 제공됨)

| 메서드 | 예상 타입 |
|--------|-----------|
| `GetPods(ctx, namespace, "pod-name")` | `*v1.Pod` |
| `GetConfigMaps(ctx, namespace, "cm-name")` | `*v1.ConfigMap` |
| `GetSecrets(ctx, namespace, "secret-name")` | `*v1.Secret` |
| `GetStorageClasses(ctx, "sc-name")` | `*storagev1.StorageClass` |

## 에러 처리

항상 에러를 확인하고 안전한 타입 어설션을 수행하세요:

```go
response, err := client.GetPods(ctx, "default", "")
if err != nil {
    return fmt.Errorf("failed to get pods: %w", err)
}

podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type: expected *v1.PodList")
}

// podList를 안전하게 사용
for _, pod := range podList.Items {
    // 파드 처리
}
```

## 네임스페이스 처리

- **특정 네임스페이스**: `"default"`, `"kube-system"` 등
- **모든 네임스페이스**: `""` (빈 문자열)
- **기본 동작**: 네임스페이스가 비어있으면 "default"로 기본 설정

## 모범 사례

1. **항상 에러 확인**: 타입 어설션 전에 에러 처리
2. **안전한 타입 어설션 사용**: 두 값 형태 `value, ok := response.(*Type)` 사용
3. **예상치 못한 타입 처리**: 타입 어설션에서 항상 `ok` 값 확인
4. **컨텍스트 사용**: 취소 및 타임아웃을 위해 항상 컨텍스트 전달

## 예제

### 모든 네임스페이스의 모든 파드 목록

```go
response, err := client.GetPods(ctx, "", "")
if err != nil {
    return err
}

podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("모든 네임스페이스에서 %d개의 파드를 찾았습니다\n", len(podList.Items))
```

### 특정 ConfigMap 조회

```go
response, err := client.GetConfigMaps(ctx, "kube-system", "kubeconfig")
if err != nil {
    return err
}

configMap, ok := response.(*v1.ConfigMap)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("ConfigMap: %s\n", configMap.Name)
```

### Storage Class 목록

```go
response, err := client.GetStorageClasses(ctx, "")
if err != nil {
    return err
}

storageClassList, ok := response.(*storagev1.StorageClassList)
if !ok {
    return fmt.Errorf("unexpected response type")
}

for _, sc := range storageClassList.Items {
    fmt.Printf("StorageClass: %s\n", sc.Name)
}
```

## 테스트

클라이언트 테스트는 다음과 같이 실행할 수 있습니다:

```bash
go test ./pkg/client/kubernetes/... -v
```

### 테스트 커버리지

현재 테스트는 다음 기능들을 커버합니다:

- ✅ `NewClient()` - 기본 클라이언트 생성
- ✅ `NewClientWithConfig()` - 설정을 통한 클라이언트 생성
- ✅ `GetPods()` - 파드 조회 (목록/단일)
- ✅ `GetConfigMaps()` - ConfigMap 조회 (목록/단일)
- ✅ `GetSecrets()` - Secret 조회 (목록/단일)
- ✅ `GetStorageClasses()` - StorageClass 조회 (목록/단일)
- ✅ `GetNamespaces()` - 네임스페이스 목록 조회
- ✅ `GetNamespace()` - 특정 네임스페이스 조회

### 테스트 실행 예제

```go
func TestKubernetesClient(t *testing.T) {
    // 기본 클라이언트 생성
    client := kubernetes.NewClient()
    if client == nil {
        t.Fatal("NewClient() returned nil")
    }

    // 파드 목록 조회 테스트
    ctx := context.Background()
    response, err := client.GetPods(ctx, "default", "")
    if err != nil {
        t.Logf("GetPods failed as expected: %v", err)
    } else {
        t.Log("GetPods succeeded - this might indicate a real cluster is available")
    }
}
```
