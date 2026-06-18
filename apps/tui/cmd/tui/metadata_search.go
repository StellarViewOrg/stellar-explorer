package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func localMetadataSearchResults(ctx context.Context, store *cache.Store, query string, profileID string, limit int) ([]app.SearchResult, error) {
	if store == nil || strings.TrimSpace(query) == "" || limit <= 0 {
		return nil, nil
	}

	filter, searchQuery := app.ParseMetadataSearchQuery(query)
	if searchQuery == "" && filter == "" {
		return nil, nil
	}
	if searchQuery == "" {
		searchQuery = query
	}

	results := make([]app.SearchResult, 0, limit)

	if filter == "" || filter == "bookmark" {
		bookmarks, err := store.ListBookmarks(ctx)
		if err != nil {
			return nil, err
		}
		for _, bookmark := range bookmarks {
			if len(results) >= limit {
				return results, nil
			}
			if !matchesProfile(bookmark.ProfileID, profileID) {
				continue
			}
			haystack := strings.ToLower(bookmark.Title + " " + bookmark.Kind + " " + bookmark.Target + " " + bookmark.Notes)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			command := lookupCommandForLocalTarget(bookmark.Kind, bookmark.Target)
			results = append(results, app.SearchResult{
				Kind:        "bookmark",
				Title:       bookmark.Title,
				Description: fmt.Sprintf("%s %s", bookmark.Kind, truncateMetadataSearchText(bookmark.Target, 18)),
				Command:     command,
				Enabled:     command != "",
				Source:      "local",
			})
		}
	}

	if filter == "" || filter == "note" {
		notes, err := store.ListNotes(ctx)
		if err != nil {
			return nil, err
		}
		for _, note := range notes {
			if len(results) >= limit {
				return results, nil
			}
			if !matchesProfile(note.ProfileID, profileID) {
				continue
			}
			haystack := strings.ToLower(note.Title + " " + note.Target + " " + note.Body)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			command := lookupCommandForLocalTarget("", note.Target)
			results = append(results, app.SearchResult{
				Kind:        "note",
				Title:       note.Title,
				Description: truncateMetadataSearchText(note.Target, 32),
				Command:     command,
				Enabled:     command != "",
				Source:      "local",
			})
		}
	}

	if filter == "" || filter == "label" {
		labels, err := store.ListLabels(ctx)
		if err != nil {
			return nil, err
		}
		labelTargets, err := store.ListLabelTargets(ctx)
		if err != nil {
			return nil, err
		}
		targetsByLabel := make(map[string][]cache.LabelTarget)
		for _, target := range labelTargets {
			if !matchesProfile(target.ProfileID, profileID) {
				continue
			}
			targetsByLabel[target.LabelID] = append(targetsByLabel[target.LabelID], target)
		}
		for _, label := range labels {
			if len(results) >= limit {
				return results, nil
			}
			if !matchesProfile(label.ProfileID, profileID) {
				continue
			}
			haystack := strings.ToLower(label.Name + " " + label.Color)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			targets := targetsByLabel[label.ID]
			if len(targets) == 0 {
				results = append(results, app.SearchResult{
					Kind:        "label",
					Title:       label.Name,
					Description: "local label",
					Enabled:     false,
					Source:      "local",
				})
				continue
			}
			for _, target := range targets {
				if len(results) >= limit {
					return results, nil
				}
				command := lookupCommandForLocalTarget(target.Kind, target.Target)
				results = append(results, app.SearchResult{
					Kind:        "label",
					Title:       label.Name,
					Description: fmt.Sprintf("%s %s", target.Kind, truncateMetadataSearchText(target.Target, 24)),
					Command:     command,
					Enabled:     command != "",
					Source:      "local",
				})
			}
		}
	}

	if filter == "" || filter == "cache" {
		entities, err := store.ListEntityCache(ctx, profileID, limit*2)
		if err != nil {
			return nil, err
		}
		for _, entity := range entities {
			if len(results) >= limit {
				return results, nil
			}
			haystack := strings.ToLower(entity.Kind + " " + entity.Target + " " + entity.Title + " " + entity.Summary)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			command := lookupCommandForLocalTarget(entity.Kind, entity.Target)
			desc := entity.Summary
			if !entity.UpdatedAt.IsZero() {
				desc += "  " + entity.UpdatedAt.UTC().Format("2006-01-02 15:04")
			}
			results = append(results, app.SearchResult{
				Kind:        "cache",
				Title:       entity.Title,
				Description: truncateMetadataSearchText(strings.TrimSpace(desc), 40),
				Command:     command,
				Enabled:     command != "",
				Source:      "local",
			})
		}
	}

	if filter == "" || filter == "watch" {
		settings, err := store.ListWatchSettings(ctx, profileID)
		if err != nil {
			return nil, err
		}
		for _, setting := range settings {
			if len(results) >= limit {
				return results, nil
			}
			haystack := strings.ToLower(setting.Name + " " + setting.FiltersJSON)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			results = append(results, app.SearchResult{
				Kind:        "watch",
				Title:       setting.Name,
				Description: truncateMetadataSearchText(savedViewLiveFilter(setting.FiltersJSON), 36),
				Command:     "watch open " + setting.Name,
				Enabled:     true,
				Source:      "local",
			})
		}
	}

	if filter == "" || filter == "view" {
		views, err := store.ListSavedViews(ctx, profileID)
		if err != nil {
			return nil, err
		}
		for _, view := range views {
			if len(results) >= limit {
				return results, nil
			}
			haystack := strings.ToLower(view.Name + " " + view.Command + " " + view.EntityKind + " " + view.EntityTarget)
			if !strings.Contains(haystack, searchQuery) {
				continue
			}
			results = append(results, app.SearchResult{
				Kind:        "view",
				Title:       view.Name,
				Description: truncateMetadataSearchText(view.Command, 36),
				Command:     "view open " + view.Name,
				Enabled:     strings.TrimSpace(view.Command) != "",
				Source:      "local",
			})
		}
	}

	return app.RankSearchResults(searchQuery, results), nil
}

func isSearchMoreCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	return len(fields) >= 3 && strings.EqualFold(fields[0], "search") && strings.EqualFold(fields[1], "more")
}

func executeSearchMoreCommand(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store, input string) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) < 4 {
		model.SetWarningStatus("Usage: search more <query> <offset>")
		return true, nil
	}
	offset, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil || offset < 0 {
		model.SetWarningStatus("search more offset must be a non-negative integer")
		return true, nil
	}
	query := strings.TrimSpace(strings.Trim(strings.Join(fields[2:len(fields)-1], " "), `"`))
	if query == "" {
		model.SetWarningStatus("search more requires a query")
		return true, nil
	}

	model.OpenCommandPalette("Search / Command", query)
	model.SetCommandPaletteSearchLimit(offset + app.SearchResultLimit())
	results, err := loadCommandPaletteResults(ctx, cfg, store, query, model.Snapshot().Profile.Name, model.Snapshot().Command.SearchLimit)
	if err != nil {
		model.SetWarningStatus(err.Error())
		return true, nil
	}
	model.SetCommandPaletteResults(results)
	model.SetInfoStatus(fmt.Sprintf("Loaded search results for %q (limit %d).", query, model.Snapshot().Command.SearchLimit))
	return true, nil
}

func matchesProfile(recordProfileID string, activeProfileID string) bool {
	return strings.TrimSpace(recordProfileID) == "" || strings.TrimSpace(activeProfileID) == "" || recordProfileID == activeProfileID
}

func lookupCommandForLocalTarget(kind string, target string) string {
	kind = strings.ToLower(strings.TrimSpace(kind))
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}

	switch kind {
	case "ledger":
		if isPositiveInteger(target) {
			return "lookup ledger " + target
		}
	case "transaction", "tx":
		return "lookup tx " + target
	case "account", "acct":
		return "lookup account " + target
	case "asset":
		if strings.Contains(target, ":") {
			return "lookup asset " + target
		}
	case "contract":
		return "lookup contract " + target
	}

	if isPositiveInteger(target) {
		return "lookup ledger " + target
	}
	if len(target) == 64 && isHexString(target) {
		return "lookup tx " + target
	}
	if strings.HasPrefix(target, "G") && len(target) == 56 {
		return "lookup account " + target
	}
	if strings.HasPrefix(target, "C") && len(target) == 56 {
		return "lookup contract " + target
	}
	if strings.Contains(target, ":") {
		return "lookup asset " + target
	}
	return ""
}

func isPositiveInteger(value string) bool {
	parsed, err := strconv.ParseUint(value, 10, 32)
	return err == nil && parsed > 0
}

func isHexString(value string) bool {
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func truncateMetadataSearchText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
