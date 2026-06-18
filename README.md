# Stellar Explorer

> A premium block explorer for the Stellar network — built for developers, traders, and ecosystem builders.

[![Live](https://img.shields.io/badge/status-live-brightgreen)](#) [![Next.js](https://img.shields.io/badge/Next.js-16-black)](#) [![Go](https://img.shields.io/badge/Indexer-Go-00ADD8)](#)

---

## What is Stellar Explorer?

Stellar Explorer gives you a fast, clean window into the Stellar blockchain. Browse ledgers, transactions, accounts, assets, and Soroban smart contracts across **Public**, **Testnet**, and **Futurenet** — all in one place, with a UI that doesn't get in your way.

The app is **live today**, powered entirely by the Stellar Horizon API and Soroban RPC. We're also building a custom **indexer** that will bring historical depth, richer analytics, and a broader data panorama that the Horizon API alone can't provide.

**TUI** brings Stellar Explorer into the terminal as a functional alpha client, offering a keyboard-driven way to investigate Stellar activity, monitor live data, keep local notes/bookmarks/labels, and move through indexed network context with the dedicated `services/tui-indexer` backend.

---

## Live App

The current version is fully operational and covers:

| Feature                  | Details                                            |
| ------------------------ | -------------------------------------------------- |
| 🔴 **Real-time data**    | Live ledger and transaction streaming from Horizon |
| 🌐 **Multi-network**     | Public mainnet, Testnet, and Futurenet             |
| 🪙 **Asset discovery**   | Token metadata and logos via `stellar.toml`        |
| 📜 **Smart contracts**   | Browse Soroban contract events, code, and storage  |
| 📌 **Watchlist**         | Track accounts and assets across sessions          |
| 🌗 **Dark / Light mode** | Optimized for both themes                          |
| 🌍 **9 languages**       | EN, ES, PT, FR, DE, ZH, JA, KO, RU                 |

---

## The Indexer — Expanding Our Data Horizon

The Horizon API is great for recent data, but it has limits: no deep historical coverage, no custom aggregations, and no way to serve analytics at scale. That's where the indexer comes in.

The **Stellar Explorer Indexer** is a Go service that ingests Stellar network data — ledgers, transactions, and operations — into **PostgreSQL + TimescaleDB**, with real-time event publishing via **Redis**. Once fully deployed, it will power:

- ⏳ **Full historical coverage** — from ledger 3 to present (60M+ ledgers on pubnet)
- 📊 **Richer analytics** — charts and aggregations that live APIs can't support
- 🔍 **Full-text search** — powered by Typesense
- ⚡ **Real-time event streams** — Redis pub/sub for live updates

### Architecture

```
Stellar RPC ──► source/rpc.go ──────► transform/ ──► store/postgres.go ──► PostgreSQL
                (JSON-RPC 2.0)         (XDR parsing)   (batch inserts)
                                           ▲                   │
AWS S3 ────► source/datalake.go ──────────┘                   ▼
              (public data lake)                    publisher/redis.go ──► Redis pub/sub
```

### Three Indexer Modes

The indexer ships with three operating modes designed to cover every scenario — from bootstrapping to production:

---

#### 🟢 `live` — Real-time Ingestion

Continuously polls the Stellar RPC for new ledgers and ingests them as they close (~1 ledger every 5 seconds). Designed to run indefinitely in production.

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet make run-live
```

Shuts down gracefully on `Ctrl+C` and resumes from the last ingested ledger on restart — no duplicates, no gaps.

---

#### 🔵 `backfill` — Historical Backfill via RPC

Processes a specific range of ledgers in parallel. Works on any network (pubnet, testnet, futurenet). Ideal for catching up after a gap or indexing a targeted range.

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet \
  ./bin/indexer backfill --start 1288000 --end 1288100
```

Controlled by `WORKER_COUNT` (default: 8) for parallel throughput.

---

#### 🟠 `s3backfill` — Mass Historical Backfill from AWS Data Lake

The fastest way to index the entire Stellar pubnet history. Reads directly from the [Stellar public AWS data lake](https://github.com/stellar/stellar-etl) — no RPC endpoint and no AWS credentials required. Covers ledger 3 through the latest pubnet ledger (60M+).

```bash
# Index a million ledgers
./bin/indexer s3backfill --start 3 --end 1000000

# Resume from a checkpoint
./bin/indexer s3backfill --start 500001 --end 1000000

# Crank up workers for maximum throughput
WORKER_COUNT=16 ./bin/indexer s3backfill --start 3 --end 5000000
```

> **Note:** `s3backfill` is pubnet-only. For testnet/futurenet historical data, use `backfill` with an RPC endpoint.

---

## Tech Stack

**Frontend**

- [Next.js 16](https://nextjs.org/) (App Router) + React 19
- [TanStack Query](https://tanstack.com/query) for data fetching and caching
- [Tailwind CSS 4](https://tailwindcss.com/) + [shadcn/ui](https://ui.shadcn.com/)
- [Stellar SDK](https://stellar.github.io/js-stellar-sdk/)
- [Bun](https://bun.sh/) as package manager and runtime

**Indexer**

- [Go 1.24](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/) + [TimescaleDB](https://www.timescale.com/) for time-series data
- [Redis](https://redis.io/) for real-time pub/sub
- [Typesense](https://typesense.org/) for full-text search

---

## Monorepo Structure

```text
stellar-explorer/
├── apps/
│   ├── explorer-web/ # Next.js explorer frontend
│   ├── docs/         # Astro/Starlight documentation site
│   └── tui/          # Go terminal client for Stellar investigation
├── services/
│   ├── indexer/      # Stable Go ingestion/indexing service
│   └── tui-indexer/  # Dedicated backend for terminal workflows
├── infra/
│   ├── docker/      # Local infrastructure files
│   └── docker-compose.yml
├── .github/         # CI workflows
├── package.json     # Bun workspace root
└── bun.lock         # Shared dependency lockfile
```

## Getting Started

### Frontend

```bash
bun install           # Install workspace dependencies
bun run dev:web       # Start frontend at http://localhost:3000
bun run build:web     # Production build for apps/explorer-web
```

### TUI

```bash
bun run tui:build     # Build the terminal client
bun run tui:run       # Run the terminal client
bun run tui:test      # Run TUI module tests
```

### Indexer

Requires Docker Compose to be running:

```bash
# From project root — starts PostgreSQL, Redis, and Typesense
docker compose -f infra/docker-compose.yml up -d

# Apply database migrations
for f in services/indexer/migrations/*.up.sql; do
  cat "$f" | docker compose -f infra/docker-compose.yml exec -T postgres psql -U explorer -d stellar_explorer
done

# Build and run the indexer
make -C services/indexer build
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet make -C services/indexer run-live
```

See [`services/indexer/README.md`](./services/indexer/README.md) for the full configuration reference and all available commands.

### TUI Indexer

```bash
bun run tui-indexer:build
bun run tui-indexer:test
bun run tui-indexer:migrate
```

See [`services/tui-indexer/README.md`](./services/tui-indexer/README.md) for the dedicated backend that enriches `apps/tui` with indexed reads, search, timelines, and live-feed data.

### Docs

```bash
bun run dev:docs
bun run build:docs
```

---

## Scripts

| Command          | Description               |
| ---------------- | ------------------------- |
| `bun run dev:web` | Start frontend development server |
| `bun run build:web` | Build the frontend |
| `bun run tui:build` | Build the Go terminal client |
| `bun run tui:run` | Run the Go terminal client |
| `bun run tui:test` | Run TUI tests |
| `bun run dev:docs` | Start the documentation site |
| `bun run build:docs` | Build the documentation site |
| `bun run indexer:build` | Build the stable Go indexer |
| `bun run indexer:test` | Run stable indexer tests |
| `bun run tui-indexer:build` | Build the dedicated TUI backend |
| `bun run tui-indexer:test` | Run TUI backend tests |
| `bun run tui-indexer:run:live` | Run the TUI backend in live mode |
| `bun run tui-indexer:infra:up` | Start isolated Docker services for `tui-indexer` |

## Vercel

For the frontend deployment, point the Vercel project to `apps/explorer-web` as the Root Directory. That isolates the production build from the rest of the monorepo and avoids accidental breakage when `services/` or `apps/docs/` change.

Recommended Vercel settings:

- Framework Preset: `Next.js`
- Root Directory: `apps/explorer-web`
- Install Command: leave default so Vercel detects the Bun workspace from repo root
- Build Command: leave default, or set `bun run build`
- Output Directory: leave empty

If you later deploy docs on a separate Vercel project, use `apps/docs` as its Root Directory.

---

## License

MIT
