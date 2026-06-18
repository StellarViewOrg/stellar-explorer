---
title: TUI Technical Roadmap
description: Planned technical implementation tracks for the Stellar Explorer terminal interface.
---

# TUI Technical Roadmap

This roadmap describes the planned technical work behind the Stellar Explorer TUI. It is organized by implementation track so future changes can be planned without turning the documentation into a status audit.

The current foundation includes:

- Bubble Tea runtime and Lip Gloss shell
- RPC, indexed, and hybrid data modes
- source metadata and fallback labels
- lookup flows for ledgers, transactions, accounts, assets, and contracts
- indexed lists, sublists, and timelines
- local SQLite profiles, metadata, scrollback, and visited entity cache
- first decoded/raw Soroban views for events, specs, and storage

## Track 1: Navigation And View Composition

Purpose:

- make terminal exploration feel continuous as entity views get deeper

Planned implementation:

- Add breadcrumb state to the app model.
- Preserve list selection and scroll offsets across nested explorer transitions.
- Add context headers for explorer sublists.
- Normalize related-entity actions so every row exposes a predictable command, label, and clipboard value.
- Split more complex detail views into reusable view sections while keeping keyboard ownership local to each view model.

Technical considerations:

- Breadcrumbs should be derived from navigation actions, not manually duplicated in render code.
- View-owned state should remain scoped to Bubble Tea models to avoid coupling render details to root application state.
- Existing `LookupExplorerSnapshot` can remain the bridge for list-style subviews, but deeper detail views may need a richer route descriptor.

## Track 2: Transaction And Operation Depth

Purpose:

- turn transaction lookup into a complete investigation surface

Planned implementation:

- Add effects to transaction detail when indexed data is available.
- Add operation detail views with type-specific fields.
- Link operation rows to source accounts, destinations, assets, contracts, and parent transactions.
- Improve operation summaries for Soroban and asset-related activity.
- Add tests for transaction-to-operation traversal and operation command selection.

Technical considerations:

- Effects should be loaded through indexed read APIs where possible.
- Direct RPC mode should remain clear about unavailable indexed-only sections.
- Operation renderers should prefer typed fields over string-only summaries.

## Track 3: Soroban Inspection

Purpose:

- make contracts and Soroban activity understandable from the terminal

Planned implementation:

- Add invocation detail models.
- Surface function name, arguments, result status, resources, and related transaction context.
- Add authorization views when the source payload can support it.
- Expand event detail beyond compact summaries into topic/value sections.
- Add storage pagination and drill-down commands.
- Keep decoded and raw modes available for every Soroban-heavy section.

Technical considerations:

- Decoded views should never hide raw fallback values.
- Storage and event pagination should use the existing limit/offset command pattern.
- Contract spec parsing should remain tolerant of partially decoded or unavailable specs.

## Track 4: Live Monitoring

Purpose:

- evolve live feed from refreshed recent activity into stream-native monitoring

Planned implementation:

- Add a stream consumer for `services/tui-indexer` live channels.
- Keep the existing polling path as fallback.
- Add replay controls over retained scrollback.
- Add filters by account, asset, contract, operation type, transaction class, and Soroban activity.
- Preserve selected row and viewport position while new rows arrive.
- Add tests for stream update ordering, deduplication, pause behavior, and selection stability.

Technical considerations:

- The live feed should keep a bounded scrollback to avoid unbounded memory growth.
- Stream ingestion should dedupe by transaction hash and preserve ledger/application ordering.
- Pause should stop UI mutation from live updates without dropping retained events needed for replay.

## Track 5: Local Workspace And Cache

Purpose:

- make local state useful for real investigation workflows

Planned implementation:

- Add create/edit/delete flows for bookmarks, labels, and notes.
- Add local metadata filters in search and lookup views.
- Add open-from-cache for recently inspected entities.
- Add cache fallback when backend/RPC reads are unavailable and a recent local payload exists.
- Add saved views that preserve command, filters, entity, and profile context.
- Add migration tests for every SQLite schema change.

Technical considerations:

- Cached entity payloads should include source label, updated timestamp, summary, and enough typed metadata to render a degraded view.
- Cache fallback must be explicit in source metadata so users know they are seeing local data.
- Saved views should reference commands and filters rather than storing duplicated UI state where possible.

## Track 6: Search And Ranking

Purpose:

- make search the fastest path into both Stellar data and local context

Planned implementation:

- Improve local and backend result ranking for partial input.
- Group search results by type and source in the command palette.
- Add result pagination or refinement for large indexed result sets.
- Make label, bookmark, note, and cached-entity results consistently executable.
- Add tests for ranking, deduplication, disabled rows, and metadata attachment.

Technical considerations:

- Ranking should prefer exact target matches, then prefix matches, then metadata matches.
- Local results should remain available when the backend is unavailable.
- Backend result limits should be explicit so the command palette remains responsive.

## Track 7: Reliability And Release Readiness

Purpose:

- ensure the terminal product remains predictable as more read paths and local state are added

Planned implementation:

- Expand tests around hybrid fallback and single-source resolution.
- Add integration tests for timelines, list pagination, and Soroban decoded/raw rendering.
- Add migration coverage for local SQLite schemas.
- Add read API tests for new effects, invocation, storage, and event endpoints.
- Add fixture-based tests for rendering degraded states.

Technical considerations:

- Tests that require Stellar RPC, public data lake access, or local PostgreSQL should be clearly separated from pure unit tests.
- Network-backed tests should skip or document their prerequisites when local infrastructure is unavailable.
- Documentation should stay aligned with implemented behavior after each milestone.

## Track 8: Documentation And Release Alignment

Purpose:

- keep architecture docs, setup guides, and CI workflows aligned with the implemented alpha after milestone delivery

Planned implementation:

- Update architecture and setup docs to reflect current TUI capabilities and limitations.
- Document test tiers and infrastructure prerequisites for `apps/tui` and `services/tui-indexer`.
- Add CI workflows for terminal build, lint, unit, and integration coverage.
- Extend read API tests for live feed, search, and paginated Soroban subresources.

Technical considerations:

- Documentation should describe capability direction without turning roadmap pages into release audits.
- CI should run fast, deterministic tiers by default and keep network-backed checks optional or documented.
- English and Spanish docs should stay in sync for architecture and development setup pages.

## Implementation Order

The recommended order is:

1. Navigation and view composition.
2. Transaction and operation depth.
3. Soroban invocation and authorization views.
4. Stream-native live monitoring.
5. Editable local workspace and open-from-cache.
6. Search ranking and result grouping.
7. Expanded reliability coverage.
8. Documentation and release alignment.

This order keeps the TUI useful at every step while steadily increasing the depth expected from a professional Stellar terminal explorer.
