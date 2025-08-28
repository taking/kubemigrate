package validation

import (
	"fmt"
	"regexp"
	"strings"
	"taking.kr/velero/utils"

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

func (v *RequestValidator) ValidateKubeConfigRequest(req *models.KubeConfig) (string, error) {
	if req.KubeConfig == "" {
		return "", fmt.Errorf("KubeConfig is required")
	}

	if len(req.KubeConfig) > 100000 { // 100KB limit
		return "", fmt.Errorf("KubeConfig too large (max 100KB)")
	}

	decodeSourceKubeConfig, _ := utils.DecodeIfBase64(req.KubeConfig)

	// Basic kubeconfig format validation
	if !strings.Contains(decodeSourceKubeConfig, "apiVersion") {
		return "", fmt.Errorf("KubeConfig appears to be invalid (missing apiVersion)")
	}

	if req.Namespace != "" {
		if !v.isValidNamespace(req.Namespace) {
			return "", fmt.Errorf("invalid namespace format: must be lowercase alphanumeric with hyphens")
		}
	}

	return decodeSourceKubeConfig, nil
}

func (v *RequestValidator) isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return v.namespacePattern.MatchString(namespace)
}
