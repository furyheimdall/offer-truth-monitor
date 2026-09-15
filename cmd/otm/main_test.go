package main

import (
	"strings"
	"testing"
)

func TestSeats(t *testing.T) {
	var b strings.Builder
	if code := run(nil, &b); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := b.String()
	for _, seat := range []string{"catalog", "engines", "truth", "parity", "alert", "report"} {
		if !strings.Contains(got, seat) {
			t.Fatalf("missing seat %q in %q", seat, got)
		}
	}
}

func TestHelp(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"help"}, &b); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(b.String(), "Quoted ≠ checkout") {
		t.Fatalf("help = %q", b.String())
	}
}

func TestUnknown(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"dashboard"}, &b); code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestSeatsList(t *testing.T) {
	if len(seats()) != 6 {
		t.Fatalf("seats = %v", seats())
	}
}

func TestReport(t *testing.T) {
	var b strings.Builder
	if code := run([]string{"report"}, &b); code != 0 {
		t.Fatalf("exit %d, out=%q", code, b.String())
	}
	out := b.String()
	for _, want := range []string{"agency-white-label", "shopify-mid", "price-delta", "stock-flip"} {
		if !strings.Contains(out, want) {
			t.Fatalf("report missing %q:\n%s", want, out)
		}
	}
}
