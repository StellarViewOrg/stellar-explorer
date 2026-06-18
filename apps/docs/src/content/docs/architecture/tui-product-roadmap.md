---
title: TUI Product Roadmap
description: Planned product direction for the Stellar Explorer terminal interface.
---

# TUI Product Roadmap

The Stellar Explorer TUI is planned as a professional terminal workspace for exploring Stellar activity. The product direction is focused on fast investigation, source-aware data access, Soroban clarity, live monitoring, and local working context.

This roadmap describes the user-facing capabilities planned for upcoming iterations. It is intentionally written as product direction, not as an audit checklist.

## Product Vision

The TUI should let a user start from any Stellar object and move through the surrounding network context without leaving the terminal. A typical workflow should feel like this:

1. Open the TUI with only Stellar RPC configured, or with the indexed TUI backend available.
2. Search for a ledger, transaction, account, asset, contract, label, bookmark, or note.
3. Inspect the entity in a compact detail view.
4. Jump into related transactions, operations, holders, events, contracts, accounts, or assets.
5. Save local context while investigating.
6. Return later and recover enough local state to continue the same line of work.

The current implementation already provides the foundation: interactive runtime, lookup, source-aware reads, hybrid fallback, local metadata, live feed browsing, timelines, contract specs, contract storage, and first decoded/raw Soroban views. The roadmap below describes how those surfaces should mature.

## Roadmap Themes

| Theme | Planned Outcome |
|---|---|
| Connected exploration | Entity views should feel connected, not like isolated lookups. |
| Soroban understanding | Contracts, events, specs, storage, and invocations should be readable from the terminal. |
| Live monitoring | Recent activity should evolve into a stream-native monitoring workflow. |
| Local workspace | Bookmarks, labels, notes, cached entities, and saved views should become primary workflows. |
| Search quality | Search should work well for partial input, local context, and indexed Stellar data. |
| Operational trust | Data source, fallback, cache state, and unavailable features should be visible and predictable. |

## Connected Exploration

The TUI should continue moving from point lookup toward graph-style Stellar exploration. A user should be able to inspect an entity, understand its immediate context, and open related entities with minimal keystrokes.

Planned improvements:

- Add stronger breadcrumbs that show how the user arrived at the current view.
- Preserve selection and scroll position when moving into and out of nested explorer lists.
- Deepen transaction details with effects and richer operation context.
- Add dedicated operation detail views instead of only compact operation rows.
- Make related-entity traversal consistent across ledgers, transactions, accounts, assets, and contracts.
- Add clearer context headers for list views such as account transactions, asset holders, and contract events.

Expected result:

- A user can move from a ledger to transactions, from a transaction to operations, from operations to accounts/assets/contracts, and back to the original context without restarting the workflow.

## Soroban Workspace

The TUI already surfaces contract metadata, recent events, spec summaries, storage summaries, and decoded/raw display modes. The next product step is to turn that foundation into a contract investigation workspace.

Planned improvements:

- Add invocation views that explain called functions, arguments, result status, fees, and related transaction context.
- Show authorization context where available, including the account or contract relationships involved in a Soroban operation.
- Expand contract method browsing beyond compact summaries.
- Add deeper event views with topic/value sections, decoded output, and raw fallback.
- Add storage drill-down with pagination, expiration context, durability, decoded display values, and raw XDR copy paths.
- Link contract views back to related accounts, assets, transactions, and events.

Expected result:

- A user can inspect a Stellar contract from the terminal with enough decoded context for daily investigation, while still having raw payload access for advanced debugging.

## Live Monitoring

The current live feed supports recent transactions, retained scrollback, pause/resume, and coarse filters for all, Soroban, and classic transactions. The roadmap is to make this a true monitoring view.

Planned improvements:

- Consume stream-oriented backend updates instead of relying only on refreshed summaries.
- Add replay controls over retained activity.
- Add filters by account, asset, contract, transaction class, operation type, and Soroban activity.
- Keep row selection stable while new activity arrives.
- Allow opening a live row and returning to the same monitoring context.
- Add profile-level watch settings for recurring monitoring sessions.

Expected result:

- A user can keep the TUI open during network observation, filter the activity stream, inspect an item, and return to monitoring without losing position.

## Local Workspace

The local SQLite layer already stores profiles, session state, bookmarks, labels, notes, live-feed scrollback, and visited lookup payloads. The next step is to make local context editable and useful inside the main experience.

Planned improvements:

- Add keyboard flows to create, edit, and remove bookmarks, labels, and notes.
- Show local metadata in entity headers, search results, and related lists.
- Add filters for bookmarked, labeled, noted, and recently visited entities.
- Add open-from-cache flows for previously inspected payloads.
- Add saved views for recurring investigations.
- Add profile-level workspace isolation so teams and users can separate contexts cleanly.

Expected result:

- The TUI becomes a persistent investigation workspace, not only a live lookup client.

## Search And Discovery

Search already combines entity inference, backend results, and local metadata. The next product step is making search more forgiving and more useful when users type partial or ambiguous input.

Planned improvements:

- Improve ranking for partial account, asset, contract, and transaction input.
- Group results by entity type and source.
- Add clearer summaries for indexed and local results.
- Support refinement or pagination for large result sets.
- Make saved local context searchable through attached Stellar entities.
- Keep disabled or incomplete results visibly distinct from executable rows.

Expected result:

- Search becomes the fastest entry point into the TUI, whether the user starts from a hash, address, label, note, asset code, or partial identifier.

## Release Direction

The next mature release should emphasize:

- connected entity navigation
- deeper transaction and operation detail
- Soroban invocation and authorization views
- stream-native monitoring
- editable local workspace flows
- reliable cache-aware revisit workflows
- stronger tests around hybrid fallback, local persistence, search, and decoded/raw rendering

The product should remain useful in RPC-only mode, but indexed mode should clearly provide richer exploration, better search, and more complete timelines.
