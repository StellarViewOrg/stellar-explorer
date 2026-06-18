package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const schemaVersion = 8

var errNoRows = errors.New("cache: no rows returned")

// ErrNoRows exposes the sentinel used when state is not yet present.
func ErrNoRows() error {
	return errNoRows
}

// Store owns the local persistence layer for TUI state and profiles.
type Store struct {
	db *sql.DB
}

// Open prepares an existing database handle for TUI usage and applies schema migrations.
func Open(ctx context.Context, db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("cache: database handle is required")
	}

	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		return nil, fmt.Errorf("cache: migrate schema: %w", err)
	}

	return store, nil
}

// OpenSQLite opens a SQLite-style DSN using a pre-registered database/sql driver.
func OpenSQLite(ctx context.Context, driverName, dsn string) (*Store, error) {
	if driverName == "" {
		return nil, errors.New("cache: driver name is required")
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("cache: open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("cache: ping database: %w", err)
	}

	store, err := Open(ctx, db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

// Close releases the underlying database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}

// DB exposes the underlying SQL handle for future package integrations.
func (s *Store) DB() *sql.DB {
	if s == nil {
		return nil
	}

	return s.db
}

// SchemaVersion returns the latest applied local schema version.
func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("cache: store is not initialized")
	}

	version, err := currentSchemaVersion(ctx, s.db)
	if err != nil {
		return 0, fmt.Errorf("cache: load schema version: %w", err)
	}

	return version, nil
}
