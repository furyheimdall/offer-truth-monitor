package engines

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

type perplexityPayload struct {
	SKU       string `json:"sku"`
	Citations []struct {
		Kind         string      `json:"kind"`
		Price        json.Number `json:"price"`
		Currency     string      `json:"currency"`
		Availability string      `json:"availability"`
		Sale         *bool       `json:"sale"`
	} `json:"citations"`
}

// ParsePerplexity maps a Perplexity cited-offer fixture to offer.Offer.
func ParsePerplexity(sku string, raw []byte) (offer.Offer, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var p perplexityPayload
	if err := dec.Decode(&p); err != nil {
		return offer.Offer{}, fmt.Errorf("offer: unparseable perplexity fixture")
	}
	bound, err := offer.BindSKU(sku, p.SKU)
	if err != nil {
		return offer.Offer{}, err
	}
	for _, c := range p.Citations {
		if c.Kind != "offer" {
			continue
		}
		return offer.Normalize(offer.Offer{
			SKU:          bound,
			Price:        c.Price.String(),
			Currency:     c.Currency,
			Availability: c.Availability,
			Sale:         c.Sale,
		})
	}
	return offer.Offer{}, fmt.Errorf("offer: missing required field price")
}
