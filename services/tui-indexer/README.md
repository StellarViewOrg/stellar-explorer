# Stellar Explorer TUI Indexer

`services/tui-indexer` is the Stellar Explorer backend dedicated to the terminal interface in `apps/tui`.

It prepares Stellar network data for terminal workflows: ingestion, semantic normalization, read APIs, search, timelines, and live feed data. The service is optimized for views that need more context than a single Stellar RPC lookup can provide.

## Role In Stellar Explorer

- `apps/tui` provides the user-facing terminal experience.
- `services/tui-indexer` provides indexed Stellar Explorer data for terminal views.
- Local SQLite in `apps/tui` keeps user state such as profiles, labels, notes, bookmarks, and session context.

The TUI can run directly against Stellar RPC. When `services/tui-indexer` is available, the terminal gains richer entity lists, timelines, related records, search results, holders, operations, events, and live feed data.

## Runtime Isolation Defaults

The TUI backend uses dedicated local defaults:

- PostgreSQL database: `stellar_explorer_tui` on local port `54330`
- Redis URL: `redis://localhost:63890`
- Typesense URL: `http://localhost:18118`
- Redis channels: `tui-indexer:stream:ledgers`, `tui-indexer:stream:transactions`

For local infrastructure, use the overlay at [`infra/docker-compose.tui-indexer.yml`](../../infra/docker-compose.tui-indexer.yml).

The service ingests Stellar network data into PostgreSQL and can publish real-time events through Redis. It supports Stellar RPC for live and range ingestion, plus the Stellar public data lake for public-network historical backfill.

## Prerequisites

- Go 1.24+ (managed via asdf, see `.tool-versions`)
- Docker Compose services running:

```bash
# from project root
docker compose -f infra/docker-compose.tui-indexer.yml up -d
```

This starts PostgreSQL+TimescaleDB (port 54330), Redis (port 63890), and Typesense (port 18118).

- Database migrations applied:

```bash
# from services/tui-indexer/
make build
./bin/tui-indexer migrate
```

## Configuration

| Variable       | Default                                                                               | Required | Description                                               |
| -------------- | ------------------------------------------------------------------------------------- | -------- | --------------------------------------------------------- |
| `RPC_ENDPOINT` | —                                                                                     | **Yes**  | Stellar RPC endpoint                                      |
| `NETWORK`      | `public`                                                                              | No       | `public`, `testnet`, or `futurenet`                       |
| `DATABASE_URL` | `postgresql://explorer:explorer_dev@localhost:54330/stellar_explorer_tui?sslmode=disable` | No   | PostgreSQL connection                                     |
| `REDIS_URL`    | `redis://localhost:63890`                                                             | No       | Redis connection (optional — logs warning if unavailable) |
| `SEARCH_BACKEND` | `postgres`                                                                          | No       | `postgres` for direct SQL search, `typesense` once dedicated search-index thresholds are met |
| `TYPESENSE_URL` | `http://localhost:18118`                                                            | No       | Dedicated search service URL, used when `SEARCH_BACKEND=typesense` |
| `TYPESENSE_KEY` | `explorer_dev_key`                                                                  | No       | Dedicated search service API key                          |
| `READ_API_ADDR`| `:8081`                                                                               | No       | HTTP listen address for the read API server               |
| `BATCH_SIZE`   | `100`                                                                                 | No       | Ledgers per batch                                         |
| `WORKER_COUNT` | `8`                                                                                   | No       | Parallel workers for backfill                             |

### Search Backend Policy

`SEARCH_BACKEND=postgres` is the default and remains the right choice while TUI search is primarily exact or prefix lookup over indexed entity tables. Promote to `SEARCH_BACKEND=typesense` when fuzzy/ranked discovery or saved-investigation search becomes required, or when observed Postgres search reaches any of these thresholds:

- at least 10 million indexed searchable entities
- p95 search latency at or above 150 ms
- sustained search traffic at or above 300 queries per minute

