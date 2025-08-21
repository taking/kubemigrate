package validation

import (
	"fmt"
	"regexp"
	"strings"

	"taking.kr/velero/models"
)

type RequestValidator struct {
	kubeconfigPattern *regexp.Regexp
	namespacePattern  *regexp.Regexp
}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		kubeconfigPattern: regexp.MustCompile(`apiVersion:\s*v1`),
		namespacePattern:  regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
	}
}

func (v *RequestValidator) ValidateVeleroRequest(req *models.VeleroRequest) error {
	if req.SourceKubeconfig == "" {
		return fmt.Errorf("sourceKubeconfig is required")
	}

	if len(req.SourceKubeconfig) > 100000 { // 100KB limit
		return fmt.Errorf("sourceKubeconfig too large (max 100KB)")
	}

	// Basic kubeconfig format validation
	if !strings.Contains(req.SourceKubeconfig, "apiVersion") {
		return fmt.Errorf("sourceKubeconfig appears to be invalid (missing apiVersion)")
	}

	if req.Namespace != "" {
		if !v.isValidNamespace(req.Namespace) {
			return fmt.Errorf("invalid namespace format: must be lowercase alphanumeric with hyphens")
		}
	}

	// Validate destination kubeconfig if provided
	if req.DestinationKubeconfig != "" {
		if len(req.DestinationKubeconfig) > 100000 {
			return fmt.Errorf("destinationKubeconfig too large (max 100KB)")
		}

		if !strings.Contains(req.DestinationKubeconfig, "apiVersion") {
			return fmt.Errorf("destinationKubeconfig appears to be invalid (missing apiVersion)")
		}
	}

	return nil
}

func (v *RequestValidator) isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return v.namespacePattern.MatchString(namespace)
}
