package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SetStateJSON stores arbitrary JSON-serializable state under a key.
func (s *Store) SetStateJSON(ctx context.Context, key string, value any) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if key == "" {
		return errors.New("cache: state key is required")
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal state %q: %w", key, err)
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO app_state (key, value, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at;`,
		key,
		string(payload),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("cache: upsert state %q: %w", key, err)
	}

	return nil
}

// GetStateJSON loads state by key and decodes it into dest.
func (s *Store) GetStateJSON(ctx context.Context, key string, dest any) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if key == "" {
		return errors.New("cache: state key is required")
	}

	var payload string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?;`, key).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w for key %q", errNoRows, key)
		}

		return fmt.Errorf("cache: load state %q: %w", key, err)
	}

	if err := json.Unmarshal([]byte(payload), dest); err != nil {
		return fmt.Errorf("cache: decode state %q: %w", key, err)
	}

	return nil
}
