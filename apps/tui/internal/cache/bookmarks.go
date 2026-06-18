package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// UpsertBookmark persists a bookmark row keyed by bookmark ID.
func (s *Store) UpsertBookmark(ctx context.Context, bookmark Bookmark) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}

	if bookmark.ID == "" {
		return errors.New("cache: bookmark ID is required")
	}

	if bookmark.ProfileID == "" {
		return errors.New("cache: bookmark profile ID is required")
	}

	if bookmark.Kind == "" {
		return errors.New("cache: bookmark kind is required")
	}

	if bookmark.Target == "" {
		return errors.New("cache: bookmark target is required")
	}

	if bookmark.Title == "" {
		return errors.New("cache: bookmark title is required")
	}

	now := time.Now().UTC()
	if bookmark.CreatedAt.IsZero() {
		bookmark.CreatedAt = now
	}
	bookmark.UpdatedAt = now

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO bookmarks (id, profile_id, kind, target, title, notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   profile_id = excluded.profile_id,
		   kind = excluded.kind,
		   target = excluded.target,
		   title = excluded.title,
		   notes = excluded.notes,
		   updated_at = excluded.updated_at;`,
		bookmark.ID,
		bookmark.ProfileID,
		bookmark.Kind,
		bookmark.Target,
		bookmark.Title,
		bookmark.Notes,
		bookmark.CreatedAt,
		bookmark.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("cache: upsert bookmark %q: %w", bookmark.ID, err)
	}

	return nil
}

// DeleteBookmark removes a bookmark by its primary key.
func (s *Store) DeleteBookmark(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if id == "" {
		return errors.New("cache: bookmark ID is required")
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ?;`, id)
	if err != nil {
		return fmt.Errorf("cache: delete bookmark %q: %w", id, err)
	}
	return nil
}

// ListBookmarks returns all bookmarks ordered by title then creation time.
func (s *Store) ListBookmarks(ctx context.Context) ([]Bookmark, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, profile_id, kind, target, title, notes, created_at, updated_at
		 FROM bookmarks
		 ORDER BY title ASC, created_at ASC;`,
	)
	if err != nil {
		return nil, fmt.Errorf("cache: list bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var bookmark Bookmark
		if err := rows.Scan(
			&bookmark.ID,
			&bookmark.ProfileID,
			&bookmark.Kind,
			&bookmark.Target,
			&bookmark.Title,
			&bookmark.Notes,
			&bookmark.CreatedAt,
			&bookmark.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan bookmark: %w", err)
		}

		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate bookmarks: %w", err)
	}

	return bookmarks, nil
}
