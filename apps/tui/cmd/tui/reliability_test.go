package main

import (
	"context"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestLedgerPaginationChainAdvancesBeforeCursor(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})

	backend := stubExplorerBackend{
		stubLookupBackend: stubLookupBackend{},
		ledgers: []backendclient.LedgerSummary{
			{Sequence: 56, Hash: "ledger-56"},
			{Sequence: 55, Hash: "ledger-55"},
		},
	}
	if _, err := executeOpenCommand(context.Background(), model, backend, "open ledgers limit 2 before 57"); err != nil {
		t.Fatalf("open ledgers page 1: %v", err)
	}
	next := model.Snapshot().Lookup.Explorer.Results[2].Command
	if next != "open ledgers limit 2 before 55" {
		t.Fatalf("unexpected next page command %q", next)
	}

	backend.ledgers = []backendclient.LedgerSummary{
		{Sequence: 54, Hash: "ledger-54"},
		{Sequence: 53, Hash: "ledger-53"},
	}
	if _, err := executeOpenCommand(context.Background(), model, backend, next); err != nil {
		t.Fatalf("open ledgers page 2: %v", err)
	}
	results := model.Snapshot().Lookup.Explorer.Results
	if len(results) < 2 || results[0].Command != "lookup ledger 54" {
		t.Fatalf("expected second page ledgers, got %#v", results)
	}
}

func TestTimelinePaginationUsesOffset(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
		},
	)

	backend := stubExplorerBackend{
		stubLookupBackend: stubLookupBackend{},
		accountTimeline: []backendclient.TimelineItem{
			{Kind: "tx", Title: "Transaction tx-2", Description: "ledger 2", Command: "lookup tx tx-2"},
		},
	}
	if _, err := executeOpenCommand(context.Background(), model, backend, "open timeline limit 1 offset 1"); err != nil {
		t.Fatalf("open timeline offset page: %v", err)
	}
	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.ListOffset != 1 {
		t.Fatalf("expected timeline offset 1, got %#v", explorer)
	}
	if len(explorer.Results) != 1 || explorer.Results[0].Title != "Transaction tx-2" {
		t.Fatalf("expected offset timeline row, got %#v", explorer.Results)
	}
}

func TestHybridTransactionFallbackUsesRPCPayload(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			err:   backendclient.HTTPError{StatusCode: 503, Message: "indexer unavailable"},
		},
		testLookupBackend{
			label: "http://rpc.test",
			transaction: backendclient.TransactionLookupResponse{
				Transaction: &backendclient.TransactionDetail{Hash: "tx-fallback", LedgerSequence: 99},
			},
		},
	)

	response, err := backend.Transaction(context.Background(), "tx-fallback")
	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}
	if response.Transaction == nil || response.Transaction.Hash != "tx-fallback" {
		t.Fatalf("expected fallback transaction, got %#v", response)
	}
	meta := backend.SourceMetadata()
	if !meta.FallbackUsed || meta.Actual != "rpc" || !meta.Degraded {
		t.Fatalf("expected rpc fallback metadata, got %#v", meta)
	}
}

func TestHybridPrimarySuccessDoesNotMarkDegraded(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			account: backendclient.AccountLookupResponse{
				Account: &backendclient.AccountDetail{ID: "GACC", Balance: "12"},
			},
		},
		testLookupBackend{label: "http://rpc.test"},
	)

	response, err := backend.Account(context.Background(), "GACC")
	if err != nil {
		t.Fatalf("Account() error = %v", err)
	}
	if response.Account == nil || response.Account.ID != "GACC" {
		t.Fatalf("expected indexer account payload, got %#v", response)
	}
	meta := backend.SourceMetadata()
	if meta.FallbackUsed || meta.Degraded || meta.Actual != "indexer" {
		t.Fatalf("expected clean indexer metadata, got %#v", meta)
	}
}

func TestProfileWorkspaceIsolationIncludesWatchSettings(t *testing.T) {
	store := openTestWorkspaceStore(t)
	defaultModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	opsCfg := configWithProfile("ops")
	otherModel := app.NewModel(opsCfg, "/tmp/config.json", app.CacheSnapshot{})

	_ = defaultModel.SetScreen(app.ScreenLiveFeed)
	if _, err := executeWatchCommand(context.Background(), config.Default(), defaultModel, store, "watch save default-watch"); err != nil {
		t.Fatalf("save default watch: %v", err)
	}
	_ = otherModel.SetScreen(app.ScreenLiveFeed)
	if _, err := executeWatchCommand(context.Background(), opsCfg, otherModel, store, "watch save ops-watch"); err != nil {
		t.Fatalf("save ops watch: %v", err)
	}

	if _, err := executeWorkspaceCommand(context.Background(), defaultModel, store, "open watches"); err != nil {
		t.Fatalf("open watches default: %v", err)
	}
	for _, result := range defaultModel.Snapshot().Lookup.Explorer.Results {
		if result.Title == "ops-watch" {
			t.Fatalf("default profile should not see ops watch preset: %#v", result)
		}
	}

	if _, err := executeWorkspaceCommand(context.Background(), otherModel, store, "open watches"); err != nil {
		t.Fatalf("open watches ops: %v", err)
	}
	found := false
	for _, result := range otherModel.Snapshot().Lookup.Explorer.Results {
		if result.Title == "ops-watch" {
			found = true
		}
		if result.Title == "default-watch" {
			t.Fatal("ops profile should not see default watch preset")
		}
	}
	if !found {
		t.Fatal("expected ops watch preset in ops profile browser")
	}
}

func configWithProfile(name string) config.Config {
	cfg := config.Default()
	cfg.DefaultProfile = name
	cfg.Profiles = []config.Profile{
		{
			Name:        name,
			Network:     "testnet",
			RPCEndpoint: "https://soroban-testnet.stellar.org",
			BackendMode: config.BackendModeRPC,
		},
	}
	return cfg
}
