package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestTryRestoreLookupFromCacheLoadsTransaction(t *testing.T) {
	store := openTestMetadataStore(t)
	payload, err := json.Marshal(backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{Hash: "tx-cache", LedgerSequence: 42},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID:   "default",
		Kind:        "transaction",
		Target:      "tx-cache",
		Title:       "transaction tx-cache",
		Summary:     "ledger 42 / 1 ops",
		Payload:     string(payload),
		SourceLabel: "tui-indexer",
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if !tryRestoreLookupFromCache(context.Background(), model, store, app.LookupTransaction, "tx-cache", lookupCacheHit) {
		t.Fatal("expected cache restore to succeed")
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Transaction == nil || snapshot.Lookup.Transaction.Transaction == nil {
		t.Fatal("expected cached transaction lookup")
	}
	if snapshot.Lookup.Source.CacheState != "hit" || snapshot.Lookup.Source.Actual != "cache" {
		t.Fatalf("expected cache hit source metadata, got %#v", snapshot.Lookup.Source)
	}
}

func TestExecuteOpenCacheCommandUsesActiveLookup(t *testing.T) {
	store := openTestMetadataStore(t)
	payload, err := json.Marshal(backendclient.AccountLookupResponse{
		Account: &backendclient.AccountDetail{ID: "GABC", Balance: "10"},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID: "default",
		Kind:      "account",
		Target:    "GABC",
		Payload:   string(payload),
		Summary:   "balance 10",
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount("GABC", backendclient.AccountLookupResponse{
		Account: &backendclient.AccountDetail{ID: "GABC", Balance: "1"},
	})

	if _, err := executeOpenCacheCommand(context.Background(), model, store, nil); err != nil {
		t.Fatalf("executeOpenCacheCommand() error = %v", err)
	}
	if model.Snapshot().Lookup.Account == nil || model.Snapshot().Lookup.Account.Account.Balance != "10" {
		t.Fatalf("expected cached account payload, got %#v", model.Snapshot().Lookup.Account)
	}
}
