package engines

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

// chatgptPayload is a dig-sourced shopping-card-like fixture.
// Not a shopping-card API guarantee; engine behavior may change.
type chatgptPayload struct {
	MerchantSKU  string `json:"merchant_sku"`
	ShoppingCard *struct {
		Price *struct {
			Value    json.Number `json:"value"`
			Currency string      `json:"currency"`
		} `json:"price"`
		Availability string `json:"availability"`
		IsSale       *bool  `json:"is_sale"`
	} `json:"shopping_card"`
}

// ParseChatGPTShopping maps a ChatGPT Shopping fixture to offer.Offer.
func ParseChatGPTShopping(sku string, raw []byte) (offer.Offer, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var p chatgptPayload
	if err := dec.Decode(&p); err != nil {
		return offer.Offer{}, fmt.Errorf("offer: unparseable chatgpt shopping fixture")
	}
	if p.ShoppingCard == nil || p.ShoppingCard.Price == nil {
		return offer.Offer{}, fmt.Errorf("offer: missing required field price")
	}
	bound, err := offer.BindSKU(sku, p.MerchantSKU)
	if err != nil {
		return offer.Offer{}, err
	}
	return offer.Normalize(offer.Offer{
		SKU:          bound,
		Price:        p.ShoppingCard.Price.Value.String(),
		Currency:     p.ShoppingCard.Price.Currency,
		Availability: p.ShoppingCard.Availability,
		Sale:         p.ShoppingCard.IsSale,
	})
}
