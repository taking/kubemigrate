package utils

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// ExtractPods extracts pods from a Kubernetes client response
func ExtractPods(response interface{}) ([]v1.Pod, error) {
	podList, ok := response.(*v1.PodList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to PodList")
	}
	return podList.Items, nil
}

// ExtractServices extracts services from a Kubernetes client response
func ExtractServices(response interface{}) ([]v1.Service, error) {
	serviceList, ok := response.(*v1.ServiceList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to ServiceList")
	}
	return serviceList.Items, nil
}

// ExtractConfigMaps extracts config maps from a Kubernetes client response
func ExtractConfigMaps(response interface{}) ([]v1.ConfigMap, error) {
	configMapList, ok := response.(*v1.ConfigMapList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to ConfigMapList")
	}
	return configMapList.Items, nil
}

// ExtractSecrets extracts secrets from a Kubernetes client response
func ExtractSecrets(response interface{}) ([]v1.Secret, error) {
	secretList, ok := response.(*v1.SecretList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to SecretList")
	}
	return secretList.Items, nil
}

// ExtractNamespaces extracts namespaces from a Kubernetes client response
func ExtractNamespaces(response interface{}) ([]v1.Namespace, error) {
	namespaceList, ok := response.(*v1.NamespaceList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to NamespaceList")
	}
	return namespaceList.Items, nil
}

// ExtractStorageClasses extracts storage classes from a Kubernetes client response
func ExtractStorageClasses(response interface{}) ([]storagev1.StorageClass, error) {
	storageClassList, ok := response.(*storagev1.StorageClassList)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to StorageClassList")
	}
	return storageClassList.Items, nil
}

// ExtractBuckets extracts buckets from a MinIO client response
func ExtractBuckets(response interface{}) ([]minio.BucketInfo, error) {
	buckets, ok := response.([]minio.BucketInfo)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to []BucketInfo")
	}
	return buckets, nil
}

// ExtractPod extracts a single pod from a Kubernetes client response
func ExtractPod(response interface{}) (*v1.Pod, error) {
	pod, ok := response.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to Pod")
	}
	return pod, nil
}

// ExtractService extracts a single service from a Kubernetes client response
func ExtractService(response interface{}) (*v1.Service, error) {
	service, ok := response.(*v1.Service)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to Service")
	}
	return service, nil
}

// ExtractConfigMap extracts a single config map from a Kubernetes client response
func ExtractConfigMap(response interface{}) (*v1.ConfigMap, error) {
	configMap, ok := response.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to ConfigMap")
	}
	return configMap, nil
}

// ExtractSecret extracts a single secret from a Kubernetes client response
func ExtractSecret(response interface{}) (*v1.Secret, error) {
	secret, ok := response.(*v1.Secret)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to Secret")
	}
	return secret, nil
}

// ExtractNamespace extracts a single namespace from a Kubernetes client response
func ExtractNamespace(response interface{}) (*v1.Namespace, error) {
	namespace, ok := response.(*v1.Namespace)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to Namespace")
	}
	return namespace, nil
}

// ExtractStorageClass extracts a single storage class from a Kubernetes client response
func ExtractStorageClass(response interface{}) (*storagev1.StorageClass, error) {
	storageClass, ok := response.(*storagev1.StorageClass)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to StorageClass")
	}
	return storageClass, nil
}
