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

func isViewCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	return len(fields) > 0 && strings.EqualFold(fields[0], "view")
}

func executeViewCommand(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, input string) (bool, error) {
	if store == nil {
		model.SetWarningStatus("Local workspace is unavailable: cache is not configured.")
		return true, nil
	}

	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) < 2 {
		model.SetWarningStatus("Usage: view save <name> | view open <name> | view delete <name>")
		return true, nil
	}

	switch strings.ToLower(fields[1]) {
	case "save":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: view save <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return true, saveCurrentView(ctx, model, store, name)
	case "open":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: view open <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return openSavedView(ctx, cfg, model, store, name)
	case "delete", "remove", "rm":
		if len(fields) < 3 {
			model.SetWarningStatus("Usage: view delete <name>")
			return true, nil
		}
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		return true, deleteSavedView(ctx, model, store, name)
	default:
		model.SetWarningStatus("Usage: view save|open|delete <name>")
		return true, nil
	}
}

func saveCurrentView(ctx context.Context, model *app.Model, store *cache.Store, name string) error {
	snapshot := model.Snapshot()
	command, screen, kind, target, filtersJSON := captureSavedViewContext(snapshot)
	if strings.TrimSpace(command) == "" {
		model.SetWarningStatus("Nothing saveable is active on the current screen.")
		return nil
	}

	view := cache.SavedView{
		ID:           generateWorkspaceID(),
		ProfileID:    snapshot.Profile.Name,
		Name:         name,
		Command:      command,
		Screen:       screen,
		EntityKind:   kind,
		EntityTarget: target,
		FiltersJSON:  filtersJSON,
	}
	if err := store.UpsertSavedView(ctx, view); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to save view: %v", err))
		return nil
	}
	model.SetInfoStatus(fmt.Sprintf("Saved view %q.", name))
	return nil
}

func openSavedView(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, name string) (bool, error) {
	views, err := store.ListSavedViews(ctx, model.Snapshot().Profile.Name)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load saved views: %v", err))
		return true, nil
	}
	for _, view := range views {
		if strings.EqualFold(view.Name, name) {
			if screen, err := app.ParseScreen(view.Screen); err == nil {
				_ = model.SetScreen(screen)
			}
			if filter := savedViewLiveFilter(view.FiltersJSON); filter != "" {
				_ = model.SetLiveFeedFilter(filter)
			}
			keepRunning, err := executeCommand(ctx, cfg, model, view.Command, store)
			if err != nil {
				return false, err
			}
			model.SetInfoStatus(fmt.Sprintf("Opened saved view %q.", view.Name))
			return keepRunning, nil
		}
	}
	model.SetWarningStatus(fmt.Sprintf("Saved view %q not found.", name))
	return true, nil
}

func deleteSavedView(ctx context.Context, model *app.Model, store *cache.Store, name string) error {
	if err := store.DeleteSavedView(ctx, model.Snapshot().Profile.Name, name); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to delete view: %v", err))
		return nil
	}
	model.SetInfoStatus(fmt.Sprintf("Deleted saved view %q.", name))
	return nil
}

func captureSavedViewContext(snapshot app.Snapshot) (command string, screen string, kind string, target string, filtersJSON string) {
	screen = string(snapshot.Current)
	switch snapshot.Current {
	case app.ScreenLookup:
		if snapshot.Lookup.State == app.ViewStateReady {
			kind = string(snapshot.Lookup.Kind)
			target = strings.TrimSpace(snapshot.Lookup.Query)
			command = lookupCommandForLocalTarget(kind, target)
		}
	case app.ScreenLiveFeed:
		command = "live"
		if filter := strings.TrimSpace(snapshot.LiveFeed.Filter); filter != "" && filter != app.LiveFeedFilterAll {
			filtersJSON = mustJSON(map[string]string{"live_filter": filter})
		}
	case app.ScreenHome:
		command = "home"
	case app.ScreenSettings:
		command = "settings"
	}
	return command, screen, kind, target, filtersJSON
}

func savedViewLiveFilter(filtersJSON string) string {
	if strings.TrimSpace(filtersJSON) == "" {
		return ""
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(filtersJSON), &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload["live_filter"])
}

func savedViewsToSearchResults(views []cache.SavedView) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(views))
	for _, view := range views {
		desc := strings.TrimSpace(view.Command)
		if view.EntityKind != "" && view.EntityTarget != "" {
			desc = view.EntityKind + " " + truncateCommandLabel(view.EntityTarget, 24)
		}
		results = append(results, app.SearchResult{
			Kind:        "view",
			Title:       view.Name,
			Description: desc,
			Command:     "view open " + view.Name,
			Enabled:     strings.TrimSpace(view.Command) != "",
			Source:      "local",
		})
	}
	return results
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
