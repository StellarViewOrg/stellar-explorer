# stellar-tui

**stellar-tui** is a terminal block explorer for the Stellar network. It is designed for fast, keyboard-driven investigation: live monitoring, entity lookup, related-entity traversal, source-aware reads, and local operator context without leaving the command line.

The terminal interface is useful when the browser is too slow for repeated inspection, when you are already working in a shell, or when a workflow benefits from local notes, labels, bookmarks, cached activity, and predictable keyboard navigation.

## Current Status

**v0.1.0** is the first standalone release foundation. The client runs in RPC mode without extra local services, persists workspace state in SQLite by default, and can use the dedicated `services/tui-indexer` backend for richer indexed reads when that backend is available.

Current implemented surfaces include:

- Bubble Tea interactive runtime with a Lip Gloss shell
- RPC, indexed, and hybrid data modes
- source-aware lookup with fallback/degraded state labels
- live feed browsing with Redis stream ingestion, polling fallback, replay controls, and local scrollback
- lookup and navigation for core Stellar entities
- indexed lists, sublists, timelines, holders, operations, and events when `services/tui-indexer` is available
- local SQLite profiles, session state, labels, bookmarks, notes, saved views, visited entity cache, and cache fallback reads
- in-client Soroban decode in RPC/network mode (contract spec, invoke args/results, events, instance storage)
- first decoded/raw Soroban display paths for contract-heavy views

The roadmap documents the intended product direction. Treat features listed only in the roadmap as planned work, not release guarantees.

## Known Limitations

- The TUI is not a stable release yet; command behavior and view composition may still change.
- Hybrid mode requires `services/tui-indexer` plus its local infrastructure for indexed reads.
- RPC mode remains intentionally narrower than indexed mode and cannot provide every list, timeline, holder, or search result.
- RPC contract events are limited to the RPC retention window (~24h); persistent contract data scans still require `tui-indexer`.
- Advanced live-feed filters for contract, asset, and operation type require indexed/stream metadata from `tui-indexer`; account and class filters work in all live-feed modes.
- Redis stream ingestion requires `redis_url` on the active profile and a running `tui-indexer` live publisher.
- Saved views restore commands and screen context; they do not replay full UI selection state beyond stored filters.
- RPC mode composes Horizon (history, assets, operations) with Stellar RPC (contract state, Soroban meta). Indexed mode still provides the richest lists, search, and live-stream metadata.

## Installation

### Prerequisites

- Go 1.25+
- A Stellar RPC endpoint

### Quick start (environment only)

```bash
STELLAR_RPC_URL=https://soroban-testnet.stellar.org stellar-tui
```

On first run, stellar-tui creates `~/.config/stellar-tui/config.json` and `~/.config/stellar-tui/cache.db`.

### Install from source (monorepo)

```bash
bun run tui:install
# binary lands in $(go env GOPATH)/bin/stellar-tui
```

### Install from source (this directory)

```bash
make install
```

### Build a local binary

```bash
make build
./bin/stellar-tui
```

### Release binaries

Tagged releases are published as `stellar-tui-v*` GitHub release assets. Example:

```bash
# after downloading and extracting stellar-tui_<version>_darwin_arm64.tar.gz
./stellar-tui -version
```

## Configuration

stellar-tui resolves config in this order:

1. `-config <path>`
2. `STELLAR_TUI_CONFIG`
3. `./config.json` in the current working directory
4. `~/.config/stellar-tui/config.json`

The same search order applies to `labels.toml` (`-labels`, `STELLAR_TUI_LABELS`, cwd, then user config dir).

### Quick setup

```bash
mkdir -p ~/.config/stellar-tui
cp config/config.example.json ~/.config/stellar-tui/config.json
cp labels.example.toml ~/.config/stellar-tui/labels.toml
```

### Environment variables

