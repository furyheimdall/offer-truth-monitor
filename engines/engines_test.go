package engines

import "testing"

func TestSeatAndDay1(t *testing.T) {
	if Seat != "engines" {
		t.Fatalf("Seat = %q", Seat)
	}
	if len(Day1) < 3 || len(Day1) > 4 {
		t.Fatalf("Day1 engine count = %d, want 3–4", len(Day1))
	}
}

func TestFixtureCitedOffer(t *testing.T) {
	f := Fixture{
		Engine: Perplexity,
		Offers: map[string]CitedOffer{
			"sku-1": {
				SKU:          "sku-1",
				Engine:       Perplexity,
				Price:        "19.00",
				Currency:     "USD",
				Availability: "InStock",
			},
		},
	}
	o, err := f.CitedOffer("sku-1")
	if err != nil {
		t.Fatal(err)
	}
	if o.Price != "19.00" || o.Currency != "USD" {
		t.Fatalf("unexpected offer: %+v", o)
	}
}

func TestFixtureUnknownSKUFailClosed(t *testing.T) {
	f := Fixture{Engine: Claude, Offers: map[string]CitedOffer{}}
	if _, err := f.CitedOffer("missing"); err == nil {
		t.Fatal("expected fail-closed on unknown sku")
	}
}

func TestValidateRequiredFailClosed(t *testing.T) {
	err := ValidateRequired(CitedOffer{SKU: "x", Engine: Gemini, Price: "1"})
	if err == nil {
		t.Fatal("expected missing currency/availability to fail closed")
	}
}
