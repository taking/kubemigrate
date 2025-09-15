# MinIO 클라이언트

MinIO 객체 스토리지와 상호작용하기 위한 통합 클라이언트입니다. 버킷 관리, 객체 업로드/다운로드, Presigned URL 생성 등의 기능을 제공합니다.

## 📋 개요

MinIO 클라이언트는 MinIO 객체 스토리지와 상호작용하기 위한 고수준 인터페이스를 제공합니다. 이 클라이언트는 복잡한 MinIO API 호출을 단순화하고, 일관된 에러 처리를 제공하며, 타입 안전성을 보장합니다.

## ✨ 주요 특징

- **🪣 버킷 관리**: 버킷 생성, 삭제, 존재 확인, 목록 조회
- **📁 객체 관리**: 객체 업로드, 다운로드, 삭제, 복사
- **🔗 Presigned URL**: 보안 URL 생성으로 직접 접근 허용
- **📊 객체 정보**: 객체 메타데이터 조회
- **⚠️ 에러 처리**: 포괄적인 에러 처리
- **⚡ 성능 최적화**: 효율적인 객체 관리 및 캐싱
- **🔧 설정 유연성**: 다양한 설정 옵션 지원

## 🌐 RESTful API 엔드포인트

### 버킷 관리
- `GET /api/v1/minio/buckets` - 버킷 목록 조회
- `POST /api/v1/minio/buckets/{bucket}` - 버킷 생성
- `GET /api/v1/minio/buckets/{bucket}` - 버킷 존재 확인
- `DELETE /api/v1/minio/buckets/{bucket}` - 버킷 삭제
- `HEAD /api/v1/minio/buckets/{bucket}` - 버킷 존재 확인 (HEAD)

### 객체 관리
- `GET /api/v1/minio/buckets/{bucket}/objects` - 객체 목록 조회
- `GET /api/v1/minio/buckets/{bucket}/objects/{object}` - 객체 다운로드
- `PUT /api/v1/minio/buckets/{bucket}/objects/{object}` - 객체 업로드
- `DELETE /api/v1/minio/buckets/{bucket}/objects/{object}` - 객체 삭제
- `HEAD /api/v1/minio/buckets/{bucket}/objects/{object}` - 객체 정보 조회
- `POST /api/v1/minio/buckets/{bucket}/objects/{object}/copy` - 객체 복사

### Presigned URL
- `GET /api/v1/minio/buckets/{bucket}/objects/{object}/presigned-get` - Presigned GET URL
- `PUT /api/v1/minio/buckets/{bucket}/objects/{object}/presigned-put` - Presigned PUT URL

## 빠른 시작

```go
import "github.com/taking/kubemigrate/pkg/client/minio"

// 클라이언트 생성
client := minio.NewClient()

// 버킷 생성
err := client.CreateBucketIfNotExists(ctx, "my-bucket")
if err != nil {
    return err
}

// 객체 업로드
file, err := os.Open("example.txt")
if err != nil {
    return err
}
defer file.Close()

fileInfo, _ := file.Stat()
uploadInfo, err := client.PutObject(ctx, "my-bucket", "example.txt", file, fileInfo.Size())
if err != nil {
    return err
}

fmt.Printf("업로드 완료: ETag=%s, Size=%d\n", uploadInfo.ETag, uploadInfo.Size)
```

## API 참조

### 버킷 관리

#### BucketExists

버킷이 존재하는지 확인합니다.

```go
func (c *client) BucketExists(ctx context.Context, bucketName string) (bool, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `bucketName`: 버킷 이름

**반환값:**
- `(bool, error)`: 존재 여부, 에러

#### CreateBucket

새로운 버킷을 생성합니다.

```go
func (c *client) CreateBucket(ctx context.Context, bucketName string) error
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `bucketName`: 버킷 이름

#### MakeBucket

옵션을 포함하여 버킷을 생성합니다.