## Commands

```bash
make build          # Compile to bin/tui-indexer
make migrate        # Apply pending database migrations
make test           # Run all tests
make fmt            # Format code
make lint           # Run go vet
make run-serve      # Start the read API server
make clean          # Remove bin/
```

### Read API

Start the HTTP read server:

```bash
make run-serve
```

Available endpoints in the first cut:

- `GET /healthz`
- `GET /v1/feed/live/summary`
- `GET /v1/search`
- `GET /v1/ledgers?limit=20&before=<sequence>`
- `GET /v1/ledgers/:sequence`
- `GET /v1/ledgers/:sequence/transactions?limit=20&offset=0`
- `GET /v1/accounts`
- `GET /v1/accounts/:id`
- `GET /v1/accounts/:id/transactions`
- `GET /v1/accounts/:id/operations?limit=20&offset=0`
- `GET /v1/accounts/:id/timeline?limit=20&offset=0&type=tx|op`
- `GET /v1/assets`
- `GET /v1/assets/:code::issuer`
- `GET /v1/assets/:code::issuer/transactions`
- `GET /v1/assets/:code::issuer/holders?limit=20&offset=0`
- `GET /v1/assets/:code::issuer/timeline?limit=20&offset=0&type=tx|holder`
- `GET /v1/contracts`
- `GET /v1/contracts/:id`
- `GET /v1/contracts/:id/transactions`
- `GET /v1/contracts/:id/events?limit=20&offset=0`
- `GET /v1/contracts/:id/timeline?limit=20&offset=0&type=tx|event`
- `GET /v1/transactions/:hash`

### Live ingestion

Continuously polls the RPC for new ledgers and ingests them in real-time (~1 ledger every 5 seconds).

> **Important:** `NETWORK` must match your RPC endpoint. It determines the network passphrase used to compute transaction hashes. Defaults to `public` (mainnet). Set `NETWORK=testnet` for testnet or `NETWORK=futurenet` for futurenet.

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet make run-live
```

Or directly:

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet ./bin/tui-indexer live
```

Stop with `Ctrl+C` — the indexer shuts down gracefully and resumes from the last ingested ledger on restart.

### Historical backfill (RPC)

Processes a range of ledgers using parallel workers. Works with any network (pubnet, testnet, futurenet):

```bash
RPC_ENDPOINT=https://soroban-testnet.stellar.org NETWORK=testnet ./bin/tui-indexer backfill --start 1288000 --end 1288100
```

### S3 data lake backfill (pubnet only)

