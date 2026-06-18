package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
)

func isWorkspaceCommand(input string) bool {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return false
	}
	switch strings.ToLower(fields[0]) {
	case "bookmark", "note", "label":
		return true
	case "open":
		if len(fields) >= 2 {
			switch strings.ToLower(fields[1]) {
			case "recent", "bookmarks", "bookmarked", "notes", "noted", "labels", "labeled", "views", "watches", "cache":
				return true
			}
		}
		return false
	default:
		return false
	}
}

func executeWorkspaceCommand(ctx context.Context, model *app.Model, store *cache.Store, input string) (bool, error) {
	if store == nil {
		model.SetWarningStatus("Local workspace is unavailable: cache is not configured.")
		return true, nil
	}

	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		model.SetWarningStatus("Usage: bookmark|note|label add|remove [args]  or  open recent|bookmarks|notes|labels")
		return true, nil
	}

	verb := strings.ToLower(fields[0])

	switch verb {
	case "open":
		if len(fields) < 2 {
			model.SetWarningStatus("Usage: open recent|bookmarks|bookmarked|notes|noted|labels|labeled|views|watches")
			return true, nil
		}
		return executeWorkspaceOpenCommand(ctx, model, store, fields)
	case "bookmark", "note", "label":
		if len(fields) < 2 {
			model.SetWarningStatus("Usage: bookmark|note|label add|remove [args]")
			return true, nil
		}
		subcommand := strings.ToLower(fields[1])
		switch verb {
		case "bookmark":
			return true, executeBookmarkCommand(ctx, model, store, subcommand, fields[2:])
		case "note":
			return true, executeNoteCommand(ctx, model, store, subcommand, fields[2:])
		case "label":
			return true, executeLabelCommand(ctx, model, store, subcommand, fields[2:])
		}
	}

	model.SetWarningStatus(fmt.Sprintf("Unknown workspace command %q.", verb))
	return true, nil
}

