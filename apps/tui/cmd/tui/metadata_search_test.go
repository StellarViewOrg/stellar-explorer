package main

import (
	"context"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
)

func TestLocalMetadataSearchResultsFindsExecutableBookmark(t *testing.T) {
	store := openTestMetadataStore(t)
	if err := store.UpsertBookmark(context.Background(), cache.Bookmark{
		ID:        "bookmark-1",
		ProfileID: "default",
		Kind:      "account",
		Target:    "GABC",
		Title:     "Whale account",
		Notes:     "watch closely",
	}); err != nil {
		t.Fatalf("UpsertBookmark() error = %v", err)
	}

	results, err := localMetadataSearchResults(context.Background(), store, "whale", "default", 8)
	if err != nil {
		t.Fatalf("localMetadataSearchResults() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].Command != "lookup account GABC" || !results[0].Enabled {
		t.Fatalf("expected executable bookmark result, got %#v", results[0])
	}
}

func TestLocalMetadataSearchResultsFindsPendingLabel(t *testing.T) {
	store := openTestMetadataStore(t)
	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID:        "label-1",
		ProfileID: "default",
		Name:      "suspicious",
		Color:     "red",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}

	results, err := localMetadataSearchResults(context.Background(), store, "suspicious", "default", 8)
	if err != nil {
		t.Fatalf("localMetadataSearchResults() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].Kind != "label" || results[0].Enabled {
		t.Fatalf("expected pending label result, got %#v", results[0])
	}
}

func TestLocalMetadataSearchResultsFindsExecutableLabelTarget(t *testing.T) {
	store := openTestMetadataStore(t)
	if err := store.UpsertLabel(context.Background(), cache.Label{
		ID:        "label-1",
		ProfileID: "default",
		Name:      "suspicious",
		Color:     "red",
	}); err != nil {
		t.Fatalf("UpsertLabel() error = %v", err)
	}
	if err := store.UpsertLabelTarget(context.Background(), cache.LabelTarget{
		ID:        "label-target-1",
		LabelID:   "label-1",
		ProfileID: "default",
		Kind:      "account",
		Target:    "GABC",
	}); err != nil {
		t.Fatalf("UpsertLabelTarget() error = %v", err)
	}

	results, err := localMetadataSearchResults(context.Background(), store, "suspicious", "default", 8)
	if err != nil {
		t.Fatalf("localMetadataSearchResults() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].Command != "lookup account GABC" || !results[0].Enabled {
		t.Fatalf("expected executable label result, got %#v", results[0])
	}
}

func TestLocalMetadataSearchResultsFindsCachedEntity(t *testing.T) {
	store := openTestMetadataStore(t)
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID: "default",
		Kind:      "account",
		Target:    "GWHALE",
		Title:     "account GWHALE",
		Summary:   "balance 100",
		Payload:   `{}`,
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	results, err := localMetadataSearchResults(context.Background(), store, "cache:whale", "default", 8)
	if err != nil {
		t.Fatalf("localMetadataSearchResults() error = %v", err)
	}
	if len(results) != 1 || results[0].Kind != "cache" || results[0].Command != "lookup account GWHALE" {
		t.Fatalf("expected cached entity result, got %#v", results)
	}
}

func TestLookupCommandForLocalTargetInfersNoteTarget(t *testing.T) {
	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if got := lookupCommandForLocalTarget("", hash); got != "lookup tx "+hash {
		t.Fatalf("expected tx command, got %q", got)
	}
	if got := lookupCommandForLocalTarget("", "USDC:GISS"); got != "lookup asset USDC:GISS" {
		t.Fatalf("expected asset command, got %q", got)
	}
}

func openTestMetadataStore(t *testing.T) *cache.Store {
	t.Helper()
	driverName := registerFakeSQLiteDriver(t)
	store, err := cache.OpenSQLite(context.Background(), driverName, "metadata-search")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}
