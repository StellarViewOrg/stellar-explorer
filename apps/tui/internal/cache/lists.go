package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ListCache stores one paginated explorer list payload.
type ListCache struct {
	ProfileID   string
	ListKind    string
	QueryKey    string
	Payload     string
	SourceLabel string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertListCache stores one paginated list payload.
func (s *Store) UpsertListCache(ctx context.Context, entry ListCache) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if entry.ProfileID == "" || entry.ListKind == "" || entry.QueryKey == "" || entry.Payload == "" {
		return errors.New("cache: list cache profile, kind, query key, and payload are required")
	}

	now := time.Now().UTC()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, `
INSERT INTO list_cache (
    profile_id, list_kind, query_key, payload, source_label, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(profile_id, list_kind, query_key) DO UPDATE SET
    payload = excluded.payload,
    source_label = excluded.source_label,
    updated_at = excluded.updated_at;`,
		entry.ProfileID,
		entry.ListKind,
		entry.QueryKey,
		entry.Payload,
		entry.SourceLabel,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert list %s:%s: %w", entry.ListKind, entry.QueryKey, err)
	}
	return nil
}

// GetListCache returns one cached list payload when present.
func (s *Store) GetListCache(ctx context.Context, profileID, listKind, queryKey string) (*ListCache, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}
	if profileID == "" || listKind == "" || queryKey == "" {
		return nil, errors.New("cache: list cache profile, kind, and query key are required")
	}

	var entry ListCache
	err := s.db.QueryRowContext(ctx, `
SELECT profile_id, list_kind, query_key, payload, source_label, created_at, updated_at
FROM list_cache
WHERE profile_id = ? AND list_kind = ? AND query_key = ?
LIMIT 1;`, profileID, listKind, queryKey).Scan(
		&entry.ProfileID,
		&entry.ListKind,
		&entry.QueryKey,
		&entry.Payload,
		&entry.SourceLabel,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRows()
		}
		return nil, fmt.Errorf("cache: get list %s:%s: %w", listKind, queryKey, err)
	}
	return &entry, nil
}
