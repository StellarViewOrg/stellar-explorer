---
title: Project Structure
description: Overview of the codebase organization.
---

## Top-Level Directories

``` 
stellar-explorer/
├── apps/
│   ├── explorer-web/ # Next.js explorer frontend
│   ├── docs/         # Astro/Starlight documentation site
│   └── tui/          # Stellar Explorer terminal interface
├── services/
│   ├── indexer/      # Stable Go data ingestion service
│   └── tui-indexer/  # Dedicated backend for terminal workflows
├── infra/
│   ├── docker/       # Docker configuration
│   └── docker-compose.yml
└── .github/          # CI workflows
```

## Frontend Source Code (`apps/explorer-web/src/`)

```
apps/explorer-web/src/
├── app/              # Next.js App Router pages
│   ├── [locale]/[network]/(explorer)/
│   └── api/          # API routes (TOML fetcher)
├── components/       # React components
│   ├── ui/           # shadcn/ui base components
│   ├── layout/       # Header, navigation, sidebars
│   ├── cards/        # Transaction, operation, contract cards
│   ├── charts/       # Recharts visualizations
│   ├── transactions/ # Transaction-specific components
│   ├── contracts/    # Soroban contract components
│   ├── assets/       # Asset browsing components
│   ├── search/       # Search UI
│   └── common/       # Shared components
├── lib/
│   ├── stellar/      # SDK clients and query definitions
│   ├── hooks/        # Custom React hooks
│   ├── providers/    # Context providers
│   ├── constants/    # App-wide constants
│   ├── utils/        # Utility functions
│   └── types/        # TypeScript types
└── i18n/             # Internationalization config
```

## Key Constants

| Constant | Value | Purpose |
|---|---|---|
| `STROOPS_PER_XLM` | `10,000,000` | Conversion factor for XLM amounts |
| `DEFAULT_PAGE_SIZE` | `20` | Items per paginated list |
| `LIVE_LEDGER_POLL_INTERVAL` | `5,000 ms` | Polling interval for live data |
| `STALE_TIME` | `10,000 ms` | Default TanStack Query stale time |

## UI Stack

- **Component library:** shadcn/ui (new-york style)
- **Styling:** Tailwind CSS 4
- **Charts:** Recharts
- **Path alias:** `@/` maps to `apps/explorer-web/src/`

## Terminal Product (`apps/tui/`)

`apps/tui` is the Go-based terminal interface for Stellar Explorer. It supports keyboard-driven lookup, live monitoring, related-entity traversal, local metadata, and source-aware reads through Stellar RPC or the dedicated TUI backend.

```
apps/tui/
├── cmd/tui/          # CLI entrypoint
├── internal/app/     # Application orchestration
├── internal/config/  # Local config/profile loading
├── internal/ui/      # Terminal view layer
└── internal/cache/   # Local SQLite/cache integration
```

### TUI Test Tiers

| Command | Scope | Prerequisites |
|---|---|---|
| `bun run tui:test` | Full TUI suite (unit + integration) | Go toolchain only |
| `bun run tui:test:unit` | Unit and fixture rendering tests only | Go toolchain only |
| `bun run tui:test:integration` | Reliability chains only (`integration` build tag) | Go toolchain only |
| `bun run tui-indexer:test` | Indexer read API handlers and store tests | Go toolchain; PostgreSQL for store integration tests |

CI workflows:

- `.github/workflows/tui-ci.yml` — `apps/tui` build, lint, unit, and integration tests
- `.github/workflows/tui-indexer-ci.yml` — `services/tui-indexer` build, lint, migrations, and tests

## TUI Backend (`services/tui-indexer/`)

`services/tui-indexer` prepares indexed Stellar Explorer data for terminal workflows. It provides read APIs, search, timelines, related records, and live feed data that enrich the terminal experience beyond direct RPC lookups.
