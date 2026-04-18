package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/k8s"
	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

var detectCmd = &cobra.Command{
	Use:     "detect <name> [name...]",
	Aliases: []string{"auto", "find"},
	Short:   "Auto-detect whether each name is a ConfigMap or Secret and find referencing workloads",
	Args:    cobra.MinimumNArgs(1),
	RunE:    runDetect,
}

func init() {
	rootCmd.AddCommand(detectCmd)
}

func runDetect(_ *cobra.Command, args []string) error {
	if err := validateOutputFlag(); err != nil {
		return err
	}
	opts, err := buildScanOpts()
	if err != nil {
		return err
	}
	client, contextNS, err := k8s.NewClient(flagContext, flagTimeout)
	if err != nil {
		return fmt.Errorf("connecting to cluster: %w", err)
	}
	ns := resolveNamespace(contextNS)

	return dispatch(func(ctx context.Context) ([]matcher.Result, error) {
		var all []matcher.Result
		for _, name := range args {
			resourceType, err := autoDetectType(ctx, client, name, ns)
			if err != nil {
				return nil, err
			}
			res, err := flatScan(ctx, client, []string{name}, resourceType, ns, opts)
			if err != nil {
				return nil, err
			}
			all = append(all, res...)
		}
		return all, nil
	})
}

// autoDetectType resolves whether resourceName is a ConfigMap or Secret.
// When namespace is empty (all-namespaces), it searches cluster-wide via a
// field selector on metadata.name.
func autoDetectType(ctx context.Context, client kubernetes.Interface, resourceName, namespace string) (string, error) {
	if namespace == "" {
		selector := metav1.ListOptions{FieldSelector: "metadata.name=" + resourceName, Limit: 1}
		cms, cmErr := client.CoreV1().ConfigMaps("").List(ctx, selector)
		if cmErr == nil && len(cms.Items) > 0 {
			return matcher.TypeConfigMap, nil
		}
		secs, secErr := client.CoreV1().Secrets("").List(ctx, selector)
		if secErr == nil && len(secs.Items) > 0 {
			return matcher.TypeSecret, nil
		}
		return "", fmt.Errorf("no ConfigMap or Secret named %q found in any namespace", resourceName)
	}

	_, cmErr := client.CoreV1().ConfigMaps(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if cmErr == nil {
		return matcher.TypeConfigMap, nil
	}
	_, secErr := client.CoreV1().Secrets(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if secErr == nil {
		return matcher.TypeSecret, nil
	}
	return "", fmt.Errorf("no ConfigMap or Secret named %q found in namespace %q", resourceName, namespace)
}
