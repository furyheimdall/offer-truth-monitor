# Offer Truth Monitor

SKU × AI-assistant cited price/availability vs live PDP + JSON-LD + feed — alert on drift.

**Quoted ≠ checkout. / Visibility ≠ truth.**

English is the single source of truth for this marketing MVP canon. Do not invent scope beyond the locked IN / OUT tables.

This repository is a thin OSS Go core. It is **not a GEO dashboard** and **not a full PMS-style console**.

> Not a shopping-card API guarantee; dig-sourced engine behavior may change.

## Anchors

| Anchor | Meaning |
| --- | --- |
| Quoted ≠ checkout | A cited price/availability is not the live checkout offer. |
| Visibility ≠ truth | Being cited or visible in an assistant is not ground truth. |

## ICP

Shopify mid-market + agency white-label

## IN (day-1)

| In scope | Locked detail |
| --- | --- |
| Watch list | Top 20–50 SKU × engines 3–4: ChatGPT Shopping / Perplexity / Gemini / Claude |
| Cited fields (required) | `price` · `currency` · `availability` (`sale?` optional) |
| Live truth | PDP HTML · JSON-LD Offer · GMC/feed · Shopify `products.json` / Admin read |
| Batch + alerts | Daily/weekly batch; `\|Δprice\| > ε` or stock-flip → Slack/email |
| Agency report | White-label report, 1 kind |

## OUT

| Out of scope | Note |
| --- | --- |
| GEO / visibility dashboards | Not this product |
| Full catalog day-1 | Watch list is 20–50 SKU only |
| geo × currency matrix | Out |
| Auto-fix citations | Out |
| Merchant auto-fix | Out |
| Shopping-card API guarantee | Out — see disclaimer |
| Deadbugz / DRC coupling | Out |

## Package seats

| Package | Seat |
| --- | --- |
| [`catalog/`](catalog/) | 20–50 SKU watch list config |
| [`engines/`](engines/) | Cited-offer adapters (ChatGPT Shopping, Perplexity, Gemini, Claude) |
| [`truth/`](truth/) | Live PDP HTML, JSON-LD Offer, GMC/feed, Shopify `products.json` |
| [`parity/`](parity/) | Compare cited vs live; `\|Δprice\| > ε` and stock-flip |
| [`alert/`](alert/) | Slack + email notifiers |
| [`report/`](report/) | Agency white-label report (1 kind) from parity results; Shopify mid config/docs skin |
| [`cmd/otm/`](cmd/otm/) | CLI stub (prints seat names or help; no network in tests) |

Stubs and fixtures are intentional until the tracked GitHub issues land. Do not add OUT-scope packages.

## Develop

```bash
go test ./...
```

CLI (no network):

```bash
go run ./cmd/otm
go run ./cmd/otm help
go run ./cmd/otm seats
go run ./cmd/otm report
```

## Disclaimer

This project is **not a shopping-card API guarantee**. Dig-sourced engine behavior may change.
