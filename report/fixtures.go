package report

import "github.com/furyheimdall/offer-truth-monitor/parity"

// FixturePairs is a stub cited-vs-live batch for the Shopify mid skin.
// One price-delta, one stock-flip, one in-parity SKU. No network I/O.
func FixturePairs() []Pair {
	return []Pair{
		{
			Cited: parity.Offer{SKU: "sku-hoodie", Price: "49.00", Currency: "USD", Availability: "InStock"},
			Live:  parity.Offer{SKU: "sku-hoodie", Price: "42.00", Currency: "USD", Availability: "InStock"},
		},
		{
			Cited: parity.Offer{SKU: "sku-mug", Price: "12.00", Currency: "USD", Availability: "InStock"},
			Live:  parity.Offer{SKU: "sku-mug", Price: "12.00", Currency: "USD", Availability: "OutOfStock"},
		},
		{
			Cited: parity.Offer{SKU: "sku-tee", Price: "24.00", Currency: "USD", Availability: "InStock"},
			Live:  parity.Offer{SKU: "sku-tee", Price: "24.00", Currency: "USD", Availability: "InStock"},
		},
	}
}
