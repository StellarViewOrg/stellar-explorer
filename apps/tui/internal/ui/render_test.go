package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestRenderIncludesCurrentScreen(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	view := RenderWithSize(model.Snapshot(), 140, 60)

	if !strings.Contains(view, "Screen: home") {
		t.Fatalf("expected rendered view to include current screen, got %q", view)
	}
}

func TestRenderLookupIncludesLookupResult(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount(
		"GACCOUNT",
		backendclient.AccountLookupResponse{
			Account:    &backendclient.AccountDetail{ID: "GACCOUNT", Balance: "12.5", Sequence: 44},
			Trustlines: []backendclient.TrustlineSummary{{AssetCode: "USDC"}},
			Signers:    []backendclient.AccountSignerSummary{{SignerKey: "GACCOUNT"}},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-account-1", LedgerSequence: 55, OperationCount: 2, Status: 1},
			},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-account-1", TypeName: "payment", Details: "{}"},
			},
		},
	)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	if !strings.Contains(view, "Route:") || !strings.Contains(view, "account GACCOUNT") {
		t.Fatalf("expected lookup route to render, got %q", view)
	}
	if !strings.Contains(view, "Balance       12.5") {
		t.Fatalf("expected account lookup summary to render, got %q", view)
	}
	if !strings.Contains(view, "Seq Ledger    unknown") {
		t.Fatalf("expected enriched account sequence details to render, got %q", view)
	}
	if !strings.Contains(view, "Account tx 1") || !strings.Contains(view, "Account op 1") {
		t.Fatalf("expected related account activity slices to render, got %q", view)
	}
	if strings.Index(view, "Recent Transactions") > strings.Index(view, "Relations") {
		t.Fatalf("expected recent account activity section to render before relations, got %q", view)
	}
	if strings.Index(view, "Account tx 1") > strings.Index(view, "Signer 1") {
		t.Fatalf("expected recent account rows to render before signer relations, got %q", view)
	}
}

func TestRenderLookupIncludesDeepTransactionSections(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	memo := "hello memo"
	resources := "{\"footprint\":\"rw\"}"
	model.UpdateLookupTransaction(
		"tx-deep-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:             "tx-deep-1",
				LedgerSequence:   77,
				Account:          "GSOURCE",
				AccountSequence:  1234,
				FeeCharged:       100,
				MaxFee:           200,
				OperationCount:   2,
				MemoType:         1,
				MemoText:         &memo,
				IsSoroban:        true,
				SorobanResources: &resources,
				CreatedAt:        time.Unix(1700000100, 0).UTC(),
			},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-deep-1", TypeName: "invoke_host_function", Details: "{\"fn\":\"mint\"}"},
			},
		},
	)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Summary", "Navigation", "Execution", "Operations (1)", "Memo", "Resources"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected transaction detail to include %q, got %q", needle, view)
		}
	}
	if !strings.Contains(view, "tx-deep-1  invoke_host_function") {
		t.Fatalf("expected operation summary to include transaction hash context, got %q", view)
	}
}

func TestRenderLookupIncludesTransactionEffectsForIndexedSource(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-effects-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-effects-1",
				LedgerSequence: 77,
				Account:        "GSOURCE",
				Status:         1,
			},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-effects-1", TypeName: "payment", Details: "{}"},
			},
			Effects: []backendclient.EffectSummary{
				{TypeName: "account_credited", Account: "GDEST", Details: `{"asset":"USDC"}`},
			},
		},
		app.SourceMetadata{Label: "tui-indexer"},
	)

	view := RenderWithSize(model.Snapshot(), 140, 80)
	if !strings.Contains(view, "Effects") || !strings.Contains(view, "account_credited") {
		t.Fatalf("expected indexed effects to render, got %q", view)
	}
}

