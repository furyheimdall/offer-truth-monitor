package engines

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/furyheimdall/offer-truth-monitor/offer"
)

// parseCited maps one engine's fixture bytes to the shared offer shape.
type parseCited func(sku string, raw []byte) (offer.Offer, error)

// FileAdapter loads per-SKU fixture files (no network).
// Files are named {sku}.json under FS.
type FileAdapter struct {
	Engine Engine
	FS     fs.FS
	Parse  parseCited
}

// Name implements Adapter.
func (a FileAdapter) Name() Engine { return a.Engine }

// CitedOffer implements Adapter. Unknown SKUs and incomplete payloads fail closed.
func (a FileAdapter) CitedOffer(sku string) (CitedOffer, error) {
	if a.Parse == nil {
		return CitedOffer{}, fmt.Errorf("engines: %s: missing parser (fail closed)", a.Engine)
	}
	sku = strings.TrimSpace(sku)
	if sku == "" {
		return CitedOffer{}, fmt.Errorf("engines: missing required field sku")
	}
	raw, err := fs.ReadFile(a.FS, sku+".json")
	if err != nil {
		return CitedOffer{}, fmt.Errorf("engines: %s: unknown sku %q (fail closed)", a.Engine, sku)
	}
	o, err := a.Parse(sku, raw)
	if err != nil {
		return CitedOffer{}, fmt.Errorf("engines: %s: %w", a.Engine, err)
	}
	cited := FromShared(a.Engine, o)
	if err := ValidateRequired(cited); err != nil {
		return CitedOffer{}, err
	}
	return cited, nil
}

// NewChatGPTShopping returns a fixture-backed ChatGPT Shopping adapter.
// This is not a shopping-card API guarantee; the fixture shape may change.
func NewChatGPTShopping(fsys fs.FS) Adapter {
	return FileAdapter{Engine: ChatGPTShopping, FS: fsys, Parse: ParseChatGPTShopping}
}

// NewPerplexity returns a fixture-backed Perplexity cited-offer adapter.
func NewPerplexity(fsys fs.FS) Adapter {
	return FileAdapter{Engine: Perplexity, FS: fsys, Parse: ParsePerplexity}
}

// NewGemini returns a fixture-backed Gemini cited-offer adapter.
func NewGemini(fsys fs.FS) Adapter {
	return FileAdapter{Engine: Gemini, FS: fsys, Parse: ParseGemini}
}

// NewClaude returns a fixture-backed Claude cited-offer adapter.
func NewClaude(fsys fs.FS) Adapter {
	return FileAdapter{Engine: Claude, FS: fsys, Parse: ParseClaude}
}

// Day1FromFS wires the locked 3–4 engines to fixture subdirectories
// named after each Engine constant (chatgpt-shopping, perplexity, gemini, claude).
func Day1FromFS(root fs.FS) ([]Adapter, error) {
	out := make([]Adapter, 0, len(Day1))
	for _, eng := range Day1 {
		sub, err := fs.Sub(root, string(eng))
		if err != nil {
			return nil, fmt.Errorf("engines: %s: %w", eng, err)
		}
		switch eng {
		case ChatGPTShopping:
			out = append(out, NewChatGPTShopping(sub))
		case Perplexity:
			out = append(out, NewPerplexity(sub))
		case Gemini:
			out = append(out, NewGemini(sub))
		case Claude:
			out = append(out, NewClaude(sub))
		default:
			return nil, fmt.Errorf("engines: unknown engine %q (fail closed)", eng)
		}
	}
	return out, nil
}
