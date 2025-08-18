package clients

import (
	"fmt"
)

//type veleroClientMulti struct {
//	src  interfaces.VeleroService
//	dest interfaces.VeleroService
//}

func NewClientFromRawConfig(clientType, rawConfig string) (interface{}, error) {
	switch clientType {
	case "kubernetes":
		return NewKubeClientFromRawConfig(rawConfig)
	case "velero":
		return NewVeleroClientFromRawConfig(rawConfig)
	default:
		return nil, fmt.Errorf("unknown clientType: %s", clientType)
	}
}

//func NewVeleroClientFromSourceAndDestConfigs(srcRawConfig, destRawConfig string) (interfaces.VeleroService, error) {
//	srcClient, err := NewVeleroClientFromRawConfig(srcRawConfig)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create source velero client: %w", err)
//	}
//	destClient, err := NewVeleroClientFromRawConfig(destRawConfig)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create destination velero client: %w", err)
//	}
//
//	return &veleroClientMulti{
//		src:  srcClient,
//		dest: destClient,
//	}, nil
//}
//
//func (v *veleroClientMulti) GetBackups(ctx context.Context, namespace string) ([]velerov1.Backup, error) {
//	return v.src.GetBackups(ctx, namespace)
//}
