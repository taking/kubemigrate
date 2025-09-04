package service

// import (
// 	"context"
// 	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
// 	"sort"
// 	"taking.kr/velero/pkg/client"
// 	"taking.kr/velero/pkg/model"
// 	"time"
// )

// type ResourceService struct {
// 	factory *client.ClientFactory
// }

// func NewResourceService() *ResourceService {
// 	return &ResourceService{
// 		factory: client.NewClientFactory(),
// 	}
// }

// // BackupSummary represents backup statistics
// type BackupSummary struct {
// 	Total           int `json:"total"`
// 	Completed       int `json:"completed"`
// 	Failed          int `json:"failed"`
// 	InProgress      int `json:"inProgress"`
// 	PartiallyFailed int `json:"partiallyFailed"`
// 	Recent          int `json:"recent"`
// 	Expired         int `json:"expired"`
// }

// func (s *ResourceService) GenerateBackupSummary(backups []velerov1.Backup) BackupSummary {
// 	summary := BackupSummary{Total: len(backups)}

// 	for _, b := range backups {
// 		switch b.Status.Phase {
// 		case velerov1.BackupPhaseCompleted:
// 			summary.Completed++
// 		case velerov1.BackupPhaseFailed:
// 			summary.Failed++
// 		case velerov1.BackupPhaseInProgress:
// 			summary.InProgress++
// 		case velerov1.BackupPhasePartiallyFailed:
// 			summary.PartiallyFailed++
// 		}

// 		if time.Since(b.CreationTimestamp.Time) < 24*time.Hour {
// 			summary.Recent++
// 		}

// 		if b.Status.Expiration != nil && b.Status.Expiration.Time.Before(time.Now()) {
// 			summary.Expired++
// 		}
// 	}

// 	return summary
// }

// func (s *ResourceService) GetVeleroResourcesWithSummary(cfg model.KubeConfig) (map[string]interface{}, error) {
// 	client, err := clients.NewVeleroClient(cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ctx := context.Background()

// 	backups, err := client.GetBackups(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	restores, err := client.GetRestores(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Sort by creation time
// 	sort.Slice(backups, func(i, j int) bool {
// 		return backups[i].CreationTimestamp.Time.After(backups[j].CreationTimestamp.Time)
// 	})

// 	sort.Slice(restores, func(i, j int) bool {
// 		return restores[i].CreationTimestamp.Time.After(restores[j].CreationTimestamp.Time)
// 	})

// 	summary := s.GenerateBackupSummary(backups)

// 	return map[string]interface{}{
// 		"backups":  backups,
// 		"restores": restores,
// 		"summary":  summary,
// 	}, nil
// }
