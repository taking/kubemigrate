# Velero 클라이언트

Kubernetes 클러스터에서 Velero 백업 및 복원 작업을 관리하기 위한 통합 클라이언트입니다. 백업, 복원, 저장소 위치 관리 등의 기능을 제공합니다.

## 기능

- **백업 관리**: 백업 생성, 조회, 삭제
- **복원 관리**: 복원 생성, 조회, 삭제
- **저장소 관리**: 백업 저장소 위치 및 볼륨 스냅샷 위치 관리
- **리포지토리 관리**: 백업 리포지토리 관리
- **볼륨 복원**: Pod 볼륨 복원 관리

## 빠른 시작

```go
import "github.com/taking/kubemigrate/pkg/client/velero"

// 클라이언트 생성
client := velero.NewClient()

// 백업 목록 조회
backups, err := client.GetBackups(ctx, "velero")
if err != nil {
    return err
}

// 백업 목록 출력
for _, backup := range backups {
    fmt.Printf("Backup: %s (Status: %s)\n", backup.Name, backup.Status.Phase)
}
```

## API 참조

### 백업 관리

#### GetBackups

네임스페이스의 백업 목록을 조회합니다.

```go
func (c *client) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `namespace`: 네임스페이스 (일반적으로 "velero")

**반환값:**
- `([]velerov1.Backup, error)`: 백업 목록, 에러

**예제:**
```go
backups, err := client.GetBackups(ctx, "velero")
if err != nil {
    return err
}

