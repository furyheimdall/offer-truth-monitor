package engines

import (
	"encoding/json"
	"fmt"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

type claudePayload struct {
	SKU          string `json:"sku"`
	Quoted       string `json:"quoted"`
	Availability string `json:"availability"`
	Sale         *bool  `json:"sale"`
}

// ParseClaude maps a Claude cited-offer fixture to offer.Offer.
// Quoted is a combined "19.00 USD" string (dig-sourced; not an API guarantee).
func ParseClaude(sku string, raw []byte) (offer.Offer, error) {
	var p claudePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return offer.Offer{}, fmt.Errorf("offer: unparseable claude fixture")
	}
	bound, err := offer.BindSKU(sku, p.SKU)
	if err != nil {
		return offer.Offer{}, err
	}
	amount, currency, err := offer.SplitAmountCurrency(p.Quoted)
	if err != nil {
		return offer.Offer{}, err
	}
	return offer.Normalize(offer.Offer{
		SKU:          bound,
		Price:        amount,
		Currency:     currency,
		Availability: p.Availability,
		Sale:         p.Sale,
	})
}
