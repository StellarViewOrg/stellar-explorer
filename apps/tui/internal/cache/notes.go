package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// UpsertNote persists a free-form note keyed by note ID.
func (s *Store) UpsertNote(ctx context.Context, note Note) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if note.ID == "" {
		return errors.New("cache: note ID is required")
	}

	if note.ProfileID == "" {
		return errors.New("cache: note profile ID is required")
	}

	if note.Target == "" {
		return errors.New("cache: note target is required")
	}

	if note.Title == "" {
		return errors.New("cache: note title is required")
	}

	now := time.Now().UTC()
	if note.CreatedAt.IsZero() {
		note.CreatedAt = now
	}
	note.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO notes (id, profile_id, target, title, body, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   profile_id = excluded.profile_id,
		   target = excluded.target,
		   title = excluded.title,
		   body = excluded.body,
		   updated_at = excluded.updated_at;`,
		note.ID,
		note.ProfileID,
		note.Target,
		note.Title,
		note.Body,
		note.CreatedAt,
		note.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert note %q: %w", note.ID, err)
	}

	return nil
}

// DeleteNote removes a note by its primary key.
func (s *Store) DeleteNote(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if id == "" {
		return errors.New("cache: note ID is required")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("cache: delete note %q: %w", id, err)
	}
	return nil
}

// ListNotes returns all notes ordered by most recently updated first.
func (s *Store) ListNotes(ctx context.Context) ([]Note, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, profile_id, target, title, body, created_at, updated_at
		 FROM notes
		 ORDER BY updated_at DESC, created_at DESC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("cache: list notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		if err := rows.Scan(
			&note.ID,
			&note.ProfileID,
			&note.Target,
			&note.Title,
			&note.Body,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan note: %w", err)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate notes: %w", err)
	}

	return notes, nil
}
