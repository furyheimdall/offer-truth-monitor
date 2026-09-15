package truth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

// ShopifyCatalog is a parsed Shopify products.json (or Admin products) fixture.
type ShopifyCatalog struct {
	src    Source
	offers map[string]offer.Offer
}

type shopifyEnvelope struct {
	Currency string           `json:"currency"`
	Products []shopifyProduct `json:"products"`
}

type shopifyProduct struct {
	Variants []shopifyVariant `json:"variants"`
}

type shopifyVariant struct {
	SKU               string      `json:"sku"`
	Price             json.Number `json:"price"`
	Available         *bool       `json:"available"`
	CompareAtPrice    *string     `json:"compare_at_price"`
	InventoryQuantity *int        `json:"inventory_quantity"`
	PresentmentPrices []struct {
		Price struct {
			Amount       json.Number `json:"amount"`
			CurrencyCode string      `json:"currency_code"`
		} `json:"price"`
	} `json:"presentment_prices"`
}

// LoadShopifyProducts parses a storefront products.json fixture.
// Currency must be present at the envelope or on presentment_prices (fail closed).
func LoadShopifyProducts(raw []byte) (*ShopifyCatalog, error) {
	return loadShopify(ShopifyProducts, raw)
}

// LoadShopifyAdmin parses an optional Admin products fixture.
// An empty body fails closed — Admin read is optional/stub.
func LoadShopifyAdmin(raw []byte) (*ShopifyCatalog, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, fmt.Errorf("truth: %s: admin read is optional/stub (fail closed)", ShopifyAdminRead)
	}
	return loadShopify(ShopifyAdminRead, raw)
}

// ShopifyAdminStub is the optional Admin extractor with no payload.
type ShopifyAdminStub struct{}

// Source implements Extractor.
func (ShopifyAdminStub) Source() Source { return ShopifyAdminRead }

// Extract always fails closed unless a real Admin fixture is loaded via LoadShopifyAdmin.
func (ShopifyAdminStub) Extract(sku string) (LiveOffer, error) {
	return LiveOffer{}, fmt.Errorf("truth: %s: admin read is optional/stub (fail closed)", ShopifyAdminRead)
}

// Source implements Extractor.
func (s *ShopifyCatalog) Source() Source { return s.src }

// Extract implements Extractor.
func (s *ShopifyCatalog) Extract(sku string) (LiveOffer, error) {
	if s == nil || s.offers == nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", ShopifyProducts, sku)
	}
	o, ok := s.offers[sku]
	if !ok {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", s.src, sku)
	}
	live := FromShared(s.src, o)
	if err := ValidateRequired(live); err != nil {
		return LiveOffer{}, err
	}
	return live, nil
}

func loadShopify(src Source, raw []byte) (*ShopifyCatalog, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var env shopifyEnvelope
	if err := dec.Decode(&env); err != nil {
		return nil, fmt.Errorf("truth: %s: unparseable products.json (fail closed)", src)
	}
	cat := &ShopifyCatalog{src: src, offers: map[string]offer.Offer{}}
	for _, p := range env.Products {
		for _, v := range p.Variants {
			sku := strings.TrimSpace(v.SKU)
			if sku == "" {
				continue
			}
			price := v.Price.String()
			currency := env.Currency
			if len(v.PresentmentPrices) > 0 {
				if price == "" {
					price = v.PresentmentPrices[0].Price.Amount.String()
				}
				if currency == "" {
					currency = v.PresentmentPrices[0].Price.CurrencyCode
				}
			}
			avail := shopifyAvailability(v)
			var sale *bool
			if v.CompareAtPrice != nil && strings.TrimSpace(*v.CompareAtPrice) != "" && strings.TrimSpace(*v.CompareAtPrice) != "0.00" {
				sale = offer.Bool(true)
			}
			o, err := offer.Normalize(offer.Offer{
				SKU:          sku,
				Price:        price,
				Currency:     currency,
				Availability: avail,
				Sale:         sale,
			})
			if err != nil {
				continue
			}
			cat.offers[sku] = o
		}
	}
	if len(cat.offers) == 0 {
		return nil, fmt.Errorf("truth: %s: no valid offers (fail closed)", src)
	}
	return cat, nil
}

func shopifyAvailability(v shopifyVariant) string {
	if v.Available != nil {
		if *v.Available {
			return "InStock"
		}
		return "OutOfStock"
	}
	if v.InventoryQuantity != nil {
		if *v.InventoryQuantity > 0 {
			return "InStock"
		}
		return "OutOfStock"
	}
	return ""
}
