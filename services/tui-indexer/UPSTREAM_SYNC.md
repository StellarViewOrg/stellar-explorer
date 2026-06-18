# Source Alignment Policy

`services/tui-indexer` should stay aligned with useful Stellar Explorer ingestion and parsing improvements while preserving terminal-specific read models, search behavior, and operating defaults.

## Rules

1. Review ingestion, parser, and data model improvements intentionally.
2. Apply only changes that remain valid for terminal workflows.
3. Re-run tests and integration checks after every alignment batch.
4. Keep TUI-specific read APIs, search behavior, live feed channels, and operational defaults explicit.

## Suggested Workflow

1. Compare the relevant Stellar Explorer service code.
2. Apply the changes that improve `services/tui-indexer`.
3. Re-run tests and integration checks.
4. Document major behavior changes in this file or in commit messages.