func TestRenderLookupOperationDetail(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupOperation(
		"tx-op-1:1",
		backendclient.OperationLookupSnapshot{
			ParentTransactionHash: "tx-op-1",
			Operation: backendclient.OperationSummary{
				TransactionHash:  "tx-op-1",
				ApplicationOrder: 1,
				TypeName:         "payment",
				Details:          "{}",
			},
		},
	)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Route:", "operation tx-op-1:1", "payment #1", "Parent Tx", "tx-op-1"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected operation detail to include %q, got %q", needle, view)
		}
	}
}

func TestRenderLookupIncludesContractExplorerSlices(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{
				ContractID:        "CCONTRACT",
				ContractType:      0,
				StorageEntryCount: 2,
			},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-contract-1", LedgerSequence: 88, OperationCount: 1, Status: 1},
			},
			RecentEvents: []backendclient.ContractEventSummary{
				{TransactionHash: "tx-contract-1", LedgerSequence: 88, Type: 0},
			},
		},
	)

	overview := lookupSections(model.Snapshot().Lookup)
	foundWorkspace := false
	for _, section := range overview {
		if section.Command == "open txs" || section.Command == "open events" {
			foundWorkspace = true
		}
	}
	if !foundWorkspace {
		t.Fatalf("expected overview workspace shortcuts, got %#v", overview)
	}

	lookup := model.Snapshot().Lookup
	lookup.ContractTab = app.ContractWorkspaceTabTransactions
	txSections := lookupContractTabSections(lookup)
	foundTx := false
	for _, section := range txSections {
		if section.Title == "Contract tx 1" {
			foundTx = true
		}
	}
	if !foundTx {
		t.Fatalf("expected transaction tab sections, got %#v", txSections)
	}

	lookup.ContractTab = app.ContractWorkspaceTabEvents
	eventSections := lookupContractTabSections(lookup)
	foundEvent := false
	for _, section := range eventSections {
		if section.Title == "Event 1" {
			foundEvent = true
		}
	}
	if !foundEvent {
		t.Fatalf("expected events tab sections, got %#v", eventSections)
	}
}

func TestRenderLookupTransactionExplorerMode(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"789",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 789, Hash: "ledger-hash"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-1", LedgerSequence: 789, Account: "GAAA", OperationCount: 2, Status: 1},
			},
		},
	)
	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-ledger-1", LedgerSequence: 789, Account: "GAAA", OperationCount: 2, Status: 1},
	}, 10, 0)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Explorer", "Context", "Ledger Transactions", "Back", "Transactions (1)", "Tx 1"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected explorer view to include %q, got %q", needle, view)
		}
	}
	if strings.Contains(view, "Previous") {
		t.Fatalf("expected explorer mode to replace parent ledger detail, got %q", view)
	}
}

func TestRenderLookupOperationExplorerMode(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", TypeName: "payment", Details: "{}"},
			},
		},
	)
	model.OpenLookupOperationExplorer("Account Operations", "open detail", []backendclient.OperationSummary{
		{TransactionHash: "tx-op-1", TypeName: "payment", Details: "{}"},
	}, 10, 0)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Explorer", "Account Operations", "Operations (1)", "Op 1"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected operation explorer view to include %q, got %q", needle, view)
		}
	}
	if strings.Contains(view, "Balance") {
		t.Fatalf("expected explorer mode to replace parent account detail, got %q", view)
	}
}

func TestRenderLookupTimelineExplorerMode(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"contract-1",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{
				ContractID:        "contract-1",
				ContractType:      1,
				StorageEntryCount: 2,
			},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-contract-1", LedgerSequence: 22, CreatedAt: time.Unix(1715000000, 0).UTC()},
			},
			RecentEvents: []backendclient.ContractEventSummary{
				{TransactionHash: "tx-event-1", LedgerSequence: 22, Type: 1, CreatedAt: time.Unix(1715000060, 0).UTC()},
			},
		},
	)
	model.OpenLookupTimelineExplorer("Contract Timeline", "open detail", []app.SearchResult{
		{Kind: "event", Title: "Event type 1", Description: "ledger 22", Command: "lookup tx tx-event-1", Enabled: true},
		{Kind: "tx", Title: "Transaction tx-contract-1", Description: "ledger 22", Command: "lookup tx tx-contract-1", Enabled: true},
	}, 20, 0)

	view := RenderWithSize(model.Snapshot(), 140, 60)
	for _, needle := range []string{"Explorer", "Contract Timeline", "Timeline (2)", "Event 1", "Tx 2"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected timeline explorer view to include %q, got %q", needle, view)
		}
	}
	if strings.Contains(view, "Storage") {
		t.Fatalf("expected timeline mode to replace parent contract detail, got %q", view)
	}
}

