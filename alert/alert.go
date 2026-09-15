// Package alert is the Slack + email notifier seat.
//
// Day-1 plugs are Slack and email. Missing credentials stay on the in-memory
// stub path. Tests inject fake clocks, recorders, and HTTP/mail senders —
// no live network I/O.
package alert

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Seat is the package seat name used by the CLI.
const Seat = "alert"

// Channel is a day-1 notifier plug.
type Channel string

const (
	Slack Channel = "slack"
	Email Channel = "email"
)

// Day1 channels: Slack/email on |Δprice| > ε or stock-flip.
var Day1 = []Channel{Slack, Email}

// Event is one parity drift to notify.
type Event struct {
	SKU     string
	Engine  string
	Reason  string
	Summary string
	At      time.Time
}

// Notifier delivers an alert. Implementations must be safe for tests
// without network.
type Notifier interface {
	Channel() Channel
	Notify(Event) error
}

// Recorder is an in-memory notifier used as the Slack/email plug stub.
type Recorder struct {
	Ch   Channel
	mu   sync.Mutex
	sent []Event
}

// Channel implements Notifier.
func (r *Recorder) Channel() Channel { return r.Ch }

// Notify implements Notifier.
func (r *Recorder) Notify(e Event) error {
	if err := validateEvent(e); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sent = append(r.sent, e)
	return nil
}

// Sent returns a copy of recorded events.
func (r *Recorder) Sent() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Event, len(r.sent))
	copy(out, r.sent)
	return out
}

// SlackStub and EmailStub return fixture notifiers for the two day-1 plugs.
func SlackStub() *Recorder { return &Recorder{Ch: Slack} }
func EmailStub() *Recorder { return &Recorder{Ch: Email} }

// Dispatch fans each event out to every notifier. Partial failures join.
func Dispatch(notifiers []Notifier, events []Event) error {
	var errs []error
	for _, ev := range events {
		for _, n := range notifiers {
			if n == nil {
				errs = append(errs, fmt.Errorf("alert: nil notifier"))
				continue
			}
			if err := n.Notify(ev); err != nil {
				errs = append(errs, fmt.Errorf("alert: %s: %w", n.Channel(), err))
			}
		}
	}
	return errors.Join(errs...)
}

func validateEvent(e Event) error {
	if e.SKU == "" || e.Reason == "" {
		return fmt.Errorf("alert: sku and reason required")
	}
	return nil
}

// FormatText is the shared Slack/email body for one event.
func FormatText(e Event) string {
	eng := e.Engine
	if eng == "" {
		eng = "-"
	}
	at := ""
	if !e.At.IsZero() {
		at = " at " + e.At.UTC().Format(time.RFC3339)
	}
	sum := e.Summary
	if sum == "" {
		sum = e.Reason
	}
	return fmt.Sprintf("otm %s %s engine=%s%s: %s", e.SKU, e.Reason, eng, at, sum)
}
