// Package truth is the live-offer extractor seat.
//
// Sources: PDP HTML · JSON-LD Offer · GMC/feed · Shopify products.json
// (Admin read optional/stub). Extractors normalize to the same offer shape.
// Scaffold extractors are fixture-backed; no live network I/O.
package truth

import "fmt"

// Seat is the package seat name used by the CLI.
const Seat = "truth"

// Source is a live truth origin.
type Source string

const (
	PDPHTML          Source = "pdp-html"
	JSONLDOffer      Source = "json-ld-offer"
	GMCFeed          Source = "gmc-feed"
	ShopifyProducts  Source = "shopify-products-json"
	ShopifyAdminRead Source = "shopify-admin-read"
)

// Day1 sources that must normalize to LiveOffer.
var Day1 = []Source{PDPHTML, JSONLDOffer, GMCFeed, ShopifyProducts}

// LiveOffer is the normalized live shape (same required fields as cited).
type LiveOffer struct {
	SKU          string
	Source       Source
	Price        string
	Currency     string
	Availability string
	Sale         *bool
}

// Extractor returns a live offer for one SKU from a single source.
type Extractor interface {
	Source() Source
	Extract(sku string) (LiveOffer, error)
}

// ValidateRequired fails closed when a required live field is missing.
func ValidateRequired(o LiveOffer) error {
	if o.SKU == "" {
		return fmt.Errorf("truth: missing required field sku")
	}
	if o.Source == "" {
		return fmt.Errorf("truth: missing required field source")
	}
	if o.Price == "" {
		return fmt.Errorf("truth: missing required field price")
	}
	if o.Currency == "" {
		return fmt.Errorf("truth: missing required field currency")
	}
	if o.Availability == "" {
		return fmt.Errorf("truth: missing required field availability")
	}
	return nil
}

// Fixture is an in-memory extractor used until live extractors land.
type Fixture struct {
	Src    Source
	Offers map[string]LiveOffer
}

// Source implements Extractor.
func (f Fixture) Source() Source { return f.Src }

// Extract implements Extractor. Missing SKUs fail closed.
func (f Fixture) Extract(sku string) (LiveOffer, error) {
	o, ok := f.Offers[sku]
	if !ok {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", f.Src, sku)
	}
	if err := ValidateRequired(o); err != nil {
		return LiveOffer{}, err
	}
	return o, nil
}
