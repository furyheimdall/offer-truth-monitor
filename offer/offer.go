// Package offer is the thin shared domain shape used by engines (cited)
// and truth (live). Required fields: price, currency, availability.
// Sale is optional. Missing or unusable required fields fail closed.
package offer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Offer is the shared cited/live offer shape.
type Offer struct {
	SKU          string
	Price        string
	Currency     string
	Availability string
	Sale         *bool
}

// Bool returns a *bool for the optional Sale field.
func Bool(v bool) *bool { return &v }

// Validate fails closed when a required field is missing or unusable.
// Sale may be nil. SKU is required so a price can be attributed.
func Validate(o Offer) error {
	if strings.TrimSpace(o.SKU) == "" {
		return fmt.Errorf("offer: missing required field sku")
	}
	if _, err := NormalizePrice(o.Price); err != nil {
		return err
	}
	if _, err := NormalizeCurrency(o.Currency); err != nil {
		return err
	}
	if _, err := NormalizeAvailability(o.Availability); err != nil {
		return err
	}
	return nil
}

// Normalize trims and canonicalizes required fields. Sale is passed through.
func Normalize(o Offer) (Offer, error) {
	out := Offer{Sale: o.Sale, SKU: strings.TrimSpace(o.SKU)}
	if out.SKU == "" {
		return Offer{}, fmt.Errorf("offer: missing required field sku")
	}
	var err error
	if out.Price, err = NormalizePrice(o.Price); err != nil {
		return Offer{}, err
	}
	if out.Currency, err = NormalizeCurrency(o.Currency); err != nil {
		return Offer{}, err
	}
	if out.Availability, err = NormalizeAvailability(o.Availability); err != nil {
		return Offer{}, err
	}
	return out, nil
}

// BindSKU requires the payload SKU (when present) to match the requested SKU.
// An empty payload SKU inherits the requested SKU (single-product sources).
func BindSKU(requested, payload string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return "", fmt.Errorf("offer: missing required field sku")
	}
	payload = strings.TrimSpace(payload)
	if payload != "" && payload != requested {
		return "", fmt.Errorf("offer: sku mismatch %q vs %q (fail closed)", payload, requested)
	}
	return requested, nil
}

// NormalizePrice fails closed on empty or non-numeric price. Commas are stripped.
// The trimmed input is kept when it parses so fixtures like "19.00" stay "19.00".
func NormalizePrice(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("offer: missing required field price")
	}
	cleaned := strings.ReplaceAll(s, ",", "")
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return "", fmt.Errorf("offer: missing required field price")
	}
	return s, nil
}

// NormalizeCurrency fails closed unless the value is a 3-letter ISO-4217-like code.
func NormalizeCurrency(s string) (string, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return "", fmt.Errorf("offer: missing required field currency")
	}
	if len(s) != 3 {
		return "", fmt.Errorf("offer: missing required field currency")
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return "", fmt.Errorf("offer: missing required field currency")
		}
	}
	return s, nil
}

// NormalizeAvailability maps common aliases and schema.org URLs to a short token.
// Empty availability fails closed. Unrecognized non-empty tokens pass through trimmed.
func NormalizeAvailability(s string) (string, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return "", fmt.Errorf("offer: missing required field availability")
	}
	key := canonicalAvailKey(raw)
	if mapped, ok := availabilityAliases[key]; ok {
		return mapped, nil
	}
	return raw, nil
}

func canonicalAvailKey(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "/"); i >= 0 && i+1 < len(s) {
		s = s[i+1:]
	}
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var availabilityAliases = map[string]string{
	"instock":             "InStock",
	"available":           "InStock",
	"outofstock":          "OutOfStock",
	"unavailable":         "OutOfStock",
	"soldout":             "OutOfStock",
	"oos":                 "OutOfStock",
	"limitedavailability": "LimitedAvailability",
	"limited":             "LimitedAvailability",
	"preorder":            "PreOrder",
	"backorder":           "BackOrder",
}

// SplitAmountCurrency parses GMC-style "19.99 USD" or "USD 19.99".
func SplitAmountCurrency(s string) (amount, currency string, err error) {
	fields := strings.Fields(strings.TrimSpace(s))
	switch len(fields) {
	case 2:
		if looksCurrency(fields[1]) {
			return fields[0], fields[1], nil
		}
		if looksCurrency(fields[0]) {
			return fields[1], fields[0], nil
		}
	case 1:
		if _, e := strconv.ParseFloat(strings.ReplaceAll(fields[0], ",", ""), 64); e == nil {
			return fields[0], "", fmt.Errorf("offer: missing required field currency")
		}
	}
	return "", "", fmt.Errorf("offer: missing required field price")
}

func looksCurrency(s string) bool {
	_, err := NormalizeCurrency(s)
	return err == nil
}
