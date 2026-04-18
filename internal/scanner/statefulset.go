package scanner

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

func scanStatefulSets(ctx context.Context, client kubernetes.Interface, namespace, resourceName, resourceType, labelSelector string) ([]matcher.Result, error) {
	list, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing statefulsets in %q: %w", namespace, err)
	}

	var results []matcher.Result
	for _, s := range list.Items {
		refs := matcher.FindRefs(s.Namespace, "StatefulSet", s.Name, s.Spec.Template.Spec, resourceName, resourceType)
		results = append(results, refs...)
	}
	return results, nil
}

func scanStatefulSetsMulti(ctx context.Context, client kubernetes.Interface, namespace string, resourceNames []string, resourceType, labelSelector string) (map[string][]matcher.Result, error) {
	list, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing statefulsets in %q: %w", namespace, err)
	}
	out := make(map[string][]matcher.Result)
	for _, s := range list.Items {
		m := matcher.FindRefsMulti(s.Namespace, "StatefulSet", s.Name, s.Spec.Template.Spec, resourceNames, resourceType)
		for name, refs := range m {
			out[name] = append(out[name], refs...)
		}
	}
	return out, nil
}
