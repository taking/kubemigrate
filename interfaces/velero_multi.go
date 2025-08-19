package interfaces

import (
	"context"
)

// MultiClusterConfig holds configurations for source and destination clusters
type MultiClusterConfig struct {
	SourceKubeconfig      string `json:"sourceKubeconfig"`
	DestinationKubeconfig string `json:"destinationKubeconfig,omitempty"`
	Namespace             string `json:"namespace,omitempty"`
}

// MultiClusterVeleroService supports operations across multiple clusters
type MultiClusterVeleroService interface {
	VeleroService

	// Cross-cluster operations
	ValidateDestinationCluster(ctx context.Context, destConfig string) error
	MigrateBackup(ctx context.Context, backupName string, destConfig string) error
	CompareStorageClasses(ctx context.Context, destConfig string) (*StorageClassComparison, error)
}

// StorageClassComparison represents comparison result between clusters
type StorageClassComparison struct {
	SourceClasses      []string `json:"sourceClasses"`
	DestinationClasses []string `json:"destinationClasses"`
	MissingInDest      []string `json:"missingInDest"`
	Compatible         bool     `json:"compatible"`
}
