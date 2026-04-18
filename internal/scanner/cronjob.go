package scanner

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

func scanCronJobs(ctx context.Context, client kubernetes.Interface, namespace, resourceName, resourceType, labelSelector string) ([]matcher.Result, error) {
	list, err := client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing cronjobs in %q: %w", namespace, err)
	}

	var results []matcher.Result
	for _, cj := range list.Items {
		refs := matcher.FindRefs(cj.Namespace, "CronJob", cj.Name, cj.Spec.JobTemplate.Spec.Template.Spec, resourceName, resourceType)
		results = append(results, refs...)
	}
	return results, nil
}

func scanCronJobsMulti(ctx context.Context, client kubernetes.Interface, namespace string, resourceNames []string, resourceType, labelSelector string) (map[string][]matcher.Result, error) {
	list, err := client.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, fmt.Errorf("listing cronjobs in %q: %w", namespace, err)
	}
	out := make(map[string][]matcher.Result)
	for _, cj := range list.Items {
		m := matcher.FindRefsMulti(cj.Namespace, "CronJob", cj.Name, cj.Spec.JobTemplate.Spec.Template.Spec, resourceNames, resourceType)
		for name, refs := range m {
			out[name] = append(out[name], refs...)
		}
	}
	return out, nil
}
