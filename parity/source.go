package parity

// CitedSource is a thin cited-offer port.
//
// engines.Adapter wires in later by mapping its CitedOffer to Offer.
// This seat does not implement engine adapters.
type CitedSource interface {
	Engine() string
	Cited(sku string) (Offer, error)
}

// LiveSource is a thin live-offer port.
//
// truth.Extractor wires in later by mapping its LiveOffer to Offer.
// This seat does not implement live extractors.
type LiveSource interface {
	Source() string
	Live(sku string) (Offer, error)
}

// CitedFunc adapts a function to CitedSource.
type CitedFunc struct {
	Name string
	Fn   func(sku string) (Offer, error)
}

// Engine implements CitedSource.
func (c CitedFunc) Engine() string { return c.Name }

// Cited implements CitedSource.
func (c CitedFunc) Cited(sku string) (Offer, error) {
	if c.Fn == nil {
		return Offer{}, errf("parity: cited %s: nil fetch (fail closed)", c.Name)
	}
	return c.Fn(sku)
}

// LiveFunc adapts a function to LiveSource.
type LiveFunc struct {
	Name string
	Fn   func(sku string) (Offer, error)
}

// Source implements LiveSource.
func (l LiveFunc) Source() string { return l.Name }

// Live implements LiveSource.
func (l LiveFunc) Live(sku string) (Offer, error) {
	if l.Fn == nil {
		return Offer{}, errf("parity: live %s: nil fetch (fail closed)", l.Name)
	}
	return l.Fn(sku)
}

// MapCited is an in-memory cited source (fixture / stub).
type MapCited struct {
	Name   string
	Offers map[string]Offer
}

// Engine implements CitedSource.
func (m MapCited) Engine() string { return m.Name }

// Cited implements CitedSource. Unknown SKUs fail closed.
func (m MapCited) Cited(sku string) (Offer, error) {
	o, ok := m.Offers[sku]
	if !ok {
		return Offer{}, errf("parity: cited %s: unknown sku %q (fail closed)", m.Name, sku)
	}
	if o.SKU == "" {
		o.SKU = sku
	}
	return o, nil
}

// MapLive is an in-memory live source (fixture / stub).
type MapLive struct {
	Name   string
	Offers map[string]Offer
}

// Source implements LiveSource.
func (m MapLive) Source() string { return m.Name }

// Live implements LiveSource. Unknown SKUs fail closed.
func (m MapLive) Live(sku string) (Offer, error) {
	o, ok := m.Offers[sku]
	if !ok {
		return Offer{}, errf("parity: live %s: unknown sku %q (fail closed)", m.Name, sku)
	}
	if o.SKU == "" {
		o.SKU = sku
	}
	return o, nil
}

// Day1Engines is the locked 3–4 cited-offer set. Names match the engines seat
// so #3 can wire adapters without renaming.
var Day1Engines = []string{
	"chatgpt-shopping",
	"perplexity",
	"gemini",
	"claude",
}
