package alert

import (
	"strings"
	"testing"
	"time"
)

func TestSeatAndDay1(t *testing.T) {
	if Seat != "alert" {
		t.Fatalf("Seat = %q", Seat)
	}
	if len(Day1) != 2 {
		t.Fatalf("Day1 = %v", Day1)
	}
}

func TestSlackAndEmailStubs(t *testing.T) {
	slack := SlackStub()
	email := EmailStub()
	ev := Event{SKU: "sku-1", Reason: "price-delta", Summary: "|Δprice| > ε"}
	if err := slack.Notify(ev); err != nil {
		t.Fatal(err)
	}
	if err := email.Notify(ev); err != nil {
		t.Fatal(err)
	}
	if slack.Channel() != Slack || email.Channel() != Email {
		t.Fatal("unexpected channels")
	}
	if len(slack.Sent()) != 1 || len(email.Sent()) != 1 {
		t.Fatal("expected one event each")
	}
}

func TestNotifyRequiresFields(t *testing.T) {
	if err := SlackStub().Notify(Event{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestDispatchFanout(t *testing.T) {
	slack := SlackStub()
	email := EmailStub()
	at := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	evs := []Event{
		{SKU: "sku-1", Engine: "claude", Reason: "price-delta", Summary: "|Δprice|=1", At: at},
		{SKU: "sku-2", Engine: "gemini", Reason: "stock-flip", Summary: "stock-flip", At: at},
	}
	if err := Dispatch([]Notifier{slack, email}, evs); err != nil {
		t.Fatal(err)
	}
	if len(slack.Sent()) != 2 || len(email.Sent()) != 2 {
		t.Fatalf("slack=%d email=%d", len(slack.Sent()), len(email.Sent()))
	}
}

func TestDispatchNilNotifier(t *testing.T) {
	if err := Dispatch([]Notifier{nil}, []Event{{SKU: "x", Reason: "price-delta"}}); err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatText(t *testing.T) {
	got := FormatText(Event{
		SKU:     "sku-1",
		Engine:  "perplexity",
		Reason:  "stock-flip",
		Summary: "stock-flip cited=InStock live=OutOfStock",
		At:      time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC),
	})
	for _, want := range []string{"sku-1", "stock-flip", "perplexity", "2026-09-15"} {
		if !strings.Contains(got, want) {
			t.Fatalf("FormatText missing %q: %s", want, got)
		}
	}
}

func TestDay1PlugsWithoutCredentialsAreStubs(t *testing.T) {
	plugs := Day1Plugs(Options{})
	if len(plugs) != 2 {
		t.Fatalf("plugs = %d", len(plugs))
	}
	ev := Event{SKU: "sku-1", Reason: "price-delta", Summary: "delta"}
	if err := Dispatch(plugs, []Event{ev}); err != nil {
		t.Fatal(err)
	}
	slack, ok := plugs[0].(*SlackPlug)
	if !ok || slack.Channel() != Slack {
		t.Fatal("expected slack plug")
	}
	email, ok := plugs[1].(*EmailPlug)
	if !ok || email.Channel() != Email {
		t.Fatal("expected email plug")
	}
	if len(slack.Sent()) != 1 || len(email.Sent()) != 1 {
		t.Fatal("expected stub recording")
	}
}
