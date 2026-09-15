package offer

import "testing"

func TestNormalizeHappy(t *testing.T) {
	sale := Bool(true)
	got, err := Normalize(Offer{
		SKU:          "sku-1",
		Price:        "19.00",
		Currency:     "usd",
		Availability: "https://schema.org/InStock",
		Sale:         sale,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Price != "19.00" || got.Currency != "USD" || got.Availability != "InStock" {
		t.Fatalf("got %+v", got)
	}
	if got.Sale == nil || !*got.Sale {
		t.Fatal("expected sale=true")
	}
}

func TestValidateFailClosed(t *testing.T) {
	base := Offer{SKU: "sku-1", Price: "1.00", Currency: "USD", Availability: "InStock"}
	cases := []Offer{
		{Price: base.Price, Currency: base.Currency, Availability: base.Availability},
		{SKU: base.SKU, Currency: base.Currency, Availability: base.Availability},
		{SKU: base.SKU, Price: base.Price, Availability: base.Availability},
		{SKU: base.SKU, Price: base.Price, Currency: base.Currency},
		{SKU: base.SKU, Price: "not-a-price", Currency: base.Currency, Availability: base.Availability},
		{SKU: base.SKU, Price: base.Price, Currency: "US", Availability: base.Availability},
	}
	for i, o := range cases {
		if err := Validate(o); err == nil {
			t.Fatalf("case %d: expected fail-closed for %+v", i, o)
		}
	}
}

func TestSaleOptional(t *testing.T) {
	o := Offer{SKU: "x", Price: "1", Currency: "USD", Availability: "available"}
	if err := Validate(o); err != nil {
		t.Fatal(err)
	}
	if o.Sale != nil {
		t.Fatal("sale should remain unset")
	}
}

func TestBindSKU(t *testing.T) {
	got, err := BindSKU("sku-1", "")
	if err != nil || got != "sku-1" {
		t.Fatalf("inherit: %q %v", got, err)
	}
	if _, err := BindSKU("sku-1", "other"); err == nil {
		t.Fatal("expected mismatch to fail closed")
	}
	if _, err := BindSKU("", "sku-1"); err == nil {
		t.Fatal("expected empty requested sku to fail closed")
	}
}

func TestSplitAmountCurrency(t *testing.T) {
	a, c, err := SplitAmountCurrency("18.00 USD")
	if err != nil || a != "18.00" || c != "USD" {
		t.Fatalf("%s %s %v", a, c, err)
	}
	a, c, err = SplitAmountCurrency("EUR 12.50")
	if err != nil || a != "12.50" || c != "EUR" {
		t.Fatalf("%s %s %v", a, c, err)
	}
	if _, _, err := SplitAmountCurrency("18.00"); err == nil {
		t.Fatal("expected missing currency to fail closed")
	}
}

func TestAvailabilityAliases(t *testing.T) {
	for in, want := range map[string]string{
		"in_stock":     "InStock",
		"in stock":     "InStock",
		"available":    "InStock",
		"out of stock": "OutOfStock",
		"sold_out":     "OutOfStock",
		"Pre-Order":    "PreOrder",
	} {
		got, err := NormalizeAvailability(in)
		if err != nil || got != want {
			t.Fatalf("%q → %q %v, want %q", in, got, err, want)
		}
	}
}
