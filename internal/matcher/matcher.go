package matcher

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	TypeConfigMap = "configmap"
	TypeSecret    = "secret"
)

// Result represents a single reference found in a workload.
type Result struct {
	Namespace     string
	Kind          string
	Name          string
	Container     string // container name; empty for volume refs (pod-level)
	ReferenceType string // "env" | "envFrom" | "volume"
	Detail        string // e.g. "env:MY_KEY", "envFrom", "volume:my-vol"
}

// FindRefs scans a PodSpec for references to the named configmap or secret.
func FindRefs(namespace, kind, workloadName string, spec corev1.PodSpec, resourceName, resourceType string) []Result {
	var results []Result

	allContainers := make([]corev1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
	allContainers = append(allContainers, spec.InitContainers...)
	allContainers = append(allContainers, spec.Containers...)

	for _, c := range allContainers {
		for _, env := range c.Env {
			if env.ValueFrom == nil {
				continue
			}
			if resourceType == TypeConfigMap && env.ValueFrom.ConfigMapKeyRef != nil &&
				env.ValueFrom.ConfigMapKeyRef.Name == resourceName {
				results = append(results, Result{
					Namespace:     namespace,
					Kind:          kind,
					Name:          workloadName,
					Container:     c.Name,
					ReferenceType: "env",
					Detail:        "env:" + env.Name,
				})
			}
			if resourceType == TypeSecret && env.ValueFrom.SecretKeyRef != nil &&
				env.ValueFrom.SecretKeyRef.Name == resourceName {
				results = append(results, Result{
					Namespace:     namespace,
					Kind:          kind,
					Name:          workloadName,
					Container:     c.Name,
					ReferenceType: "env",
					Detail:        "env:" + env.Name,
				})
			}
		}

		for _, ef := range c.EnvFrom {
			if resourceType == TypeConfigMap && ef.ConfigMapRef != nil &&
				ef.ConfigMapRef.Name == resourceName {
				results = append(results, Result{
					Namespace:     namespace,
					Kind:          kind,
					Name:          workloadName,
					Container:     c.Name,
					ReferenceType: "envFrom",
					Detail:        "envFrom",
				})
			}
			if resourceType == TypeSecret && ef.SecretRef != nil &&
				ef.SecretRef.Name == resourceName {
				results = append(results, Result{
					Namespace:     namespace,
					Kind:          kind,
					Name:          workloadName,
					Container:     c.Name,
					ReferenceType: "envFrom",
					Detail:        "envFrom",
				})
			}
		}
	}

	// Volumes are pod-level, not container-level.
	for _, vol := range spec.Volumes {
		if resourceType == TypeConfigMap && vol.ConfigMap != nil &&
			vol.ConfigMap.Name == resourceName {
			results = append(results, Result{
				Namespace:     namespace,
				Kind:          kind,
				Name:          workloadName,
				Container:     "",
				ReferenceType: "volume",
				Detail:        "volume:" + vol.Name,
			})
		}
		if resourceType == TypeSecret && vol.Secret != nil &&
			vol.Secret.SecretName == resourceName {
			results = append(results, Result{
				Namespace:     namespace,
				Kind:          kind,
				Name:          workloadName,
				Container:     "",
				ReferenceType: "volume",
				Detail:        "volume:" + vol.Name,
			})
		}
	}

	return results
}

// FindRefsMulti scans a PodSpec once and returns references for any of the
// given resource names, keyed by name. Names not referenced are omitted.
func FindRefsMulti(namespace, kind, workloadName string, spec corev1.PodSpec, resourceNames []string, resourceType string) map[string][]Result {
	if len(resourceNames) == 0 {
		return nil
	}
	wanted := make(map[string]struct{}, len(resourceNames))
	for _, n := range resourceNames {
		wanted[n] = struct{}{}
	}

	out := make(map[string][]Result)
	add := func(name string, r Result) {
		out[name] = append(out[name], r)
	}
	addMatches := func(refName string, r Result) {
		if _, ok := wanted[refName]; ok {
			add(refName, r)
		}
		// Support namespace-qualified keys (namespace/name) so callers can
		// disambiguate same-named resources across namespaces.
		nsKey := namespace + "/" + refName
		if _, ok := wanted[nsKey]; ok {
			add(nsKey, r)
		}
	}

	allContainers := make([]corev1.Container, 0, len(spec.InitContainers)+len(spec.Containers))
	allContainers = append(allContainers, spec.InitContainers...)
	allContainers = append(allContainers, spec.Containers...)

	for _, c := range allContainers {
		for _, env := range c.Env {
			if env.ValueFrom == nil {
				continue
			}
			if resourceType == TypeConfigMap && env.ValueFrom.ConfigMapKeyRef != nil {
				addMatches(env.ValueFrom.ConfigMapKeyRef.Name, Result{
					Namespace: namespace, Kind: kind, Name: workloadName,
					Container: c.Name, ReferenceType: "env", Detail: "env:" + env.Name,
				})
			}
			if resourceType == TypeSecret && env.ValueFrom.SecretKeyRef != nil {
				addMatches(env.ValueFrom.SecretKeyRef.Name, Result{
					Namespace: namespace, Kind: kind, Name: workloadName,
					Container: c.Name, ReferenceType: "env", Detail: "env:" + env.Name,
				})
			}
		}

		for _, ef := range c.EnvFrom {
			if resourceType == TypeConfigMap && ef.ConfigMapRef != nil {
				addMatches(ef.ConfigMapRef.Name, Result{
					Namespace: namespace, Kind: kind, Name: workloadName,
					Container: c.Name, ReferenceType: "envFrom", Detail: "envFrom",
				})
			}
			if resourceType == TypeSecret && ef.SecretRef != nil {
				addMatches(ef.SecretRef.Name, Result{
					Namespace: namespace, Kind: kind, Name: workloadName,
					Container: c.Name, ReferenceType: "envFrom", Detail: "envFrom",
				})
			}
		}
	}

	for _, vol := range spec.Volumes {
		if resourceType == TypeConfigMap && vol.ConfigMap != nil {
			addMatches(vol.ConfigMap.Name, Result{
				Namespace: namespace, Kind: kind, Name: workloadName,
				Container: "", ReferenceType: "volume", Detail: "volume:" + vol.Name,
			})
		}
		if resourceType == TypeSecret && vol.Secret != nil {
			addMatches(vol.Secret.SecretName, Result{
				Namespace: namespace, Kind: kind, Name: workloadName,
				Container: "", ReferenceType: "volume", Detail: "volume:" + vol.Name,
			})
		}
	}

	return out
}
