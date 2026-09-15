package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ShopifyMidProfile is the only day-1 skin: a thin Shopify mid-market
// config/docs overlay. It is not a console and not a GEO dashboard.
const ShopifyMidProfile = "shopify-mid"

// DefaultShopifyMidAccent is the locked Shopify mid config accent.
const DefaultShopifyMidAccent = "#96bf48"

// Skin is the thin Shopify mid-market white-label config.
// Load it from JSON or use DefaultShopifyMidSkin. There is no UI surface.
type Skin struct {
	Profile    string `json:"profile,omitempty"`
	Agency     string `json:"agency"`
	Merchant   string `json:"merchant"`
	Storefront string `json:"storefront,omitempty"`
	Accent     string `json:"accent,omitempty"`
	LogoText   string `json:"logo_text,omitempty"`
}

// DefaultShopifyMidSkin returns the fixture Shopify mid-market skin.
func DefaultShopifyMidSkin() Skin {
	return Skin{
		Profile:    ShopifyMidProfile,
		Agency:     "North Agency",
		Merchant:   "MidShop",
		Storefront: "midshop.myshopify.com",
		Accent:     DefaultShopifyMidAccent,
		LogoText:   "North × MidShop",
	}
}

// Validate enforces the locked Shopify mid profile and required labels.
func (s Skin) Validate() error {
	if s.Agency == "" {
		return fmt.Errorf("report: agency required")
	}
	if s.Merchant == "" {
		return fmt.Errorf("report: merchant required")
	}
	profile := s.Profile
	if profile == "" {
		profile = ShopifyMidProfile
	}
	if profile != ShopifyMidProfile {
		return fmt.Errorf("report: unsupported skin %q (only %s)", s.Profile, ShopifyMidProfile)
	}
	return nil
}

// LoadSkinJSON reads a Shopify mid skin from r. Unknown fields fail closed
// so dashboard / console / GEO keys cannot sneak into the config.
func LoadSkinJSON(r io.Reader) (Skin, error) {
	var s Skin
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return Skin{}, fmt.Errorf("report: decode skin: %w", err)
	}
	if s.Profile == "" {
		s.Profile = ShopifyMidProfile
	}
	if s.Accent == "" {
		s.Accent = DefaultShopifyMidAccent
	}
	if err := s.Validate(); err != nil {
		return Skin{}, err
	}
	return s, nil
}

// LoadSkinFile reads a Shopify mid skin JSON file.
func LoadSkinFile(path string) (Skin, error) {
	f, err := os.Open(path)
	if err != nil {
		return Skin{}, fmt.Errorf("report: open %s: %w", path, err)
	}
	defer f.Close()
	return LoadSkinJSON(f)
}
