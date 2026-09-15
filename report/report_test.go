package report

import (
	"strings"
	"testing"

	"github.com/furyheimdall/offer-truth-monitor/parity"
)

func TestSeat(t *testing.T) {
	if Seat != "report" {
		t.Fatalf("Seat = %q", Seat)
	}
	if Kind != "agency-white-label" {
		t.Fatalf("Kind = %q", Kind)
	}
}

func TestRenderFromParityResults(t *testing.T) {
	in, err := FromCompare(DefaultShopifyMidSkin(), FixturePairs(), 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if in.SKUCount != 3 {
		t.Fatalf("SKUCount = %d", in.SKUCount)
	}
	if len(in.Drifts) != 2 {
		t.Fatalf("drifts = %+v", in.Drifts)
	}

	var b strings.Builder
	if err := Render(&b, in); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		"agency-white-label",
		"shopify-mid",
		"config/docs only",
		"North Agency",
		"MidShop",
		"midshop.myshopify.com",
		"sku-hoodie",
		"price-delta",
		"|Δ|=7.00",
		"sku-mug",
		"stock-flip",
		"Quoted ≠ checkout",
		"Not a GEO dashboard",
		"Not a PMS-style console",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("report missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "sku-tee") {
		t.Fatalf("in-parity SKU should not appear as a drift:\n%s", out)
	}
}

func TestRenderNoDrift(t *testing.T) {
	pairs := []Pair{{
		Cited: parity.Offer{SKU: "sku-1", Price: "10.00", Currency: "USD", Availability: "InStock"},
		Live:  parity.Offer{SKU: "sku-1", Price: "10.10", Currency: "USD", Availability: "InStock"},
	}}
	in, err := FromCompare(DefaultShopifyMidSkin(), pairs, 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Drifts) != 0 {
		t.Fatalf("drifts = %+v", in.Drifts)
	}
	var b strings.Builder
	if err := Render(&b, in); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "No drift") {
		t.Fatalf("expected clean report:\n%s", b.String())
	}
}

func TestRenderRequiresAgency(t *testing.T) {
	err := Render(&strings.Builder{}, Input{Skin: Skin{Merchant: "x"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFromComparePropagatesParityError(t *testing.T) {
	_, err := FromCompare(DefaultShopifyMidSkin(), []Pair{{
		Cited: parity.Offer{SKU: "a", Price: "1", Currency: "USD", Availability: "InStock"},
		Live:  parity.Offer{SKU: "b", Price: "1", Currency: "USD", Availability: "InStock"},
	}}, 0)
	if err == nil {
		t.Fatal("expected sku mismatch")
	}
}

func TestRenderRejectsNegativeSKUCount(t *testing.T) {
	err := Render(&strings.Builder{}, Input{Skin: DefaultShopifyMidSkin(), SKUCount: -1})
	if err == nil {
		t.Fatal("expected error")
	}
}
