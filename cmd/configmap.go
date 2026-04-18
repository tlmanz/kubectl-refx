package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tlmanz/kubectl-refx/internal/k8s"
	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

var configmapCmd = &cobra.Command{
	Use:     "configmap <name> [name...]",
	Aliases: []string{"cm", "configmaps"},
	Short:   "Find workloads referencing one or more ConfigMaps",
	Args:    cobra.MinimumNArgs(1),
	RunE:    runConfigmap,
}

func init() {
	rootCmd.AddCommand(configmapCmd)
}

func runConfigmap(_ *cobra.Command, args []string) error {
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
		return flatScan(ctx, client, args, matcher.TypeConfigMap, ns, opts)
	})
}