for _, backup := range backups {
    fmt.Printf("Backup: %s (Status: %s)\n", backup.Name, backup.Status.Phase)
}
```

#### GetBackup

특정 백업의 상세 정보를 조회합니다.

```go
func (c *client) GetBackup(ctx context.Context, namespace, name string) (*velerov1.Backup, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `namespace`: 네임스페이스
- `name`: 백업 이름

**반환값:**
- `(*velerov1.Backup, error)`: 백업 상세 정보, 에러

#### CreateBackup

새로운 백업을 생성합니다.

```go
func (c *client) CreateBackup(ctx context.Context, namespace string, backup *velerov1.Backup) error
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `namespace`: 네임스페이스
- `backup`: 백업 객체

#### DeleteBackup

백업을 삭제합니다.

```go
func (c *client) DeleteBackup(ctx context.Context, namespace, name string) error
```

### 복원 관리

#### GetRestores

네임스페이스의 복원 목록을 조회합니다.

```go
func (c *client) GetRestores(ctx context.Context, namespace string) ([]velerov1.Restore, error)
```

#### GetRestore

특정 복원의 상세 정보를 조회합니다.

```go
func (c *client) GetRestore(ctx context.Context, namespace, name string) (*velerov1.Restore, error)
```

#### CreateRestore

새로운 복원을 생성합니다.

```go
func (c *client) CreateRestore(ctx context.Context, namespace string, restore *velerov1.Restore) error
```

#### DeleteRestore

복원을 삭제합니다.

```go
func (c *client) DeleteRestore(ctx context.Context, namespace, name string) error
```

### 백업 리포지토리 관리

#### GetBackupRepositories

네임스페이스의 백업 리포지토리 목록을 조회합니다.

```go
func (c *client) GetBackupRepositories(ctx context.Context, namespace string) ([]velerov1.BackupRepository, error)
```

#### GetBackupRepository

특정 백업 리포지토리의 상세 정보를 조회합니다.

```go
func (c *client) GetBackupRepository(ctx context.Context, namespace, name string) (*velerov1.BackupRepository, error)
```

### 백업 저장소 위치 관리

#### GetBackupStorageLocations

네임스페이스의 백업 저장소 위치 목록을 조회합니다.

```go
func (c *client) GetBackupStorageLocations(ctx context.Context, namespace string) ([]velerov1.BackupStorageLocation, error)
```

#### GetBackupStorageLocation

특정 백업 저장소 위치의 상세 정보를 조회합니다.

```go
func (c *client) GetBackupStorageLocation(ctx context.Context, namespace, name string) (*velerov1.BackupStorageLocation, error)
```

### 볼륨 스냅샷 위치 관리

#### GetVolumeSnapshotLocations

네임스페이스의 볼륨 스냅샷 위치 목록을 조회합니다.

```go
func (c *client) GetVolumeSnapshotLocations(ctx context.Context, namespace string) ([]velerov1.VolumeSnapshotLocation, error)
```

#### GetVolumeSnapshotLocation

특정 볼륨 스냅샷 위치의 상세 정보를 조회합니다.

```go
func (c *client) GetVolumeSnapshotLocation(ctx context.Context, namespace, name string) (*velerov1.VolumeSnapshotLocation, error)
```

### Pod 볼륨 복원 관리

#### GetPodVolumeRestores

네임스페이스의 Pod 볼륨 복원 목록을 조회합니다.

```go
func (c *client) GetPodVolumeRestores(ctx context.Context, namespace string) ([]velerov1.PodVolumeRestore, error)
```

#### GetPodVolumeRestore

특정 Pod 볼륨 복원의 상세 정보를 조회합니다.

```go
func (c *client) GetPodVolumeRestore(ctx context.Context, namespace, name string) (*velerov1.PodVolumeRestore, error)
```

## 클라이언트 생성

### 기본 클라이언트

```go
client := velero.NewClient()
```

### 설정을 통한 클라이언트 생성

```go
import "github.com/taking/kubemigrate/internal/config"

cfg := config.VeleroConfig{
    KubeConfig: config.KubeConfig{
        KubeConfig: "base64-encoded-kubeconfig",
        Namespace:  "velero",
    },
}

client, err := velero.NewClientWithConfig(cfg)
if err != nil {
    return err
}
```

## 에러 처리

모든 메서드는 적절한 에러 처리를 포함합니다:

```go
// 백업 조회
backup, err := client.GetBackup(ctx, "velero", "my-backup")
if err != nil {
    return fmt.Errorf("백업 조회 실패: %w", err)
}

// 백업 상태 확인
if backup.Status.Phase == velerov1.BackupPhaseCompleted {
    fmt.Println("백업이 성공적으로 완료되었습니다.")
} else if backup.Status.Phase == velerov1.BackupPhaseFailed {
    fmt.Printf("백업이 실패했습니다: %s\n", backup.Status.Message)
}
```

## 모범 사례

1. **네임스페이스 확인**: Velero는 일반적으로 "velero" 네임스페이스에 설치됨
2. **상태 확인**: 백업/복원 작업 후 상태를 확인하여 성공 여부 판단
3. **에러 처리**: 모든 Velero 작업에 대한 적절한 에러 처리
4. **컨텍스트 사용**: 취소 및 타임아웃을 위해 항상 컨텍스트 전달

## 예제

### 백업 생성

```go
import velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"

// 백업 객체 생성
backup := &velerov1.Backup{
    ObjectMeta: metav1.ObjectMeta{
        Name: "my-backup",
    },
    Spec: velerov1.BackupSpec{
        IncludedNamespaces: []string{"default"},
        StorageLocation:    "default",
        VolumeSnapshotLocations: []string{"default"},
    },
}

// 백업 생성
err := client.CreateBackup(ctx, "velero", backup)
if err != nil {
    return err
}

fmt.Println("백업이 생성되었습니다.")
```

### 백업 목록 조회

```go
// 백업 목록 조회
backups, err := client.GetBackups(ctx, "velero")
if err != nil {
    return err
}

fmt.Printf("총 %d개의 백업이 있습니다:\n", len(backups))
for _, backup := range backups {
    fmt.Printf("- %s (Status: %s, Created: %s)\n", 
        backup.Name, backup.Status.Phase, backup.CreationTimestamp)
}
```

### 복원 생성

```go
// 복원 객체 생성
restore := &velerov1.Restore{
    ObjectMeta: metav1.ObjectMeta{
        Name: "my-restore",
    },
    Spec: velerov1.RestoreSpec{
        BackupName: "my-backup",
    },
}

// 복원 생성
err := client.CreateRestore(ctx, "velero", restore)
if err != nil {
    return err
}

fmt.Println("복원이 생성되었습니다.")
```

### 백업 저장소 위치 조회

```go
// 백업 저장소 위치 목록 조회
locations, err := client.GetBackupStorageLocations(ctx, "velero")
if err != nil {
    return err
}

for _, location := range locations {
    fmt.Printf("저장소 위치: %s (Provider: %s, Access Mode: %s)\n", 
        location.Name, location.Spec.Provider, location.Spec.AccessMode)
}
```

### 볼륨 스냅샷 위치 조회

```go
// 볼륨 스냅샷 위치 목록 조회
snapshotLocations, err := client.GetVolumeSnapshotLocations(ctx, "velero")
if err != nil {
    return err
}

for _, location := range snapshotLocations {
    fmt.Printf("스냅샷 위치: %s (Provider: %s)\n", 
        location.Name, location.Spec.Provider)
}
```

### 백업 리포지토리 조회

```go
// 백업 리포지토리 목록 조회
repositories, err := client.GetBackupRepositories(ctx, "velero")
if err != nil {
    return err
}

for _, repo := range repositories {
    fmt.Printf("리포지토리: %s (Status: %s)\n", 
        repo.Name, repo.Status.Phase)
}
```

### Pod 볼륨 복원 조회

```go
// Pod 볼륨 복원 목록 조회
podRestores, err := client.GetPodVolumeRestores(ctx, "velero")
if err != nil {
    return err
}

for _, restore := range podRestores {
    fmt.Printf("Pod 볼륨 복원: %s (Status: %s)\n", 
        restore.Name, restore.Status.Phase)
}
```

### 백업 삭제

```go
// 백업 삭제
err := client.DeleteBackup(ctx, "velero", "my-backup")
if err != nil {
    return err
}

fmt.Println("백업이 삭제되었습니다.")
```

## 테스트

클라이언트 테스트는 다음과 같이 실행할 수 있습니다:

```bash
go test ./pkg/client/velero/... -v
```

### 테스트 커버리지

현재 테스트는 다음 기능들을 커버합니다:

- ✅ `NewClient()` - 기본 클라이언트 생성
- ✅ `NewClientWithConfig()` - 설정을 통한 클라이언트 생성
- ✅ `GetBackups()` - 백업 목록 조회
- ✅ `GetBackup()` - 특정 백업 조회
- ✅ `GetRestores()` - 복원 목록 조회
- ✅ `GetRestore()` - 특정 복원 조회
- ✅ `GetBackupRepositories()` - 백업 리포지토리 목록 조회
- ✅ `GetBackupRepository()` - 특정 백업 리포지토리 조회
- ✅ `GetBackupStorageLocations()` - 백업 저장소 위치 목록 조회
- ✅ `GetBackupStorageLocation()` - 특정 백업 저장소 위치 조회
- ✅ `GetVolumeSnapshotLocations()` - 볼륨 스냅샷 위치 목록 조회
- ✅ `GetVolumeSnapshotLocation()` - 특정 볼륨 스냅샷 위치 조회
- ✅ `GetPodVolumeRestores()` - Pod 볼륨 복원 목록 조회
- ✅ `GetPodVolumeRestore()` - 특정 Pod 볼륨 복원 조회

### 테스트 실행 예제

```go
func TestVeleroClient(t *testing.T) {
    // 기본 클라이언트 생성
    client := velero.NewClient()
    if client == nil {
        t.Fatal("NewClient() returned nil")
    }

    // 백업 목록 조회 테스트
    ctx := context.Background()
    backups, err := client.GetBackups(ctx, "velero")
    if err != nil {
        t.Logf("GetBackups failed as expected: %v", err)
    } else {
        t.Log("GetBackups succeeded - this might indicate a real cluster is available")
    }
}
```