Backfills historical pubnet data directly from the [Stellar AWS public data lake](https://github.com/stellar/stellar-etl) -- no RPC endpoint or AWS credentials needed. The data lake covers ledger 3 through the latest pubnet ledger (~61.5M+).

```bash
./bin/tui-indexer s3backfill --start 3 --end 1000000
```

Key details:

- **Pubnet only** -- for testnet/futurenet historical data, use the `backfill` command with an RPC endpoint instead.
- **No `RPC_ENDPOINT` required** -- data is read from a public S3 bucket using anonymous access.
- **No AWS credentials required** -- the bucket is publicly accessible.
- **Resume support** -- if interrupted, re-run with `--start` set to the last successfully ingested ledger.
- **Worker count** -- controlled by the `WORKER_COUNT` env var (default `8`).

```bash
# Resume from where you left off
./bin/tui-indexer s3backfill --start 500001 --end 1000000

# Use more workers for faster throughput
WORKER_COUNT=16 ./bin/tui-indexer s3backfill --start 3 --end 5000000
```

## Migrations

Migrations live in `services/tui-indexer/migrations/` and are embedded in the binary at build time.

### Running migrations

```bash
make migrate
```

This applies all pending migrations in order. Running it again when already up to date is safe — it exits cleanly with no changes.

### Creating a new migration

Install the `migrate` CLI if you don't have it:

```bash
brew install golang-migrate
```

Then run:

```bash
migrate create -ext sql -dir migrations -seq your_description
```

This generates two files in `services/tui-indexer/migrations/`:

```
000014_your_description.up.sql    # forward change (CREATE TABLE, ALTER TABLE, etc.)
000014_your_description.down.sql  # rollback (DROP TABLE IF EXISTS ... CASCADE)
```

The version is zero-padded to 6 digits by the CLI (`000014`, `000015`, ...).

Fill in both files, then apply:

```bash
make migrate
```

## Resetting the database

To wipe all ingested data and start fresh (useful after testing with different networks or ledger ranges):

```bash
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui -c "
  TRUNCATE ledgers, transactions, operations, ingestion_state CASCADE;
"
```

To reset only the ingestion cursor (keeps existing data but allows re-ingestion):

```bash
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui -c "
  DELETE FROM ingestion_state;
"
```

## Validating it works

After running the indexer for a few seconds, check that data landed in PostgreSQL:

```bash
# Ledgers
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui \
  -c "SELECT sequence, transaction_count, operation_count, protocol_version FROM ledgers ORDER BY sequence DESC LIMIT 5;"

# Transactions
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui \
  -c "SELECT hash, ledger_sequence, account, operation_count, is_soroban FROM transactions ORDER BY created_at DESC LIMIT 5;"

# Operations
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui \
  -c "SELECT transaction_hash, type_name, destination, amount FROM operations ORDER BY created_at DESC LIMIT 5;"

# Ingestion cursor
docker compose -f infra/docker-compose.tui-indexer.yml exec postgres-tui psql -U explorer -d stellar_explorer_tui \
  -c "SELECT * FROM ingestion_state;"
```

To verify Redis pub/sub, subscribe in one terminal:

```bash
docker compose -f infra/docker-compose.tui-indexer.yml exec redis-tui redis-cli SUBSCRIBE tui-indexer:stream:ledgers
```

Then run the indexer in another terminal — you should see JSON messages as ledgers are ingested.

## Testing

Tests run against a live Stellar testnet RPC and local Docker services:

```bash
make test
```

Run a specific package:

```bash
go test ./internal/source/ -v      # RPC client (hits testnet)
go test ./internal/transform/ -v   # XDR parsing (hits testnet)
go test ./internal/store/ -v       # PostgreSQL (requires Docker)
go test ./internal/publisher/ -v   # Redis pub/sub (requires Docker)
go test ./internal/pipeline/ -v    # End-to-end (requires both)
```

Tests skip gracefully if Docker services are unavailable. Override connection strings with `TEST_DATABASE_URL` and `TEST_RPC_ENDPOINT`.

## Architecture

```
Stellar RPC ──> source/rpc.go ──────> transform/ ──> store/postgres.go ──> PostgreSQL
                (JSON-RPC 2.0)        (XDR parsing)   (batch inserts)
                                          ▲                   │
AWS S3 ─────> source/datalake.go ─────────┘                   ▼
              (public data lake)                       publisher/redis.go ──> Redis pub/sub
                                                       (stream:ledgers,
                                                        stream:transactions)
```

| Package              | Purpose                                                                 |
| -------------------- | ----------------------------------------------------------------------- |
| `cmd/indexer`        | Entry point with `live`, `backfill`, `migrate` commands                 |
| `internal/config`    | Environment variable loading and validation                             |
| `internal/source`    | Stellar RPC client (`getLedgers`, `getTransactions`, `getLatestLedger`) |
| `internal/transform` | XDR parsing into database models (ledgers, transactions, operations)    |
| `internal/store`     | PostgreSQL writer with batch inserts and ingestion cursor               |
| `internal/pipeline`  | Live ingestion loop and parallel backfill orchestration                 |
| `internal/publisher` | Redis pub/sub for real-time event streaming                             |
