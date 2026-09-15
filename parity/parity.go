// Package parity compares cited vs live offers.
// Alert when |Δprice| > ε or availability stock-flips.
package parity

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Seat is the package seat name used by the CLI.
const Seat = "parity"

// Offer is the compared slice of a cited or live offer.
type Offer struct {
	SKU          string
	Price        string
	Currency     string
	Availability string
}

// Reason is why a pair drifted.
type Reason string

const (
	ReasonPrice    Reason = "price-delta"
	ReasonStock    Reason = "stock-flip"
	ReasonCurrency Reason = "currency-mismatch"
)

// Drift is one cited-vs-live mismatch.
type Drift struct {
	SKU        string
	Engine     string
	LiveSource string
	Reason     Reason
	Cited      Offer
	Live       Offer
	Delta      float64
}

// Summary is a stable one-line description for notifier plugs.
func (d Drift) Summary() string {
	switch d.Reason {
	case ReasonPrice:
		return fmt.Sprintf("|Δprice|=%.4g > ε (cited %s %s vs live %s %s)", d.Delta, d.Cited.Price, d.Cited.Currency, d.Live.Price, d.Live.Currency)
	case ReasonStock:
		return fmt.Sprintf("stock-flip cited=%s live=%s", d.Cited.Availability, d.Live.Availability)
	case ReasonCurrency:
		return fmt.Sprintf("currency-mismatch cited=%s live=%s", d.Cited.Currency, d.Live.Currency)
	default:
		return string(d.Reason)
	}
}

// Compare returns drifts for |Δprice| > epsilon or a stock-flip.
// Currency mismatch is treated as fail-closed drift (no silent FX).
func Compare(cited, live Offer, epsilon float64) ([]Drift, error) {
	if cited.SKU == "" || live.SKU == "" {
		return nil, fmt.Errorf("parity: sku required on both offers")
	}
	if cited.SKU != live.SKU {
		return nil, fmt.Errorf("parity: sku mismatch %q vs %q", cited.SKU, live.SKU)
	}
	if cited.Currency == "" || live.Currency == "" {
		return nil, fmt.Errorf("parity: currency required")
	}

	var out []Drift
	if !strings.EqualFold(cited.Currency, live.Currency) {
		out = append(out, Drift{SKU: cited.SKU, Reason: ReasonCurrency, Cited: cited, Live: live})
		return out, nil
	}

	cp, err := parsePrice(cited.Price)
	if err != nil {
		return nil, fmt.Errorf("parity: cited price: %w", err)
	}
	lp, err := parsePrice(live.Price)
	if err != nil {
		return nil, fmt.Errorf("parity: live price: %w", err)
	}
	delta := math.Abs(cp - lp)
	if delta > epsilon {
		out = append(out, Drift{SKU: cited.SKU, Reason: ReasonPrice, Cited: cited, Live: live, Delta: delta})
	}
	if stockFlip(cited.Availability, live.Availability) {
		out = append(out, Drift{SKU: cited.SKU, Reason: ReasonStock, Cited: cited, Live: live})
	}
	return out, nil
}

func parsePrice(s string) (float64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.ParseFloat(s, 64)
}

func stockFlip(cited, live string) bool {
	return inStock(cited) != inStock(live)
}

func inStock(a string) bool {
	switch strings.ToLower(strings.TrimSpace(a)) {
	case "instock", "in_stock", "in stock", "available":
		return true
	default:
		return false
	}
}