func executeWorkspaceOpenCommand(ctx context.Context, model *app.Model, store *cache.Store, fields []string) (bool, error) {
	profileID := model.Snapshot().Profile.Name
	subcommand := ""
	if len(fields) >= 2 {
		subcommand = strings.ToLower(fields[1])
	}

	switch subcommand {
	case "recent":
		entities, err := store.ListEntityCache(ctx, profileID, 20)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load entity cache: %v", err))
			return true, nil
		}
		if len(entities) == 0 {
			model.SetInfoStatus("No recently visited entities.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Recently Visited", "", "open recent", entityCacheToSearchResults(entities), 20, 0)
		return true, nil
	case "bookmarked", "bookmarks":
		bookmarks, err := store.ListBookmarks(ctx)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load bookmarks: %v", err))
			return true, nil
		}
		results := make([]app.SearchResult, 0, len(bookmarks))
		for _, b := range bookmarks {
			if !matchesProfile(b.ProfileID, profileID) {
				continue
			}
			results = append(results, bookmarkToSearchResult(b))
		}
		if len(results) == 0 {
			model.SetInfoStatus("No bookmarks saved yet.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Bookmarks", "", "open bookmarks", results, 20, 0)
		return true, nil
	case "noted", "notes":
		notes, err := store.ListNotes(ctx)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load notes: %v", err))
			return true, nil
		}
		results := make([]app.SearchResult, 0, len(notes))
		for _, n := range notes {
			if !matchesProfile(n.ProfileID, profileID) {
				continue
			}
			results = append(results, noteToSearchResult(n))
		}
		if len(results) == 0 {
			model.SetInfoStatus("No notes saved yet.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Notes", "", "open notes", results, 20, 0)
		return true, nil
	case "views":
		views, err := store.ListSavedViews(ctx, profileID)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load saved views: %v", err))
			return true, nil
		}
		results := savedViewsToSearchResults(views)
		if len(results) == 0 {
			model.SetInfoStatus("No saved views yet.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Saved Views", "", "open views", results, 20, 0)
		return true, nil
	case "labeled":
		return executeWorkspaceLabeledCommand(ctx, model, store, fields[2:])
	case "watches":
		settings, err := store.ListWatchSettings(ctx, profileID)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load watch settings: %v", err))
			return true, nil
		}
		results := watchSettingsToSearchResults(settings)
		if len(results) == 0 {
			model.SetInfoStatus("No watch settings saved yet.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Watch Settings", "", "open watches", results, 20, 0)
		return true, nil
	case "labels":
		labels, err := store.ListLabels(ctx)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
			return true, nil
		}
		labelTargets, err := store.ListLabelTargets(ctx)
		if err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to load label targets: %v", err))
			return true, nil
		}
		targetsByLabel := make(map[string][]cache.LabelTarget)
		for _, lt := range labelTargets {
			if matchesProfile(lt.ProfileID, profileID) {
				targetsByLabel[lt.LabelID] = append(targetsByLabel[lt.LabelID], lt)
			}
		}
		results := make([]app.SearchResult, 0)
		for _, label := range labels {
			if !matchesProfile(label.ProfileID, profileID) {
				continue
			}
			targets := targetsByLabel[label.ID]
			if len(targets) == 0 {
				results = append(results, app.SearchResult{
					Kind:        "label",
					Title:       label.Name,
					Description: "no entities attached",
					Enabled:     false,
					Source:      "local",
				})
				continue
			}
			for _, lt := range targets {
				command := lookupCommandForLocalTarget(lt.Kind, lt.Target)
				desc := lt.Kind + " " + truncateCommandLabel(lt.Target, 24)
				if label.Color != "" {
					desc += "  " + label.Color
				}
				results = append(results, app.SearchResult{
					Kind:        "label",
					Title:       label.Name,
					Description: desc,
					Command:     command,
					Enabled:     command != "",
					Source:      "local",
				})
			}
		}
		if len(results) == 0 {
			model.SetInfoStatus("No labels saved yet.")
			return true, nil
		}
		model.OpenLookupResultExplorer("Labels", "", "open labels", results, 20, 0)
		return true, nil
	default:
		model.SetWarningStatus(fmt.Sprintf("Unknown open workspace command %q. Use: open recent|bookmarks|bookmarked|notes|noted|labels|labeled|views|watches", subcommand))
		return true, nil
	}
}

func executeWorkspaceLabeledCommand(ctx context.Context, model *app.Model, store *cache.Store, filterArgs []string) (bool, error) {
	profileID := model.Snapshot().Profile.Name
	labelFilter := strings.ToLower(strings.TrimSpace(strings.Join(filterArgs, " ")))

	labels, err := store.ListLabels(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
		return true, nil
	}
	labelTargets, err := store.ListLabelTargets(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load label targets: %v", err))
		return true, nil
	}

	targetsByLabel := make(map[string][]cache.LabelTarget)
	for _, lt := range labelTargets {
		if matchesProfile(lt.ProfileID, profileID) {
			targetsByLabel[lt.LabelID] = append(targetsByLabel[lt.LabelID], lt)
		}
	}

	results := make([]app.SearchResult, 0)
	for _, label := range labels {
		if !matchesProfile(label.ProfileID, profileID) {
			continue
		}
		if labelFilter != "" && !strings.EqualFold(label.Name, labelFilter) && !strings.Contains(strings.ToLower(label.Name), labelFilter) {
			continue
		}
		for _, lt := range targetsByLabel[label.ID] {
			command := lookupCommandForLocalTarget(lt.Kind, lt.Target)
			desc := lt.Kind + " " + truncateCommandLabel(lt.Target, 24)
			if label.Color != "" {
				desc += "  " + label.Color
			}
			results = append(results, app.SearchResult{
				Kind:        "label",
				Title:       label.Name,
				Description: desc,
				Command:     command,
				Enabled:     command != "",
				Source:      "local",
			})
		}
	}

	if len(results) == 0 {
		if labelFilter != "" {
			model.SetInfoStatus(fmt.Sprintf("No labeled entities matching %q.", labelFilter))
		} else {
			model.SetInfoStatus("No labeled entities yet.")
		}
		return true, nil
	}

	title := "Labeled Entities"
	if labelFilter != "" {
		title = "Labeled: " + labelFilter
	}
	model.OpenLookupResultExplorer(title, "", "open labeled", results, 20, 0)
	return true, nil
}

