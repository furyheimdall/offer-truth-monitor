package report

import (
	"strings"
	"testing"
)

func TestSeat(t *testing.T) {
	if Seat != "report" {
		t.Fatalf("Seat = %q", Seat)
	}
	if Kind != "agency-white-label" {
		t.Fatalf("Kind = %q", Kind)
	}
}

func TestRender(t *testing.T) {
	var b strings.Builder
	err := Render(&b, Input{Agency: "North", Merchant: "MidShop", SKUCount: 20, Drifts: 2})
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"agency-white-label", "North", "MidShop", "Quoted ≠ checkout"} {
		if !strings.Contains(out, want) {
			t.Fatalf("report missing %q:\n%s", want, out)
		}
	}
}

func TestRenderRequiresAgency(t *testing.T) {
	if err := Render(&strings.Builder{}, Input{Merchant: "x"}); err == nil {
		t.Fatal("expected error")
	}
}
