---
title: Architecture Overview
description: High-level architecture of Stellar Explorer.
---

:::caution[Construction in Progress]
Currently, the production version of Stellar Explorer relies on the **Horizon API** for data retrieval. The custom Indexer architecture described in this section is under active development and will replace direct Horizon calls in future updates to enable advanced features.
:::

Stellar Explorer is a Next.js 16 application that reads data from the Stellar network through two main interfaces: the **Indexer API** (for optimized historical and real-time data) and **Soroban RPC** (for smart contract interactions).

![Architecture Overview](../../../assets/diagrams/architecture-overview.svg)

## Components

### Frontend

The frontend is a React 19 application built with the Next.js App Router. It uses TanStack Query for data fetching and caching, shadcn/ui for the component library, and Tailwind CSS 4 for styling.

All pages are server-side rendered on first load and then hydrated for client-side navigation.

### Backend Services (Indexer)

The Go-based indexer service processes Stellar ledger data into PostgreSQL (with TimescaleDB), Redis, and Typesense. This architecture enables advanced queries, full-text search, and analytics that are not possible with Horizon alone.

See the [Indexer Pipeline](./indexer/) for detailed implementation.

Stellar Explorer also includes a terminal interface in `apps/tui` and a dedicated data service in `services/tui-indexer`. Together they provide keyboard-driven Stellar lookup, live monitoring, search, timelines, and source-aware reads for terminal workflows. See [TUI Architecture](./tui/).

### Stellar Network

The system interacts with the Stellar network at two levels:

- **Stellar Node** — Ingested by the Indexer for high-performance data serving.
- **Soroban RPC** — Queried directly by the frontend for smart contract data (code, storage, events).

## Key Design Decisions

- **Indexer-first architecture** — Optimized for speed and complex relational queries.
- **Network-agnostic** — All data fetching includes the network parameter, enabling seamless switching.
- **Cached SDK clients** — Indexer and RPC clients are created once per network and reused.
- **Immutable data caching** — Ledgers, transactions, and contract code use `staleTime: Infinity` since they never change.
