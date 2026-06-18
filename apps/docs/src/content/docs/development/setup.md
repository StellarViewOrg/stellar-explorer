---
title: Development Setup
description: How to set up a local development environment.
---

## Prerequisites

- [Bun](https://bun.sh/) (package manager and runtime)
- [Docker](https://www.docker.com/) (for backend services, optional)
- [D2](https://d2lang.com/) (for diagram rendering, optional)

## Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/salazarsebas/stellar-explorer.git
cd stellar-explorer
```

2. **Install dependencies**

```bash
bun install
```

3. **Start the development server**

```bash
bun run dev
```

The explorer will be available at `http://localhost:3000`.

## Terminal Product

Stellar Explorer includes a Go-based terminal interface for keyboard-driven Stellar lookup and monitoring:

```bash
bun run tui:build
bun run tui:run
bun run tui:test
```

For hybrid mode with indexed reads:

```bash
bun run tui-indexer:infra:up
bun run tui-indexer:migrate
bun run tui-indexer:run:serve
bun run tui:run:hybrid
```

## Backend Services (Optional)

For indexer functionality, start the Docker services:

```bash
docker compose -f infra/docker-compose.yml up -d
```

This starts PostgreSQL (port 54320), Redis (port 63790), and Typesense (port 18108).

For the dedicated TUI backend infrastructure, use:

```bash
bun run tui-indexer:infra:up
```

## Environment Variables

Copy the example file and adjust if needed:

```bash
cp .env.local.example .env.local
```

## Available Commands

| Command | Purpose |
|---|---|
| `bun run dev` | Start development server (port 3000) |
| `bun run build` | Production build |
| `bun run lint` | Run ESLint |
| `bun run format` | Format code with Prettier |
| `bun run test` | Run tests |
| `bun run test:watch` | Run tests in watch mode |
| `bun run tui:build` | Build the TUI |
| `bun run tui:test` | Run full TUI test suite (unit + integration) |
| `bun run tui:test:unit` | Run fast TUI unit and fixture tests only |
| `bun run tui:test:integration` | Run TUI integration reliability tests only |
| `bun run tui-indexer:build` | Build the TUI backend |
| `bun run tui-indexer:test` | Run local-safe TUI indexer tests (no Docker required) |
| `bun run tui-indexer:test:all` | Run full TUI indexer suite (requires local Postgres) |
| `bun run tui-indexer:infra:up` | Start local infrastructure for `tui-indexer` |
