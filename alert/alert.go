// Package alert is the Slack + email notifier seat.
// Scaffold notifiers record in memory; no live network I/O.
package alert

import (
	"fmt"
	"sync"
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
	Reason  string
	Summary string
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
	if e.SKU == "" || e.Reason == "" {
		return fmt.Errorf("alert: sku and reason required")
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
