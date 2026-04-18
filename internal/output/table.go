package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"

	"github.com/tlmanz/kubectl-refx/internal/matcher"
)

// Options controls table rendering behaviour.
type Options struct {
	NoHeaders bool
	Wide      bool // include Container column
}

func WriteTable(w io.Writer, results []matcher.Result, opts Options) {
	table := newTable(w)

	if !opts.NoHeaders {
		if opts.Wide {
			table.Header("NAMESPACE", "KIND", "NAME", "CONTAINER", "REFERENCE TYPE", "DETAIL")
		} else {
			table.Header("NAMESPACE", "KIND", "NAME", "REFERENCE TYPE", "DETAIL")
		}
	}

	for _, r := range results {
		if opts.Wide {
			_ = table.Append(r.Namespace, r.Kind, r.Name, r.Container, r.ReferenceType, r.Detail)
		} else {
			_ = table.Append(r.Namespace, r.Kind, r.Name, r.ReferenceType, r.Detail)
		}
	}
	_ = table.Render()
}

func newTable(w io.Writer) *tablewriter.Table {
	return tablewriter.NewTable(w,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
			Row:    tw.CellConfig{Alignment: tw.CellAlignment{Global: tw.AlignLeft}},
		}),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{Left: tw.Off, Right: tw.Off, Top: tw.Off, Bottom: tw.Off},
			Settings: tw.Settings{
				Separators: tw.Separators{BetweenRows: tw.Off, BetweenColumns: tw.On},
				Lines:      tw.Lines{ShowHeaderLine: tw.Off, ShowTop: tw.Off, ShowBottom: tw.Off},
			},
		}),
	)
}
