package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"time"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tlmanz/kubectl-refx/internal/k8s"
	"github.com/tlmanz/kubectl-refx/internal/matcher"
	"github.com/tlmanz/kubectl-refx/internal/output"
	"github.com/tlmanz/kubectl-refx/internal/scanner"
)

var unusedCmd = &cobra.Command{
	Use:   "unused",
	Short: "Find ConfigMaps or Secrets not referenced by any workload",
}

var unusedConfigmapCmd = &cobra.Command{
	Use:     "configmap",
	Aliases: []string{"cm", "configmaps"},
	Short:   "Find ConfigMaps not referenced by any workload",
	RunE:    runUnusedConfigmap,
}

var unusedSecretCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"sec", "secrets"},
	Short:   "Find Secrets not referenced by any workload",
	RunE:    runUnusedSecret,
}

func init() {
	unusedCmd.AddCommand(unusedConfigmapCmd)
	unusedCmd.AddCommand(unusedSecretCmd)
	rootCmd.AddCommand(unusedCmd)
}

func runUnusedConfigmap(_ *cobra.Command, _ []string) error {
	return runUnused(matcher.TypeConfigMap)
}

func runUnusedSecret(_ *cobra.Command, _ []string) error {
	return runUnused(matcher.TypeSecret)
}

func runUnused(resourceType string) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), flagTimeout+5*time.Second)
	defer cancel()

	// List all resources of the given type.
	type resourceItem struct {
		Key       string
		Namespace string
		Name      string
	}
	var resources []resourceItem
	var resourceKeys []string
	seenKeys := make(map[string]struct{})
	makeKey := func(namespace, name string) string {
		return namespace + "/" + name
	}

	if resourceType == matcher.TypeConfigMap {
		list, err := client.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("listing configmaps: %w", err)
		}
		for _, cm := range list.Items {
			if !flagIncludeSystem && isSystemConfigMap(cm.Name) {
				continue
			}
			key := makeKey(cm.Namespace, cm.Name)
			if _, ok := seenKeys[key]; ok {
				continue
			}
			seenKeys[key] = struct{}{}
			resources = append(resources, resourceItem{
				Key:       key,
				Namespace: cm.Namespace,
				Name:      cm.Name,
			})
			resourceKeys = append(resourceKeys, key)
		}
	} else {
		list, err := client.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("listing secrets: %w", err)
		}
		for _, s := range list.Items {
			if !flagIncludeSystem && isSystemSecret(s.Type) {
				continue
			}
			key := makeKey(s.Namespace, s.Name)
			if _, ok := seenKeys[key]; ok {
				continue
			}
			seenKeys[key] = struct{}{}
			resources = append(resources, resourceItem{
				Key:       key,
				Namespace: s.Namespace,
				Name:      s.Name,
			})
			resourceKeys = append(resourceKeys, key)
		}
	}

	if len(resources) == 0 {
		fmt.Fprintln(os.Stderr, "No resources found.")
		return nil
	}

	byKey, err := scanner.ScanMultiple(ctx, client, resourceKeys, resourceType, ns, opts)
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	var unused []output.RefCount
	for _, resource := range resources {
		if len(byKey[resource.Key]) == 0 {
			unused = append(unused, output.RefCount{
				Namespace: resource.Namespace,
				Name:      resource.Name,
			})
		}
	}

	sort.Slice(unused, func(i, j int) bool {
		if unused[i].Namespace != unused[j].Namespace {
			return unused[i].Namespace < unused[j].Namespace
		}
		return unused[i].Name < unused[j].Name
	})

	switch flagOutput {
	case "json":
		return output.WriteRefCountJSON(os.Stdout, unused)
	case "yaml":
		return output.WriteRefCountYAML(os.Stdout, unused)
	default:
		output.WriteUnusedTable(os.Stdout, unused, tableOpts())
		return nil
	}
}
