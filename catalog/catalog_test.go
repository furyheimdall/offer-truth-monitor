package catalog

import (
	"strings"
	"testing"
)

func TestSeat(t *testing.T) {
	if Seat != "catalog" {
		t.Fatalf("Seat = %q", Seat)
	}
}

func TestLoadJSON(t *testing.T) {
	wl, err := LoadJSON(strings.NewReader(`{"skus":[{"id":"sku-1","pdp_url":"https://example.com/p/1"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(wl.SKUs) != 1 || wl.SKUs[0].ID != "sku-1" {
		t.Fatalf("unexpected watch list: %+v", wl)
	}
}

func TestValidateMaxSKUs(t *testing.T) {
	wl := &Watchlist{SKUs: make([]SKU, MaxSKUs+1)}
	for i := range wl.SKUs {
		wl.SKUs[i].ID = string(rune('a'+(i%26))) + string(rune('A'+(i/26)))
	}
	if err := wl.Validate(); err == nil {
		t.Fatal("expected max-SKU error")
	}
}

func TestValidateDuplicate(t *testing.T) {
	wl := &Watchlist{SKUs: []SKU{{ID: "a"}, {ID: "a"}}}
	if err := wl.Validate(); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestLoadUnknownFieldFailClosed(t *testing.T) {
	_, err := LoadJSON(strings.NewReader(`{"skus":[],"geo":true}`))
	if err == nil {
		t.Fatal("expected unknown field to fail closed")
	}
}