func executeBookmarkCommand(ctx context.Context, model *app.Model, store *cache.Store, subcommand string, args []string) error {
	switch subcommand {
	case "add":
		return addBookmark(ctx, model, store, args)
	case "remove", "rm", "delete", "del":
		return removeBookmarks(ctx, model, store)
	case "note", "annotate":
		return setBookmarkNote(ctx, model, store, args)
	default:
		model.SetWarningStatus("Usage: bookmark add [title]  |  bookmark remove  |  bookmark note <text>")
		return nil
	}
}

func addBookmark(ctx context.Context, model *app.Model, store *cache.Store, titleArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("bookmark add requires an active entity lookup.")
		return nil
	}

	title := strings.Join(titleArgs, " ")
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSpace(kind) + " " + strings.TrimSpace(target)
	}

	bookmark := cache.Bookmark{
		ID:        generateWorkspaceID(),
		ProfileID: snapshot.Profile.Name,
		Kind:      kind,
		Target:    target,
		Title:     title,
	}
	if err := store.UpsertBookmark(ctx, bookmark); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to add bookmark: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus(fmt.Sprintf("Bookmarked: %s", title))
	return nil
}

func removeBookmarks(ctx context.Context, model *app.Model, store *cache.Store) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("bookmark remove requires an active entity lookup.")
		return nil
	}

	bookmarks, err := store.ListBookmarks(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load bookmarks: %v", err))
		return nil
	}

	removed := 0
	for _, bookmark := range bookmarks {
		if !matchesProfile(bookmark.ProfileID, snapshot.Profile.Name) {
			continue
		}
		if !lookupTargetMatches(bookmark.Kind, bookmark.Target, kind, target) {
			continue
		}
		if err := store.DeleteBookmark(ctx, bookmark.ID); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to remove bookmark: %v", err))
			return nil
		}
		removed++
	}

	afterWorkspaceChange(ctx, model, store)
	if removed == 0 {
		model.SetInfoStatus("No bookmarks to remove for this entity.")
	} else {
		model.SetInfoStatus(fmt.Sprintf("Removed %d bookmark(s).", removed))
	}
	return nil
}

func executeNoteCommand(ctx context.Context, model *app.Model, store *cache.Store, subcommand string, args []string) error {
	switch subcommand {
	case "add":
		return addNote(ctx, model, store, args)
	case "remove", "rm", "delete", "del":
		return removeNotes(ctx, model, store, args)
	case "body":
		return setNoteBody(ctx, model, store, args)
	default:
		model.SetWarningStatus("Usage: note add [title] [| body]  |  note remove [title-filter]  |  note body <text>")
		return nil
	}
}

func addNote(ctx context.Context, model *app.Model, store *cache.Store, titleArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("note add requires an active entity lookup.")
		return nil
	}

	full := strings.Join(titleArgs, " ")
	title, body, _ := strings.Cut(full, "|")
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		title = "Note on " + strings.TrimSpace(kind) + " " + strings.TrimSpace(target)
	}

	noteTarget := strings.TrimSpace(kind) + ":" + strings.TrimSpace(target)
	note := cache.Note{
		ID:        generateWorkspaceID(),
		ProfileID: snapshot.Profile.Name,
		Target:    noteTarget,
		Title:     title,
		Body:      body,
	}
	if err := store.UpsertNote(ctx, note); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to add note: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus(fmt.Sprintf("Note added: %s", title))
	return nil
}

