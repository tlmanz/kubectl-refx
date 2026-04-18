package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	flagNamespace     string
	flagAllNamespaces bool
	flagOutput        string
	flagTimeout       time.Duration
	flagContext       string
	flagKinds         string
	flagSelector      string
	flagNoHeaders     bool
	flagWatch         bool
	flagWatchInterval time.Duration
	flagIncludeSystem bool
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-refx",
	Short: "Find which Kubernetes workloads reference a given ConfigMap or Secret",
	Long: `kubectl-refx scans your cluster workloads and reports every Deployment,
StatefulSet, DaemonSet, Job, CronJob, and Pod that references a named
ConfigMap or Secret via env, envFrom, or volume mounts.`,
}

// Execute runs the root cobra command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagNamespace, "namespace", "n", "", "target namespace (defaults to current context namespace)")
	rootCmd.PersistentFlags().BoolVarP(&flagAllNamespaces, "all-namespaces", "A", false, "scan all namespaces")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "table", "output format: table, wide, json, yaml")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "request timeout per scan")
	rootCmd.PersistentFlags().StringVar(&flagContext, "context", "", "kubeconfig context override")
	rootCmd.PersistentFlags().StringVarP(&flagKinds, "kind", "K", "", "comma-separated workload kinds to scan (e.g. Deployment,Job)")
	rootCmd.PersistentFlags().StringVarP(&flagSelector, "selector", "l", "", "label selector to filter workloads (e.g. app=api)")
	rootCmd.PersistentFlags().BoolVar(&flagNoHeaders, "no-headers", false, "suppress table header row")
	rootCmd.PersistentFlags().BoolVar(&flagWatch, "watch", false, "re-run scan on an interval (use --interval to set rate)")
	rootCmd.PersistentFlags().DurationVar(&flagWatchInterval, "interval", 5*time.Second, "interval between watch iterations")
	rootCmd.PersistentFlags().BoolVar(&flagIncludeSystem, "include-system", false, "include auto-generated system resources (kube-root-ca.crt, service account tokens)")
}

func resolveNamespace(contextDefault string) string {
	if flagAllNamespaces {
		return ""
	}
	if flagNamespace != "" {
		return flagNamespace
	}
	return contextDefault
}

func validateOutputFlag() error {
	switch flagOutput {
	case "table", "wide", "json", "yaml":
		return nil
	default:
		return fmt.Errorf("unknown output format %q: must be table, wide, json, or yaml", flagOutput)
	}
}
