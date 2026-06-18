package ui

import (
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestRenderBreadcrumbsIncludesExplorerStep(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
		},
	)
	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-ledger-1", LedgerSequence: 789},
	}, 10, 0)

	crumbs := renderBreadcrumbs(model.Snapshot(), 120)
	if !strings.Contains(crumbs, "ledger 789") {
		t.Fatalf("expected ledger step in breadcrumbs, got %q", crumbs)
	}
	if !strings.Contains(crumbs, "ledger transactions") {
		t.Fatalf("expected explorer step in breadcrumbs, got %q", crumbs)
	}
}

func TestLookupSectionsExposeCommandHintCopyForTraversableRows(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-deep-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-deep-1", LedgerSequence: 77, Account: "GSOURCE"},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-deep-1", TypeName: "payment", Details: "{}"},
			},
		},
	)

	sections := lookupSections(model.Snapshot().Lookup)
	found := false
	for _, section := range sections {
		if section.Title != "Op 1" {
			continue
		}
		found = true
		if section.Command != "open op 1" {
			t.Fatalf("expected open op command, got %q", section.Command)
		}
		if strings.TrimSpace(section.Hint) == "" || strings.TrimSpace(section.Copy) == "" {
			t.Fatalf("expected traversable operation row to include hint and copy, got %#v", section)
		}
	}
	if !found {
		t.Fatalf("expected operation row in lookup sections, got %#v", sections)
	}
}

func TestLookupModelWithSnapshotRestoresSelection(t *testing.T) {
	snapshot := app.Snapshot{
		Lookup: app.LookupSnapshot{
			Query:           "789",
			Kind:            app.LookupLedger,
			State:           app.ViewStateReady,
			SelectedSection: 3,
			ScrollOffset:    2,
			Ledger: &backendclient.LedgerLookupResponse{
				Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
				Transactions: []backendclient.TransactionSummary{
					{Hash: "tx-1", LedgerSequence: 789},
					{Hash: "tx-2", LedgerSequence: 789},
				},
			},
		},
	}

	model := NewLookupModel(snapshot, 120, 14).WithSnapshot(snapshot)
	if model.selectedSection != 3 {
		t.Fatalf("expected selected section 3, got %d", model.selectedSection)
	}
	if model.offset != 2 {
		t.Fatalf("expected scroll offset 2 with compact viewport, got %d", model.offset)
	}
}