func removeNotes(ctx context.Context, model *app.Model, store *cache.Store, filterArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("note remove requires an active entity lookup.")
		return nil
	}

	titleFilter := strings.ToLower(strings.TrimSpace(strings.Join(filterArgs, " ")))

	notes, err := store.ListNotes(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load notes: %v", err))
		return nil
	}

	removed := 0
	for _, note := range notes {
		if !matchesProfile(note.ProfileID, snapshot.Profile.Name) {
			continue
		}
		if !noteTargetMatches(note.Target, kind, target) {
			continue
		}
		if titleFilter != "" && !strings.Contains(strings.ToLower(note.Title), titleFilter) {
			continue
		}
		if err := store.DeleteNote(ctx, note.ID); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to remove note: %v", err))
			return nil
		}
		removed++
	}

	afterWorkspaceChange(ctx, model, store)
	if removed == 0 {
		if titleFilter != "" {
			model.SetInfoStatus(fmt.Sprintf("No notes matching %q to remove.", titleFilter))
		} else {
			model.SetInfoStatus("No notes to remove for this entity.")
		}
	} else {
		model.SetInfoStatus(fmt.Sprintf("Removed %d note(s).", removed))
	}
	return nil
}

func executeLabelCommand(ctx context.Context, model *app.Model, store *cache.Store, subcommand string, args []string) error {
	switch subcommand {
	case "add":
		return addLabel(ctx, model, store, args)
	case "remove", "rm":
		return removeLabel(ctx, model, store, args)
	case "delete", "del", "destroy":
		return deleteLabel(ctx, model, store, args)
	case "color", "colour":
		return setLabelColor(ctx, model, store, args)
	default:
		model.SetWarningStatus("Usage: label add <name>  |  label remove <name>  |  label delete <name>  |  label color <name> <color>")
		return nil
	}
}

func addLabel(ctx context.Context, model *app.Model, store *cache.Store, nameArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("label add requires an active entity lookup.")
		return nil
	}

	name := strings.TrimSpace(strings.Join(nameArgs, " "))
	if name == "" {
		model.SetWarningStatus("Usage: label add <name>")
		return nil
	}

	labels, err := store.ListLabels(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
		return nil
	}

	var labelID string
	for _, label := range labels {
		if strings.EqualFold(label.Name, name) && matchesProfile(label.ProfileID, snapshot.Profile.Name) {
			labelID = label.ID
			break
		}
	}
	if labelID == "" {
		newLabel := cache.Label{
			ID:        generateWorkspaceID(),
			ProfileID: snapshot.Profile.Name,
			Name:      name,
		}
		if err := store.UpsertLabel(ctx, newLabel); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to create label: %v", err))
			return nil
		}
		labelID = newLabel.ID
	}

	labelTarget := cache.LabelTarget{
		ID:        generateWorkspaceID(),
		LabelID:   labelID,
		ProfileID: snapshot.Profile.Name,
		Kind:      kind,
		Target:    target,
	}
	if err := store.UpsertLabelTarget(ctx, labelTarget); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to attach label: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus(fmt.Sprintf("Label %q applied to %s %s.", name, kind, target))
	return nil
}

func removeLabel(ctx context.Context, model *app.Model, store *cache.Store, nameArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("label remove requires an active entity lookup.")
		return nil
	}

	name := strings.TrimSpace(strings.Join(nameArgs, " "))
	if name == "" {
		model.SetWarningStatus("Usage: label remove <name>")
		return nil
	}

	labels, err := store.ListLabels(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
		return nil
	}

	var labelID string
	for _, label := range labels {
		if strings.EqualFold(label.Name, name) && matchesProfile(label.ProfileID, snapshot.Profile.Name) {
			labelID = label.ID
			break
		}
	}
	if labelID == "" {
		model.SetInfoStatus(fmt.Sprintf("Label %q not found.", name))
		return nil
	}

	labelTargets, err := store.ListLabelTargets(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load label targets: %v", err))
		return nil
	}

	removed := 0
	for _, lt := range labelTargets {
		if lt.LabelID != labelID || !matchesProfile(lt.ProfileID, snapshot.Profile.Name) {
			continue
		}
		if !lookupTargetMatches(lt.Kind, lt.Target, kind, target) {
			continue
		}
		if err := store.DeleteLabelTarget(ctx, lt.ID); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to remove label: %v", err))
			return nil
		}
		removed++
	}

	afterWorkspaceChange(ctx, model, store)
	if removed == 0 {
		model.SetInfoStatus(fmt.Sprintf("Label %q was not applied to this entity.", name))
	} else {
		model.SetInfoStatus(fmt.Sprintf("Removed label %q.", name))
	}
	return nil
}

