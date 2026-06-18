package main

import (
	"context"
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestIsWorkspaceCommand(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"bookmark add My Title", true},
		{"bookmark remove", true},
		{"bookmark note foo", true},
		{"BOOKMARK add foo", true},
		{"note add Investigation", true},
		{"note remove", true},
		{"label add important", true},
		{"label remove important", true},
		{"label color important red", true},
		{"open recent", true},
		{"open bookmarks", true},
		{"open notes", true},
		{"open labels", true},
		{"open bookmarked", true},
		{"open noted", true},
		{"open labeled", true},
		{"open labeled important", true},
		{"open watches", true},
		{"OPEN RECENT", true},
		{"lookup tx abc", false},
		{"open txs", false},
		{"open ledgers", false},
		{"home", false},
		{"open", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isWorkspaceCommand(tc.input); got != tc.want {
			t.Errorf("isWorkspaceCommand(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestAddBookmarkPersistsAndAppearsInMetadata(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	keepRunning, err := executeWorkspaceCommand(context.Background(), model, store, "bookmark add Whale")
	if err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected runtime to keep running")
	}

	bookmarks, err := store.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("ListBookmarks() error = %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(bookmarks))
	}
	if bookmarks[0].Title != "Whale" {
		t.Errorf("bookmark title = %q, want %q", bookmarks[0].Title, "Whale")
	}
	if bookmarks[0].Kind != "account" || bookmarks[0].Target != "GABC123" {
		t.Errorf("bookmark target = %s/%s, want account/GABC123", bookmarks[0].Kind, bookmarks[0].Target)
	}

	snapshot := model.Snapshot()
	if len(snapshot.Lookup.Metadata.Bookmarks) != 1 {
		t.Errorf("expected metadata to reflect new bookmark, got %d", len(snapshot.Lookup.Metadata.Bookmarks))
	}
	if snapshot.Lookup.Metadata.Bookmarks[0].Title != "Whale" {
		t.Errorf("metadata bookmark title = %q, want %q", snapshot.Lookup.Metadata.Bookmarks[0].Title, "Whale")
	}
}

func TestAddBookmarkDefaultsTitleToKindPlusTarget(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "bookmark add"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	bookmarks, err := store.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("ListBookmarks() error = %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(bookmarks))
	}
	if !strings.Contains(bookmarks[0].Title, "account") || !strings.Contains(bookmarks[0].Title, "GABC123") {
		t.Errorf("expected default title to contain kind+target, got %q", bookmarks[0].Title)
	}
}

func TestRemoveBookmarkDeletesMatchingBookmarks(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID: "bm-1", ProfileID: "default", Kind: "account", Target: "GABC123", Title: "test",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "bookmark remove"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	bookmarks, err := store.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("ListBookmarks() error = %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("expected 0 bookmarks after remove, got %d", len(bookmarks))
	}
}

func TestAddNotePersistsAndAppearsInMetadata(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note add Suspicious activity"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Title != "Suspicious activity" {
		t.Errorf("note title = %q, want %q", notes[0].Title, "Suspicious activity")
	}
	if !strings.Contains(notes[0].Target, "account") || !strings.Contains(notes[0].Target, "GABC123") {
		t.Errorf("note target = %q, expected account:GABC123", notes[0].Target)
	}

	snapshot := model.Snapshot()
	if len(snapshot.Lookup.Metadata.Notes) != 1 {
		t.Errorf("expected metadata to reflect new note, got %d", len(snapshot.Lookup.Metadata.Notes))
	}
}

func TestRemoveNoteDeletesMatchingNotes(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertNote(context.Background(), cache.Note{
		ID: "note-1", ProfileID: "default", Target: "account:GABC123", Title: "test",
	}); err != nil {
		t.Fatalf("UpsertNote() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note remove"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("expected 0 notes after remove, got %d", len(notes))
	}
}

func TestAddLabelCreatesLabelAndAttachesToEntity(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "label add important"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	labels, err := store.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Name != "important" {
		t.Errorf("label name = %q, want %q", labels[0].Name, "important")
	}

	targets, err := store.ListLabelTargets(context.Background())
	if err != nil {
		t.Fatalf("ListLabelTargets() error = %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 label target, got %d", len(targets))
	}
	if targets[0].Kind != "account" || targets[0].Target != "GABC123" {
		t.Errorf("label target = %s/%s, want account/GABC123", targets[0].Kind, targets[0].Target)
	}

	snapshot := model.Snapshot()
	if len(snapshot.Lookup.Metadata.Labels) != 1 {
		t.Errorf("expected metadata to reflect new label, got %d", len(snapshot.Lookup.Metadata.Labels))
	}
}

func TestAddLabelReusesExistingLabelByName(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID: "label-existing", ProfileID: "default", Name: "important",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "label add important"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	labels, err := store.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label (reused, not duplicated), got %d", len(labels))
	}
	if labels[0].ID != "label-existing" {
		t.Errorf("expected existing label to be reused, got ID %q", labels[0].ID)
	}
}

func TestRemoveLabelDetachesFromEntity(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID: "lbl-1", ProfileID: "default", Name: "flagged",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}
	if err := store.UpsertLabelTarget(context.Background(), cache.LabelTarget{
		ID: "lt-1", LabelID: "lbl-1", ProfileID: "default", Kind: "account", Target: "GABC123",
	}); err != nil {
		t.Fatalf("UpsertLabelTarget() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "label remove flagged"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	targets, err := store.ListLabelTargets(context.Background())
	if err != nil {
		t.Fatalf("ListLabelTargets() error = %v", err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected 0 label targets after remove, got %d", len(targets))
	}
}

func TestOpenRecentShowsEntityCache(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID:   "default",
		Kind:        "account",
		Target:      "GABC123",
		Title:       "account GABC123",
		Summary:     "balance 100.0000000",
		Payload:     `{}`,
		SourceLabel: "test",
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	keepRunning, err := executeWorkspaceCommand(context.Background(), model, store, "open recent")
	if err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected runtime to keep running")
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil {
		t.Fatal("expected explorer to be open after open recent")
	}
	if snapshot.Lookup.Explorer.Title != "Recently Visited" {
		t.Errorf("explorer title = %q, want %q", snapshot.Lookup.Explorer.Title, "Recently Visited")
	}
	if len(snapshot.Lookup.Explorer.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(snapshot.Lookup.Explorer.Results))
	}
	if snapshot.Lookup.Explorer.Results[0].Kind != "account" {
		t.Errorf("result kind = %q, want %q", snapshot.Lookup.Explorer.Results[0].Kind, "account")
	}
}

func TestOpenRecentWithEmptyCacheShowsInfo(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "open recent"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer != nil {
		t.Error("expected no explorer when cache is empty")
	}
	if snapshot.Status.Level != app.StatusInfo {
		t.Errorf("expected info status for empty cache, got %v", snapshot.Status.Level)
	}
}

func TestOpenBookmarksShowsSavedBookmarks(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID: "bm-1", ProfileID: "default", Kind: "account", Target: "GABC123", Title: "My Account",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	keepRunning, err := executeWorkspaceCommand(context.Background(), model, store, "open bookmarks")
	if err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected runtime to keep running")
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil {
		t.Fatal("expected explorer to be open after open bookmarks")
	}
	if snapshot.Lookup.Explorer.Title != "Bookmarks" {
		t.Errorf("explorer title = %q, want %q", snapshot.Lookup.Explorer.Title, "Bookmarks")
	}
	if len(snapshot.Lookup.Explorer.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(snapshot.Lookup.Explorer.Results))
	}
	if snapshot.Lookup.Explorer.Results[0].Title != "My Account" {
		t.Errorf("result title = %q, want %q", snapshot.Lookup.Explorer.Results[0].Title, "My Account")
	}
}

func TestOpenNotesShowsSavedNotes(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertNote(context.Background(), cache.Note{
		ID: "note-1", ProfileID: "default", Target: "account:GABC123", Title: "Investigation",
	}); err != nil {
		t.Fatalf("UpsertNote() error = %v", err)
	}

	keepRunning, err := executeWorkspaceCommand(context.Background(), model, store, "open notes")
	if err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected runtime to keep running")
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil {
		t.Fatal("expected explorer to be open after open notes")
	}
	if snapshot.Lookup.Explorer.Title != "Notes" {
		t.Errorf("explorer title = %q, want %q", snapshot.Lookup.Explorer.Title, "Notes")
	}
	if len(snapshot.Lookup.Explorer.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(snapshot.Lookup.Explorer.Results))
	}
	if snapshot.Lookup.Explorer.Results[0].Title != "Investigation" {
		t.Errorf("result title = %q, want %q", snapshot.Lookup.Explorer.Results[0].Title, "Investigation")
	}
}

