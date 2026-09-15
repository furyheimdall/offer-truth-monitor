package truth

import "testing"

func TestSeatAndDay1(t *testing.T) {
	if Seat != "truth" {
		t.Fatalf("Seat = %q", Seat)
	}
	want := map[Source]bool{PDPHTML: true, JSONLDOffer: true, GMCFeed: true, ShopifyProducts: true}
	if len(Day1) != 4 {
		t.Fatalf("Day1 = %v", Day1)
	}
	for _, s := range Day1 {
		if !want[s] {
			t.Fatalf("unexpected source %q", s)
		}
	}
}

func TestFixtureExtract(t *testing.T) {
	f := Fixture{
		Src: JSONLDOffer,
		Offers: map[string]LiveOffer{
			"sku-1": {
				SKU:          "sku-1",
				Source:       JSONLDOffer,
				Price:        "18.00",
				Currency:     "USD",
				Availability: "InStock",
			},
		},
	}
	o, err := f.Extract("sku-1")
	if err != nil {
		t.Fatal(err)
	}
	if o.Price != "18.00" {
		t.Fatalf("price = %q", o.Price)
	}
}

func TestFixtureUnknownFailClosed(t *testing.T) {
	f := Fixture{Src: GMCFeed, Offers: map[string]LiveOffer{}}
	if _, err := f.Extract("missing"); err == nil {
		t.Fatal("expected fail-closed")
	}
}