func deleteLabel(ctx context.Context, model *app.Model, store *cache.Store, nameArgs []string) error {
	snapshot := model.Snapshot()

	name := strings.TrimSpace(strings.Join(nameArgs, " "))
	if name == "" {
		model.SetWarningStatus("Usage: label delete <name>")
		return nil
	}

	labels, err := store.ListLabels(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
		return nil
	}

	var found *cache.Label
	for i := range labels {
		l := &labels[i]
		if strings.EqualFold(l.Name, name) && matchesProfile(l.ProfileID, snapshot.Profile.Name) {
			found = l
			break
		}
	}

	if found == nil {
		model.SetInfoStatus(fmt.Sprintf("Label %q not found.", name))
		return nil
	}

	labelTargets, err := store.ListLabelTargets(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load label targets: %v", err))
		return nil
	}

	for _, lt := range labelTargets {
		if lt.LabelID != found.ID {
			continue
		}
		if err := store.DeleteLabelTarget(ctx, lt.ID); err != nil {
			model.SetWarningStatus(fmt.Sprintf("Failed to remove label target: %v", err))
			return nil
		}
	}

	if err := store.DeleteLabel(ctx, found.ID); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to delete label: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus(fmt.Sprintf("Label %q deleted.", name))
	return nil
}

func setBookmarkNote(ctx context.Context, model *app.Model, store *cache.Store, textArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("bookmark note requires an active entity lookup.")
		return nil
	}

	text := strings.TrimSpace(strings.Join(textArgs, " "))

	bookmarks, err := store.ListBookmarks(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load bookmarks: %v", err))
		return nil
	}

	var found *cache.Bookmark
	for i := range bookmarks {
		b := &bookmarks[i]
		if !matchesProfile(b.ProfileID, snapshot.Profile.Name) {
			continue
		}
		if lookupTargetMatches(b.Kind, b.Target, kind, target) {
			found = b
			break
		}
	}

	if found == nil {
		model.SetWarningStatus("No bookmark found for this entity. Use 'bookmark add' first.")
		return nil
	}

	updated := *found
	updated.Notes = text
	if err := store.UpsertBookmark(ctx, updated); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to update bookmark: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus("Bookmark annotation updated.")
	return nil
}

func setLabelColor(ctx context.Context, model *app.Model, store *cache.Store, args []string) error {
	snapshot := model.Snapshot()
	if len(args) < 2 {
		model.SetWarningStatus("Usage: label color <name> <color>")
		return nil
	}

	name := strings.TrimSpace(args[0])
	color := strings.TrimSpace(strings.Join(args[1:], " "))

	labels, err := store.ListLabels(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load labels: %v", err))
		return nil
	}

	var found *cache.Label
	for i := range labels {
		l := &labels[i]
		if strings.EqualFold(l.Name, name) && matchesProfile(l.ProfileID, snapshot.Profile.Name) {
			found = l
			break
		}
	}

	if found == nil {
		model.SetWarningStatus(fmt.Sprintf("Label %q not found.", name))
		return nil
	}

	updated := *found
	updated.Color = color
	if err := store.UpsertLabel(ctx, updated); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to update label: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus(fmt.Sprintf("Label %q color set to %q.", name, color))
	return nil
}

