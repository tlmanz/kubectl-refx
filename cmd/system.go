package cmd

import (
	corev1 "k8s.io/api/core/v1"
)

// isSystemConfigMap returns true for auto-generated ConfigMaps Kubernetes
// creates in every namespace (e.g. the kube-root-ca.crt published by the
// RootCAConfigMap controller).
func isSystemConfigMap(name string) bool {
	return name == "kube-root-ca.crt"
}

// isSystemSecret returns true for Secret types that Kubernetes generates
// automatically (service account tokens, docker pull credentials).
func isSystemSecret(t corev1.SecretType) bool {
	switch t {
	case corev1.SecretTypeServiceAccountToken,
		corev1.SecretTypeDockercfg,
		corev1.SecretTypeDockerConfigJson:
		return true
	}
	return false
}
