// Package routes MinIO 관련 라우트를 관리합니다.
package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/taking/kubemigrate/internal/api/minio"
)

// SetupMinioRoutes MinIO 관련 라우트를 설정합니다.
func SetupMinioRoutes(e *echo.Echo, minioHandler *minio.Handler) {
	api := e.Group("/api/v1")
	minioGroup := api.Group("/minio")

	// 헬스체크
	minioGroup.POST("/health", minioHandler.HealthCheck)

	// 버킷 관리 라우트 (RESTful)
	minioGroup.GET("/buckets", minioHandler.GetBuckets)                // 버킷 목록 조회
	minioGroup.GET("/buckets/:bucket", minioHandler.CheckBucketExists) // 버킷 존재 확인
	minioGroup.POST("/buckets/:bucket", minioHandler.CreateBucket)     // 버킷 생성
	minioGroup.DELETE("/buckets/:bucket", minioHandler.DeleteBucket)   // 버킷 삭제

	// 객체 관리 라우트 (RESTful)
	minioGroup.GET("/buckets/:bucket/objects", minioHandler.GetObjects)                                           // 객체 목록 조회
	minioGroup.POST("/buckets/:bucket/objects/*", minioHandler.PutObject)                                         // 객체 업로드
	minioGroup.GET("/buckets/:bucket/objects/*", minioHandler.GetObject)                                          // 객체 다운로드
	minioGroup.GET("/buckets/:bucket/objects/*", minioHandler.StatObject)                                         // 객체 정보 조회
	minioGroup.POST("/buckets/:srcBucket/objects/:srcObject/copy/:dstBucket/:dstObject", minioHandler.CopyObject) // 객체 복사
	minioGroup.DELETE("/buckets/:bucket/objects/*", minioHandler.DeleteObject)                                    // 객체 삭제

	// Presigned URL 라우트
	minioGroup.GET("/buckets/:bucket/objects/:object/presigned-get", minioHandler.PresignedGetObject) // Presigned GET URL 생성
	minioGroup.PUT("/buckets/:bucket/objects/:object/presigned-put", minioHandler.PresignedPutObject) // Presigned PUT URL 생성
}
