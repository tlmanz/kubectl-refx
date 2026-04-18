package main

import (
	"os"

	"github.com/tlmanz/kubectl-refx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