func TestOpenLabelsShowsLabelsAndTargets(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID: "lbl-1", ProfileID: "default", Name: "whale",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}
	if err := store.UpsertLabelTarget(context.Background(), cache.LabelTarget{
		ID: "lt-1", LabelID: "lbl-1", ProfileID: "default", Kind: "account", Target: "GABC123",
	}); err != nil {
		t.Fatalf("UpsertLabelTarget() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "open labels"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil {
		t.Fatal("expected explorer to be open after open labels")
	}
	if snapshot.Lookup.Explorer.Title != "Labels" {
		t.Errorf("explorer title = %q, want %q", snapshot.Lookup.Explorer.Title, "Labels")
	}
	if len(snapshot.Lookup.Explorer.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(snapshot.Lookup.Explorer.Results))
	}
	if snapshot.Lookup.Explorer.Results[0].Title != "whale" {
		t.Errorf("result title = %q, want %q", snapshot.Lookup.Explorer.Results[0].Title, "whale")
	}
	if snapshot.Lookup.Explorer.BackCommand != "open labels" {
		t.Errorf("back command = %q, want %q", snapshot.Lookup.Explorer.BackCommand, "open labels")
	}
}

func TestOpenWorkspaceBrowsersHaveBackCommand(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID: "bm-1", ProfileID: "default", Kind: "account", Target: "GABC123", Title: "test",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	for _, cmd := range []string{"open bookmarks", "open notes", "open labels"} {
		if cmd == "open notes" {
			if err := store.UpsertNote(context.Background(), cache.Note{
				ID: "n-1", ProfileID: "default", Target: "account:GABC123", Title: "t",
			}); err != nil {
				t.Fatalf("UpsertNote: %v", err)
			}
		}
		if cmd == "open labels" {
			if err := store.UpsertLabel(context.Background(), cache.Label{
				ID: "l-1", ProfileID: "default", Name: "x",
			}); err != nil {
				t.Fatalf("UpsertLabel: %v", err)
			}
		}

		if _, err := executeWorkspaceCommand(context.Background(), model, store, cmd); err != nil {
			t.Fatalf("%s: executeWorkspaceCommand() error = %v", cmd, err)
		}
		snapshot := model.Snapshot()
		if snapshot.Lookup.Explorer == nil {
			t.Fatalf("%s: expected explorer to be open", cmd)
		}
		wantBack := cmd
		if snapshot.Lookup.Explorer.BackCommand != wantBack {
			t.Errorf("%s: back command = %q, want %q", cmd, snapshot.Lookup.Explorer.BackCommand, wantBack)
		}
	}
}