func setNoteBody(ctx context.Context, model *app.Model, store *cache.Store, bodyArgs []string) error {
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		model.SetWarningStatus("note body requires an active entity lookup.")
		return nil
	}

	full := strings.Join(bodyArgs, " ")
	var titleFilter, body string
	if idx := strings.Index(full, "|"); idx >= 0 {
		titleFilter = strings.ToLower(strings.TrimSpace(full[:idx]))
		body = strings.TrimSpace(full[idx+1:])
	} else {
		body = strings.TrimSpace(full)
	}

	noteTarget := strings.TrimSpace(kind) + ":" + strings.TrimSpace(target)

	notes, err := store.ListNotes(ctx)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to load notes: %v", err))
		return nil
	}

	var found *cache.Note
	for i := range notes {
		n := &notes[i]
		if !matchesProfile(n.ProfileID, snapshot.Profile.Name) {
			continue
		}
		if !noteTargetMatches(n.Target, kind, target) {
			continue
		}
		if titleFilter != "" && !strings.Contains(strings.ToLower(n.Title), titleFilter) {
			continue
		}
		found = n
		break // ListNotes orders by updated_at DESC; first match is most recent
	}

	if found == nil {
		if titleFilter != "" {
			model.SetWarningStatus(fmt.Sprintf("No note matching %q found for this entity.", titleFilter))
		} else {
			model.SetWarningStatus("No note found for this entity. Use 'note add' first.")
		}
		return nil
	}

	updated := *found
	updated.Body = body
	updated.Target = noteTarget
	if err := store.UpsertNote(ctx, updated); err != nil {
		model.SetWarningStatus(fmt.Sprintf("Failed to update note: %v", err))
		return nil
	}

	afterWorkspaceChange(ctx, model, store)
	model.SetInfoStatus("Note body updated.")
	return nil
}

func entityCacheToSearchResults(entities []cache.EntityCache) []app.SearchResult {
	results := make([]app.SearchResult, 0, len(entities))
	for _, e := range entities {
		command := lookupCommandForLocalTarget(e.Kind, e.Target)
		desc := e.Summary
		if !e.UpdatedAt.IsZero() {
			desc += "  " + e.UpdatedAt.UTC().Format("2006-01-02 15:04")
		}
		results = append(results, app.SearchResult{
			Kind:        e.Kind,
			Title:       e.Title,
			Description: strings.TrimSpace(desc),
			Command:     command,
			Enabled:     command != "",
			Source:      "local",
		})
	}
	return results
}

func bookmarkToSearchResult(b cache.Bookmark) app.SearchResult {
	command := lookupCommandForLocalTarget(b.Kind, b.Target)
	desc := b.Kind + " " + truncateCommandLabel(b.Target, 24)
	if b.Notes != "" {
		desc += " — " + truncateCommandLabel(b.Notes, 20)
	}
	return app.SearchResult{
		Kind:        "bookmark",
		Title:       b.Title,
		Description: desc,
		Command:     command,
		Enabled:     command != "",
		Source:      "local",
	}
}

func noteToSearchResult(n cache.Note) app.SearchResult {
	command := lookupCommandForLocalTarget("", n.Target)
	desc := truncateCommandLabel(n.Target, 28)
	if n.Body != "" {
		desc += " — " + truncateCommandLabel(n.Body, 20)
	}
	return app.SearchResult{
		Kind:        "note",
		Title:       n.Title,
		Description: desc,
		Command:     command,
		Enabled:     command != "",
		Source:      "local",
	}
}

func lookupEntityContext(snapshot app.Snapshot) (kind string, target string) {
	if snapshot.Current != app.ScreenLookup {
		return "", ""
	}
	if snapshot.Lookup.State != app.ViewStateReady {
		return "", ""
	}
	return string(snapshot.Lookup.Kind), strings.TrimSpace(snapshot.Lookup.Query)
}

func afterWorkspaceChange(ctx context.Context, model *app.Model, store *cache.Store) {
	if store == nil {
		return
	}
	snapshot := model.Snapshot()
	kind, target := lookupEntityContext(snapshot)
	if target == "" {
		return
	}
	metadata, err := loadLookupMetadata(ctx, store, snapshot.Profile.Name, kind, target)
	if err != nil {
		model.SetWarningStatus(fmt.Sprintf("Metadata reload failed: %v", err))
		return
	}
	model.SetLookupMetadata(metadata)
}

func generateWorkspaceID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