| Variable | Description |
| --- | --- |
| `STELLAR_RPC_URL` | Override the active profile RPC endpoint |
| `STELLAR_NETWORK` | Override network (`public`, `testnet`, `futurenet`) |
| `STELLAR_BACKEND_MODE` | `rpc`, `hybrid`, or `indexer` |
| `STELLAR_HORIZON_URL` | Horizon endpoint (auto-resolved per network when unset) |
| `STELLAR_INDEXER_URL` | TUI indexer read API URL |
| `STELLAR_REDIS_URL` | Redis URL for stream-native live feed |
| `STELLAR_TUI_CONFIG` | Explicit config file path |
| `STELLAR_TUI_PROFILE` | Active profile name |
| `STELLAR_TUI_LABELS` | Explicit labels file path |

Example:

```bash
STELLAR_RPC_URL=https://soroban-testnet.stellar.org \
STELLAR_NETWORK=testnet \
./bin/stellar-tui
```

## Cache-First Lookups

stellar-tui serves visited entities from local SQLite before calling the network:

1. **Fresh cache hit** — instant revisit with `cache:hit` in the header
2. **Network fetch** — backend read, then payload is stored locally
3. **Stale fallback** — if the backend fails, an older cached payload is shown with `cache:stale`

TTL policy:

| Entity kind | Cache TTL |
| --- | --- |
| ledger, transaction | permanent (immutable) |
| account | 5 minutes |
| asset, contract | 10 minutes |

Press `r` on the lookup screen to force a network refresh and bypass cache.

Use `open cache` or `open recent` to revisit stored payloads explicitly.

## Capability Matrix

Each cell describes what is available today. `-` means unavailable in that mode.

| Data | RPC mode (Horizon + RPC) | + `tui-indexer` |
| --- | --- | --- |
| Ledger lookup | yes | yes |
| Transaction lookup (ops/effects) | yes | yes |
| Account lookup (state + trustlines) | yes | yes |
| Contract instance lookup | yes | yes |
| Contract spec decode | yes (WASM via RPC) | yes |
| Soroban op decode (invoke args/results) | yes | yes |
| Account operations / timeline | yes | yes |
| Asset lookup / holders | yes | yes |
| Contract events (decoded) | yes (RPC window) | yes (indexed) |
| Contract storage (instance map) | yes | yes |
| Contract persistent storage scan | - | yes |
| Identifier search (tx/account/contract/asset/ledger) | yes (Horizon inference) | yes (indexed) |
| Indexed full-text search | - | yes |
| Live feed (poll) | yes | yes |
| Live feed (stream) | - | yes (Redis) |
| Live feed filters (contract/asset/operation) | - | yes (stream + poll) |
| Local labels / bookmarks / notes | yes | yes |
| Cache-first entity revisit | yes | yes |
| Visited entity cache fallback | yes | yes |

## Why It Matters

Stellar data often needs to be inspected as a graph of related entities: a ledger contains transactions, transactions contain operations, accounts connect to assets, assets connect to issuers and holders, and contracts connect to events, storage, invocations, and transactions. The TUI is built for that kind of movement.

Instead of copying identifiers between pages or repeatedly opening new views, the terminal keeps the workflow compact:

- search from anywhere with `/`
- move with `j` / `k` or arrow keys
- open the selected entity with `enter`
- go back and forward with `b` / `f`
- copy the active identifier with `y`
- paste a Stellar identifier into search with `ctrl+v`

## Benefits

- **More efficient investigation:** keyboard navigation, command search, and related-entity jumps reduce repetitive lookup work.
- **Local-first context:** profiles, session state, labels, bookmarks, notes, visited entities, and live-feed scrollback are stored locally in SQLite.
- **More private workflows:** local metadata stays on the machine. Notes, labels, bookmarks, and cached investigation context do not need to be sent to a hosted UI.
- **Clear data source visibility:** each lookup can show whether it came from Stellar RPC, `services/tui-indexer`, or an RPC fallback.
- **Useful with minimal setup:** RPC mode can run directly against Stellar RPC for basic live feed and lookup workflows.
- **Richer with indexed data:** hybrid/indexed mode adds search, timelines, holders, operations, full contract event history, persistent storage scans, and related entity lists from `services/tui-indexer`.
- **Terminal-native ergonomics:** the UI is optimized for repeated inspection, compact output, clipboard use, and shell-driven workflows.

## Data Modes

### RPC Mode

RPC mode queries Stellar RPC directly. It is the default local mode and works without running the TUI backend.

Use it for:

