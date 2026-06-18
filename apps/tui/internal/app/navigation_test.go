package app

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestLookupRouteTracksLedgerToTransactions(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
		},
	)
	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-ledger-1", LedgerSequence: 789},
	}, 10, 0)

	route := model.Snapshot().LookupRoute
	if len(route) != 2 {
		t.Fatalf("expected 2 route steps, got %#v", route)
	}
	if route[0].Kind != LookupLedger || route[0].Query != "789" || route[0].ExplorerKind != "" {
		t.Fatalf("expected ledger detail route step, got %#v", route[0])
	}
	if route[1].ExplorerKind != LookupExplorerTransactions || route[1].Title != "Ledger Transactions" {
		t.Fatalf("expected transaction explorer route step, got %#v", route[1])
	}
}

func TestBackRestoresLookupSelection(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-1", LedgerSequence: 789},
				{Hash: "tx-2", LedgerSequence: 789},
			},
		},
	)
	model.SetLookupSelection(4, 1)
	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-1", LedgerSequence: 789},
		{Hash: "tx-2", LedgerSequence: 789},
	}, 10, 0)

	if !model.Back() {
		t.Fatal("expected back navigation to succeed")
	}

	snapshot := model.Snapshot().Lookup
	if snapshot.Explorer != nil {
		t.Fatalf("expected detail view after back, got explorer %#v", snapshot.Explorer)
	}
	if snapshot.SelectedSection != 4 || snapshot.ScrollOffset != 1 {
		t.Fatalf("expected restored selection 4/1, got %d/%d", snapshot.SelectedSection, snapshot.ScrollOffset)
	}
}

func TestForwardRestoresLookupExplorerSelection(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
		},
	)
	model.SetLookupSelection(2, 0)
	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-1", LedgerSequence: 789},
	}, 10, 0)
	model.SetLookupSelection(3, 1)

	if !model.Back() {
		t.Fatal("expected back navigation to succeed")
	}
	if !model.Forward() {
		t.Fatal("expected forward navigation to succeed")
	}

	snapshot := model.Snapshot().Lookup
	if snapshot.Explorer == nil || snapshot.Explorer.Kind != LookupExplorerTransactions {
		t.Fatalf("expected transaction explorer after forward, got %#v", snapshot.Explorer)
	}
	if snapshot.SelectedSection != 3 || snapshot.ScrollOffset != 1 {
		t.Fatalf("expected restored explorer selection 3/1, got %d/%d", snapshot.SelectedSection, snapshot.ScrollOffset)
	}
}
