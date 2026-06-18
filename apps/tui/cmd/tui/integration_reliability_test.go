//go:build integration

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/ui"
)

func TestIntegrationReliabilityNavigationAndRenderChain(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-789"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-chain-1", LedgerSequence: 789, OperationCount: 1},
			},
		},
		app.DefaultSourceMetadata(config.Default().Profiles[0], "ledger"),
	)

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open txs"); err != nil {
		t.Fatalf("open txs: %v", err)
	}
	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{
		transaction: backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-chain-1", LedgerSequence: 789},
		},
	}, "open tx 1"); err != nil {
		t.Fatalf("open tx 1: %v", err)
	}
	if !model.Back() {
		t.Fatal("expected back to transaction explorer")
	}
	if !model.Back() {
		t.Fatal("expected back to ledger detail")
	}

	view := ui.RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Route:", "ledger 789", "State: ready"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected restored ledger render to include %q, got %q", needle, view)
		}
	}
}
