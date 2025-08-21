package models

import (
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"time"
)

type ErrorResponse struct {
	Error     string    `json:"error"`
	Operation string    `json:"operation"`
	Namespace string    `json:"namespace"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"requestId"`
	Duration  string    `json:"duration"`
}

type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Success   bool        `json:"success"`
	Operation string      `json:"operation"`
	Namespace string      `json:"namespace"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"requestId"`
	Duration  string      `json:"duration"`
}

type BackupValidationRequest struct {
	VeleroRequest
	BackupName string `json:"backupName" validate:"required"`
}

type BackupValidationResponse struct {
	BackupName      string     `json:"backupName"`
	IsValid         bool       `json:"isValid"`
	Warnings        []string   `json:"warnings"`
	Errors          []string   `json:"errors"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"createdAt"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	Size            string     `json:"size"`
	Timestamp       time.Time  `json:"timestamp"`
	Recommendations []string   `json:"recommendations,omitempty"`
}

type ValidationResult struct {
	IsValid         bool     `json:"isValid"`
	Warnings        []string `json:"warnings"`
	Errors          []string `json:"errors"`
	Recommendations []string `json:"recommendations,omitempty"`
}

type BackupSummary struct {
	Total           int `json:"total"`
	Completed       int `json:"completed"`
	Failed          int `json:"failed"`
	InProgress      int `json:"inProgress"`
	PartiallyFailed int `json:"partiallyFailed"`
	Recent          int `json:"recent"` // Last 24 hours
	Expired         int `json:"expired"`
}

type RestoreSummary struct {
	Total           int `json:"total"`
	Completed       int `json:"completed"`
	Failed          int `json:"failed"`
	InProgress      int `json:"inProgress"`
	PartiallyFailed int `json:"partiallyFailed"`
	Recent          int `json:"recent"` // Last 24 hours
}

type StorageLocationSummary struct {
	Total       int `json:"total"`
	Available   int `json:"available"`
	Unavailable int `json:"unavailable"`
}

type BackupDetails struct {
	Name               string                   `json:"name"`
	Namespace          string                   `json:"namespace"`
	Status             string                   `json:"status"`
	CreatedAt          time.Time                `json:"createdAt"`
	CompletedAt        *time.Time               `json:"completedAt,omitempty"`
	ExpiresAt          *time.Time               `json:"expiresAt,omitempty"`
	Size               string                   `json:"size"`
	Progress           *velerov1.BackupProgress `json:"progress,omitempty"`
	IncludedNamespaces []string                 `json:"includedNamespaces,omitempty"`
	ExcludedNamespaces []string                 `json:"excludedNamespaces,omitempty"`
	IncludedResources  []string                 `json:"includedResources,omitempty"`
	ExcludedResources  []string                 `json:"excludedResources,omitempty"`
	StorageLocation    string                   `json:"storageLocation"`
	Labels             map[string]string        `json:"labels,omitempty"`
	Annotations        map[string]string        `json:"annotations,omitempty"`
	ValidationErrors   []string                 `json:"validationErrors,omitempty"`
}

type StorageComparison struct {
	SourceBackupLocations      []string `json:"sourceBackupLocations"`
	DestinationBackupLocations []string `json:"destinationBackupLocations"`
	SourceVolumeLocations      []string `json:"sourceVolumeLocations"`
	DestinationVolumeLocations []string `json:"destinationVolumeLocations"`
	MissingBackupLocations     []string `json:"missingBackupLocations"`
	MissingVolumeLocations     []string `json:"missingVolumeLocations"`
	Compatible                 bool     `json:"compatible"`
	Recommendations            []string `json:"recommendations"`
}
