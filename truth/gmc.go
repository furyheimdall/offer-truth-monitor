package truth

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

// GMC is a parsed Google Merchant Center feed (TSV/CSV or RSS/Atom with g:).
type GMC struct {
	offers map[string]offer.Offer
}

// LoadGMC parses a GMC/feed fixture. Incomplete rows are skipped.
// Extract of an unknown SKU fails closed.
func LoadGMC(raw []byte) (*GMC, error) {
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return nil, fmt.Errorf("truth: %s: empty feed (fail closed)", GMCFeed)
	}
	if strings.Contains(s, "<rss") || strings.Contains(s, "<feed") || strings.Contains(s, "g:price") {
		return loadGMCXML(raw)
	}
	return loadGMCDelimited(raw)
}

// Source implements Extractor.
func (g *GMC) Source() Source { return GMCFeed }

// Extract implements Extractor.
func (g *GMC) Extract(sku string) (LiveOffer, error) {
	if g == nil || g.offers == nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", GMCFeed, sku)
	}
	o, ok := g.offers[sku]
	if !ok {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", GMCFeed, sku)
	}
	live := FromShared(GMCFeed, o)
	if err := ValidateRequired(live); err != nil {
		return LiveOffer{}, err
	}
	return live, nil
}

func loadGMCDelimited(raw []byte) (*GMC, error) {
	r := csv.NewReader(bytes.NewReader(raw))
	r.ReuseRecord = false
	r.LazyQuotes = true
	if bytes.Contains(raw, []byte("\t")) {
		r.Comma = '\t'
	}
	rows, err := r.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("truth: %s: unparseable feed (fail closed)", GMCFeed)
	}
	idx := map[string]int{}
	for i, h := range rows[0] {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	col := func(row []string, names ...string) string {
		for _, n := range names {
			if i, ok := idx[n]; ok && i < len(row) {
				return strings.TrimSpace(row[i])
			}
		}
		return ""
	}
	g := &GMC{offers: map[string]offer.Offer{}}
	for _, row := range rows[1:] {
		id := col(row, "id", "sku", "item_id")
		if id == "" {
			continue
		}
		priceRaw := col(row, "price")
		currency := col(row, "currency", "price_currency")
		amount := priceRaw
		if currency == "" {
			var splitErr error
			amount, currency, splitErr = offer.SplitAmountCurrency(priceRaw)
			if splitErr != nil {
				continue
			}
		}
		avail := col(row, "availability")
		var sale *bool
		if sp := col(row, "sale_price", "saleprice"); sp != "" {
			sale = offer.Bool(true)
		}
		o, err := offer.Normalize(offer.Offer{
			SKU:          id,
			Price:        amount,
			Currency:     currency,
			Availability: avail,
			Sale:         sale,
		})
		if err != nil {
			continue
		}
		g.offers[id] = o
	}
	if len(g.offers) == 0 {
		return nil, fmt.Errorf("truth: %s: no valid offers (fail closed)", GMCFeed)
	}
	return g, nil
}

type gmcRSS struct {
	Items []gmcXMLItem `xml:"channel>item"`
}

type gmcAtom struct {
	Entries []gmcXMLItem `xml:"entry"`
}

type gmcXMLItem struct {
	ID           string `xml:"http://base.google.com/ns/1.0 id"`
	Price        string `xml:"http://base.google.com/ns/1.0 price"`
	Availability string `xml:"http://base.google.com/ns/1.0 availability"`
	SalePrice    string `xml:"http://base.google.com/ns/1.0 sale_price"`
}

func loadGMCXML(raw []byte) (*GMC, error) {
	var items []gmcXMLItem
	if bytes.Contains(raw, []byte("<feed")) {
		var feed gmcAtom
		if err := xml.NewDecoder(bytes.NewReader(raw)).Decode(&feed); err != nil {
			return nil, fmt.Errorf("truth: %s: unparseable feed (fail closed)", GMCFeed)
		}
		items = feed.Entries
	} else {
		var rss gmcRSS
		if err := xml.NewDecoder(bytes.NewReader(raw)).Decode(&rss); err != nil {
			return nil, fmt.Errorf("truth: %s: unparseable feed (fail closed)", GMCFeed)
		}
		items = rss.Items
	}
	g := &GMC{offers: map[string]offer.Offer{}}
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		amount, currency, err := offer.SplitAmountCurrency(it.Price)
		if err != nil {
			continue
		}
		var sale *bool
		if strings.TrimSpace(it.SalePrice) != "" {
			sale = offer.Bool(true)
		}
		o, err := offer.Normalize(offer.Offer{
			SKU:          id,
			Price:        amount,
			Currency:     currency,
			Availability: it.Availability,
			Sale:         sale,
		})
		if err != nil {
			continue
		}
		g.offers[id] = o
	}
	if len(g.offers) == 0 {
		return nil, fmt.Errorf("truth: %s: no valid offers (fail closed)", GMCFeed)
	}
	return g, nil
}
