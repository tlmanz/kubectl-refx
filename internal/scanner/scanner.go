package scanner

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

// Options controls which workloads are scanned.
type Options struct {
	// Kinds filters workload kinds to scan. Empty means all.
	// Valid values: Deployment, StatefulSet, DaemonSet, Job, CronJob, Pod
	Kinds []string
	// LabelSelector filters workloads by label (passed to ListOptions).
	LabelSelector string
}

type scanFn func(ctx context.Context, client kubernetes.Interface, namespace, resourceName, resourceType, labelSelector string) ([]matcher.Result, error)

type scanMultiFn func(ctx context.Context, client kubernetes.Interface, namespace string, resourceNames []string, resourceType, labelSelector string) (map[string][]matcher.Result, error)

// AllKinds is the canonical ordered list of supported workload kinds.
var AllKinds = []string{"Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob", "Pod"}

var registry = map[string]scanFn{
	"deployment":  scanDeployments,
	"statefulset": scanStatefulSets,
	"daemonset":   scanDaemonSets,
	"job":         scanJobs,
	"cronjob":     scanCronJobs,
	"pod":         scanPods,
}

var multiRegistry = map[string]scanMultiFn{
	"deployment":  scanDeploymentsMulti,
	"statefulset": scanStatefulSetsMulti,
	"daemonset":   scanDaemonSetsMulti,
	"job":         scanJobsMulti,
	"cronjob":     scanCronJobsMulti,
	"pod":         scanPodsMulti,
}

// NormalizeKind converts user-supplied kind strings (including aliases) to
// the canonical form used in AllKinds.
func NormalizeKind(k string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(k)) {
	case "deployment", "deployments", "deploy":
		return "Deployment", nil
	case "statefulset", "statefulsets", "sts":
		return "StatefulSet", nil
	case "daemonset", "daemonsets", "ds":
		return "DaemonSet", nil
	case "job", "jobs":
		return "Job", nil
	case "cronjob", "cronjobs", "cj":
		return "CronJob", nil
	case "pod", "pods", "po":
		return "Pod", nil
	default:
		return "", fmt.Errorf("unknown workload kind %q (valid: Deployment, StatefulSet, DaemonSet, Job, CronJob, Pod)", k)
	}
}

// Scan concurrently scans workload kinds for references to resourceName.
// namespace="" scans all namespaces.
func Scan(ctx context.Context, client kubernetes.Interface, resourceName, resourceType, namespace string, opts Options) ([]matcher.Result, error) {
	fns, err := resolveScanners(opts.Kinds)
	if err != nil {
		return nil, err
	}

	var (
		mu      sync.Mutex
		results []matcher.Result
	)

	g, gctx := errgroup.WithContext(ctx)
	for _, fn := range fns {
		fn := fn
		g.Go(func() error {
			res, err := fn(gctx, client, namespace, resourceName, resourceType, opts.LabelSelector)
			if err != nil {
				return err
			}
			mu.Lock()
			results = append(results, res...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// ScanMultiple lists each workload kind ONCE and matches against all
// resourceNames in a single pass, returning results keyed by resource name.
// Each enabled kind fans out concurrently, so the total API call count is
// bounded by the number of kinds (not by len(resourceNames)).
func ScanMultiple(ctx context.Context, client kubernetes.Interface, resourceNames []string, resourceType, namespace string, opts Options) (map[string][]matcher.Result, error) {
	fns, err := resolveMultiScanners(opts.Kinds)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	out := make(map[string][]matcher.Result, len(resourceNames))
	// Pre-populate so callers can distinguish "no refs" from "not scanned".
	for _, name := range resourceNames {
		out[name] = nil
	}

	g, gctx := errgroup.WithContext(ctx)
	for _, fn := range fns {
		fn := fn
		g.Go(func() error {
			partial, err := fn(gctx, client, namespace, resourceNames, resourceType, opts.LabelSelector)
			if err != nil {
				return err
			}
			mu.Lock()
			for name, refs := range partial {
				out[name] = append(out[name], refs...)
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func resolveScanners(kinds []string) ([]scanFn, error) {
	if len(kinds) == 0 {
		fns := make([]scanFn, 0, len(AllKinds))
		for _, k := range AllKinds {
			fns = append(fns, registry[strings.ToLower(k)])
		}
		return fns, nil
	}

	seen := make(map[string]bool)
	var fns []scanFn
	for _, k := range kinds {
		norm, err := NormalizeKind(k)
		if err != nil {
			return nil, err
		}
		if seen[norm] {
			continue
		}
		seen[norm] = true
		fns = append(fns, registry[strings.ToLower(norm)])
	}
	return fns, nil
}

func resolveMultiScanners(kinds []string) ([]scanMultiFn, error) {
	if len(kinds) == 0 {
		fns := make([]scanMultiFn, 0, len(AllKinds))
		for _, k := range AllKinds {
			fns = append(fns, multiRegistry[strings.ToLower(k)])
		}
		return fns, nil
	}

	seen := make(map[string]bool)
	var fns []scanMultiFn
	for _, k := range kinds {
		norm, err := NormalizeKind(k)
		if err != nil {
			return nil, err
		}
		if seen[norm] {
			continue
		}
		seen[norm] = true
		fns = append(fns, multiRegistry[strings.ToLower(norm)])
	}
	return fns, nil
}
