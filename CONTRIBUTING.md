# Contributing

English README canon is the single source of truth for scope. Keep the locked package seats. Do not open PRs for OUT scope.

## Keep these seats

| Seat | Path |
| --- | --- |
| Watch list (20–50 SKU) | `catalog/` |
| Cited-offer adapters | `engines/` |
| Live truth extractors | `truth/` |
| Cited vs live compare | `parity/` |
| Slack + email | `alert/` |
| Agency white-label report | `report/` |
| CLI | `cmd/otm/` |

Add implementation inside an existing seat. Do not invent sibling product surfaces.

## Do not open PRs for OUT scope

OUT of the locked MVP (see README):

- GEO / visibility dashboards
- Full catalog day-1
- geo × currency matrix
- Auto-fix citations
- Merchant auto-fix
- Shopping-card API guarantee
- Deadbugz / DRC coupling
- Full PMS-style console

PRs that add those surfaces will be closed.

## Develop

```bash
go test ./...
```

Tests must not require live network. Engine and truth seats may use fixtures; unknown required fields fail closed.

## Issues

Work is tracked under the `[Epic] Offer Truth Monitor MVP` parent and its child issues. Reference the matching child in the PR body.
