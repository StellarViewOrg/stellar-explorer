---
title: TUI Architecture
description: Architecture for the Stellar Explorer terminal interface and its dedicated data path.
---

## Overview

The Stellar Explorer TUI is a terminal interface for inspecting Stellar network data. It is designed for users who need fast, keyboard-driven access to ledgers, transactions, accounts, assets, contracts, and Soroban activity from a command-line environment.

The TUI is made of two coordinated parts:

- `apps/tui`: the Go terminal application, responsible for the user experience, local state, navigation, and direct Stellar RPC access.
- `services/tui-indexer`: the Stellar Explorer backend dedicated to terminal workflows, responsible for indexed reads, search, timeline slices, and live feed data.

This structure keeps the terminal product focused: the app stays responsive and local-first, while the backend prepares richer Stellar data for views that need more context than a single RPC lookup can provide.

## Current Status

The TUI is a functional alpha that covers the seven implementation tracks in the technical roadmap at a first-release maturity level. It can run as a direct RPC terminal client, and it can use `services/tui-indexer` for richer indexed reads in hybrid or indexed mode.

Implemented surfaces include:

- connected navigation with breadcrumbs, selection persistence, and explorer sublists
- transaction effects, operation detail views, and related-entity traversal
- Soroban invocation, event, and storage detail with decoded/raw display modes
- stream-native live feed with Redis ingestion, polling fallback, replay controls, filters, and watch presets
- editable local workspace flows for bookmarks, labels, notes, saved views, cache fallback, and profile isolation
- grouped search with local metadata filters, ranking, and backend pagination
- reliability coverage through unit, fixture, integration, and indexer read API test tiers

The product and technical roadmaps still describe additional polish and depth beyond this alpha. They should be read as direction, not as a guarantee that every future item is already complete.

## Known Limitations

- The TUI is not a stable release yet; command behavior and view composition may still change.
- RPC mode is intentionally narrower than indexed mode.
- Hybrid/indexed workflows require `services/tui-indexer` and its local infrastructure.
- Advanced live-feed filters for contract, asset, and operation type depend on optional indexed metadata fields.
- Redis stream ingestion requires `redis_url` on the active profile and a running `tui-indexer` live publisher.
- Saved views restore commands and screen context; they do not replay full UI selection state beyond stored filters.
- Some unavailable indexed sections degrade visibly instead of silently mixing RPC and indexed payloads.

## Design Principles

- **Terminal-first UX:** keyboard navigation, compact views, command search, and predictable shortcuts.
- **Stellar-native data model:** ledgers, transactions, operations, accounts, assets, contracts, trustlines, Soroban events, and Stellar RPC metadata are first-class concepts.
- **Local-first working context:** profiles, labels, bookmarks, notes, and session state are persisted locally so investigations can continue across sessions.
- **Source-aware reads:** the UI shows whether data came from Stellar RPC, the TUI backend, or an RPC fallback.
- **Progressive data depth:** the TUI starts with direct RPC and gains richer search, lists, timelines, and related entities when `services/tui-indexer` is available.
- **Single-source resolution:** each read operation resolves to one data source, which avoids mixing fields from different payloads into a single result.

## Runtime Model

`apps/tui` uses a Bubble Tea runtime with Lip Gloss rendering. The shell is organized into stable regions:

- header with profile, network, and data-source context
- sidebar navigation for primary screens
- main explorer area for live feed, lookup, and detail views
- status area for lifecycle and fallback information
- overlays for help and command search

Each view owns its local interaction state. Live feed, lookup, settings, home, and command palette models manage their own selection, scrolling, and empty/loading/error lifecycle. This keeps terminal behavior predictable as the product adds deeper entity views.

## Data Modes

### RPC Mode

RPC mode uses Stellar RPC directly. It supports quick startup and basic lookup workflows without running additional local services.

Typical use cases:

- inspect recent ledger activity
- look up a ledger, transaction, account, or contract
- verify account and contract metadata directly from Stellar RPC

### Indexer Mode

Indexer mode reads from `services/tui-indexer`. It is used when the terminal needs indexed data, related slices, entity lists, search results, holders, timelines, operations, events, and other explorer-oriented data.

Typical use cases:

- browse recent ledgers, accounts, assets, and contracts
- inspect account, asset, and contract timelines
- search across indexed Stellar entities
- open related transactions, operations, holders, and contract events

### Hybrid Mode

Hybrid mode prefers `services/tui-indexer` and falls back to Stellar RPC for operations that the backend cannot serve. The TUI surfaces the resolved source and fallback reason so users understand when a view is complete, indexed, or degraded.

## Local Persistence

The TUI stores user-facing state in SQLite:

- profiles
- bookmarks
- labels
- notes
- session state
- local metadata attachments

This makes the terminal useful as a persistent workspace, not just a transient lookup client. The local layer is also the foundation for saved views, revisiting inspected entities, and attaching operator context to Stellar objects.

## Backend Responsibilities

`services/tui-indexer` prepares data for terminal exploration:

- ingest Stellar ledgers, transactions, operations, assets, accounts, contracts, token events, and contract events
- expose read APIs for lookup, lists, sublists, and timelines
- provide search over indexed Stellar entities
- publish live feed data for terminal monitoring
- normalize related slices so the TUI can move from one entity to another quickly

## User Benefits

- faster investigation from the command line
- fewer manual lookup loops
- clear visibility into active data source and fallback state
- useful default mode through direct Stellar RPC
- richer exploration when indexed Stellar Explorer data is available
- persistent local context for recurring analysis

## Maturity Path

The architecture is intentionally incremental:

1. Keep direct RPC mode useful for basic exploration.
2. Expand indexed reads for entity lists, timelines, and related data.
3. Deepen Soroban decoding for invocations, events, specs, and storage.
4. Add stream-native live monitoring.
5. Expand local cache and saved investigation workflows.

For the planned product direction, see the [TUI Product Roadmap](./tui-product-roadmap/). For implementation tracks, see the [TUI Technical Roadmap](./tui-technical-roadmap/).
