package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

type failingLookupBackend struct {
	err error
}

func (f failingLookupBackend) Label() string { return "failing-backend" }
func (f failingLookupBackend) Search(context.Context, string, int) (backendclient.SearchResponse, error) {
	return backendclient.SearchResponse{}, f.err
}
func (f failingLookupBackend) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return backendclient.LiveFeedSummaryResponse{}, f.err
}
func (f failingLookupBackend) Ledger(context.Context, uint32) (backendclient.LedgerLookupResponse, error) {
	return backendclient.LedgerLookupResponse{}, f.err
}
func (f failingLookupBackend) Transaction(context.Context, string) (backendclient.TransactionLookupResponse, error) {
	return backendclient.TransactionLookupResponse{}, f.err
}
func (f failingLookupBackend) Account(context.Context, string) (backendclient.AccountLookupResponse, error) {
	return backendclient.AccountLookupResponse{}, f.err
}
func (f failingLookupBackend) Asset(context.Context, string, string) (backendclient.AssetLookupResponse, error) {
	return backendclient.AssetLookupResponse{}, f.err
}
func (f failingLookupBackend) Contract(context.Context, string) (backendclient.ContractLookupResponse, error) {
	return backendclient.ContractLookupResponse{}, f.err
}

func TestPerformTransactionLookupUsesFreshCacheBeforeBackend(t *testing.T) {
	store := openTestMetadataStore(t)
	payload, err := json.Marshal(backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{Hash: "tx-hit", LedgerSequence: 99},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID:   "default",
		Kind:        "transaction",
		Target:      "tx-hit",
		Payload:     string(payload),
		SourceLabel: "tui-indexer",
		UpdatedAt:   time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	backend := failingLookupBackend{err: errors.New("network down")}

	keepRunning, err := performTransactionLookup(context.Background(), model, store, backend, "tx-hit", lookupCacheOptions{})
	if err != nil || !keepRunning {
		t.Fatalf("performTransactionLookup() = (%v, %v)", keepRunning, err)
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Transaction == nil || snapshot.Lookup.Transaction.Transaction.Hash != "tx-hit" {
		t.Fatalf("expected cached transaction, got %#v", snapshot.Lookup.Transaction)
	}
	if snapshot.Lookup.Source.CacheState != "hit" || snapshot.Lookup.Source.Degraded {
		t.Fatalf("expected fresh cache hit metadata, got %#v", snapshot.Lookup.Source)
	}
}

func openTestSQLiteStore(t *testing.T) *cache.Store {
	t.Helper()
	store, err := cache.OpenSQLite(context.Background(), "sqlite", filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func TestPerformAccountLookupFallsBackToStaleCache(t *testing.T) {
	store := openTestSQLiteStore(t)
	payload, err := json.Marshal(backendclient.AccountLookupResponse{
		Account: &backendclient.AccountDetail{ID: "GABC", Balance: "42"},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID: "default",
		Kind:      "account",
		Target:    "GABC",
		Payload:   string(payload),
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}
	staleAt := time.Now().UTC().Add(-30 * time.Minute)
	if _, err := store.DB().ExecContext(context.Background(), `
UPDATE entity_cache SET updated_at = ? WHERE profile_id = ? AND kind = ? AND target = ?;`,
		staleAt, "default", "account", "GABC"); err != nil {
		t.Fatalf("age cache row: %v", err)
	}

	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	backend := failingLookupBackend{err: errors.New("network down")}

	keepRunning, err := performAccountLookup(context.Background(), model, store, backend, "GABC", lookupCacheOptions{})
	if err != nil || !keepRunning {
		t.Fatalf("performAccountLookup() = (%v, %v)", keepRunning, err)
	}

	snapshot := model.Snapshot()
	if snapshot.Lookup.Account == nil || snapshot.Lookup.Account.Account.Balance != "42" {
		t.Fatalf("expected stale cached account, got %#v", snapshot.Lookup.Account)
	}
	if snapshot.Lookup.Source.CacheState != "stale" || !snapshot.Lookup.Source.Degraded {
		t.Fatalf("expected stale cache metadata, got %#v", snapshot.Lookup.Source)
	}
}

func TestRefreshActiveLookupBypassesCache(t *testing.T) {
	store := openTestMetadataStore(t)
	payload, err := json.Marshal(backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{Hash: "tx-refresh", LedgerSequence: 1},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := store.UpsertEntityCache(context.Background(), cache.EntityCache{
		ProfileID: "default",
		Kind:      "transaction",
		Target:    "tx-refresh",
		Payload:   string(payload),
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("UpsertEntityCache() error = %v", err)
	}

	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction("tx-refresh", backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{Hash: "tx-refresh", LedgerSequence: 1},
	})
	_ = model.SetScreen(app.ScreenLookup)

	previousFactory := openLookupBackend
	t.Cleanup(func() {
		openLookupBackend = previousFactory
	})
	openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
		return fakeLookupBackend{
			transaction: backendclient.TransactionLookupResponse{
				Transaction: &backendclient.TransactionDetail{Hash: "tx-refresh", LedgerSequence: 77},
			},
		}, nil
	}

	refreshActiveLookup(context.Background(), config.Default(), model, store)

	snapshot := model.Snapshot()
	if snapshot.Lookup.Transaction == nil || snapshot.Lookup.Transaction.Transaction.LedgerSequence != 77 {
		t.Fatalf("expected refreshed transaction payload, got %#v", snapshot.Lookup.Transaction)
	}
	if snapshot.Lookup.Source.CacheState == "hit" {
		t.Fatalf("expected network refresh, got cache hit metadata %#v", snapshot.Lookup.Source)
	}
}
