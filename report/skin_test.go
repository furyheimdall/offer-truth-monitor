package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultShopifyMidSkin(t *testing.T) {
	s := DefaultShopifyMidSkin()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.Profile != ShopifyMidProfile {
		t.Fatalf("profile = %q", s.Profile)
	}
	if s.Accent != DefaultShopifyMidAccent {
		t.Fatalf("accent = %q", s.Accent)
	}
}

func TestLoadSkinFile(t *testing.T) {
	s, err := LoadSkinFile(filepath.Join("testdata", "shopify-mid-skin.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Agency != "North Agency" || s.Merchant != "MidShop" || s.Profile != ShopifyMidProfile {
		t.Fatalf("skin = %+v", s)
	}
}

func TestLoadSkinDefaultsProfile(t *testing.T) {
	s, err := LoadSkinJSON(strings.NewReader(`{"agency":"A","merchant":"B"}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Profile != ShopifyMidProfile {
		t.Fatalf("profile = %q", s.Profile)
	}
	if s.Accent != DefaultShopifyMidAccent {
		t.Fatalf("accent = %q", s.Accent)
	}
}

func TestUnsupportedSkin(t *testing.T) {
	_, err := LoadSkinJSON(strings.NewReader(`{"profile":"geo-dashboard","agency":"A","merchant":"B"}`))
	if err == nil {
		t.Fatal("expected unsupported skin")
	}
}

func TestSkinRejectsConsoleFields(t *testing.T) {
	for _, raw := range []string{
		`{"agency":"A","merchant":"B","dashboard":true}`,
		`{"agency":"A","merchant":"B","geo":true}`,
		`{"agency":"A","merchant":"B","console":true}`,
	} {
		if _, err := LoadSkinJSON(strings.NewReader(raw)); err == nil {
			t.Fatalf("expected fail closed for %s", raw)
		}
	}
}

func TestLoadSkinRequiresLabels(t *testing.T) {
	if _, err := LoadSkinJSON(strings.NewReader(`{"merchant":"B"}`)); err == nil {
		t.Fatal("expected agency required")
	}
}

func TestLoadSkinFileMissing(t *testing.T) {
	if _, err := LoadSkinFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected open error")
	}
}

func TestLoadSkinFileUsesOS(t *testing.T) {
	// Keep LoadSkinFile exercised on a real path besides testdata.
	p := filepath.Join(t.TempDir(), "skin.json")
	if err := os.WriteFile(p, []byte(`{"agency":"Temp","merchant":"Shop"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSkinFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if s.Agency != "Temp" {
		t.Fatalf("skin = %+v", s)
	}
}