```go
func (c *client) MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `bucketName`: 버킷 이름
- `opts`: 버킷 생성 옵션

**MakeBucketOptions:**
```go
type MakeBucketOptions struct {
    Region string  // 리전 설정
}
```

#### CreateBucketIfNotExists

버킷이 없으면 생성합니다.

```go
func (c *client) CreateBucketIfNotExists(ctx context.Context, bucketName string) error
```

#### DeleteBucket

버킷을 삭제합니다.

```go
func (c *client) DeleteBucket(ctx context.Context, bucketName string) error
```

#### ListBuckets

모든 버킷 목록을 조회합니다.

```go
func (c *client) ListBuckets(ctx context.Context) (interface{}, error)
```

**반환값:**
- `([]minio.BucketInfo, error)`: 버킷 정보 목록, 에러

### 객체 관리

#### PutObject

객체를 업로드합니다.

```go
func (c *client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) (interface{}, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `bucketName`: 버킷 이름
- `objectName`: 객체 이름
- `reader`: 데이터 스트림
- `objectSize`: 객체 크기

**반환값:**
- `(minio.UploadInfo, error)`: 업로드 정보, 에러

#### GetObject

객체를 다운로드합니다.

```go
func (c *client) GetObject(ctx context.Context, bucketName, objectName string) (interface{}, error)
```

**반환값:**
- `(*minio.Object, error)`: 객체 스트림, 에러

#### DeleteObject

객체를 삭제합니다.

```go
func (c *client) DeleteObject(ctx context.Context, bucketName, objectName string) error
```

#### ListObjects

버킷 내 객체 목록을 조회합니다.

```go
func (c *client) ListObjects(ctx context.Context, bucketName string) (interface{}, error)
```

**반환값:**
- `([]minio.ObjectInfo, error)`: 객체 정보 목록, 에러

#### StatObject

객체의 메타데이터를 조회합니다.

```go
func (c *client) StatObject(ctx context.Context, bucketName, objectName string) (interface{}, error)
```

**반환값:**
- `(minio.ObjectInfo, error)`: 객체 메타데이터, 에러

#### CopyObject

객체를 복사합니다.

```go
func (c *client) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) (interface{}, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `srcBucket`: 소스 버킷
- `srcObject`: 소스 객체
- `dstBucket`: 대상 버킷
- `dstObject`: 대상 객체

**반환값:**
- `(minio.UploadInfo, error)`: 복사 결과 정보, 에러

### Presigned URL

#### PresignedGetObject

다운로드를 위한 Presigned URL을 생성합니다.

```go
func (c *client) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error)
```

**매개변수:**
- `ctx`: 요청을 위한 컨텍스트
- `bucketName`: 버킷 이름
- `objectName`: 객체 이름
- `expiry`: 만료 시간 (초)

**반환값:**
- `(string, error)`: Presigned URL, 에러

#### PresignedPutObject

업로드를 위한 Presigned URL을 생성합니다.

```go
func (c *client) PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry int) (string, error)
```

## 클라이언트 생성

### 기본 클라이언트

```go
client := minio.NewClient()
```

### 설정을 통한 클라이언트 생성

```go
import "github.com/taking/kubemigrate/internal/config"

cfg := config.MinioConfig{
    Endpoint:  "localhost:9000",
    AccessKey: "minioadmin",
    SecretKey: "minioadmin",
    UseSSL:    false,
}

client, err := minio.NewClientWithConfig(cfg)
if err != nil {
    return err
}
```

## 타입 어설션 가이드

일부 메서드는 `interface{}`를 반환하므로, 적절한 타입 어설션을 수행해야 합니다:

| 메서드 | 반환 타입 |
|--------|-----------|
| `ListBuckets(ctx)` | `[]minio.BucketInfo` |
| `ListObjects(ctx, bucketName)` | `[]minio.ObjectInfo` |
| `StatObject(ctx, bucketName, objectName)` | `minio.ObjectInfo` |

### 안전한 타입 어설션 예제

```go
// 버킷 목록 조회
buckets, err := client.ListBuckets(ctx)
if err != nil {
    return err
}

bucketList, ok := buckets.([]minio.BucketInfo)
if !ok {
    return fmt.Errorf("unexpected response type: expected []minio.BucketInfo")
}

// 버킷 목록 사용
for _, bucket := range bucketList {
    fmt.Printf("버킷: %s\n", bucket.Name)
}
```

## 에러 처리

모든 메서드는 적절한 에러 처리를 포함합니다:

```go
// 버킷 존재 확인
exists, err := client.BucketExists(ctx, "my-bucket")
if err != nil {
    return fmt.Errorf("버킷 확인 실패: %w", err)
}

if !exists {
    // 버킷 생성
    err = client.CreateBucket(ctx, "my-bucket")
    if err != nil {
        return fmt.Errorf("버킷 생성 실패: %w", err)
    }
}
```

## 모범 사례

1. **버킷 사전 확인**: 객체 작업 전에 버킷 존재 여부 확인
2. **에러 처리**: 모든 MinIO 작업에 대한 적절한 에러 처리
3. **컨텍스트 사용**: 취소 및 타임아웃을 위해 항상 컨텍스트 전달
4. **리소스 정리**: 파일 핸들 등 리소스 적절히 정리

## 예제

### 버킷 생성 및 객체 업로드

```go
// 버킷 생성 (존재하지 않는 경우에만)
err := client.CreateBucketIfNotExists(ctx, "my-bucket")
if err != nil {
    return err
}

// 파일 업로드
file, err := os.Open("example.txt")
if err != nil {
    return err
}
defer file.Close()

