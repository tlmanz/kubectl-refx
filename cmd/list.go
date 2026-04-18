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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ConfigMaps or Secrets with their reference counts",
}

var listConfigmapCmd = &cobra.Command{
	Use:     "configmap",
	Aliases: []string{"cm", "configmaps"},
	Short:   "List ConfigMaps with reference counts",
	RunE:    runListConfigmap,
}

var listSecretCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"sec", "secrets"},
	Short:   "List Secrets with reference counts",
	RunE:    runListSecret,
}

func init() {
	listCmd.AddCommand(listConfigmapCmd)
	listCmd.AddCommand(listSecretCmd)
	rootCmd.AddCommand(listCmd)
}

func runListConfigmap(_ *cobra.Command, _ []string) error {
	return runList(matcher.TypeConfigMap)
}

func runListSecret(_ *cobra.Command, _ []string) error {
	return runList(matcher.TypeSecret)
}

func runList(resourceType string) error {
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

	counts := make([]output.RefCount, 0, len(resources))
	for _, resource := range resources {
		counts = append(counts, output.RefCount{
			Namespace:  resource.Namespace,
			Name:       resource.Name,
			References: len(byKey[resource.Key]),
		})
	}

	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Namespace != counts[j].Namespace {
			return counts[i].Namespace < counts[j].Namespace
		}
		return counts[i].Name < counts[j].Name
	})

	switch flagOutput {
	case "json":
		return output.WriteRefCountJSON(os.Stdout, counts)
	case "yaml":
		return output.WriteRefCountYAML(os.Stdout, counts)
	default:
		output.WriteRefCountTable(os.Stdout, counts, tableOpts())
		return nil
	}
}