- quick lookup of ledgers, transactions, accounts, and contracts
- basic live feed workflows
- low-friction terminal exploration on testnet

### Hybrid Mode

Hybrid mode prefers `services/tui-indexer` when it can provide richer indexed data, and falls back to Stellar RPC when an indexed path is unavailable.

Use it for:

- indexed search across Stellar entities
- ledger, account, asset, and contract lists
- account, asset, and contract timelines
- asset holders and contract events
- richer contract inspection with spec and storage summaries

## Quick Commands

Run the default RPC profile:

```bash
bun run tui:run
# or: ./bin/stellar-tui
```

Equivalent explicit RPC command:

```bash
bun run tui:run:rpc
```

Render once and exit, useful for scripts or quick checks:

```bash
bun run tui:run:once
```

Run with the hybrid testnet profile:

```bash
bun run tui:run:hybrid
```

Render the hybrid profile once and exit:

```bash
bun run tui:run:hybrid:once
```

Build and test:

```bash
bun run tui:build
bun run tui:test
```

Quick smoke check (render once, no keyboard):

```bash
bun run tui:run:once
```

Optional split tiers:

```bash
bun run tui:test:unit          # unit + fixture suite only
bun run tui:test:integration   # integration reliability chains only
```

Indexer backend (only when changing `services/tui-indexer`):

```bash
bun run tui-indexer:test       # local-safe suite (no Docker required)
bun run tui-indexer:test:all   # full suite (requires Postgres on :54320)
```

### Test Tiers

- `bun run tui:test` — one command for the terminal app. Runs unit + integration tests. No Stellar RPC, PostgreSQL, or Redis required.
- `bun run tui:test:unit` — faster subset when iterating on a single package.
- `bun run tui:test:integration` — integration build-tag tests only.
- `bun run tui-indexer:test` — indexer unit and read API tests; skips DB/Redis/S3 integration checks unless infra is running.
- `bun run tui-indexer:test:all` — full indexer suite including integration tests (start infra with `bun run tui-indexer:infra:up` first).

## Hybrid Backend Setup

Hybrid mode expects `services/tui-indexer` at `http://127.0.0.1:8081`, as configured in `apps/tui/config/hybrid.testnet.json`.

Start local infrastructure:

```bash
bun run tui-indexer:infra:up
```

Apply migrations:

```bash
bun run tui-indexer:migrate
```

Start the TUI read API:

```bash
bun run tui-indexer:run:serve
```

In another terminal, run the TUI in hybrid mode:

```bash
bun run tui:run:hybrid
```

