# report/

Agency white-label report seat. **One kind:** `agency-white-label`.

The template renders from [`parity/`](../parity/) results (price-delta, stock-flip, currency-mismatch). Tests and `otm report` use fixtures/stubs — no live network.

## Shopify mid skin (config/docs only)

ICP is Shopify mid-market + agency white-label. The skin is a JSON config overlay, not a product surface:

- **In:** agency / merchant labels, optional storefront, accent hex, text logo
- **Out:** SaaS / PMS-style console, GEO / visibility dashboard, HTML app, auto-fix, extra report kinds

Profile is locked to `shopify-mid`. Unknown JSON fields fail closed so `dashboard`, `geo`, and `console` keys cannot be added.

Example (`testdata/shopify-mid-skin.json`):

```json
{
  "profile": "shopify-mid",
  "agency": "North Agency",
  "merchant": "MidShop",
  "storefront": "midshop.myshopify.com",
  "accent": "#96bf48",
  "logo_text": "North × MidShop"
}
```

```bash
go run ./cmd/otm report
```
