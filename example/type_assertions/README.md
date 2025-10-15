# Type Assertion Helper Functions Example

This example demonstrates how to use the helper functions provided in `pkg/utils/type_assertions.go` to safely extract typed data from `interface{}` return values returned by client functions.

## Problem

Client functions like `GetPods()`, `GetServices()`, `ListBuckets()` return `interface{}` values, which require explicit type assertion before using operations like `len()` or `range`. This can lead to compilation errors and runtime panics if not handled properly.

## Solution

The `pkg/utils/type_assertions.go` package provides helper functions that:

1. **Safely extract typed data** from `interface{}` return values
2. **Provide clear error messages** when type assertion fails
3. **Simplify common operations** like getting list lengths and iterating over items
4. **Reduce boilerplate code** in client usage

## Available Helper Functions

### List Extraction Functions
- `ExtractPods(response interface{}) ([]v1.Pod, error)`
- `ExtractServices(response interface{}) ([]v1.Service, error)`
- `ExtractConfigMaps(response interface{}) ([]v1.ConfigMap, error)`
- `ExtractSecrets(response interface{}) ([]v1.Secret, error)`
- `ExtractNamespaces(response interface{}) ([]v1.Namespace, error)`
- `ExtractStorageClasses(response interface{}) ([]storagev1.StorageClass, error)`
- `ExtractBuckets(response interface{}) ([]minio.BucketInfo, error)`

### Single Resource Extraction Functions
- `ExtractPod(response interface{}) (*v1.Pod, error)`
- `ExtractService(response interface{}) (*v1.Service, error)`
- `ExtractConfigMap(response interface{}) (*v1.ConfigMap, error)`
- `ExtractSecret(response interface{}) (*v1.Secret, error)`
- `ExtractNamespace(response interface{}) (*v1.Namespace, error)`
- `ExtractStorageClass(response interface{}) (*storagev1.StorageClass, error)`

## Usage Examples

### Before (Manual Type Assertion)
```go
podsResponse, err := client.GetPods(ctx, "default", "")
if err != nil {
    log.Printf("Failed to get pods: %v", err)
} else {
    podList, ok := podsResponse.(*v1.PodList)
    if !ok {
        log.Printf("Failed to cast response to PodList")
    } else {
        pods := podList.Items
        fmt.Printf("Found %d pods\n", len(pods))
        for _, pod := range pods {
            // Use pod...
        }
    }
}
```

### After (Using Helper Functions)
```go
podsResponse, err := client.GetPods(ctx, "default", "")
if err != nil {
    log.Printf("Failed to get pods: %v", err)
} else {
    pods, err := utils.ExtractPods(podsResponse)
    if err != nil {
        log.Printf("Failed to extract pods: %v", err)
    } else {
        fmt.Printf("Found %d pods\n", len(pods))
        for _, pod := range pods {
            // Use pod...
        }
    }
}
```

## Benefits

1. **Cleaner Code**: Reduces boilerplate type assertion code
2. **Better Error Handling**: Clear error messages when type assertion fails
3. **Type Safety**: Compile-time type checking for extracted data
4. **Consistency**: Standardized approach across all client functions
5. **Maintainability**: Easier to update if underlying types change

## Running the Example

```bash
cd example/type_assertions
go run main.go
```

## Dependencies

- `github.com/taking/kubemigrate/pkg/client/kubernetes`
- `github.com/taking/kubemigrate/pkg/client/minio`
- `github.com/taking/kubemigrate/pkg/utils`
- `k8s.io/api/core/v1`
- `k8s.io/api/storage/v1`
- `github.com/minio/minio-go/v7`

## Notes

- All helper functions return errors when type assertion fails
- Helper functions are designed to work with the current client interface
- Functions are optimized for common use cases and provide good error messages
- Consider using these helpers in production code for better maintainability
