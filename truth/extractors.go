package truth

import (
	"fmt"
	"io/fs"
)

// Day1FromFS wires the four locked live sources to fixture files:
//
//	pdp-html/{sku}.html
//	json-ld-offer/{sku}.json (or {sku}.html)
//	gmc-feed/feed.tsv (or feed.xml / feed.csv)
//	shopify-products-json/products.json
func Day1FromFS(root fs.FS) ([]Extractor, error) {
	pdp, err := fs.Sub(root, string(PDPHTML))
	if err != nil {
		return nil, fmt.Errorf("truth: %s: %w", PDPHTML, err)
	}
	ld, err := fs.Sub(root, string(JSONLDOffer))
	if err != nil {
		return nil, fmt.Errorf("truth: %s: %w", JSONLDOffer, err)
	}

	gmcRaw, _, err := readFirst(root,
		string(GMCFeed)+"/feed.tsv",
		string(GMCFeed)+"/feed.csv",
		string(GMCFeed)+"/feed.xml",
	)
	if err != nil {
		return nil, fmt.Errorf("truth: %s: missing feed fixture (fail closed)", GMCFeed)
	}
	gmc, err := LoadGMC(gmcRaw)
	if err != nil {
		return nil, err
	}

	shopRaw, err := fs.ReadFile(root, string(ShopifyProducts)+"/products.json")
	if err != nil {
		return nil, fmt.Errorf("truth: %s: missing products.json (fail closed)", ShopifyProducts)
	}
	shop, err := LoadShopifyProducts(shopRaw)
	if err != nil {
		return nil, err
	}

	return []Extractor{
		PDPHTMLDir{FS: pdp},
		JSONLDDir{FS: ld},
		gmc,
		shop,
	}, nil
}
