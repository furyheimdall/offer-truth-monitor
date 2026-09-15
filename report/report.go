// Package report is the agency white-label report seat (1 kind).
// Thin Shopify mid-market config/docs skin belongs here later — not a SaaS console.
package report

import (
	"fmt"
	"io"
	"strings"
)

// Seat is the package seat name used by the CLI.
const Seat = "report"

// Kind is the locked day-1 report: one agency white-label template.
const Kind = "agency-white-label"

// Input is the stub payload for the single report kind.
type Input struct {
	Agency   string
	Merchant string
	SKUCount int
	Drifts   int
}

// Render writes a text stub of the agency white-label report.
func Render(w io.Writer, in Input) error {
	if in.Agency == "" {
		return fmt.Errorf("report: agency required")
	}
	if in.Merchant == "" {
		return fmt.Errorf("report: merchant required")
	}
	_, err := fmt.Fprintf(w, strings.TrimSpace(`
Offer Truth Monitor — agency white-label report
Kind: %s
Agency: %s
Merchant: %s
SKUs watched: %d
Drifts: %d
Quoted ≠ checkout. / Visibility ≠ truth.
`)+"\n", Kind, in.Agency, in.Merchant, in.SKUCount, in.Drifts)
	return err
}
