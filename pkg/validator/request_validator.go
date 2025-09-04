package validator

import (
	"fmt"
	"github.com/taking/velero/pkg"
	"regexp"
	"strings"

	"github.com/taking/velero/internal/model"
)

type KubernetesValidator struct {
	kubeconfigPattern *regexp.Regexp
	namespacePattern  *regexp.Regexp
}

func NewKubernetesValidator() *KubernetesValidator {
	return &KubernetesValidator{
		kubeconfigPattern: regexp.MustCompile(`apiVersion:\s*v1`),
		namespacePattern:  regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
	}
}

func (v *KubernetesValidator) ValidateKubernetesConfig(req *model.KubeConfig) (string, error) {
	if req.KubeConfig == "" {
		return "", fmt.Errorf("kubeconfig is required")
	}

	if len(req.KubeConfig) > 100000 { // 100KB limit
		return "", fmt.Errorf("kubeconfig too large (max 100KB)")
	}

	decodeSourceKubeConfig, _ := pkg.DecodeIfBase64(req.KubeConfig)

	if !strings.Contains(decodeSourceKubeConfig, "apiVersion") {
		return "", fmt.Errorf("kubeconfig appears to be invalid (missing apiVersion)")
	}

	if req.Namespace != "" {
		if !v.isValidNamespace(req.Namespace) {
			return "", fmt.Errorf("invalid namespace format: must be lowercase alphanumeric with hyphens")
		}
	}

	return decodeSourceKubeConfig, nil
}

func (v *KubernetesValidator) isValidNamespace(namespace string) bool {
	if len(namespace) == 0 || len(namespace) > 63 {
		return false
	}

	return v.namespacePattern.MatchString(namespace)
}
