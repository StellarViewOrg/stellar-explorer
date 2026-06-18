package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// UpsertProfile persists a local profile record.
func (s *Store) UpsertProfile(ctx context.Context, profile Profile) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if profile.ID == "" {
		return errors.New("cache: profile ID is required")
	}

	if profile.Name == "" {
		return errors.New("cache: profile name is required")
	}

	now := time.Now().UTC()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO profiles (id, name, network, rpc_url, indexer_url, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   network = excluded.network,
		   rpc_url = excluded.rpc_url,
		   indexer_url = excluded.indexer_url,
		   updated_at = excluded.updated_at;`,
		profile.ID,
		profile.Name,
		profile.Network,
		profile.RPCURL,
		profile.IndexerURL,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert profile %q: %w", profile.ID, err)
	}

	return nil
}

// ListProfiles returns all locally persisted profiles ordered by name.
func (s *Store) ListProfiles(ctx context.Context) ([]Profile, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name, network, rpc_url, indexer_url, created_at, updated_at
		 FROM profiles
		 ORDER BY name ASC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("cache: list profiles: %w", err)
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var profile Profile
		if err := rows.Scan(
			&profile.ID,
			&profile.Name,
			&profile.Network,
			&profile.RPCURL,
			&profile.IndexerURL,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan profile: %w", err)
		}

		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate profiles: %w", err)
	}

	return profiles, nil
}
