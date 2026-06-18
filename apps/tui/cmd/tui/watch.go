package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func isWatchCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	return len(fields) > 0 && strings.EqualFold(fields[0], "watch")
}

func executeWatchCommand(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, input string) (bool, error) {
	if store == nil {
		model.SetWarningStatus("Local workspace is unavailable: cache is not configured.")
		return true, nil
	}

	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) < 2 {
		model.SetWarningStatus("Usage: watch save <name> | watch open <name> | watch delete <name> | watch auto <name>")
		return true, nil
	}

	switch strings.ToLower(fields[1]) {
	case "save":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: watch save <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return true, saveWatchSetting(ctx, model, store, name, false)
	case "open":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: watch open <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return openWatchSetting(ctx, model, store, name)
	case "delete", "remove", "rm":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: watch delete <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return true, deleteWatchSetting(ctx, model, store, name)
	case "auto":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: watch auto <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return true, saveWatchSetting(ctx, model, store, name, true)
	default:
		model.SetWarningStatus("Usage: watch save|open|delete|auto <name>")
		return true, nil
	}
}

func saveWatchSetting(ctx context.Context, model *app.Model, store *cache.Store, name string, autoApply bool) error {
	snapshot := model.Snapshot()
	if snapshot.Current != app.ScreenLiveFeed {
		model.SetWarningStatus("watch save requires the live feed screen.")
		return nil
	}

	setting := cache.WatchSetting{
		ID:          generateWorkspaceID(),
		ProfileID:   snapshot.Profile.Name,
		Name:        name,
		FiltersJSON: captureWatchFiltersJSON(snapshot),
		Paused:      snapshot.LiveFeed.Paused,
		AutoApply:   autoApply,
	}
	if err := store.UpsertWatchSetting(ctx, setting); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to save watch setting: %v", err))
		return nil
	}

	if autoApply {
		model.SetInfoStatus(fmt.Sprintf("Saved watch %q and marked it auto-apply.", name))
	} else {
		model.SetInfoStatus(fmt.Sprintf("Saved watch %q.", name))
	}
	return nil
}

func openWatchSetting(ctx context.Context, model *app.Model, store *cache.Store, name string) (bool, error) {
	settings, err := store.ListWatchSettings(ctx, model.Snapshot().Profile.Name)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load watch settings: %v", err))
		return true, nil
	}

	for _, setting := range settings {
		if strings.EqualFold(setting.Name, name) {
			_ = model.SetScreen(app.ScreenLiveFeed)
			applyWatchSetting(model, setting)
			model.SetInfoStatus(fmt.Sprintf("Opened watch %q.", setting.Name))
			return true, nil
		}
	}

	if profileWatch := strings.TrimSpace(model.Snapshot().Profile.DefaultWatch); profileWatch != "" && strings.EqualFold(profileWatch, name) {
		for _, setting := range settings {
			if strings.EqualFold(setting.Name, profileWatch) {
				_ = model.SetScreen(app.ScreenLiveFeed)
				applyWatchSetting(model, setting)
				model.SetInfoStatus(fmt.Sprintf("Opened profile watch %q.", setting.Name))
				return true, nil
			}
		}
	}

	model.SetWarningStatus(fmt.Sprintf("Watch setting %q not found.", name))
	return true, nil
}

func deleteWatchSetting(ctx context.Context, model *app.Model, store *cache.Store, name string) error {
	if err := store.DeleteWatchSetting(ctx, model.Snapshot().Profile.Name, name); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to delete watch setting: %v", err))
		return nil
	}
	model.SetInfoStatus(fmt.Sprintf("Deleted watch %q.", name))
	return nil
}

func applyWatchSetting(model *app.Model, setting cache.WatchSetting) {
	spec := parseWatchFilterSpec(setting.FiltersJSON)
	model.ApplyWatchSettings(spec, setting.Paused)
}

func captureWatchFiltersJSON(snapshot app.Snapshot) string {
	payload := map[string]string{
		"live_filter": strings.TrimSpace(snapshot.LiveFeed.Filter),
	}
	return mustJSON(payload)
}

func parseWatchFilterSpec(filtersJSON string) app.LiveFeedFilterSpec {
	if strings.TrimSpace(filtersJSON) == "" {
		return app.LiveFeedFilterSpec{Class: app.LiveFeedFilterAll}
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(filtersJSON), &payload); err != nil {
		return app.LiveFeedFilterSpec{Class: app.LiveFeedFilterAll}
	}
	filter := strings.TrimSpace(payload["live_filter"])
	if filter == "" || filter == app.LiveFeedFilterAll {
		return app.LiveFeedFilterSpec{Class: app.LiveFeedFilterAll}
	}
	spec, err := app.ParseLiveFeedFilterValue(filter)
	if err != nil {
		return app.LiveFeedFilterSpec{Class: app.LiveFeedFilterAll}
	}
	return spec
}

func watchSettingsToSearchResults(settings []cache.WatchSetting) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(settings))
	for _, setting := range settings {
		desc := strings.TrimSpace(setting.FiltersJSON)
		if filter := savedViewLiveFilter(setting.FiltersJSON); filter != "" {
			desc = filter
		}
		if setting.Paused {
			desc += "  paused"
		}
		if setting.AutoApply {
			desc += "  auto"
		}
		results = append(results, app.SearchResult{
			Kind:        "watch",
			Title:       setting.Name,
			Description: desc,
			Command:     "watch open " + setting.Name,
			Enabled:     true,
			Source:      "local",
		})
	}
	return results
}

func applyProfileWatchOnLiveScreen(ctx context.Context, model *app.Model, store *cache.Store, profile config.Profile) {
	if store == nil || model.Snapshot().Current != app.ScreenLiveFeed || !model.ShouldApplyDefaultWatch() {
		return
	}

	name := strings.TrimSpace(profile.DefaultWatch)
	if name != "" {
		settings, err := store.ListWatchSettings(ctx, profile.Name)
		if err == nil {
			for _, setting := range settings {
				if strings.EqualFold(setting.Name, name) {
					applyWatchSetting(model, setting)
					model.MarkDefaultWatchApplied()
					return
				}
			}
		}
	}

	setting, err := store.FindAutoApplyWatchSetting(ctx, profile.Name)
	if err != nil {
		return
	}
	applyWatchSetting(model, *setting)
	model.MarkDefaultWatchApplied()
}