func TestNoteAddWithBodySeparator(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note add Investigation | Whale bought 1M XLM"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Title != "Investigation" {
		t.Errorf("note title = %q, want %q", notes[0].Title, "Investigation")
	}
	if notes[0].Body != "Whale bought 1M XLM" {
		t.Errorf("note body = %q, want %q", notes[0].Body, "Whale bought 1M XLM")
	}
}

func TestNoteRemoveWithTitleFilterOnlyDeletesMatching(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	for _, note := range []cache.Note{
		{ID: "n-1", ProfileID: "default", Target: "account:GABC123", Title: "Whale activity"},
		{ID: "n-2", ProfileID: "default", Target: "account:GABC123", Title: "Routine check"},
	} {
		if err := store.UpsertNote(context.Background(), note); err != nil {
			t.Fatalf("UpsertNote() error = %v", err)
		}
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note remove whale"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note after filtered remove, got %d", len(notes))
	}
	if notes[0].Title != "Routine check" {
		t.Errorf("remaining note = %q, want %q", notes[0].Title, "Routine check")
	}
}

func TestBookmarkNoteAnnotatesMatchingBookmark(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID: "bm-1", ProfileID: "default", Kind: "account", Target: "GABC123", Title: "Whale",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "bookmark note Suspicious large holder"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	bookmarks, err := store.ListBookmarks(context.Background())
	if err != nil {
		t.Fatalf("ListBookmarks() error = %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("expected 1 bookmark, got %d", len(bookmarks))
	}
	if bookmarks[0].Notes != "Suspicious large holder" {
		t.Errorf("bookmark notes = %q, want %q", bookmarks[0].Notes, "Suspicious large holder")
	}
}

func TestLabelColorSetsColorField(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID: "lbl-1", ProfileID: "default", Name: "important",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "label color important red"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	labels, err := store.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Color != "red" {
		t.Errorf("label color = %q, want %q", labels[0].Color, "red")
	}
}

func TestOpenWorkspaceBrowsersRespectProfile(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	// Insert records belonging to a different profile — should not appear
	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID: "bm-other", ProfileID: "other-profile", Kind: "account", Target: "GABC123", Title: "Other",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "open bookmarks"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer != nil {
		t.Error("expected no explorer when all bookmarks belong to another profile")
	}
	if snapshot.Status.Level != app.StatusInfo {
		t.Errorf("expected info status, got %v", snapshot.Status.Level)
	}
}

func TestNoteBodyUpdatesExistingNote(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if err := store.UpsertNote(context.Background(), cache.Note{
		ID: "note-1", ProfileID: "default", Target: "account:GABC123", Title: "Investigation",
	}); err != nil {
		t.Fatalf("UpsertNote() error = %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note body Suspicious whale activity"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	notes, err := store.ListNotes(context.Background())
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if notes[0].Body != "Suspicious whale activity" {
		t.Errorf("note body = %q, want %q", notes[0].Body, "Suspicious whale activity")
	}
}

func TestNoteBodyWithNoExistingNoteWarns(t *testing.T) {
	store := openTestWorkspaceStore(t)
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, store, "note body some text"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if snapshot.Status.Level != app.StatusWarn {
		t.Errorf("expected warning when no note exists, got %v", snapshot.Status.Level)
	}
}

func TestWorkspaceCommandRequiresActiveLookup(t *testing.T) {
	store := openTestWorkspaceStore(t)
	cfg := config.Default()
	model := app.NewModel(cfg, "/tmp/config.json", app.CacheSnapshot{})

	for _, cmd := range []string{"bookmark add", "bookmark remove", "note add", "note remove", "label add x", "label remove x"} {
		if _, err := executeWorkspaceCommand(context.Background(), model, store, cmd); err != nil {
			t.Fatalf("executeWorkspaceCommand(%q) error = %v", cmd, err)
		}
		snapshot := model.Snapshot()
		if snapshot.Status.Level != app.StatusWarn {
			t.Errorf("executeWorkspaceCommand(%q): expected warning status on home screen, got %v", cmd, snapshot.Status.Level)
		}
	}
}

func TestWorkspaceCommandWithNilStoreWarns(t *testing.T) {
	model := newLookupReadyModel(t)

	if _, err := executeWorkspaceCommand(context.Background(), model, nil, "bookmark add test"); err != nil {
		t.Fatalf("executeWorkspaceCommand() error = %v", err)
	}
	snapshot := model.Snapshot()
	if snapshot.Status.Level != app.StatusWarn {
		t.Errorf("expected warning status with nil store, got %v", snapshot.Status.Level)
	}
}

func openTestWorkspaceStore(t *testing.T) *cache.Store {
	t.Helper()
	driverName := registerFakeSQLiteDriver(t)
	store, err := cache.OpenSQLite(context.Background(), driverName, "workspace-test")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func newLookupReadyModel(t *testing.T) *app.Model {
	t.Helper()
	cfg := config.Default()
	model := app.NewModel(cfg, "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount("GABC123", backendclient.AccountLookupResponse{
		Account: &backendclient.AccountDetail{
			ID:      "GABC123",
			Balance: "100.0000000",
		},
	})
	return model
}
