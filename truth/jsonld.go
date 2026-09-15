package truth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

var reJSONLDScript = regexp.MustCompile(`(?is)<script\b[^>]*type\s*=\s*["']application/ld\+json["'][^>]*>(.*?)</script>`)

// JSONLDDir reads {sku}.json or {sku}.html fixtures containing schema.org
// JSON-LD Product / Offer nodes.
type JSONLDDir struct {
	FS fs.FS
}

// Source implements Extractor.
func (j JSONLDDir) Source() Source { return JSONLDOffer }

// Extract implements Extractor.
func (j JSONLDDir) Extract(sku string) (LiveOffer, error) {
	raw, name, err := readFirst(j.FS, sku+".json", sku+".html")
	if err != nil {
		return LiveOffer{}, fmt.Errorf("truth: %s: unknown sku %q (fail closed)", JSONLDOffer, sku)
	}
	_ = name
	return parseJSONLD(sku, raw)
}

func readFirst(fsys fs.FS, names ...string) ([]byte, string, error) {
	var last error
	for _, n := range names {
		b, err := fs.ReadFile(fsys, n)
		if err == nil {
			return b, n, nil
		}
		last = err
	}
	return nil, "", last
}

func parseJSONLD(sku string, raw []byte) (LiveOffer, error) {
	blobs := jsonLDBlobs(raw)
	if len(blobs) == 0 {
		return LiveOffer{}, fmt.Errorf("truth: %s: missing required field price", JSONLDOffer)
	}
	var last error
	for _, blob := range blobs {
		nodes, err := decodeJSONLDNodes(blob)
		if err != nil {
			last = err
			continue
		}
		if o, err := offerFromJSONLDNodes(sku, nodes); err == nil {
			live := FromShared(JSONLDOffer, o)
			if err := ValidateRequired(live); err != nil {
				return LiveOffer{}, err
			}
			return live, nil
		} else {
			last = err
		}
	}
	if last == nil {
		last = fmt.Errorf("offer: missing required field price")
	}
	return LiveOffer{}, fmt.Errorf("truth: %s: %w", JSONLDOffer, last)
}

func jsonLDBlobs(raw []byte) [][]byte {
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return [][]byte{[]byte(s)}
	}
	var out [][]byte
	for _, m := range reJSONLDScript.FindAllStringSubmatch(s, -1) {
		out = append(out, []byte(strings.TrimSpace(m[1])))
	}
	return out
}

func decodeJSONLDNodes(raw []byte) ([]map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("offer: unparseable json-ld")
	}
	return flattenJSONLD(v), nil
}

func flattenJSONLD(v any) []map[string]any {
	var out []map[string]any
	switch t := v.(type) {
	case []any:
		for _, item := range t {
			out = append(out, flattenJSONLD(item)...)
		}
	case map[string]any:
		out = append(out, t)
		if g, ok := t["@graph"]; ok {
			out = append(out, flattenJSONLD(g)...)
		}
		if o, ok := t["offers"]; ok {
			out = append(out, flattenJSONLD(o)...)
		}
	}
	return out
}

func offerFromJSONLDNodes(sku string, nodes []map[string]any) (offer.Offer, error) {
	var productSKU string
	for _, n := range nodes {
		if hasJSONLDType(n, "Product") {
			if s := jsonString(n["sku"]); s != "" {
				productSKU = s
			}
		}
	}
	for _, n := range nodes {
		if !hasJSONLDType(n, "Offer") && jsonString(n["price"]) == "" && n["priceCurrency"] == nil {
			continue
		}
		payloadSKU := firstNonEmpty(jsonString(n["sku"]), productSKU)
		bound, err := offer.BindSKU(sku, payloadSKU)
		if err != nil {
			return offer.Offer{}, err
		}
		price := jsonString(n["price"])
		currency := jsonString(n["priceCurrency"])
		if price == "" && n["priceSpecification"] != nil {
			if spec, ok := n["priceSpecification"].(map[string]any); ok {
				price = jsonString(spec["price"])
				if currency == "" {
					currency = jsonString(spec["priceCurrency"])
				}
			}
		}
		avail := jsonString(n["availability"])
		var sale *bool
		if sp := jsonString(n["salePrice"]); sp != "" {
			sale = offer.Bool(true)
		}
		return offer.Normalize(offer.Offer{
			SKU:          bound,
			Price:        price,
			Currency:     currency,
			Availability: avail,
			Sale:         sale,
		})
	}
	return offer.Offer{}, fmt.Errorf("offer: missing required field price")
}

func hasJSONLDType(n map[string]any, want string) bool {
	v, ok := n["@type"]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case string:
		return strings.EqualFold(t, want)
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && strings.EqualFold(s, want) {
				return true
			}
		}
	}
	return false
}

func jsonString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}
