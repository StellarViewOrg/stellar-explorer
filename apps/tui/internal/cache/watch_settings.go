package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// WatchSetting stores a reusable live-feed monitoring preset for one profile.
type WatchSetting struct {
	ID          string
	ProfileID   string
	Name        string
	FiltersJSON string
	Paused      bool
	AutoApply   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpsertWatchSetting stores or updates a named watch preset.
func (s *Store) UpsertWatchSetting(ctx context.Context, setting WatchSetting) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if setting.ProfileID == "" {
		return errors.New("cache: watch setting profile ID is required")
	}
	if setting.Name == "" {
		return errors.New("cache: watch setting name is required")
	}

	now := time.Now().UTC()
	if setting.CreatedAt.IsZero() {
		setting.CreatedAt = now
	}
	setting.UpdatedAt = now
	if setting.FiltersJSON == "" {
		setting.FiltersJSON = "{}"
	}

	paused := 0
	if setting.Paused {
		paused = 1
	}
	autoApply := 0
	if setting.AutoApply {
		autoApply = 1
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO watch_settings (
    id, profile_id, name, filters_json, paused, auto_apply, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(profile_id, name) DO UPDATE SET
    filters_json = excluded.filters_json,
    paused = excluded.paused,
    auto_apply = excluded.auto_apply,
    updated_at = excluded.updated_at;`,
		setting.ID,
		setting.ProfileID,
		setting.Name,
		setting.FiltersJSON,
		paused,
		autoApply,
		setting.CreatedAt,
		setting.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert watch setting %q: %w", setting.Name, err)
	}
	return nil
}

// ListWatchSettings returns watch presets for one profile ordered by recent updates.
func (s *Store) ListWatchSettings(ctx context.Context, profileID string) ([]WatchSetting, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, profile_id, name, filters_json, paused, auto_apply, created_at, updated_at
FROM watch_settings
WHERE profile_id = ?
ORDER BY updated_at DESC;`, profileID)
	if err != nil {
		return nil, fmt.Errorf("cache: list watch settings: %w", err)
	}
	defer rows.Close()

	settings := make([]WatchSetting, 0)
	for rows.Next() {
		var setting WatchSetting
		var paused int
		var autoApply int
		if err := rows.Scan(
			&setting.ID,
			&setting.ProfileID,
			&setting.Name,
			&setting.FiltersJSON,
			&paused,
			&autoApply,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan watch setting: %w", err)
		}
		setting.Paused = paused != 0
		setting.AutoApply = autoApply != 0
		settings = append(settings, setting)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate watch settings: %w", err)
	}
	return settings, nil
}

// DeleteWatchSetting removes a watch preset by profile and name.
func (s *Store) DeleteWatchSetting(ctx context.Context, profileID string, name string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if profileID == "" || name == "" {
		return errors.New("cache: watch setting profile ID and name are required")
	}

	if _, err := s.db.ExecContext(ctx, `DELETE FROM watch_settings WHERE profile_id = ? AND name = ?;`, profileID, name); err != nil {
		return fmt.Errorf("cache: delete watch setting %q: %w", name, err)
	}
	return nil
}

// FindAutoApplyWatchSetting returns the most recently updated auto-apply preset for a profile.
func (s *Store) FindAutoApplyWatchSetting(ctx context.Context, profileID string) (*WatchSetting, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	var setting WatchSetting
	var paused int
	var autoApply int
	err := s.db.QueryRowContext(ctx, `
SELECT id, profile_id, name, filters_json, paused, auto_apply, created_at, updated_at
FROM watch_settings
WHERE profile_id = ? AND auto_apply = 1
ORDER BY updated_at DESC
LIMIT 1;`, profileID).Scan(
		&setting.ID,
		&setting.ProfileID,
		&setting.Name,
		&setting.FiltersJSON,
		&paused,
		&autoApply,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w for auto-apply watch setting", errNoRows)
		}
		return nil, fmt.Errorf("cache: find auto-apply watch setting: %w", err)
	}
	setting.Paused = paused != 0
	setting.AutoApply = autoApply != 0
	return &setting, nil
}
