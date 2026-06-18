package main

import (
	"context"
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestIsWatchCommand(t *testing.T) {
	if !isWatchCommand("watch save ops") {
		t.Fatal("expected watch command")
	}
	if isWatchCommand("view save ops") {
		t.Fatal("did not expect view command to match watch")
	}
}

func TestSaveAndOpenWatchSetting(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	_ = model.SetScreen(app.ScreenLiveFeed)
	model.SetLiveFeedPaused(true)
	_ = model.SetLiveFeedFilter(app.LiveFeedFilterSoroban)

	if _, err := executeWatchCommand(context.Background(), config.Default(), model, store, "watch save soroban-watch"); err != nil {
		t.Fatalf("executeWatchCommand(save) error = %v", err)
	}

	settings, err := store.ListWatchSettings(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListWatchSettings() error = %v", err)
	}
	if len(settings) != 1 {
		t.Fatalf("expected 1 watch setting, got %d", len(settings))
	}
	if !settings[0].Paused {
		t.Fatal("expected paused watch setting")
	}

	_ = model.SetScreen(app.ScreenHome)
	keepRunning, err := executeWatchCommand(context.Background(), config.Default(), model, store, "watch open soroban-watch")
	if err != nil {
		t.Fatalf("executeWatchCommand(open) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected runtime to keep running")
	}
	if model.Snapshot().Current != app.ScreenLiveFeed {
		t.Fatalf("expected live feed screen, got %s", model.Snapshot().Current)
	}
	if model.Snapshot().LiveFeed.Filter != app.LiveFeedFilterSoroban {
		t.Fatalf("expected soroban filter, got %q", model.Snapshot().LiveFeed.Filter)
	}
	if !model.Snapshot().LiveFeed.Paused {
		t.Fatal("expected paused state after opening watch")
	}
}

func TestParseWatchFilterSpec(t *testing.T) {
	spec := parseWatchFilterSpec(`{"live_filter":"soroban account:GABC"}`)
	if spec.Class != app.LiveFeedFilterSoroban {
		t.Fatalf("class = %q, want soroban", spec.Class)
	}
	if !strings.Contains(spec.Account, "GABC") {
		t.Fatalf("account = %q, want GABC fragment", spec.Account)
	}
}
