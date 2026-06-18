package cache

import (
	"context"
	"database/sql"
	"testing"
)

func TestUpsertAndListWatchSettings(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)
	db, err := sql.Open(driverName, "watch-settings-test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	setting := WatchSetting{
		ID:          "watch-1",
		ProfileID:   "default",
		Name:        "soroban-watch",
		FiltersJSON: `{"live_filter":"soroban"}`,
		Paused:      true,
		AutoApply:   true,
	}
	if err := store.UpsertWatchSetting(context.Background(), setting); err != nil {
		t.Fatalf("UpsertWatchSetting() error = %v", err)
	}

	settings, err := store.ListWatchSettings(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListWatchSettings() error = %v", err)
	}
	if len(settings) != 1 {
		t.Fatalf("expected 1 watch setting, got %d", len(settings))
	}
	if !settings[0].Paused || !settings[0].AutoApply {
		t.Fatalf("expected paused auto-apply watch setting, got %#v", settings[0])
	}

	found, err := store.FindAutoApplyWatchSetting(context.Background(), "default")
	if err != nil {
		t.Fatalf("FindAutoApplyWatchSetting() error = %v", err)
	}
	if found.Name != "soroban-watch" {
		t.Fatalf("expected auto-apply watch soroban-watch, got %q", found.Name)
	}
}

func TestDeleteWatchSetting(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)
	db, err := sql.Open(driverName, "watch-settings-delete-test")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store, err := Open(context.Background(), db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.UpsertWatchSetting(context.Background(), WatchSetting{
		ID:        "watch-del",
		ProfileID: "default",
		Name:      "temp",
	}); err != nil {
		t.Fatalf("UpsertWatchSetting() error = %v", err)
	}
	if err := store.DeleteWatchSetting(context.Background(), "default", "temp"); err != nil {
		t.Fatalf("DeleteWatchSetting() error = %v", err)
	}
	settings, err := store.ListWatchSettings(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListWatchSettings() error = %v", err)
	}
	if len(settings) != 0 {
		t.Fatalf("expected watch setting deleted, got %#v", settings)
	}
}
