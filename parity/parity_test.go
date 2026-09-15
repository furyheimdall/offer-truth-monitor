package parity

import "testing"

func TestSeat(t *testing.T) {
	if Seat != "parity" {
		t.Fatalf("Seat = %q", Seat)
	}
}

func TestComparePriceDelta(t *testing.T) {
	cited := Offer{SKU: "sku-1", Price: "19.00", Currency: "USD", Availability: "InStock"}
	live := Offer{SKU: "sku-1", Price: "18.00", Currency: "USD", Availability: "InStock"}
	d, err := Compare(cited, live, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 || d[0].Reason != ReasonPrice || d[0].Delta != 1 {
		t.Fatalf("drifts = %+v", d)
	}
}

func TestCompareWithinEpsilon(t *testing.T) {
	cited := Offer{SKU: "sku-1", Price: "19.00", Currency: "USD", Availability: "InStock"}
	live := Offer{SKU: "sku-1", Price: "19.10", Currency: "USD", Availability: "InStock"}
	d, err := Compare(cited, live, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatalf("expected no drift, got %+v", d)
	}
}

func TestCompareStockFlip(t *testing.T) {
	cited := Offer{SKU: "sku-1", Price: "10.00", Currency: "USD", Availability: "InStock"}
	live := Offer{SKU: "sku-1", Price: "10.00", Currency: "USD", Availability: "OutOfStock"}
	d, err := Compare(cited, live, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 || d[0].Reason != ReasonStock {
		t.Fatalf("drifts = %+v", d)
	}
}

func TestCompareCurrencyMismatch(t *testing.T) {
	cited := Offer{SKU: "sku-1", Price: "10.00", Currency: "USD", Availability: "InStock"}
	live := Offer{SKU: "sku-1", Price: "10.00", Currency: "EUR", Availability: "InStock"}
	d, err := Compare(cited, live, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 1 || d[0].Reason != ReasonCurrency {
		t.Fatalf("drifts = %+v", d)
	}
}
