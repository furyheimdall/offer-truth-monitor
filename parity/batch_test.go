package parity

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/furyheimdall/offer-truth-monitor/alert"
	"github.com/furyheimdall/offer-truth-monitor/catalog"
	"github.com/furyheimdall/offer-truth-monitor/engines"
	"github.com/furyheimdall/offer-truth-monitor/truth"
)

func TestBatchWatchListDay1Engines(t *testing.T) {
	wl, err := catalog.LoadFile(filepath.Join("testdata", "watchlist_20.json"))
	if err != nil {
		t.Fatal(err)
	}
	if n := len(wl.SKUs); n < 20 || n > catalog.MaxSKUs {
		t.Fatalf("watch list size %d not in 20–50", n)
	}

	cited, live := watchFixture(SKUIds(wl))
	clk := NewFakeClock(time.Date(2026, 9, 15, 9, 0, 0, 0, time.UTC))
	slack := alert.SlackStub()
	email := alert.EmailStub()

	res, err := Run(context.Background(), Input{
		SKUs:    SKUIds(wl),
		Epsilon: 0.5,
		Cadence: Daily,
		Clock:   clk,
		Cited:   cited,
		Live:    live,
		Notify:  []alert.Notifier{slack, email},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped {
		t.Fatal("first run should not skip")
	}
	if res.Compared != 20*len(Day1Engines) {
		t.Fatalf("compared = %d, want %d", res.Compared, 20*len(Day1Engines))
	}
	if len(res.Engines) != 4 {
		t.Fatalf("engines = %v", res.Engines)
	}

	var prices, stocks, currencies int
	for _, d := range res.Drifts {
		switch d.Reason {
		case ReasonPrice:
			prices++
		case ReasonStock:
			stocks++
		case ReasonCurrency:
			currencies++
		}
	}
	if prices != 2 || stocks != 2 || currencies != 1 {
		t.Fatalf("drifts price=%d stock=%d currency=%d all=%+v", prices, stocks, currencies, reasons(res.Drifts))
	}

	alerts := AlertDrifts(res.Drifts)
	if len(alerts) != 4 {
		t.Fatalf("alert drifts = %d, want 4 (|Δprice| and stock-flip only)", len(alerts))
	}
	if len(slack.Sent()) != 4 || len(email.Sent()) != 4 {
		t.Fatalf("slack=%d email=%d, want 4 each", len(slack.Sent()), len(email.Sent()))
	}
	for _, ev := range slack.Sent() {
		if ev.Reason != string(ReasonPrice) && ev.Reason != string(ReasonStock) {
			t.Fatalf("unexpected alert reason %q", ev.Reason)
		}
		if ev.At.IsZero() {
			t.Fatal("expected event timestamp from fake clock")
		}
	}
}

func TestBatchSkipsWhenNotDue(t *testing.T) {
	start := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	clk := NewFakeClock(start)
	clk.Advance(time.Hour)
	res, err := Run(context.Background(), Input{
		SKUs:    []string{"sku-01"},
		Cadence: Daily,
		LastRun: start,
		Clock:   clk,
		Cited:   []CitedSource{MapCited{Name: "claude", Offers: map[string]Offer{}}},
		Live:    MapLive{Name: "json-ld-offer", Offers: map[string]Offer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped || res.Compared != 0 {
		t.Fatalf("result = %+v", res)
	}
}

func TestBatchWeeklyDueThenCompare(t *testing.T) {
	start := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	clk := NewFakeClock(start)
	clk.Advance(7 * 24 * time.Hour)
	cited := MapCited{Name: "perplexity", Offers: map[string]Offer{
		"sku-1": {SKU: "sku-1", Price: "10.00", Currency: "USD", Availability: "InStock"},
	}}
	live := MapLive{Name: "pdp-html", Offers: map[string]Offer{
		"sku-1": {SKU: "sku-1", Price: "12.00", Currency: "USD", Availability: "InStock"},
	}}
	res, err := Run(context.Background(), Input{
		SKUs:    []string{"sku-1"},
		Epsilon: 0.01,
		Cadence: Weekly,
		LastRun: start,
		Clock:   clk,
		Cited:   []CitedSource{cited},
		Live:    live,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped || res.Compared != 1 || len(res.Drifts) != 1 || res.Drifts[0].Reason != ReasonPrice {
		t.Fatalf("result = %+v", res)
	}
}

func TestBatchRejectsOverMaxSKUs(t *testing.T) {
	ids := make([]string, catalog.MaxSKUs+1)
	for i := range ids {
		ids[i] = "sku-" + itoa(i)
	}
	_, err := Run(context.Background(), Input{
		SKUs:  ids,
		Clock: NewFakeClock(time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)),
		Cited: []CitedSource{MapCited{Name: "gemini"}},
		Live:  MapLive{Name: "gmc-feed"},
	})
	if err == nil || !strings.Contains(err.Error(), "exceeds day-1 max") {
		t.Fatalf("err = %v", err)
	}
}

func TestBatchLiveFetchFailClosed(t *testing.T) {
	res, err := Run(context.Background(), Input{
		SKUs:  []string{"missing"},
		Clock: NewFakeClock(time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)),
		Cited: []CitedSource{MapCited{Name: "claude", Offers: map[string]Offer{}}},
		Live:  MapLive{Name: "json-ld-offer", Offers: map[string]Offer{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Compared != 0 || len(res.FetchErrs) != 1 {
		t.Fatalf("result = %+v", res)
	}
}

func TestBatchWireScaffoldEngineAndTruthFixtures(t *testing.T) {
	eng := engines.Fixture{
		Engine: engines.Gemini,
		Offers: map[string]engines.CitedOffer{
			"sku-1": {
				SKU: "sku-1", Engine: engines.Gemini,
				Price: "19.00", Currency: "USD", Availability: "InStock",
			},
		},
	}
	tr := truth.Fixture{
		Src: truth.JSONLDOffer,
		Offers: map[string]truth.LiveOffer{
			"sku-1": {
				SKU: "sku-1", Source: truth.JSONLDOffer,
				Price: "19.00", Currency: "USD", Availability: "OutOfStock",
			},
		},
	}
	res, err := Run(context.Background(), Input{
		SKUs:  []string{"sku-1"},
		Clock: NewFakeClock(time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)),
		Cited: []CitedSource{CitedFunc{
			Name: string(eng.Name()),
			Fn: func(sku string) (Offer, error) {
				o, err := eng.CitedOffer(sku)
				if err != nil {
					return Offer{}, err
				}
				return Offer{SKU: o.SKU, Price: o.Price, Currency: o.Currency, Availability: o.Availability}, nil
			},
		}},
		Live: LiveFunc{
			Name: string(tr.Source()),
			Fn: func(sku string) (Offer, error) {
				o, err := tr.Extract(sku)
				if err != nil {
					return Offer{}, err
				}
				return Offer{SKU: o.SKU, Price: o.Price, Currency: o.Currency, Availability: o.Availability}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Drifts) != 1 || res.Drifts[0].Reason != ReasonStock {
		t.Fatalf("drifts = %+v", res.Drifts)
	}
}

func TestBatchCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, Input{})
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestSKUIdsNil(t *testing.T) {
	if SKUIds(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestAlertDriftsFiltersCurrency(t *testing.T) {
	got := AlertDrifts([]Drift{
		{Reason: ReasonPrice},
		{Reason: ReasonCurrency},
		{Reason: ReasonStock},
	})
	if len(got) != 2 {
		t.Fatalf("got %d", len(got))
	}
}

func watchFixture(skus []string) ([]CitedSource, LiveSource) {
	liveOffers := make(map[string]Offer, len(skus))
	cited := make([]map[string]Offer, len(Day1Engines))
	for i := range cited {
		cited[i] = make(map[string]Offer, len(skus))
	}
	for _, id := range skus {
		base := Offer{SKU: id, Price: "20.00", Currency: "USD", Availability: "InStock"}
		liveOffers[id] = base
		for i := range Day1Engines {
			cited[i][id] = base
		}
	}
	// sku-01: ChatGPT price delta
	c0 := cited[0]["sku-01"]
	c0.Price = "25.00"
	cited[0]["sku-01"] = c0
	// sku-02: Perplexity stock-flip
	c1 := cited[1]["sku-02"]
	c1.Availability = "OutOfStock"
	cited[1]["sku-02"] = c1
	// sku-03: Gemini price + stock
	c2 := cited[2]["sku-03"]
	c2.Price = "18.00"
	c2.Availability = "OutOfStock"
	cited[2]["sku-03"] = c2
	// sku-04: Claude currency mismatch (drift, not an alert)
	c3 := cited[3]["sku-04"]
	c3.Currency = "EUR"
	cited[3]["sku-04"] = c3

	src := make([]CitedSource, len(Day1Engines))
	for i, name := range Day1Engines {
		src[i] = MapCited{Name: name, Offers: cited[i]}
	}
	return src, MapLive{Name: "json-ld-offer", Offers: liveOffers}
}

func reasons(ds []Drift) []Reason {
	out := make([]Reason, len(ds))
	for i, d := range ds {
		out[i] = d.Reason
	}
	return out
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}
