package parity

import (
	"github.com/furyheimdall/offer-truth-monitor/engines"
	"github.com/furyheimdall/offer-truth-monitor/offer"
	"github.com/furyheimdall/offer-truth-monitor/truth"
)

// FromShared maps the #3/#4 shared offer.Offer into the compared slice.
func FromShared(o offer.Offer) Offer {
	return Offer{
		SKU:          o.SKU,
		Price:        o.Price,
		Currency:     o.Currency,
		Availability: o.Availability,
	}
}

// AdaptCited wraps an engines.Adapter (fixture-backed; no live network).
func AdaptCited(a engines.Adapter) CitedSource {
	if a == nil {
		return CitedFunc{Name: "", Fn: nil}
	}
	return CitedFunc{
		Name: string(a.Name()),
		Fn: func(sku string) (Offer, error) {
			o, err := a.CitedOffer(sku)
			if err != nil {
				return Offer{}, err
			}
			return FromShared(o.Shared()), nil
		},
	}
}

// AdaptCitedAll wraps a set of cited-offer adapters (engines.Day1FromFS).
func AdaptCitedAll(as []engines.Adapter) []CitedSource {
	out := make([]CitedSource, 0, len(as))
	for _, a := range as {
		out = append(out, AdaptCited(a))
	}
	return out
}

// AdaptLive wraps a truth.Extractor (fixture-backed; no live network).
func AdaptLive(e truth.Extractor) LiveSource {
	if e == nil {
		return LiveFunc{Name: "", Fn: nil}
	}
	return LiveFunc{
		Name: string(e.Source()),
		Fn: func(sku string) (Offer, error) {
			o, err := e.Extract(sku)
			if err != nil {
				return Offer{}, err
			}
			return FromShared(o.Shared()), nil
		},
	}
}
