package output

import (
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

func WriteYAML(w io.Writer, results []matcher.Result) error {
	data, err := yaml.Marshal(results)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}
	return nil
}
