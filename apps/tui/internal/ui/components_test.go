package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestLiveFeedModelEmitsSelectionAction(t *testing.T) {
	model := NewLiveFeedModel(app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{}).Snapshot(), 100, 20)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd == nil {
		t.Fatal("expected live feed key update to return an action command")
	}

	action, ok := cmd().(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", cmd())
	}
	if action.Kind != ActionMoveSelection || action.Delta != 1 {
		t.Fatalf("unexpected action %#v", action)
	}
}

func TestCommandPaletteModelEmitsAppendAction(t *testing.T) {
	snapshot := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{}).Snapshot()
	model := NewCommandPaletteModel(snapshot, 100)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd == nil {
		t.Fatal("expected command palette key update to return an action command")
	}

	action, ok := cmd().(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", cmd())
	}
	if action.Kind != ActionAppendCommandInput || action.Text != "x" {
		t.Fatalf("unexpected action %#v", action)
	}
}

func TestCommandPaletteModelEmitsPasteClipboardAction(t *testing.T) {
	snapshot := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{}).Snapshot()
	model := NewCommandPaletteModel(snapshot, 100).Open("Search / Command", "")

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlV})
	if cmd == nil {
		t.Fatal("expected clipboard paste action command")
	}

	action, ok := cmd().(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", cmd())
	}
	if action.Kind != ActionPasteClipboard {
		t.Fatalf("unexpected action %#v", action)
	}
}

func TestLiveFeedModelKeepsSelectionVisibleWithOffset(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if err := appModel.SetScreen(app.ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}
	if err := appModel.RefreshLiveFeed(context.Background(), fakeComponentLiveFeedService{}); err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	model := NewLiveFeedModel(appModel.Snapshot(), 100, 4)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd == nil {
		t.Fatal("expected action command")
	}
	model = updated.(LiveFeedModel)
	if model.offset != 1 {
		t.Fatalf("expected offset to advance, got %d", model.offset)
	}
}

func TestLiveFeedModelTracksCopyField(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if err := appModel.SetScreen(app.ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}
	if err := appModel.RefreshLiveFeed(context.Background(), fakeComponentLiveFeedService{}); err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	model := NewLiveFeedModel(appModel.Snapshot(), 100, 8)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(LiveFeedModel)
	if got := model.ClipboardValue(); got != "0" {
		t.Fatalf("ClipboardValue() = %q, want %q", got, "0")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(LiveFeedModel)
	if got := model.ClipboardValue(); got != "" {
		t.Fatalf("ClipboardValue() = %q, want empty account", got)
	}
}

func TestLiveFeedModelSupportsPageAndJumpNavigation(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	if err := appModel.SetScreen(app.ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}
	if err := appModel.RefreshLiveFeed(context.Background(), fakeComponentLiveFeedService{}); err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	model := NewLiveFeedModel(appModel.Snapshot(), 100, 4)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	model = updated.(LiveFeedModel)
	if model.selected <= 0 {
		t.Fatalf("expected page navigation to advance selection, got %d", model.selected)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model = updated.(LiveFeedModel)
	if model.selected != len(appModel.Snapshot().LiveFeed.RecentTransactions)-1 {
		t.Fatalf("expected end to jump to last row, got %d", model.selected)
	}
}

func TestLookupModelMaintainsOwnSectionSelection(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.UpdateLookupTransaction(
		"tx-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-1",
				LedgerSequence: 10,
				OperationCount: 2,
				Status:         1,
				Account:        "GTEST",
				IsSoroban:      true,
			},
			Operations: []backendclient.OperationSummary{
				{TypeName: "invoke_host_function"},
			},
		},
	)

	model := NewLookupModel(appModel.Snapshot(), 100, 6)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(LookupModel)
	if model.selectedSection == 0 {
		t.Fatal("expected local lookup selection to move")
	}
}

func TestLookupModelSkipsDividerSectionsWhenMoving(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.UpdateLookupLedger(
		"10",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{
				Sequence:          10,
				Hash:              "ledger-10",
				TransactionCount:  2,
				OperationCount:    3,
				SuccessfulTxCount: 2,
				FailedTxCount:     0,
			},
		},
	)

	model := NewLookupModel(appModel.Snapshot(), 100, 8)
	if model.selectedSection != 2 {
		t.Fatalf("expected first selectable section to be 2, got %d", model.selectedSection)
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(LookupModel)
	if model.selectedSection != 3 {
		t.Fatalf("expected next selectable section to be 3, got %d", model.selectedSection)
	}
}

