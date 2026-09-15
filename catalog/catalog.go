// Package catalog holds the day-1 20–50 SKU watch list config seat.
package catalog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Seat is the package seat name used by the CLI.
const Seat = "catalog"

// MaxSKUs is the day-1 watch-list ceiling (top 20–50 SKU).
const MaxSKUs = 50

// SKU is one watched product identifier plus optional live PDP URL.
type SKU struct {
	ID     string `json:"id"`
	PDPURL string `json:"pdp_url,omitempty"`
}

// Watchlist is the configured SKU set. Day-1 target is 20–50 SKUs; the
// loader rejects more than MaxSKUs. An empty list is valid for tests.
type Watchlist struct {
	SKUs []SKU `json:"skus"`
}

// LoadJSON reads a watch list from r.
func LoadJSON(r io.Reader) (*Watchlist, error) {
	var wl Watchlist
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&wl); err != nil {
		return nil, fmt.Errorf("catalog: decode watch list: %w", err)
	}
	if err := wl.Validate(); err != nil {
		return nil, err
	}
	return &wl, nil
}

// LoadFile reads a watch list JSON file.
func LoadFile(path string) (*Watchlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("catalog: open %s: %w", path, err)
	}
	defer f.Close()
	return LoadJSON(f)
}

// Validate enforces the day-1 SKU ceiling and non-empty IDs.
func (w *Watchlist) Validate() error {
	if w == nil {
		return fmt.Errorf("catalog: nil watch list")
	}
	if len(w.SKUs) > MaxSKUs {
		return fmt.Errorf("catalog: %d SKUs exceeds day-1 max %d", len(w.SKUs), MaxSKUs)
	}
	seen := make(map[string]struct{}, len(w.SKUs))
	for i, s := range w.SKUs {
		if s.ID == "" {
			return fmt.Errorf("catalog: sku[%d] missing id", i)
		}
		if _, ok := seen[s.ID]; ok {
			return fmt.Errorf("catalog: duplicate sku id %q", s.ID)
		}
		seen[s.ID] = struct{}{}
	}
	return nil
}
