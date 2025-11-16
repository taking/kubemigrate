# 코드 주석 표준

## 함수 주석

모든 public 함수는 다음 형식의 주석을 가져야 합니다:

```go
// FunctionName : 함수의 간단한 설명
// 상세한 설명이 필요한 경우 추가로 작성
func FunctionName() error {
    // 구현
}
```

## 구조체 주석

```go
// StructName : 구조체의 간단한 설명
// 상세한 설명이 필요한 경우 추가로 작성
type StructName struct {
    Field1 string // 필드 설명
    Field2 int    // 필드 설명
}
```

## 인터페이스 주석

```go
// InterfaceName : 인터페이스의 간단한 설명
// 상세한 설명이 필요한 경우 추가로 작성
type InterfaceName interface {
    Method1() error // 메서드 설명
    Method2() error // 메서드 설명
}
```

## 패키지 주석

```go
// Package packagename 패키지의 간단한 설명
// 상세한 설명이 필요한 경우 추가로 작성
package packagename
```

## 예시

```go
// Package client 클라이언트 인터페이스를 정의합니다.
// Kubernetes, Helm, MinIO, Velero 클라이언트의 공통 인터페이스를 제공합니다.
package client

// Client : 통합 클라이언트 인터페이스
// 모든 클라이언트의 공통 메서드를 정의합니다.
type Client interface {
    // Kubernetes : Kubernetes 클라이언트를 반환합니다
    Kubernetes() kubernetes.Client
    
    // Helm : Helm 클라이언트를 반환합니다
    Helm() helm.Client
    
    // Velero : Velero 클라이언트를 반환합니다
    Velero() velero.Client
    
    // Minio : MinIO 클라이언트를 반환합니다
    Minio() minio.Client
}

// NewClient : 새로운 통합 클라이언트를 생성합니다
// 모든 하위 클라이언트를 초기화하고 통합 클라이언트를 반환합니다.
func NewClient() (Client, error) {
    // 구현
}
```
