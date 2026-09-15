package parity

import (
	"testing"
	"time"
)

func TestDueFirstRun(t *testing.T) {
	now := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	due, err := Due(time.Time{}, now, Daily)
	if err != nil || !due {
		t.Fatalf("first run due=%v err=%v", due, err)
	}
}

func TestDueDailyFakeClock(t *testing.T) {
	start := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	clk := NewFakeClock(start)
	due, err := Due(start, clk.Now(), Daily)
	if err != nil || due {
		t.Fatalf("same instant should not be due: due=%v err=%v", due, err)
	}
	clk.Advance(23 * time.Hour)
	due, err = Due(start, clk.Now(), Daily)
	if err != nil || due {
		t.Fatalf("23h should not be due: due=%v err=%v", due, err)
	}
	clk.Advance(time.Hour)
	due, err = Due(start, clk.Now(), Daily)
	if err != nil || !due {
		t.Fatalf("24h should be due: due=%v err=%v", due, err)
	}
}

func TestDueWeeklyFakeClock(t *testing.T) {
	start := time.Date(2026, 9, 15, 8, 0, 0, 0, time.UTC)
	clk := NewFakeClock(start)
	clk.Advance(6 * 24 * time.Hour)
	due, err := Due(start, clk.Now(), Weekly)
	if err != nil || due {
		t.Fatalf("6d should not be due: due=%v err=%v", due, err)
	}
	clk.Advance(24 * time.Hour)
	due, err = Due(start, clk.Now(), Weekly)
	if err != nil || !due {
		t.Fatalf("7d should be due: due=%v err=%v", due, err)
	}
}

func TestDueUnknownCadence(t *testing.T) {
	if _, err := Due(time.Time{}, time.Now(), "hourly"); err == nil {
		t.Fatal("expected unknown cadence error")
	}
}

func TestFakeClockSet(t *testing.T) {
	clk := NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	next := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	clk.Set(next)
	if !clk.Now().Equal(next) {
		t.Fatalf("Now = %s", clk.Now())
	}
}
