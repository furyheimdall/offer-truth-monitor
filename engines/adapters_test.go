package engines

import (
	"net/http"
	"os"
	"testing"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

func TestMain(m *testing.M) {
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		panic("engines tests must not use the network")
	})
	os.Exit(m.Run())
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testdataAdapters(t *testing.T) []Adapter {
	t.Helper()
	as, err := Day1FromFS(os.DirFS("testdata"))
	if err != nil {
		t.Fatal(err)
	}
	if len(as) < 3 || len(as) > 4 {
		t.Fatalf("day-1 adapter count = %d, want 3–4", len(as))
	}
	return as
}

func TestDay1AdaptersImplementSeat(t *testing.T) {
	as := testdataAdapters(t)
	seen := map[Engine]bool{}
	for _, a := range as {
		seen[a.Name()] = true
		o, err := a.CitedOffer("sku-1")
		if err != nil {
			t.Fatalf("%s: %v", a.Name(), err)
		}
		if err := ValidateRequired(o); err != nil {
			t.Fatalf("%s: %v", a.Name(), err)
		}
		if o.Price != "19.00" || o.Currency != "USD" || o.Availability != "InStock" {
			t.Fatalf("%s: shared shape = %+v", a.Name(), o)
		}
		if o.Sale != nil {
			t.Fatalf("%s: sale should be optional/unset, got %v", a.Name(), *o.Sale)
		}
		shared := o.Shared()
		if err := offer.Validate(shared); err != nil {
			t.Fatalf("%s shared: %v", a.Name(), err)
		}
	}
	for _, eng := range Day1 {
		if !seen[eng] {
			t.Fatalf("missing day-1 adapter %s", eng)
		}
	}
}

func TestChatGPTShoppingSaleOptional(t *testing.T) {
	a := NewChatGPTShopping(os.DirFS("testdata/chatgpt-shopping"))
	o, err := a.CitedOffer("sku-sale")
	if err != nil {
		t.Fatal(err)
	}
	if o.Sale == nil || !*o.Sale {
		t.Fatalf("expected sale=true, got %+v", o)
	}
}

func TestAdaptersUnknownSKUFailClosed(t *testing.T) {
	for _, a := range testdataAdapters(t) {
		if _, err := a.CitedOffer("missing"); err == nil {
			t.Fatalf("%s: expected unknown sku to fail closed", a.Name())
		}
	}
}

func TestChatGPTMissingCurrencyFailClosed(t *testing.T) {
	a := NewChatGPTShopping(os.DirFS("testdata/chatgpt-shopping"))
	if _, err := a.CitedOffer("incomplete"); err == nil {
		t.Fatal("expected missing currency to fail closed")
	}
}

func TestParseFailClosed(t *testing.T) {
	cases := []struct {
		name  string
		parse parseCited
		raw   string
	}{
		{"chatgpt", ParseChatGPTShopping, `{"merchant_sku":"sku-1"}`},
		{"perplexity", ParsePerplexity, `{"sku":"sku-1","citations":[]}`},
		{"gemini", ParseGemini, `{"productSku":"sku-1"}`},
		{"claude", ParseClaude, `{"sku":"sku-1","quoted":"19.00","availability":"InStock"}`},
		{"bad-json", ParseChatGPTShopping, `{`},
	}
	for _, tc := range cases {
		if _, err := tc.parse("sku-1", []byte(tc.raw)); err == nil {
			t.Fatalf("%s: expected fail-closed", tc.name)
		}
	}
}

func TestSKUMismatchFailClosed(t *testing.T) {
	raw := []byte(`{"sku":"other","quoted":"19.00 USD","availability":"InStock"}`)
	if _, err := ParseClaude("sku-1", raw); err == nil {
		t.Fatal("expected sku mismatch to fail closed")
	}
}
