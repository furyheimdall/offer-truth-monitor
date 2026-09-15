package parity

import (
	"context"
	"fmt"
	"time"

	"github.com/furyheimdall/offer-truth-monitor/alert"
	"github.com/furyheimdall/offer-truth-monitor/catalog"
)

// Input is one batch compare of a watch list × cited sources vs live.
//
// SKUs should be the catalog watch-list IDs (day-1: 20–50). Cited sources
// are thin ports over engine adapters; Live is a thin port over a resolved
// truth offer. Neither engines nor truth seats are implemented here.
type Input struct {
	SKUs    []string
	Epsilon float64
	Cadence Cadence
	LastRun time.Time
	Clock   Clock
	Cited   []CitedSource
	Live    LiveSource
	Notify  []alert.Notifier // optional Slack + email plugs
}

// FetchError is a fail-closed fetch for one SKU × source.
type FetchError struct {
	SKU    string
	Engine string
	Source string
	Err    error
}

func (e FetchError) Error() string {
	switch {
	case e.Engine != "":
		return fmt.Sprintf("parity: sku %s engine %s: %v", e.SKU, e.Engine, e.Err)
	case e.Source != "":
		return fmt.Sprintf("parity: sku %s live %s: %v", e.SKU, e.Source, e.Err)
	default:
		return fmt.Sprintf("parity: sku %s: %v", e.SKU, e.Err)
	}
}

// Result is the outcome of one batch pass.
type Result struct {
	RanAt     time.Time
	Cadence   Cadence
	SKUCount  int
	Engines   []string
	Compared  int
	Skipped   bool
	Drifts    []Drift
	FetchErrs []FetchError
}

// SKUIds extracts catalog watch-list IDs for the batch path.
func SKUIds(wl *catalog.Watchlist) []string {
	if wl == nil {
		return nil
	}
	ids := make([]string, len(wl.SKUs))
	for i, s := range wl.SKUs {
		ids[i] = s.ID
	}
	return ids
}

// Run compares every watch-list SKU × cited source against live.
// When Notify is set, |Δprice| > ε and stock-flip drifts are dispatched
// to the plugged notifiers. Currency-mismatch stays on the result only
// (fail-closed, not a day-1 alert).
func Run(ctx context.Context, in Input) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if in.Epsilon < 0 {
		return Result{}, errf("parity: epsilon must be >= 0")
	}
	clock := in.Clock
	if clock == nil {
		clock = System()
	}
	now := clock.Now()
	cadence := in.Cadence
	if cadence == "" {
		cadence = Daily
	}
	due, err := Due(in.LastRun, now, cadence)
	if err != nil {
		return Result{}, err
	}

	out := Result{
		RanAt:    now,
		Cadence:  cadence,
		SKUCount: len(in.SKUs),
		Engines:  engineNames(in.Cited),
	}
	if !due {
		out.Skipped = true
		return out, nil
	}
	if err := validateSKUs(in.SKUs); err != nil {
		return Result{}, err
	}
	if len(in.Cited) == 0 {
		return Result{}, errf("parity: at least one cited source required")
	}
	if in.Live == nil {
		return Result{}, errf("parity: live source required")
	}

	liveName := in.Live.Source()
	for _, sku := range in.SKUs {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		live, err := in.Live.Live(sku)
		if err != nil {
			out.FetchErrs = append(out.FetchErrs, FetchError{SKU: sku, Source: liveName, Err: err})
			continue
		}
		live.SKU = sku
		for _, src := range in.Cited {
			if src == nil {
				return Result{}, errf("parity: nil cited source")
			}
			eng := src.Engine()
			cited, err := src.Cited(sku)
			if err != nil {
				out.FetchErrs = append(out.FetchErrs, FetchError{SKU: sku, Engine: eng, Err: err})
				continue
			}
			cited.SKU = sku
			drifts, err := Compare(cited, live, in.Epsilon)
			if err != nil {
				out.FetchErrs = append(out.FetchErrs, FetchError{SKU: sku, Engine: eng, Source: liveName, Err: err})
				continue
			}
			out.Compared++
			for _, d := range drifts {
				d.Engine = eng
				d.LiveSource = liveName
				out.Drifts = append(out.Drifts, d)
			}
		}
	}

	if len(in.Notify) > 0 {
		if err := Notify(in.Notify, AlertDrifts(out.Drifts), now); err != nil {
			return out, err
		}
	}
	return out, nil
}

// AlertDrifts returns the day-1 alert set: |Δprice| > ε or stock-flip.
func AlertDrifts(drifts []Drift) []Drift {
	out := make([]Drift, 0, len(drifts))
	for _, d := range drifts {
		if d.Reason == ReasonPrice || d.Reason == ReasonStock {
			out = append(out, d)
		}
	}
	return out
}

// Notify sends drifts to Slack + email (or any) notifier plugs.
func Notify(notifiers []alert.Notifier, drifts []Drift, at time.Time) error {
	evs := make([]alert.Event, 0, len(drifts))
	for _, d := range drifts {
		evs = append(evs, alert.Event{
			SKU:     d.SKU,
			Engine:  d.Engine,
			Reason:  string(d.Reason),
			Summary: d.Summary(),
			At:      at,
		})
	}
	return alert.Dispatch(notifiers, evs)
}

func validateSKUs(skus []string) error {
	if len(skus) > catalog.MaxSKUs {
		return errf("parity: %d SKUs exceeds day-1 max %d", len(skus), catalog.MaxSKUs)
	}
	seen := make(map[string]struct{}, len(skus))
	for i, id := range skus {
		if id == "" {
			return errf("parity: sku[%d] missing id", i)
		}
		if _, ok := seen[id]; ok {
			return errf("parity: duplicate sku id %q", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func engineNames(src []CitedSource) []string {
	out := make([]string, 0, len(src))
	for _, s := range src {
		if s == nil {
			continue
		}
		out = append(out, s.Engine())
	}
	return out
}
