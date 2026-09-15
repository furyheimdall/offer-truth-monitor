package truth

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"unicode"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

var (
	reOpenTag = regexp.MustCompile(`(?is)<([a-z][\w:-]*)\b([^>]*)>`)
	reAttr    = regexp.MustCompile(`(?i)([^\s=]+)\s*=\s*("([^"]*)"|'([^']*)')`)
	reTags    = regexp.MustCompile(`(?s)<[^>]+>`)
)

var voidHTML = map[string]bool{
	"meta": true, "link": true, "img": true, "input": true, "br": true, "hr": true, "area": true, "base": true,
}

// PDPHTMLDir reads {sku}.html fixtures and extracts schema.org microdata
// (itemprop) or data-price / data-currency / data-availability attributes.
type PDPHTMLDir struct {
	FS fs.FS
}

// Source implements Extractor.
func (p PDPHTMLDir) Source() Source { return PDPHTML }

// Extract implements Extractor.
func (p PDPHTMLDir) Extract(sku string) (LiveOffer, error) {
	raw, err := fs.ReadFile(p.FS, sku+".html")
	if err != nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", PDPHTML, sku)
	}
	return parsePDPHTML(sku, raw)
}

func parsePDPHTML(sku string, raw []byte) (LiveOffer, error) {
	html := string(raw)
	props := collectItemprops(html)
	data := collectDataAttrs(html)

	payloadSKU := firstNonEmpty(props["sku"], data["sku"])
	bound, err := offer.BindSKU(sku, payloadSKU)
	if err != nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: %w", PDPHTML, err)
	}

	price := firstNonEmpty(props["price"], data["price"])
	currency := firstNonEmpty(props["pricecurrency"], data["currency"], data["pricecurrency"])
	avail := firstNonEmpty(props["availability"], data["availability"])
	var sale *bool
	if v, ok := props["sale"]; ok {
		sale = parseLooseBool(v)
	} else if v, ok := data["sale"]; ok {
		sale = parseLooseBool(v)
	}

	o, err := offer.Normalize(offer.Offer{
		SKU:          bound,
		Price:        price,
		Currency:     currency,
		Availability: avail,
		Sale:         sale,
	})
	if err != nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: %w", PDPHTML, err)
	}
	live := FromShared(PDPHTML, o)
	if err := ValidateRequired(live); err != nil {
		return LiveOffer{}, err
	}
	return live, nil
}

func collectItemprops(html string) map[string]string {
	out := make(map[string]string)
	for _, m := range reOpenTag.FindAllStringSubmatchIndex(html, -1) {
		tag := strings.ToLower(html[m[2]:m[3]])
		attrs := parseAttrs(html[m[4]:m[5]])
		prop, ok := attrs["itemprop"]
		if !ok {
			continue
		}
		val := firstNonEmpty(attrs["content"], attrs["href"])
		if val == "" && !voidHTML[tag] {
			rest := html[m[1]:]
			closeAt := strings.Index(strings.ToLower(rest), "</"+tag)
			if closeAt >= 0 {
				val = stripTags(rest[:closeAt])
			}
		}
		if strings.TrimSpace(val) == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(prop))
		if _, exists := out[key]; !exists {
			out[key] = strings.TrimSpace(val)
		}
	}
	return out
}

func collectDataAttrs(html string) map[string]string {
	out := make(map[string]string)
	for _, m := range reAttr.FindAllStringSubmatch(html, -1) {
		name := strings.ToLower(m[1])
		if !strings.HasPrefix(name, "data-") {
			continue
		}
		out[strings.TrimPrefix(name, "data-")] = strings.TrimSpace(firstNonEmpty(m[3], m[4]))
	}
	return out
}

func parseAttrs(s string) map[string]string {
	out := make(map[string]string)
	for _, m := range reAttr.FindAllStringSubmatch(s, -1) {
		out[strings.ToLower(m[1])] = firstNonEmpty(m[3], m[4])
	}
	return out
}

func stripTags(s string) string {
	return collapseSpace(strings.TrimSpace(reTags.ReplaceAllString(s, "")))
}

func collapseSpace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseLooseBool(s string) *bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return offer.Bool(true)
	case "0", "false", "no", "off":
		return offer.Bool(false)
	default:
		return nil
	}
}
