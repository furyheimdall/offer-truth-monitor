package parity

import "time"

// Clock is the time source for batch cadence. Tests inject a fake clock;
// production uses System.
type Clock interface {
	Now() time.Time
}

// System returns a Clock backed by time.Now.
func System() Clock { return systemClock{} }

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// FakeClock is a test clock. Advance to exercise daily/weekly due checks
// without waiting or using the wall clock.
type FakeClock struct {
	t time.Time
}

// NewFakeClock returns a fake clock frozen at t.
func NewFakeClock(t time.Time) *FakeClock { return &FakeClock{t: t} }

// Now implements Clock.
func (f *FakeClock) Now() time.Time { return f.t }

// Advance moves the fake clock forward by d.
func (f *FakeClock) Advance(d time.Duration) { f.t = f.t.Add(d) }

// Set jumps the fake clock to t.
func (f *FakeClock) Set(t time.Time) { f.t = t }

// Cadence is the day-1 batch rhythm: daily or weekly.
type Cadence string

const (
	Daily  Cadence = "daily"
	Weekly Cadence = "weekly"
)

// Due reports whether a batch should run at now given the last run time.
// A zero last run is always due (first pass). Empty cadence defaults to daily.
func Due(last, now time.Time, cadence Cadence) (bool, error) {
	period, err := periodOf(cadence)
	if err != nil {
		return false, err
	}
	if last.IsZero() {
		return true, nil
	}
	return !now.Before(last.Add(period)), nil
}

func periodOf(cadence Cadence) (time.Duration, error) {
	switch cadence {
	case Daily, "":
		return 24 * time.Hour, nil
	case Weekly:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmtCadence(cadence)
	}
}

func fmtCadence(c Cadence) error {
	return errf("parity: unknown cadence %q (want daily or weekly)", c)
}
