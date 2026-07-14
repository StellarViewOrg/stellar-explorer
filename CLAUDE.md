# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
bun run dev          # Start dev server (port 3000)
bun run build        # Production build
bun run lint         # ESLint
bun run format       # Prettier format
bun run format:check # Prettier check (used in CI)
bun run test         # Vitest run
bun run test:watch   # Vitest watch mode
```

Run a single test file: `bunx vitest run src/lib/utils/format.test.ts`

Tests use happy-dom environment. Test files live alongside source as `*.test.{ts,tsx}`.

## Architecture

StellarView Explorer is a Next.js 16 App Router application that provides a blockchain explorer for the Stellar network. It uses Bun as its package manager.

### Routing: `src/app/[locale]/[network]/(explorer)/...`

Every page is nested under `[locale]/[network]`, making URLs like `/en/public/tx/abc123` or `/es/testnet/accounts`. The `(explorer)` route group organizes the main pages (transactions, ledgers, accounts, assets, contracts, search, watchlist, analytics, learn).

Three networks are supported: `public`, `testnet`, `futurenet` — configured in `src/lib/constants/networks.ts`.

Nine locales (en, es, pt, fr, de, zh, ja, ko, ru) via next-intl. Translation files are in `messages/*.json`. **CI validates that all translation files have identical key counts**, so when adding UI text, update all 9 files.

### Data Flow

- **Stellar SDK clients** (`src/lib/stellar/client.ts`): Factory functions `getHorizonClient()` / `getRpcClient()` return cached instances per network. Horizon is used for most data; RPC is used for Soroban contract data (events, code, storage).
- **Query definitions** (`src/lib/stellar/queries.ts`): `stellarKeys` factory produces hierarchical cache keys. `stellarQueries` contains TanStack Query option factories. Immutable entities (ledgers, transactions, contract code) use `staleTime: Infinity`.
- **Custom hooks** (`src/lib/hooks/`): `use-stellar-query.ts` wraps TanStack Query. `use-streaming.ts` provides real-time Horizon streaming for ledgers/transactions.

### Providers (wrap order matters)

`src/lib/providers/index.tsx`: ThemeProvider → QueryProvider → NetworkProvider → DeveloperModeProvider → AnalyticsModeProvider

Context hooks exported: `useNetwork`, `useTheme`, `useDeveloperMode`, `useAnalyticsMode`.

### UI

- shadcn/ui components in `src/components/ui/` (new-york style, Tailwind CSS 4)
- Domain components organized by: `layout/`, `common/`, `cards/`, `charts/`, `transactions/`, `search/`, `contracts/`, `assets/`, `stats/`, `glossary/`
- Charts use recharts
- Path alias: `@/` → `src/`

### Key Constants (`src/lib/constants/index.ts`)

- `STROOPS_PER_XLM = 10_000_000`
- `DEFAULT_PAGE_SIZE = 20`
- `LIVE_LEDGER_POLL_INTERVAL = 5000` (ms)
- `STALE_TIME = 10_000` (ms)

### API Route

`src/app/api/toml/route.ts` — Stellar TOML fetcher with SSRF protection and rate limiting (30 req/min).

### Prettier Config

Double quotes, semicolons, 100 char print width, trailing commas (es5), Tailwind CSS plugin for class sorting.

## Git

- Never add `Co-Authored-By` to commit messages.
