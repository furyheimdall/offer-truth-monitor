// Package engines is the cited-offer adapter seat.
//
// Day-1 engines (3–4): ChatGPT Shopping / Perplexity / Gemini / Claude.
// Required cited fields: price · currency · availability. sale is optional.
// Stubs may return fixtures. Unknown or missing required fields fail closed.
// This is not a shopping-card API guarantee; dig-sourced engine behavior may change.
package engines

import "fmt"

// Seat is the package seat name used by the CLI.
const Seat = "engines"

// Engine is a cited-offer source name.
type Engine string

const (
	ChatGPTShopping Engine = "chatgpt-shopping"
	Perplexity      Engine = "perplexity"
	Gemini          Engine = "gemini"
	Claude          Engine = "claude"
)

// Day1 is the locked 3–4 engine set.
var Day1 = []Engine{ChatGPTShopping, Perplexity, Gemini, Claude}

// CitedOffer is the required cited shape. Sale is optional.
type CitedOffer struct {
	SKU          string
	Engine       Engine
	Price        string
	Currency     string
	Availability string
	Sale         *bool
}

// Adapter fetches a cited offer for one SKU. Implementations in this
// scaffold are fixture-backed and must not perform live network I/O.
type Adapter interface {
	Name() Engine
	CitedOffer(sku string) (CitedOffer, error)
}

// ValidateRequired fails closed when a required cited field is missing.
func ValidateRequired(o CitedOffer) error {
	if o.SKU == "" {
		return fmt.Errorf("engines: missing required field sku")
	}
	if o.Engine == "" {
		return fmt.Errorf("engines: missing required field engine")
	}
	if o.Price == "" {
		return fmt.Errorf("engines: missing required field price")
	}
	if o.Currency == "" {
		return fmt.Errorf("engines: missing required field currency")
	}
	if o.Availability == "" {
		return fmt.Errorf("engines: missing required field availability")
	}
	return nil
}

// Fixture is an in-memory adapter used until live adapters land.
type Fixture struct {
	Engine Engine
	Offers map[string]CitedOffer
}

// Name implements Adapter.
func (f Fixture) Name() Engine { return f.Engine }

// CitedOffer implements Adapter. Missing SKUs and incomplete rows fail closed.
func (f Fixture) CitedOffer(sku string) (CitedOffer, error) {
	o, ok := f.Offers[sku]
	if !ok {
		return CitedOffer{}, fmt.Errorf("engines: %s: unknown sku %q (fail closed)", f.Engine, sku)
	}
	if err := ValidateRequired(o); err != nil {
		return CitedOffer{}, err
	}
	return o, nil
}
