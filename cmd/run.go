package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
	"github.com/tlmanz/kubectl-refx/internal/output"
	"github.com/tlmanz/kubectl-refx/internal/scanner"
)

// buildScanOpts converts CLI flags into scanner.Options.
func buildScanOpts() (scanner.Options, error) {
	var kinds []string
	if flagKinds != "" {
		for _, k := range strings.Split(flagKinds, ",") {
			norm, err := scanner.NormalizeKind(k)
			if err != nil {
				return scanner.Options{}, err
			}
			kinds = append(kinds, norm)
		}
	}
	return scanner.Options{
		Kinds:         kinds,
		LabelSelector: flagSelector,
	}, nil
}

func tableOpts() output.Options {
	return output.Options{
		NoHeaders: flagNoHeaders,
		Wide:      flagOutput == "wide",
	}
}

func printResults(results []matcher.Result) error {
	switch flagOutput {
	case "json":
		return output.WriteJSON(os.Stdout, results)
	case "yaml":
		return output.WriteYAML(os.Stdout, results)
	default: // table or wide
		output.WriteTable(os.Stdout, results, tableOpts())
		return nil
	}
}

// scanFnCtx is a scan function that accepts a fresh context per iteration.
type scanFnCtx func(ctx context.Context) ([]matcher.Result, error)

// runOnce executes scanFn with a timeout-bounded context and prints results.
func runOnce(scanFn scanFnCtx) error {
	ctx, cancel := context.WithTimeout(context.Background(), flagTimeout+5*time.Second)
	defer cancel()
	results, err := scanFn(ctx)
	if err != nil {
		return err
	}
	return printResults(results)
}

// runWatch loops scanFn on flagWatchInterval until the process receives SIGINT/SIGTERM.
func runWatch(scanFn scanFnCtx) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(flagWatchInterval)
	defer ticker.Stop()

	exec := func() {
		fmt.Print("\033[H\033[2J") // clear screen
		fmt.Printf("Every %s - %s  (Ctrl+C to stop)\n\n",
			flagWatchInterval, time.Now().Format("15:04:05"))

		scanCtx, cancel := context.WithTimeout(ctx, flagTimeout+5*time.Second)
		defer cancel()

		results, err := scanFn(scanCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		_ = printResults(results)
	}

	exec()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			exec()
		}
	}
}

// dispatch runs once or starts watch mode depending on --watch flag.
func dispatch(scanFn scanFnCtx) error {
	if flagWatch {
		return runWatch(scanFn)
	}
	return runOnce(scanFn)
}

// flatScan runs Scan for each name and returns the combined results.
func flatScan(ctx context.Context, client kubernetes.Interface, names []string, resourceType, namespace string, opts scanner.Options) ([]matcher.Result, error) {
	if len(names) == 1 {
		return scanner.Scan(ctx, client, names[0], resourceType, namespace, opts)
	}
	byName, err := scanner.ScanMultiple(ctx, client, names, resourceType, namespace, opts)
	if err != nil {
		return nil, err
	}
	var all []matcher.Result
	for _, name := range names {
		all = append(all, byName[name]...)
	}
	return all, nil
}
