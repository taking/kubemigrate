package types

import (
	"fmt"

	"github.com/minio/minio-go/v7"
)

// MinIO 리소스 타입 정의
type (
	// Bucket 관련
	BucketList     = []string
	Bucket         = string
	BucketResource interface {
		BucketList | Bucket
	}

	// Object 관련
	ObjectList     = []string
	Object         = string
	ObjectResource interface {
		ObjectList | Object
	}

	// Object 정보
	ObjectInfo = minio.ObjectInfo

	// Bucket 정보
	BucketInfo = minio.BucketInfo

	// MakeBucket 옵션
	MakeBucketOptions struct {
		Region string
	}
)

// MinIO 타입 어설션 헬퍼 함수들

// 타입 어설션 헬퍼 함수들
func AssertBucketList(v interface{}) (BucketList, bool) {
	bucketList, ok := v.(BucketList)
	return bucketList, ok
}

func AssertBucket(v interface{}) (Bucket, bool) {
	bucket, ok := v.(Bucket)
	return bucket, ok
}

func AssertObjectList(v interface{}) (ObjectList, bool) {
	objectList, ok := v.(ObjectList)
	return objectList, ok
}

func AssertObject(v interface{}) (Object, bool) {
	object, ok := v.(Object)
	return object, ok
}

func AssertObjectInfo(v interface{}) (ObjectInfo, bool) {
	objectInfo, ok := v.(ObjectInfo)
	return objectInfo, ok
}

func AssertBucketInfo(v interface{}) (BucketInfo, bool) {
	bucketInfo, ok := v.(BucketInfo)
	return bucketInfo, ok
}

// 안전한 타입 어설션을 위한 래퍼 함수들
func SafeGetBucketList(v interface{}) (BucketList, error) {
	if bucketList, ok := AssertBucketList(v); ok {
		return bucketList, nil
	}
	return nil, fmt.Errorf("expected BucketList, got %T", v)
}

func SafeGetBucket(v interface{}) (Bucket, error) {
	if bucket, ok := AssertBucket(v); ok {
		return bucket, nil
	}
	return "", fmt.Errorf("expected Bucket, got %T", v)
}

func SafeGetObjectList(v interface{}) (ObjectList, error) {
	if objectList, ok := AssertObjectList(v); ok {
		return objectList, nil
	}
	return nil, fmt.Errorf("expected ObjectList, got %T", v)
}

func SafeGetObject(v interface{}) (Object, error) {
	if object, ok := AssertObject(v); ok {
		return object, nil
	}
	return "", fmt.Errorf("expected Object, got %T", v)
}

func SafeGetObjectInfo(v interface{}) (ObjectInfo, error) {
	if objectInfo, ok := AssertObjectInfo(v); ok {
		return objectInfo, nil
	}
	return ObjectInfo{}, fmt.Errorf("expected ObjectInfo, got %T", v)
}

func SafeGetBucketInfo(v interface{}) (BucketInfo, error) {
	if bucketInfo, ok := AssertBucketInfo(v); ok {
		return bucketInfo, nil
	}
	return BucketInfo{}, fmt.Errorf("expected BucketInfo, got %T", v)
}