Optional: ingest live Stellar data into the TUI backend:

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet bun run tui-indexer:run:live
```

## What You Can Inspect

The current TUI supports first-class lookup and navigation for:

- ledgers
- transactions
- accounts
- assets
- contracts
- operations
- asset holders
- contract events
- account, asset, and contract timelines
- contract specs and storage summaries

The UI also exposes source metadata, degraded/fallback states, local metadata attachments, clipboard workflows, and decoded/raw contract display modes where available.

## Keyboard Flows

Global:

- `h` home
- `l` live feed
- `u` lookup
- `s` settings
- `b` back (top-level screen or lookup route step)
- `f` forward (top-level screen or lookup route step)
- `r` refresh
- `/` open search or command palette
- `tab` cycle focus
- `?` toggle help
- `y` copy the current entity identifier to the clipboard
- `q` quit

Explorer navigation:

- `Route:` line shows lookup breadcrumbs (entity detail and explorer sublists)
- `b` / `f` walk the lookup route without losing list selection when returning
- `j` / `k` or arrow keys move selection
- `pgup` / `pgdn` and `home` / `end` move through longer views
- `enter` follows the selected row command (lookup, open list, or related entity)
- in live feed, `left/right` switches the value that `y` will copy: `hash`, `ledger`, or `account`
- in live feed, `p` pauses or resumes updates
- in live feed, `t` cycles between all, Soroban, and classic transaction filters
- in live feed, `[` and `]` replay retained scrollback while paused
- in live feed, `enter` opens the selected transaction; `b` or `live return` restores monitoring position, filters, and pause state
- in live feed, use `live filter account <id>`, `live filter contract <id>`, `live filter asset <code:issuer>`, or `live filter operation <type>` for advanced filters
- in live feed, `watch save <name>` stores profile watch presets (filters and pause state); `watch open <name>` restores them

Search and command palette:

- type a Stellar lookup target or command directly
- use `open ledgers`, `open accounts`, `open assets`, or `open contracts` to browse recent indexed entities
- use `open ledgers limit 20 before <sequence>` or the `Next Page` row to page backward through ledger history
- use `open txs`, `open ops`, `open op <n>`, `open holders`, `open events`, `open storage`, or `open invocations` from compatible lookup details
- use `open event <n>`, `open storage <n>`, or `open invocation <n>` to drill into Soroban event, storage, and invocation detail views
- use `open decode raw|decoded` on contract, event, and storage detail views to switch Soroban-heavy sections between decoded and raw payloads
- use `lookup op <txhash>:<index>` to open a dedicated operation detail view from account activity or the command palette
- transaction lookups show indexed effects when the active backend is indexer or hybrid; RPC-only backends show a degradation notice instead
- use `open timeline` from account, asset, or contract details to inspect normalized activity
- use `open timeline type tx|op|holder|event` for category-specific timeline views supported by the current entity
- use `open decode raw` or `open decode decoded` from contract views to switch Soroban-heavy sections between raw and decoded display
- add `limit <n> offset <n>` to paged sublists, or use the `Next Page` row when it is shown
- `ctrl+v` pastes the system clipboard into the palette
- `enter` executes the highlighted result or typed command
- `esc` closes the palette

## Local Workspace Commands

Workspace commands operate on the active lookup entity unless they open a local browser. They are stored in the local SQLite cache for the active profile.

Bookmarks:

- `bookmark add [title]` saves the active entity.
- `bookmark remove` removes bookmarks attached to the active entity.
- `bookmark note <text>` updates the annotation on the active entity bookmark.
- `open bookmarks` or `open bookmarked` browses saved bookmarks for the active profile.

Notes:

- `note add [title] [| body]` creates a note on the active entity.
- `note remove [title-filter]` removes notes from the active entity, optionally matching the title.
- `note body [title-filter |] <text>` updates the body of the most recent matching note.
- `open notes` or `open noted` browses saved notes for the active profile.

Labels:

- `label add <name>` applies a label to the active entity, creating the label if needed.
- `label remove <name>` detaches a label from the active entity.
- `label delete <name>` deletes a label definition and its attachments for the active profile.
- `label color <name> <color>` stores a display color for the label.
- `open labels` browses labels and their attached entities.
- `open labeled [name]` browses entities with labels, optionally filtered by label name.

Recent local cache:

- `open recent` browses recently visited cached entities.
- `open cache` reloads the cached payload for the active lookup (or `open cache <kind> <target>`).
- failed backend lookups automatically fall back to a recent cached payload when one exists.

Saved views:

- `view save <name>` stores the current screen, lookup target, and live-feed filter context.
- `view open <name>` restores a saved view.
- `view delete <name>` removes a saved view.
- `open views` browses saved views for the active profile.

Watch settings:

- `watch save <name>` stores the current live-feed filter and pause state for the active profile.
- `watch open <name>` restores a saved watch preset and switches to the live feed.
- `watch delete <name>` removes a watch preset.
- `watch auto <name>` saves a preset and marks it auto-apply on live-feed entry.
- `open watches` browses saved watch presets for the active profile.
- profiles may set `default_watch` in config to apply a named preset on first live-feed entry.

Lookup workspace shortcuts:

- `m` opens the bookmark command palette.
- `,` quick-adds a bookmark for the active entity.
- `-` removes bookmarks from the active entity.
- `;` opens the note command palette.
- `.` opens the label command palette.

Search refinements:

- command palette results are grouped by source and kind.
- prefix filters like `label:`, `bookmark:`, `note:`, `cache:`, `view:`, and `watch:` narrow local metadata search.
- command palette groups results by source (`INDEXER`, `HORIZON`, `LOCAL`) and kind, with ranked exact/prefix matches first.
- `search more <query> <offset>` loads additional indexed results when the backend returns a full page.
- `search more <query> <offset>` loads additional backend search results when the palette shows a load-more row.
