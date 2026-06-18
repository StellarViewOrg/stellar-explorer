package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// UpsertEntityCache stores the latest payload observed for one entity lookup.
func (s *Store) UpsertEntityCache(ctx context.Context, entity EntityCache) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if entity.ProfileID == "" {
		return errors.New("cache: entity profile ID is required")
	}
	if entity.Kind == "" {
		return errors.New("cache: entity kind is required")
	}
	if entity.Target == "" {
		return errors.New("cache: entity target is required")
	}
	if entity.Payload == "" {
		return errors.New("cache: entity payload is required")
	}

	now := time.Now().UTC()
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = now
	}
	entity.UpdatedAt = now

	if err := s.PruneEntityCache(ctx, entity.ProfileID, DefaultEntityCacheLimit); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO entity_cache (
    profile_id, kind, target, title, summary, payload, source_label, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(profile_id, kind, target) DO UPDATE SET
    title = excluded.title,
    summary = excluded.summary,
    payload = excluded.payload,
    source_label = excluded.source_label,
    updated_at = excluded.updated_at;`,
		entity.ProfileID,
		entity.Kind,
		entity.Target,
		entity.Title,
		entity.Summary,
		entity.Payload,
		entity.SourceLabel,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert entity %s:%s: %w", entity.Kind, entity.Target, err)
	}

	return nil
}

// ListEntityCache returns recently visited entities for one profile.
func (s *Store) ListEntityCache(ctx context.Context, profileID string, limit int) ([]EntityCache, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT profile_id, kind, target, title, summary, payload, source_label, created_at, updated_at
FROM entity_cache
WHERE profile_id = ?
ORDER BY updated_at DESC
LIMIT ?;`, profileID, limit)
	if err != nil {
		return nil, fmt.Errorf("cache: list entity cache: %w", err)
	}
	defer rows.Close()

	entities := make([]EntityCache, 0, limit)
	for rows.Next() {
		var entity EntityCache
		if err := rows.Scan(
			&entity.ProfileID,
			&entity.Kind,
			&entity.Target,
			&entity.Title,
			&entity.Summary,
			&entity.Payload,
			&entity.SourceLabel,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan entity cache: %w", err)
		}
		entities = append(entities, entity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate entity cache: %w", err)
	}

	return entities, nil
}

// GetEntityCache returns one cached entity payload when present.
func (s *Store) GetEntityCache(ctx context.Context, profileID string, kind string, target string) (*EntityCache, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}
	if profileID == "" || kind == "" || target == "" {
		return nil, errors.New("cache: entity profile ID, kind, and target are required")
	}

	var entity EntityCache
	err := s.db.QueryRowContext(ctx, `
SELECT profile_id, kind, target, title, summary, payload, source_label, created_at, updated_at
FROM entity_cache
WHERE profile_id = ? AND kind = ? AND target = ?
LIMIT 1;`, profileID, kind, target).Scan(
		&entity.ProfileID,
		&entity.Kind,
		&entity.Target,
		&entity.Title,
		&entity.Summary,
		&entity.Payload,
		&entity.SourceLabel,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRows()
		}
		return nil, fmt.Errorf("cache: get entity %s:%s: %w", kind, target, err)
	}
	return &entity, nil
}

// PruneEntityCache keeps only the most recent entity payloads for one profile.
func (s *Store) PruneEntityCache(ctx context.Context, profileID string, maxEntries int) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if profileID == "" {
		return errors.New("cache: entity profile ID is required")
	}
	if maxEntries <= 0 {
		return nil
	}

	_, err := s.db.ExecContext(ctx, `
DELETE FROM entity_cache
WHERE profile_id = ?
  AND rowid NOT IN (
    SELECT rowid
    FROM entity_cache
    WHERE profile_id = ?
    ORDER BY updated_at DESC
    LIMIT ?
  );`, profileID, profileID, maxEntries)
	if err != nil {
		return fmt.Errorf("cache: prune entity cache: %w", err)
	}
	return nil
}
