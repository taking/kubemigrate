# Kubernetes Client

A unified client for Kubernetes resource operations that simplifies common tasks by providing a single interface for both list and single resource retrieval.

## Features

- **Unified API**: Single method for both list and single resource operations
- **Type Safety**: Clear documentation for type assertions
- **Namespace Support**: Full namespace support including "all" namespaces
- **Error Handling**: Comprehensive error handling with detailed messages

## Quick Start

```go
import "github.com/taking/kubemigrate/pkg/client/kubernetes"

// Create a client
client := kubernetes.NewClient()

// List all pods in default namespace
response, err := client.GetPods(ctx, "default", "")
if err != nil {
    return err
}

// Type assertion for list response
podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type")
}

// Iterate through pods
for _, pod := range podList.Items {
    fmt.Printf("Pod: %s\n", pod.Name)
}
```

## API Reference

### GetPods

Retrieves pods from a specified namespace.

```go
func (c *client) GetPods(ctx context.Context, namespace, name string) (interface{}, error)
```

**Parameters:**
- `ctx`: Context for the request
- `namespace`: Namespace name (use "" for all namespaces)
- `name`: Pod name (use "" for list, specific name for single pod)

**Returns:**
- `(*v1.PodList, error)` when name is empty (list all pods)
- `(*v1.Pod, error)` when name is provided (single pod)

**Example:**
```go
// List all pods
response, err := client.GetPods(ctx, "default", "")
podList := response.(*v1.PodList)

// Get specific pod
response, err = client.GetPods(ctx, "default", "my-pod")
pod := response.(*v1.Pod)
```

### GetConfigMaps

Retrieves ConfigMaps from a specified namespace.

```go
func (c *client) GetConfigMaps(ctx context.Context, namespace, name string) (interface{}, error)
```

**Returns:**
- `(*v1.ConfigMapList, error)` when name is empty (list all configmaps)
- `(*v1.ConfigMap, error)` when name is provided (single configmap)

### GetSecrets

Retrieves Secrets from a specified namespace.

```go
func (c *client) GetSecrets(ctx context.Context, namespace, name string) (interface{}, error)
```

**Returns:**
- `(*v1.SecretList, error)` when name is empty (list all secrets)
- `(*v1.Secret, error)` when name is provided (single secret)

### GetStorageClasses

Retrieves StorageClasses (cluster-scoped resource).

```go
func (c *client) GetStorageClasses(ctx context.Context, name string) (interface{}, error)
```

**Returns:**
- `(*storagev1.StorageClassList, error)` when name is empty (list all storage classes)
- `(*storagev1.StorageClass, error)` when name is provided (single storage class)

## Type Assertion Guide

Since all methods return `interface{}`, you need to perform type assertions based on the parameters:

### List Operations (name is empty)

| Method | Expected Type |
|--------|---------------|
| `GetPods(ctx, namespace, "")` | `*v1.PodList` |
| `GetConfigMaps(ctx, namespace, "")` | `*v1.ConfigMapList` |
| `GetSecrets(ctx, namespace, "")` | `*v1.SecretList` |
| `GetStorageClasses(ctx, "")` | `*storagev1.StorageClassList` |

### Single Resource Operations (name is provided)

| Method | Expected Type |
|--------|---------------|
| `GetPods(ctx, namespace, "pod-name")` | `*v1.Pod` |
| `GetConfigMaps(ctx, namespace, "cm-name")` | `*v1.ConfigMap` |
| `GetSecrets(ctx, namespace, "secret-name")` | `*v1.Secret` |
| `GetStorageClasses(ctx, "sc-name")` | `*storagev1.StorageClass` |

## Error Handling

Always check for errors and perform safe type assertions:

```go
response, err := client.GetPods(ctx, "default", "")
if err != nil {
    return fmt.Errorf("failed to get pods: %w", err)
}

podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type: expected *v1.PodList")
}

// Use podList safely
for _, pod := range podList.Items {
    // Process pod
}
```

## Namespace Handling

- **Specific namespace**: `"default"`, `"kube-system"`, etc.
- **All namespaces**: `""` (empty string)
- **Default behavior**: If namespace is empty, it defaults to "default"

## Best Practices

1. **Always check errors**: Handle errors before type assertions
2. **Use safe type assertions**: Use the two-value form `value, ok := response.(*Type)`
3. **Handle unexpected types**: Always check the `ok` value from type assertions
4. **Use context**: Always pass a context for cancellation and timeouts

## Examples

### List All Pods in All Namespaces

```go
response, err := client.GetPods(ctx, "", "")
if err != nil {
    return err
}

podList, ok := response.(*v1.PodList)
if !ok {
    return fmt.Errorf("unexpected response type")
}

fmt.Printf("Found %d pods across all namespaces\n", len(podList.Items))
```

### Get Specific ConfigMap

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

### List Storage Classes

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
