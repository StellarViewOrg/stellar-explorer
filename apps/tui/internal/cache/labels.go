package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// UpsertLabel persists a label definition keyed by label ID.
func (s *Store) UpsertLabel(ctx context.Context, label Label) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if label.ID == "" {
		return errors.New("cache: label ID is required")
	}

	if label.ProfileID == "" {
		return errors.New("cache: label profile ID is required")
	}

	if label.Name == "" {
		return errors.New("cache: label name is required")
	}

	now := time.Now().UTC()
	if label.CreatedAt.IsZero() {
		label.CreatedAt = now
	}
	label.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO labels (id, profile_id, name, color, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   profile_id = excluded.profile_id,
		   name = excluded.name,
		   color = excluded.color,
		   updated_at = excluded.updated_at;`,
		label.ID,
		label.ProfileID,
		label.Name,
		label.Color,
		label.CreatedAt,
		label.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert label %q: %w", label.ID, err)
	}

	return nil
}

// ListLabels returns all labels ordered by name.
func (s *Store) ListLabels(ctx context.Context) ([]Label, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, profile_id, name, color, created_at, updated_at
		 FROM labels
		 ORDER BY name ASC, created_at ASC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("cache: list labels: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var label Label
		if err := rows.Scan(
			&label.ID,
			&label.ProfileID,
			&label.Name,
			&label.Color,
			&label.CreatedAt,
			&label.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan label: %w", err)
		}

		labels = append(labels, label)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate labels: %w", err)
	}

	return labels, nil
}

// UpsertLabelTarget attaches a label to one searchable entity target.
func (s *Store) UpsertLabelTarget(ctx context.Context, target LabelTarget) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if target.ID == "" {
		return errors.New("cache: label target ID is required")
	}

	if target.LabelID == "" {
		return errors.New("cache: label target label ID is required")
	}

	if target.ProfileID == "" {
		return errors.New("cache: label target profile ID is required")
	}

	if target.Kind == "" {
		return errors.New("cache: label target kind is required")
	}

	if target.Target == "" {
		return errors.New("cache: label target value is required")
	}

	now := time.Now().UTC()
	if target.CreatedAt.IsZero() {
		target.CreatedAt = now
	}
	target.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO label_targets (id, label_id, profile_id, kind, target, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(label_id, profile_id, kind, target) DO UPDATE SET
		   id = excluded.id,
		   updated_at = excluded.updated_at;`,
		target.ID,
		target.LabelID,
		target.ProfileID,
		target.Kind,
		target.Target,
		target.CreatedAt,
		target.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert label target %q: %w", target.ID, err)
	}

	return nil
}

// DeleteLabel removes a label definition by its primary key.
func (s *Store) DeleteLabel(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if id == "" {
		return errors.New("cache: label ID is required")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM labels WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("cache: delete label %q: %w", id, err)
	}
	return nil
}

// DeleteLabelTarget removes one label-to-entity attachment by its primary key.
func (s *Store) DeleteLabelTarget(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if id == "" {
		return errors.New("cache: label target ID is required")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM label_targets WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("cache: delete label target %q: %w", id, err)
	}
	return nil
}

// ListLabelTargets returns all label attachments ordered by label then entity.
func (s *Store) ListLabelTargets(ctx context.Context) ([]LabelTarget, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, label_id, profile_id, kind, target, created_at, updated_at
		 FROM label_targets
		 ORDER BY label_id ASC, kind ASC, target ASC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("cache: list label targets: %w", err)
	}
	defer rows.Close()

	var targets []LabelTarget
	for rows.Next() {
		var target LabelTarget
		if err := rows.Scan(
			&target.ID,
			&target.LabelID,
			&target.ProfileID,
			&target.Kind,
			&target.Target,
			&target.CreatedAt,
			&target.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan label target: %w", err)
		}

		targets = append(targets, target)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate label targets: %w", err)
	}

	return targets, nil
}
