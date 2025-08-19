package models

type VeleroRequest struct {
	SourceKubeconfig      string `json:"sourceKubeconfig"`
	DestinationKubeconfig string `json:"destinationKubeconfig,omitempty"`
	Namespace             string `json:"namespace,omitempty"`
}
