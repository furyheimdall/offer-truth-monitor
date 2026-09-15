package truth

import (
	"net/http"
	"os"
	"testing"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

func TestMain(m *testing.M) {
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		panic("truth tests must not use the network")
	})
	os.Exit(m.Run())
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testdataExtractors(t *testing.T) []Extractor {
	t.Helper()
	es, err := Day1FromFS(os.DirFS("testdata"))
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 4 {
		t.Fatalf("day-1 extractors = %d, want 4", len(es))
	}
	return es
}

func TestDay1ExtractorsSharedShape(t *testing.T) {
	seen := map[Source]bool{}
	for _, ex := range testdataExtractors(t) {
		seen[ex.Source()] = true
		o, err := ex.Extract("sku-1")
		if err != nil {
			t.Fatalf("%s: %v", ex.Source(), err)
		}
		if err := ValidateRequired(o); err != nil {
			t.Fatalf("%s: %v", ex.Source(), err)
		}
		if o.Price != "18.00" || o.Currency != "USD" || o.Availability != "InStock" {
			t.Fatalf("%s: shared shape = %+v", ex.Source(), o)
		}
		if o.Sale != nil {
			t.Fatalf("%s: sale should be unset, got %v", ex.Source(), *o.Sale)
		}
		if err := offer.Validate(o.Shared()); err != nil {
			t.Fatalf("%s shared: %v", ex.Source(), err)
		}
	}
	for _, src := range Day1 {
		if !seen[src] {
			t.Fatalf("missing day-1 source %s", src)
		}
	}
}

func TestExtractorsUnknownSKUFailClosed(t *testing.T) {
	for _, ex := range testdataExtractors(t) {
		if _, err := ex.Extract("missing"); err == nil {
			t.Fatalf("%s: expected unknown sku to fail closed", ex.Source())
		}
	}
}

func TestPDPHTMLMissingPriceFailClosed(t *testing.T) {
	ex := PDPHTMLDir{FS: os.DirFS("testdata/pdp-html")}
	if _, err := ex.Extract("missing-price"); err == nil {
		t.Fatal("expected missing price to fail closed")
	}
}

func TestJSONLDMissingCurrencyFailClosed(t *testing.T) {
	ex := JSONLDDir{FS: os.DirFS("testdata/json-ld-offer")}
	if _, err := ex.Extract("missing-currency"); err == nil {
		t.Fatal("expected missing currency to fail closed")
	}
}

func TestJSONLDFromHTMLScript(t *testing.T) {
	html := []byte(`<html><script type="application/ld+json">
	{"@type":"Offer","sku":"sku-1","price":"18.00","priceCurrency":"USD","availability":"InStock"}
	</script></html>`)
	o, err := parseJSONLD("sku-1", html)
	if err != nil {
		t.Fatal(err)
	}
	if o.Price != "18.00" || o.Currency != "USD" {
		t.Fatalf("got %+v", o)
	}
}

func TestGMCXMLAndSale(t *testing.T) {
	xmlFeed := []byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:g="http://base.google.com/ns/1.0">
  <channel>
    <item>
      <g:id>sku-1</g:id>
      <g:price>18.00 USD</g:price>
      <g:availability>in stock</g:availability>
    </item>
    <item>
      <g:id>sku-sale</g:id>
      <g:price>17.99 USD</g:price>
      <g:availability>in stock</g:availability>
      <g:sale_price>15.00 USD</g:sale_price>
    </item>
  </channel>
</rss>`)
	g, err := LoadGMC(xmlFeed)
	if err != nil {
		t.Fatal(err)
	}
	o, err := g.Extract("sku-1")
	if err != nil {
		t.Fatal(err)
	}
	if o.Price != "18.00" || o.Availability != "InStock" {
		t.Fatalf("xml sku-1 = %+v", o)
	}
	sale, err := g.Extract("sku-sale")
	if err != nil {
		t.Fatal(err)
	}
	if sale.Sale == nil || !*sale.Sale {
		t.Fatalf("expected sale, got %+v", sale)
	}

	tsv, err := os.ReadFile("testdata/gmc-feed/feed.tsv")
	if err != nil {
		t.Fatal(err)
	}
	gt, err := LoadGMC(tsv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gt.Extract("incomplete"); err == nil {
		t.Fatal("expected incomplete GMC row to fail closed")
	}
}

func TestShopifyNoCurrencyFailClosed(t *testing.T) {
	raw := []byte(`{"products":[{"variants":[{"sku":"sku-1","price":"18.00","available":true}]}]}`)
	if _, err := LoadShopifyProducts(raw); err == nil {
		t.Fatal("expected missing currency to fail closed")
	}
}

func TestShopifyPresentmentCurrencyAndAdminStub(t *testing.T) {
	raw := []byte(`{
	  "products":[{
	    "variants":[{
	      "sku":"sku-1",
	      "available":true,
	      "presentment_prices":[{"price":{"amount":"18.00","currency_code":"USD"}}]
	    }]
	  }]
	}`)
	cat, err := LoadShopifyProducts(raw)
	if err != nil {
		t.Fatal(err)
	}
	o, err := cat.Extract("sku-1")
	if err != nil {
		t.Fatal(err)
	}
	if o.Price != "18.00" || o.Currency != "USD" {
		t.Fatalf("got %+v", o)
	}

	if _, err := LoadShopifyAdmin(nil); err == nil {
		t.Fatal("expected empty admin stub to fail closed")
	}
	if _, err := (ShopifyAdminStub{}).Extract("sku-1"); err == nil {
		t.Fatal("expected admin stub extract to fail closed")
	}

	admin, err := LoadShopifyAdmin([]byte(`{
	  "products":[{
	    "variants":[{
	      "sku":"sku-1",
	      "price":"18.00",
	      "inventory_quantity":3,
	      "presentment_prices":[{"price":{"amount":"18.00","currency_code":"USD"}}]
	    }]
	  }]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if admin.Source() != ShopifyAdminRead {
		t.Fatalf("source = %s", admin.Source())
	}
	if _, err := admin.Extract("sku-1"); err != nil {
		t.Fatal(err)
	}
}

func TestShopifySaleFromCompareAt(t *testing.T) {
	ex := testdataExtractors(t)
	var shop Extractor
	for _, e := range ex {
		if e.Source() == ShopifyProducts {
			shop = e
		}
	}
	o, err := shop.Extract("sku-sale")
	if err != nil {
		t.Fatal(err)
	}
	if o.Sale == nil || !*o.Sale {
		t.Fatalf("expected sale, got %+v", o)
	}
}

func TestValidateRequiredFailClosed(t *testing.T) {
	if err := ValidateRequired(LiveOffer{SKU: "x", Source: PDPHTML, Price: "1"}); err == nil {
		t.Fatal("expected missing currency/availability to fail closed")
	}
}
