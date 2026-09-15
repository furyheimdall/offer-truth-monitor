package alert

import "testing"

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
