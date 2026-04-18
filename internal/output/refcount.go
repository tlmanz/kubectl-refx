package output

import (
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

// RefCount holds a resource name and how many workloads reference it.
type RefCount struct {
	Namespace  string `json:"namespace" yaml:"namespace"`
	Name       string `json:"name"      yaml:"name"`
	References int    `json:"references" yaml:"references"`
}

// WriteRefCountTable renders a namespace/name/references table.
func WriteRefCountTable(w io.Writer, counts []RefCount, opts Options) {
	table := newTable(w)
	if !opts.NoHeaders {
		table.Header("NAMESPACE", "NAME", "REFERENCES")
	}
	for _, c := range counts {
		_ = table.Append(c.Namespace, c.Name, fmt.Sprintf("%d", c.References))
	}
	_ = table.Render()
}

// WriteUnusedTable renders a namespace/name table (no references column).
func WriteUnusedTable(w io.Writer, counts []RefCount, opts Options) {
	table := newTable(w)
	if !opts.NoHeaders {
		table.Header("NAMESPACE", "NAME")
	}
	for _, c := range counts {
		_ = table.Append(c.Namespace, c.Name)
	}
	_ = table.Render()
}

func WriteRefCountJSON(w io.Writer, counts []RefCount) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(counts); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

func WriteRefCountYAML(w io.Writer, counts []RefCount) error {
	data, err := yaml.Marshal(counts)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}
	return nil
}
