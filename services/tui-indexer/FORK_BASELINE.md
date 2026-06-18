# TUI Indexer Maintenance Notes

`services/tui-indexer` is the Stellar Explorer data service dedicated to terminal workflows.

## Intent

- provide indexed Stellar data for `apps/tui`
- support terminal-specific read models, search, timelines, and live feed behavior
- keep operational defaults separate from other Stellar Explorer services where the terminal workflow benefits from isolation

## Current Scope

- ingest Stellar ledgers, transactions, operations, accounts, assets, contracts, token events, and contract events
- expose read APIs that are shaped for terminal exploration
- provide search over indexed Stellar Explorer entities
- publish terminal-oriented live feed data

## Shared Implementation Areas

- migrations
- source adapters
- base ingestion pipelines
- low-level storage models

## Terminal-Specific Implementation Areas

- semantic decode
- read APIs
- search strategy
- Redis channel model
- operational defaults for DB and infra