fileInfo, err := file.Stat()
if err != nil {
    return err
}

result, err := client.PutObject(ctx, "my-bucket", "example.txt", file, fileInfo.Size())
if err != nil {
    return err
}

fmt.Printf("파일이 성공적으로 업로드되었습니다: %+v\n", result)
```

### 객체 다운로드

```go
// 객체 다운로드
object, err := client.GetObject(ctx, "my-bucket", "example.txt")
if err != nil {
    return err
}
defer object.Close()

// 객체 내용 읽기
data, err := io.ReadAll(object)
if err != nil {
    return err
}

fmt.Printf("다운로드 완료: %d bytes\n", len(data))
```

### Presigned URL 생성

```go
// 다운로드용 Presigned URL 생성 (1시간 유효)
url, err := client.PresignedGetObject(ctx, "my-bucket", "example.txt", 3600)
if err != nil {
    return err
}

fmt.Printf("다운로드 URL: %s\n", url)

// 업로드용 Presigned URL 생성 (30분 유효)
uploadURL, err := client.PresignedPutObject(ctx, "my-bucket", "new-file.txt", 1800)
if err != nil {
    return err
}

fmt.Printf("업로드 URL: %s\n", uploadURL)
```

### 객체 복사

```go
// 객체 복사
copyInfo, err := client.CopyObject(ctx, "source-bucket", "source-file.txt", "dest-bucket", "dest-file.txt")
if err != nil {
    return err
}

fmt.Printf("복사 완료: ETag=%s, Size=%d\n", copyInfo.ETag, copyInfo.Size)
```

### 버킷 목록 조회

```go
// 모든 버킷 목록 조회
buckets, err := client.ListBuckets(ctx)
if err != nil {
    return err
}

bucketList, ok := buckets.([]minio.BucketInfo)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("총 %d개의 버킷이 있습니다:\n", len(bucketList))
for _, bucket := range bucketList {
    fmt.Printf("- %s (생성일: %s)\n", bucket.Name, bucket.CreationDate)
}
```

### 객체 목록 조회

```go
// 특정 버킷의 객체 목록 조회
objects, err := client.ListObjects(ctx, "my-bucket")
if err != nil {
    return err
}

objectList, ok := objects.([]minio.ObjectInfo)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("총 %d개의 객체가 있습니다:\n", len(objectList))
for _, obj := range objectList {
    fmt.Printf("- %s (크기: %d bytes, 수정일: %s)\n", obj.Key, obj.Size, obj.LastModified)
}
```

### 객체 정보 조회

```go
// 객체 메타데이터 조회
info, err := client.StatObject(ctx, "my-bucket", "example.txt")
if err != nil {
    return err
}

objectInfo, ok := info.(minio.ObjectInfo)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("객체 정보: 이름=%s, 크기=%d bytes, ETag=%s, 수정일=%s\n", 
    objectInfo.Key, objectInfo.Size, objectInfo.ETag, objectInfo.LastModified)
```

## 테스트

클라이언트 테스트는 다음과 같이 실행할 수 있습니다:

```bash
go test ./pkg/client/minio/... -v
```

### 테스트 커버리지

현재 테스트는 다음 기능들을 커버합니다:

- ✅ `NewClient()` - 기본 클라이언트 생성
- ✅ `NewClientWithConfig()` - 설정을 통한 클라이언트 생성
- ✅ `BucketExists()` - 버킷 존재 확인
- ✅ `CreateBucket()` - 버킷 생성
- ✅ `CreateBucketIfNotExists()` - 버킷 생성 (없는 경우에만)
- ✅ `DeleteBucket()` - 버킷 삭제
- ✅ `ListBuckets()` - 버킷 목록 조회
- ✅ `PutObject()` - 객체 업로드
- ✅ `GetObject()` - 객체 다운로드
- ✅ `DeleteObject()` - 객체 삭제
- ✅ `ListObjects()` - 객체 목록 조회
- ✅ `StatObject()` - 객체 정보 조회
- ✅ `CopyObject()` - 객체 복사
- ✅ `PresignedGetObject()` - Presigned GET URL 생성
- ✅ `PresignedPutObject()` - Presigned PUT URL 생성

### 테스트 실행 예제

```go
func TestMinioClient(t *testing.T) {
    // 기본 클라이언트 생성
    client := minio.NewClient()
    if client == nil {
        t.Fatal("NewClient() returned nil")
    }

    // 버킷 존재 확인 테스트
    ctx := context.Background()
    exists, err := client.BucketExists(ctx, "test-bucket")
    if err != nil {
        t.Logf("BucketExists failed as expected: %v", err)
    } else {
        t.Logf("BucketExists succeeded: bucket exists=%v", exists)
    }
}
```
