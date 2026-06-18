package main

import (
	"context"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestSaveCurrentViewPersistsLookupContext(t *testing.T) {
	store := openTestMetadataStore(t)
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction("tx-view", backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{Hash: "tx-view"},
	})

	if err := saveCurrentView(context.Background(), model, store, "investigation"); err != nil {
		t.Fatalf("saveCurrentView() error = %v", err)
	}

	views, err := store.ListSavedViews(context.Background(), "default")
	if err != nil {
		t.Fatalf("ListSavedViews() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected one saved view, got %d", len(views))
	}
	if views[0].Command != "lookup tx tx-view" {
		t.Fatalf("expected lookup command to be saved, got %q", views[0].Command)
	}
	if views[0].EntityKind != "transaction" || views[0].EntityTarget != "tx-view" {
		t.Fatalf("unexpected saved entity context: %#v", views[0])
	}
}
