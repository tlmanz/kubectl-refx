package scanner

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

func scanPods(ctx context.Context, client kubernetes.Interface, namespace, resourceName, resourceType, labelSelector string) ([]matcher.Result, error) {
	list, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing pods in %q: %w", namespace, err)
	}

	var results []matcher.Result
	for _, p := range list.Items {
		// Skip pods owned by higher-level controllers to avoid duplicate results.
		if isOwned(p.OwnerReferences) {
			continue
		}
		refs := matcher.FindRefs(p.Namespace, "Pod", p.Name, p.Spec, resourceName, resourceType)
		results = append(results, refs...)
	}
	return results, nil
}

func scanPodsMulti(ctx context.Context, client kubernetes.Interface, namespace string, resourceNames []string, resourceType, labelSelector string) (map[string][]matcher.Result, error) {
	list, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing pods in %q: %w", namespace, err)
	}
	out := make(map[string][]matcher.Result)
	for _, p := range list.Items {
		if isOwned(p.OwnerReferences) {
			continue
		}
		m := matcher.FindRefsMulti(p.Namespace, "Pod", p.Name, p.Spec, resourceNames, resourceType)
		for name, refs := range m {
			out[name] = append(out[name], refs...)
		}
	}
	return out, nil
}

func isOwned(owners []metav1.OwnerReference) bool {
	for _, o := range owners {
		switch o.Kind {
		case "ReplicaSet", "StatefulSet", "DaemonSet", "Job":
			return true
		}
	}
	return false
}
