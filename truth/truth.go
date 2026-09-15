// Package truth is the live-offer extractor seat.
//
// Sources: PDP HTML · JSON-LD Offer · GMC/feed · Shopify products.json
// (Admin read optional/stub). Extractors normalize to offer.Offer.
// Default extractors are fixture-backed; no live network I/O.
package truth

import (
	"fmt"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

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

// Day1 sources that must normalize to the shared live offer shape.
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

// Shared returns the common offer shape used by engines and truth.
func (o LiveOffer) Shared() offer.Offer {
	return offer.Offer{
		SKU:          o.SKU,
		Price:        o.Price,
		Currency:     o.Currency,
		Availability: o.Availability,
		Sale:         o.Sale,
	}
}

// FromShared wraps a normalized offer.Offer with source provenance.
func FromShared(src Source, o offer.Offer) LiveOffer {
	return LiveOffer{
		SKU:          o.SKU,
		Source:       src,
		Price:        o.Price,
		Currency:     o.Currency,
		Availability: o.Availability,
		Sale:         o.Sale,
	}
}

// Extractor returns a live offer for one SKU from a single source.
type Extractor interface {
	Source() Source
	Extract(sku string) (LiveOffer, error)
}

// ValidateRequired fails closed when a required live field is missing.
func ValidateRequired(o LiveOffer) error {
	if o.Source == "" {
		return fmt.Errorf("truth: missing required field source")
	}
	if err := offer.Validate(o.Shared()); err != nil {
		return fmt.Errorf("truth: %w", err)
	}
	return nil
}

// Fixture is an in-memory extractor used for simple table tests.
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
