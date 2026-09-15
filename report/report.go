// Package report is the agency white-label report seat (1 kind).
//
// The single kind renders from parity results. The Shopify mid-market
// skin is config/docs only — not a SaaS / PMS-style console and not a
// GEO dashboard.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/furyheimdall/offer-truth-monitor/parity"
)

// Seat is the package seat name used by the CLI.
const Seat = "report"

// Kind is the locked day-1 report: one agency white-label template.
const Kind = "agency-white-label"

// Pair is one cited-vs-live offer used to produce parity results.
type Pair struct {
	Cited parity.Offer
	Live  parity.Offer
}

// Input is the payload for the single report kind.
type Input struct {
	Skin     Skin
	SKUCount int
	Drifts   []parity.Drift
}

// FromCompare runs parity.Compare on each pair and builds Input.
// Callers supply fixtures or stubs; this path does not touch the network.
func FromCompare(skin Skin, pairs []Pair, epsilon float64) (Input, error) {
	if err := skin.Validate(); err != nil {
		return Input{}, err
	}
	var drifts []parity.Drift
	for i, p := range pairs {
		d, err := parity.Compare(p.Cited, p.Live, epsilon)
		if err != nil {
			return Input{}, fmt.Errorf("report: pair[%d]: %w", i, err)
		}
		drifts = append(drifts, d...)
	}
	return Input{Skin: skin, SKUCount: len(pairs), Drifts: drifts}, nil
}

// Render writes the agency white-label report from parity results.
func Render(w io.Writer, in Input) error {
	if err := in.Skin.Validate(); err != nil {
		return err
	}
	if in.SKUCount < 0 {
		return fmt.Errorf("report: sku count must be >= 0")
	}

	profile := in.Skin.Profile
	if profile == "" {
		profile = ShopifyMidProfile
	}

	var b strings.Builder
	fmt.Fprintln(&b, "Offer Truth Monitor — agency white-label report")
	fmt.Fprintf(&b, "Kind: %s\n", Kind)
	fmt.Fprintf(&b, "Skin: %s (config/docs only — not a console)\n", profile)
	if in.Skin.LogoText != "" {
		fmt.Fprintf(&b, "Label: %s\n", in.Skin.LogoText)
	}
	fmt.Fprintf(&b, "Agency: %s\n", in.Skin.Agency)
	fmt.Fprintf(&b, "Merchant: %s\n", in.Skin.Merchant)
	if in.Skin.Storefront != "" {
		fmt.Fprintf(&b, "Storefront: %s\n", in.Skin.Storefront)
	}
	if in.Skin.Accent != "" {
		fmt.Fprintf(&b, "Accent: %s\n", in.Skin.Accent)
	}
	fmt.Fprintf(&b, "SKUs watched: %d\n", in.SKUCount)
	fmt.Fprintf(&b, "Drifts: %d\n", len(in.Drifts))
	fmt.Fprintln(&b)

	if len(in.Drifts) == 0 {
		fmt.Fprintln(&b, "No drift. Cited offers match live truth within ε.")
	} else {
		for _, d := range in.Drifts {
			fmt.Fprintln(&b, formatDrift(d))
		}
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "Quoted ≠ checkout. / Visibility ≠ truth.")
	fmt.Fprintln(&b, "Not a GEO dashboard. Not a PMS-style console.")
	_, err := io.WriteString(w, b.String())
	return err
}

func formatDrift(d parity.Drift) string {
	line := fmt.Sprintf("- %s  %s  cited %s  live %s",
		d.SKU, d.Reason, formatOffer(d.Cited), formatOffer(d.Live))
	if d.Reason == parity.ReasonPrice {
		line += fmt.Sprintf("  |Δ|=%.2f", d.Delta)
	}
	return line
}

func formatOffer(o parity.Offer) string {
	return strings.TrimSpace(fmt.Sprintf("%s %s %s", o.Price, o.Currency, o.Availability))
}