func TestLookupModelSupportsPageNavigation(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.UpdateLookupContract(
		"contract-1",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{
				ContractID:        "contract-1",
				ContractType:      1,
				StorageEntryCount: 2,
				InvocationCount:   5,
				EventCount:        7,
			},
		},
	)

	model := NewLookupModel(appModel.Snapshot(), 100, 6)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	model = updated.(LookupModel)
	if model.selectedSection <= 1 {
		t.Fatalf("expected page navigation to advance selection, got %d", model.selectedSection)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyHome})
	model = updated.(LookupModel)
	if model.selectedSection != 2 {
		t.Fatalf("expected home to jump to first selectable section, got %d", model.selectedSection)
	}
}

func TestLookupModelExposesSelectedActionCommand(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.UpdateLookupAsset(
		"USDC:GISS",
		backendclient.AssetLookupResponse{
			Asset: &backendclient.AssetDetail{
				AssetCode:     "USDC",
				AssetIssuer:   "GISS",
				SACContractID: stringPtr("CSAC"),
			},
		},
	)

	model := NewLookupModel(appModel.Snapshot(), 100, 10)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(LookupModel)
	if got := model.ActionCommand(); got != "lookup account GISS" {
		t.Fatalf("ActionCommand() = %q, want issuer command", got)
	}
}

func TestLookupModelSelectsOperationDetailCommand(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{
				ID:      "GACC",
				Balance: "10",
			},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", ApplicationOrder: 2, TypeName: "manage_data", Details: "{}"},
			},
		},
	)

	model := NewLookupModel(appModel.Snapshot(), 100, 40)
	for i := 0; i < 20 && model.ActionCommand() != "lookup op tx-op-1:2"; i++ {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(LookupModel)
	}

	if got := model.ActionCommand(); got != "lookup op tx-op-1:2" {
		t.Fatalf("ActionCommand() = %q, want operation detail command", got)
	}
}

func TestDashboardFocusOrderIncludesSidebarOnWideLayouts(t *testing.T) {
	model := NewDashboardModel(app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{}).Snapshot(), 100, 20)

	order := model.FocusOrder()
	if len(order) != 3 {
		t.Fatalf("expected 3 focus areas on wide layout, got %d", len(order))
	}
	if order[1] != app.FocusSidebar {
		t.Fatalf("expected sidebar as second focus area, got %q", order[1])
	}
}

func TestDashboardFocusOrderSkipsSidebarOnNarrowLayouts(t *testing.T) {
	model := NewDashboardModel(app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{}).Snapshot(), 80, 20)

	order := model.FocusOrder()
	if len(order) != 2 {
		t.Fatalf("expected 2 focus areas on narrow layout, got %d", len(order))
	}
	if order[1] != app.FocusStatus {
		t.Fatalf("expected status as second focus area, got %q", order[1])
	}
}

func TestSidebarModelEmitsScreenActionOnEnter(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	snapshot := appModel.Snapshot()
	model := NewSidebarModel(snapshot, 30, 20)
	model.selected = app.ScreenLookup

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected sidebar enter to return an action")
	}
	model = updated.(SidebarModel)
	if model.selected != app.ScreenLookup {
		t.Fatalf("expected lookup to stay selected, got %q", model.selected)
	}

	action, ok := cmd().(ActionMsg)
	if !ok {
		t.Fatalf("expected ActionMsg, got %T", cmd())
	}
	if action.Kind != ActionLookup {
		t.Fatalf("expected lookup action, got %#v", action)
	}
}

func TestDashboardSidebarHandlesMovementOnlyWhenSidebarFocused(t *testing.T) {
	appModel := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	appModel.FocusNextAvailable(app.FocusMain, app.FocusSidebar, app.FocusStatus)
	model := NewDashboardModel(appModel.Snapshot(), 100, 20)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd != nil {
		t.Fatalf("expected sidebar movement to stay local, got cmd %#v", cmd)
	}

	dashboard := updated.(DashboardModel)
	if dashboard.sidebar.selected != app.ScreenLiveFeed {
		t.Fatalf("expected live-feed to be selected after moving down, got %q", dashboard.sidebar.selected)
	}
}

type fakeComponentLiveFeedService struct{}

func (fakeComponentLiveFeedService) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return backendclient.LiveFeedSummaryResponse{
		RecentTransactions: []backendclient.TransactionSummary{
			{Hash: "tx-1"},
			{Hash: "tx-2"},
			{Hash: "tx-3"},
			{Hash: "tx-4"},
		},
	}, nil
}

func stringPtr(value string) *string {
	return &value
}
