package types

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
)

// Kubernetes 리소스 타입 정의
type (
	// Pod 관련
	PodList     = v1.PodList
	Pod         = v1.Pod
	PodResource interface {
		*PodList | *Pod
	}

	// ConfigMap 관련
	ConfigMapList     = v1.ConfigMapList
	ConfigMap         = v1.ConfigMap
	ConfigMapResource interface {
		*ConfigMapList | *ConfigMap
	}

	// Secret 관련
	SecretList     = v1.SecretList
	Secret         = v1.Secret
	SecretResource interface {
		*SecretList | *Secret
	}

	// Namespace 관련
	NamespaceList     = v1.NamespaceList
	Namespace         = v1.Namespace
	NamespaceResource interface {
		*NamespaceList | *Namespace
	}

	// StorageClass 관련
	StorageClassList     = storagev1.StorageClassList
	StorageClass         = storagev1.StorageClass
	StorageClassResource interface {
		*StorageClassList | *StorageClass
	}
)

// Kubernetes 타입 어설션 헬퍼 함수들

// 타입 어설션 헬퍼 함수들
func AssertPodList(v interface{}) (*PodList, bool) {
	podList, ok := v.(*PodList)
	return podList, ok
}

func AssertPod(v interface{}) (*Pod, bool) {
	pod, ok := v.(*Pod)
	return pod, ok
}

func AssertConfigMapList(v interface{}) (*ConfigMapList, bool) {
	cmList, ok := v.(*ConfigMapList)
	return cmList, ok
}

func AssertConfigMap(v interface{}) (*ConfigMap, bool) {
	cm, ok := v.(*ConfigMap)
	return cm, ok
}

func AssertSecretList(v interface{}) (*SecretList, bool) {
	secretList, ok := v.(*SecretList)
	return secretList, ok
}

func AssertSecret(v interface{}) (*Secret, bool) {
	secret, ok := v.(*Secret)
	return secret, ok
}

func AssertNamespaceList(v interface{}) (*NamespaceList, bool) {
	nsList, ok := v.(*NamespaceList)
	return nsList, ok
}

func AssertNamespace(v interface{}) (*Namespace, bool) {
	ns, ok := v.(*Namespace)
	return ns, ok
}

func AssertStorageClassList(v interface{}) (*StorageClassList, bool) {
	scList, ok := v.(*StorageClassList)
	return scList, ok
}

func AssertStorageClass(v interface{}) (*StorageClass, bool) {
	sc, ok := v.(*StorageClass)
	return sc, ok
}

// 안전한 타입 어설션을 위한 래퍼 함수들
func SafeGetPodList(v interface{}) (*PodList, error) {
	if podList, ok := AssertPodList(v); ok {
		return podList, nil
	}
	return nil, fmt.Errorf("expected *PodList, got %T", v)
}

func SafeGetPod(v interface{}) (*Pod, error) {
	if pod, ok := AssertPod(v); ok {
		return pod, nil
	}
	return nil, fmt.Errorf("expected *Pod, got %T", v)
}

func SafeGetConfigMapList(v interface{}) (*ConfigMapList, error) {
	if cmList, ok := AssertConfigMapList(v); ok {
		return cmList, nil
	}
	return nil, fmt.Errorf("expected *ConfigMapList, got %T", v)
}

func SafeGetConfigMap(v interface{}) (*ConfigMap, error) {
	if cm, ok := AssertConfigMap(v); ok {
		return cm, nil
	}
	return nil, fmt.Errorf("expected *ConfigMap, got %T", v)
}

func SafeGetSecretList(v interface{}) (*SecretList, error) {
	if secretList, ok := AssertSecretList(v); ok {
		return secretList, nil
	}
	return nil, fmt.Errorf("expected *SecretList, got %T", v)
}

func SafeGetSecret(v interface{}) (*Secret, error) {
	if secret, ok := AssertSecret(v); ok {
		return secret, nil
	}
	return nil, fmt.Errorf("expected *Secret, got %T", v)
}

func SafeGetNamespaceList(v interface{}) (*NamespaceList, error) {
	if nsList, ok := AssertNamespaceList(v); ok {
		return nsList, nil
	}
	return nil, fmt.Errorf("expected *NamespaceList, got %T", v)
}

func SafeGetNamespace(v interface{}) (*Namespace, error) {
	if ns, ok := AssertNamespace(v); ok {
		return ns, nil
	}
	return nil, fmt.Errorf("expected *Namespace, got %T", v)
}

func SafeGetStorageClassList(v interface{}) (*StorageClassList, error) {
	if scList, ok := AssertStorageClassList(v); ok {
		return scList, nil
	}
	return nil, fmt.Errorf("expected *StorageClassList, got %T", v)
}

func SafeGetStorageClass(v interface{}) (*StorageClass, error) {
	if sc, ok := AssertStorageClass(v); ok {
		return sc, nil
	}
	return nil, fmt.Errorf("expected *StorageClass, got %T", v)
}
