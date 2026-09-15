// Command otm is the Offer Truth Monitor CLI stub.
// It prints seat names or help. No network I/O.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/furyheimdall/offer-truth-monitor/alert"
	"github.com/furyheimdall/offer-truth-monitor/catalog"
	"github.com/furyheimdall/offer-truth-monitor/engines"
	"github.com/furyheimdall/offer-truth-monitor/parity"
	"github.com/furyheimdall/offer-truth-monitor/report"
	"github.com/furyheimdall/offer-truth-monitor/truth"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, w io.Writer) int {
	cmd := "seats"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "help", "-h", "--help":
		return writeHelp(w)
	case "seats", "":
		return writeSeats(w)
	case "report":
		return writeReport(w)
	default:
		fmt.Fprintf(w, "unknown command %q\n", cmd)
		writeHelp(w)
		return 2
	}
}

func writeHelp(w io.Writer) int {
	fmt.Fprint(w, `Offer Truth Monitor — otm

SKU × AI-assistant cited price/availability vs live PDP + JSON-LD + feed — alert on drift.
Quoted ≠ checkout. / Visibility ≠ truth.

Usage:
  otm          print package seats
  otm seats    print package seats
  otm report   render the agency white-label report from fixture parity results
  otm help     print this help

No network. Not a GEO dashboard / not a full PMS-style console.
`)
	return 0
}

func writeReport(w io.Writer) int {
	in, err := report.FromCompare(report.DefaultShopifyMidSkin(), report.FixturePairs(), 0.5)
	if err != nil {
		fmt.Fprintf(w, "report: %v\n", err)
		return 1
	}
	if err := report.Render(w, in); err != nil {
		fmt.Fprintf(w, "report: %v\n", err)
		return 1
	}
	return 0
}

func writeSeats(w io.Writer) int {
	for _, name := range seats() {
		fmt.Fprintln(w, name)
	}
	return 0
}

func seats() []string {
	return []string{
		catalog.Seat,
		engines.Seat,
		truth.Seat,
		parity.Seat,
		alert.Seat,
		report.Seat,
	}
}
