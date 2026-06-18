package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// SavedView stores a reusable command and screen context for one profile.
type SavedView struct {
	ID           string
	ProfileID    string
	Name         string
	Command      string
	Screen       string
	EntityKind   string
	EntityTarget string
	FiltersJSON  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UpsertSavedView stores or updates a named saved view.
func (s *Store) UpsertSavedView(ctx context.Context, view SavedView) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if view.ProfileID == "" {
		return errors.New("cache: saved view profile ID is required")
	}
	if view.Name == "" {
		return errors.New("cache: saved view name is required")
	}
	if view.Command == "" {
		return errors.New("cache: saved view command is required")
	}

	now := time.Now().UTC()
	if view.CreatedAt.IsZero() {
		view.CreatedAt = now
	}
	view.UpdatedAt = now
	if view.FiltersJSON == "" {
		view.FiltersJSON = "{}"
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO saved_views (
    id, profile_id, name, command, screen, entity_kind, entity_target, filters_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(profile_id, name) DO UPDATE SET
    command = excluded.command,
    screen = excluded.screen,
    entity_kind = excluded.entity_kind,
    entity_target = excluded.entity_target,
    filters_json = excluded.filters_json,
    updated_at = excluded.updated_at;`,
		view.ID,
		view.ProfileID,
		view.Name,
		view.Command,
		view.Screen,
		view.EntityKind,
		view.EntityTarget,
		view.FiltersJSON,
		view.CreatedAt,
		view.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert saved view %q: %w", view.Name, err)
	}
	return nil
}

// ListSavedViews returns saved views for one profile ordered by recent updates.
func (s *Store) ListSavedViews(ctx context.Context, profileID string) ([]SavedView, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, profile_id, name, command, screen, entity_kind, entity_target, filters_json, created_at, updated_at
FROM saved_views
WHERE profile_id = ?
ORDER BY updated_at DESC;`, profileID)
	if err != nil {
		return nil, fmt.Errorf("cache: list saved views: %w", err)
	}
	defer rows.Close()

	views := make([]SavedView, 0)
	for rows.Next() {
		var view SavedView
		if err := rows.Scan(
			&view.ID,
			&view.ProfileID,
			&view.Name,
			&view.Command,
			&view.Screen,
			&view.EntityKind,
			&view.EntityTarget,
			&view.FiltersJSON,
			&view.CreatedAt,
			&view.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan saved view: %w", err)
		}
		views = append(views, view)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate saved views: %w", err)
	}
	return views, nil
}

// DeleteSavedView removes a saved view by profile and name.
func (s *Store) DeleteSavedView(ctx context.Context, profileID string, name string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if profileID == "" || name == "" {
		return errors.New("cache: saved view profile ID and name are required")
	}

	if _, err := s.db.ExecContext(ctx, `DELETE FROM saved_views WHERE profile_id = ? AND name = ?;`, profileID, name); err != nil {
		return fmt.Errorf("cache: delete saved view %q: %w", name, err)
	}
	return nil
}
