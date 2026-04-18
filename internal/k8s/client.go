package k8s

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient creates a Kubernetes clientset from the given kubeconfig context.
// An empty contextOverride uses the current context.
func NewClient(contextOverride string, timeout time.Duration) (kubernetes.Interface, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	configOverrides := &clientcmd.ConfigOverrides{}
	if contextOverride != "" {
		configOverrides.CurrentContext = contextOverride
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("building REST config: %w", err)
	}
	restConfig.Timeout = timeout

	ns, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("resolving namespace: %w", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, "", fmt.Errorf("creating clientset: %w", err)
	}

	return client, ns, nil
}
