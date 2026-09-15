package engines

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

type geminiPayload struct {
	ProductSKU string `json:"productSku"`
	Offer      *struct {
		Amount *struct {
			Value        json.Number `json:"value"`
			CurrencyCode string      `json:"currencyCode"`
		} `json:"amount"`
		Availability string `json:"availability"`
		OnSale       *bool  `json:"onSale"`
	} `json:"offer"`
}

// ParseGemini maps a Gemini cited-offer fixture to offer.Offer.
func ParseGemini(sku string, raw []byte) (offer.Offer, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var p geminiPayload
	if err := dec.Decode(&p); err != nil {
		return offer.Offer{}, fmt.Errorf("offer: unparseable gemini fixture")
	}
	if p.Offer == nil || p.Offer.Amount == nil {
		return offer.Offer{}, fmt.Errorf("offer: missing required field price")
	}
	bound, err := offer.BindSKU(sku, p.ProductSKU)
	if err != nil {
		return offer.Offer{}, err
	}
	return offer.Normalize(offer.Offer{
		SKU:          bound,
		Price:        p.Offer.Amount.Value.String(),
		Currency:     p.Offer.Amount.CurrencyCode,
		Availability: p.Offer.Availability,
		Sale:         p.Offer.OnSale,
	})
}
