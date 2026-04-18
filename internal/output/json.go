package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

func WriteJSON(w io.Writer, results []matcher.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