func TestRenderLiveFeedIncludesBackendData(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if err := model.SetScreen(app.ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}

	err := model.RefreshLiveFeed(context.Background(), fakeRenderLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 123,
			LatestLedger: &backendclient.LedgerSummary{
				Sequence:         123,
				ClosedAt:         time.Unix(1700000000, 0).UTC(),
				TransactionCount: 2,
				OperationCount:   5,
			},
			RecentTransactions: []backendclient.TransactionSummary{
				{
					Hash:           "abcdef1234567890",
					LedgerSequence: 123,
					Account:        "GABC",
					OperationCount: 2,
					IsSoroban:      true,
					CreatedAt:      time.Unix(1700000000, 0).UTC(),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	view := Render(model.Snapshot())
	if !strings.Contains(view, "State: ready") {
		t.Fatalf("expected ready state in live feed view, got %q", view)
	}
	if !strings.Contains(view, "Last ingested ledger: 123") {
		t.Fatalf("expected live feed ledger in view, got %q", view)
	}
	if !strings.Contains(view, "abcdef123456...") {
		t.Fatalf("expected truncated transaction hash in view, got %q", view)
	}
}

func TestRenderLookupLoadingStateUsesConsistentLabel(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.SetLookupLoading(app.LookupAccount, "GACCOUNT")

	view := Render(model.Snapshot())
	if !strings.Contains(view, "State: loading") {
		t.Fatalf("expected loading state label, got %q", view)
	}
	if !strings.Contains(view, "Waiting for backend response") {
		t.Fatalf("expected loading detail, got %q", view)
	}
}

func TestRenderNarrowLayoutStillIncludesNavigationPanel(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})

	view := RenderWithSize(model.Snapshot(), 80, 24)
	if !strings.Contains(view, "Navigation") {
		t.Fatalf("expected navigation panel in narrow layout, got %q", view)
	}
	if !strings.Contains(view, "Screen: home") {
		t.Fatalf("expected sidebar content in narrow layout, got %q", view)
	}
}

func TestRenderShowsSelectionMarkerAndHelpOverlay(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if err := model.SetScreen(app.ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}

	err := model.RefreshLiveFeed(context.Background(), fakeRenderLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-a"},
				{Hash: "tx-b"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	model.MoveSelection(1)
	model.ToggleHelpOverlay()
	view := Render(model.Snapshot())
	if !strings.Contains(view, "> [2] tx-b") {
		t.Fatalf("expected selected row marker, got %q", view)
	}
	if !strings.Contains(view, "Keyboard Help") {
		t.Fatalf("expected help overlay, got %q", view)
	}
}

func TestRenderShowsCommandPaletteOverlay(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.OpenCommandPalette("Search / Command", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML")

	view := Render(model.Snapshot())
	if !strings.Contains(view, "Search / Command") {
		t.Fatalf("expected command palette title, got %q", view)
	}
	if !strings.Contains(view, "> GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML") {
		t.Fatalf("expected command palette input, got %q", view)
	}
	if !strings.Contains(view, "LOCAL / ACCOUNT") || !strings.Contains(view, "> Account:") {
		t.Fatalf("expected grouped inferred account result, got %q", view)
	}
}

type fakeRenderLiveFeedService struct {
	response backendclient.LiveFeedSummaryResponse
}

func (s fakeRenderLiveFeedService) LiveFeedSummary(_ context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return s.response, nil
}
