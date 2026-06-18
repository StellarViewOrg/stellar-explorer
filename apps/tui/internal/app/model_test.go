package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

type fakeLiveFeedService struct {
	response backendclient.LiveFeedSummaryResponse
	err      error
}

func (s fakeLiveFeedService) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	if s.err != nil {
		return backendclient.LiveFeedSummaryResponse{}, s.err
	}
	return s.response, nil
}

func TestHandleCommandSwitchesScreen(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if keepRunning := model.HandleCommand("live"); !keepRunning {
		t.Fatal("expected app to keep running")
	}

	if got := model.Snapshot().Current; got != ScreenLiveFeed {
		t.Fatalf("expected current screen %q, got %q", ScreenLiveFeed, got)
	}
}

func TestHandleCommandQuitStopsLoop(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if keepRunning := model.HandleCommand("quit"); keepRunning {
		t.Fatal("expected app to stop running")
	}
}

func TestHandleCommandBackReturnsToPreviousScreen(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if err := model.SetScreen(ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}

	if keepRunning := model.HandleCommand("back"); !keepRunning {
		t.Fatal("expected app to keep running")
	}

	if got := model.Snapshot().Current; got != ScreenHome {
		t.Fatalf("expected current screen %q, got %q", ScreenHome, got)
	}
}

func TestHandleCommandForwardReturnsToNextScreen(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if err := model.SetScreen(ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen(live) error = %v", err)
	}
	if err := model.SetScreen(ScreenSettings); err != nil {
		t.Fatalf("SetScreen(settings) error = %v", err)
	}
	if keepRunning := model.HandleCommand("back"); !keepRunning {
		t.Fatal("expected app to keep running")
	}

	snapshot := model.Snapshot()
	if snapshot.Current != ScreenLiveFeed {
		t.Fatalf("expected back to land on live-feed, got %q", snapshot.Current)
	}
	if len(snapshot.Forward) != 1 {
		t.Fatalf("expected one forward entry, got %d", len(snapshot.Forward))
	}

	if keepRunning := model.HandleCommand("forward"); !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Current; got != ScreenSettings {
		t.Fatalf("expected forward to land on settings, got %q", got)
	}
}

func TestNewNavigationClearsForwardHistory(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if err := model.SetScreen(ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen(live) error = %v", err)
	}
	if err := model.SetScreen(ScreenSettings); err != nil {
		t.Fatalf("SetScreen(settings) error = %v", err)
	}
	model.Back()
	if len(model.Snapshot().Forward) != 1 {
		t.Fatal("expected forward history after back")
	}

	if err := model.SetScreen(ScreenLookup); err != nil {
		t.Fatalf("SetScreen(lookup) error = %v", err)
	}
	if len(model.Snapshot().Forward) != 0 {
		t.Fatalf("expected forward history to clear after new navigation, got %d", len(model.Snapshot().Forward))
	}
}

func TestParseScreenRejectsUnknownValue(t *testing.T) {
	if _, err := ParseScreen("unknown"); err == nil {
		t.Fatal("expected parse to fail")
	}
}

func TestBuildCacheSnapshotCapturesProfileState(t *testing.T) {
	cfg := config.Default()
	snapshot := BuildCacheSnapshot(cfg, nil, "/tmp/cache.db", 1, nil, "home", "cache ready")

	if snapshot.Path != "/tmp/cache.db" {
		t.Fatalf("expected cache path to be preserved, got %q", snapshot.Path)
	}

	if snapshot.LastScreen != "home" {
		t.Fatalf("expected last screen to be tracked, got %q", snapshot.LastScreen)
	}

	if snapshot.Status != "cache ready" {
		t.Fatalf("expected cache status to be tracked, got %q", snapshot.Status)
	}
}

func TestUpdateLookupTransactionStoresResult(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupTransaction(
		"abc123",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "abc123"},
		},
	)

	snapshot := model.Snapshot()
	if snapshot.Current != ScreenLookup {
		t.Fatalf("expected lookup screen, got %q", snapshot.Current)
	}

	if snapshot.Lookup.Transaction == nil || snapshot.Lookup.Transaction.Transaction == nil {
		t.Fatal("expected transaction lookup to be stored")
	}

	if snapshot.Lookup.Query != "abc123" {
		t.Fatalf("expected query to be stored, got %q", snapshot.Lookup.Query)
	}
}

func TestUpdateLookupLedgerStoresResult(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupLedger(
		"12345",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 12345, Hash: "ledger-hash"},
		},
	)

	snapshot := model.Snapshot()
	if snapshot.Current != ScreenLookup {
		t.Fatalf("expected lookup screen, got %q", snapshot.Current)
	}
	if snapshot.Lookup.Ledger == nil || snapshot.Lookup.Ledger.Ledger == nil {
		t.Fatal("expected ledger lookup to be stored")
	}
	if snapshot.Lookup.Ledger.Ledger.Sequence != 12345 {
		t.Fatalf("expected ledger 12345, got %d", snapshot.Lookup.Ledger.Ledger.Sequence)
	}
}

func TestLookupTransactionExplorerTransitions(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupLedger(
		"12345",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 12345, Hash: "ledger-hash"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-1", LedgerSequence: 12345},
			},
		},
	)

	model.OpenLookupTransactionExplorer("Ledger Transactions", "open detail", []backendclient.TransactionSummary{
		{Hash: "tx-ledger-1", LedgerSequence: 12345},
	}, 10, 0)
	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil || snapshot.Lookup.Explorer.Kind != LookupExplorerTransactions {
		t.Fatalf("expected transaction explorer to be open, got %#v", snapshot.Lookup.Explorer)
	}
	if len(snapshot.Lookup.Explorer.Transactions) != 1 || snapshot.Lookup.Explorer.Transactions[0].Hash != "tx-ledger-1" {
		t.Fatalf("expected explorer transactions to be preserved, got %#v", snapshot.Lookup.Explorer.Transactions)
	}

	model.CloseLookupExplorer()
	if model.Snapshot().Lookup.Explorer != nil {
		t.Fatalf("expected explorer to close, got %#v", model.Snapshot().Lookup.Explorer)
	}
}

func TestLookupOperationExplorerTransitions(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", TypeName: "payment"},
			},
		},
	)

	model.OpenLookupOperationExplorer("Account Operations", "open detail", []backendclient.OperationSummary{
		{TransactionHash: "tx-op-1", TypeName: "payment"},
	}, 10, 0)
	snapshot := model.Snapshot()
	if snapshot.Lookup.Explorer == nil || snapshot.Lookup.Explorer.Kind != LookupExplorerOperations {
		t.Fatalf("expected operation explorer to be open, got %#v", snapshot.Lookup.Explorer)
	}
	if len(snapshot.Lookup.Explorer.Operations) != 1 || snapshot.Lookup.Explorer.Operations[0].TransactionHash != "tx-op-1" {
		t.Fatalf("expected explorer operations to be preserved, got %#v", snapshot.Lookup.Explorer.Operations)
	}
}

func TestUpdateLookupAssetStoresResult(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupAsset(
		"USDC:GISS",
		backendclient.AssetLookupResponse{
			Asset: &backendclient.AssetDetail{AssetCode: "USDC", AssetIssuer: "GISS", NumAccounts: 10},
		},
	)

	snapshot := model.Snapshot()
	if snapshot.Current != ScreenLookup {
		t.Fatalf("expected lookup screen, got %q", snapshot.Current)
	}
	if snapshot.Lookup.Asset == nil || snapshot.Lookup.Asset.Asset == nil {
		t.Fatal("expected asset lookup to be stored")
	}
	if snapshot.Lookup.Asset.Asset.AssetCode != "USDC" {
		t.Fatalf("expected USDC asset, got %#v", snapshot.Lookup.Asset.Asset)
	}
}

func TestRefreshLiveFeedPopulatesSnapshot(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 77,
			LatestLedger: &backendclient.LedgerSummary{
				Sequence: 77,
				ClosedAt: time.Now().UTC(),
			},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-1", LedgerSequence: 77, Account: "GABC"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	snapshot := model.Snapshot()
	if !snapshot.LiveFeed.Available {
		t.Fatal("expected live feed to be available")
	}
	if snapshot.LiveFeed.LastIngestedLedger != 77 {
		t.Fatalf("expected live feed ledger 77, got %d", snapshot.LiveFeed.LastIngestedLedger)
	}
	if len(snapshot.LiveFeed.RecentTransactions) != 1 {
		t.Fatalf("expected one live feed transaction, got %d", len(snapshot.LiveFeed.RecentTransactions))
	}
	if snapshot.LiveFeed.ScrollbackCount != 1 {
		t.Fatalf("expected one scrollback transaction, got %d", snapshot.LiveFeed.ScrollbackCount)
	}
}

func TestRefreshLiveFeedMergesIncrementalScrollback(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	now := time.Unix(1715000000, 0).UTC()

	if err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 77,
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-1", LedgerSequence: 77, CreatedAt: now},
			},
		},
	}); err != nil {
		t.Fatalf("RefreshLiveFeed() first error = %v", err)
	}
	if err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 78,
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-2", LedgerSequence: 78, CreatedAt: now.Add(time.Minute)},
				{Hash: "tx-1", LedgerSequence: 77, CreatedAt: now},
			},
		},
	}); err != nil {
		t.Fatalf("RefreshLiveFeed() second error = %v", err)
	}

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 2 {
		t.Fatalf("expected merged scrollback, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-2" {
		t.Fatalf("expected newest transaction first, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.LastUpdateCount != 1 {
		t.Fatalf("expected one new item on second refresh, got %d", snapshot.LiveFeed.LastUpdateCount)
	}
}

func TestLiveFeedFilterUsesScrollback(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	if err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-soroban", IsSoroban: true},
				{Hash: "tx-classic", IsSoroban: false},
			},
		},
	}); err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}
	if err := model.SetLiveFeedFilter(LiveFeedFilterSoroban); err != nil {
		t.Fatalf("SetLiveFeedFilter() error = %v", err)
	}

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 1 || snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-soroban" {
		t.Fatalf("expected soroban-only filtered feed, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.ScrollbackCount != 2 {
		t.Fatalf("expected unfiltered scrollback count 2, got %d", snapshot.LiveFeed.ScrollbackCount)
	}
}

func TestRefreshLiveFeedTracksBackendError(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		err: errors.New("backend down"),
	})
	if err == nil {
		t.Fatal("expected refresh error")
	}

	snapshot := model.Snapshot()
	if snapshot.LiveFeed.Available {
		t.Fatal("expected live feed to be unavailable")
	}
	if snapshot.LiveFeed.Error != "backend down" {
		t.Fatalf("expected backend error to be preserved, got %q", snapshot.LiveFeed.Error)
	}
}

func TestMoveSelectionTracksLiveFeedRow(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	if err := model.SetScreen(ScreenLiveFeed); err != nil {
		t.Fatalf("SetScreen() error = %v", err)
	}

	err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 88,
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-1"},
				{Hash: "tx-2"},
			},
		},
	})
	if err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	model.MoveSelection(1)
	snapshot := model.Snapshot()
	if snapshot.Selection.LiveFeedIndex != 1 {
		t.Fatalf("expected selected row 1, got %d", snapshot.Selection.LiveFeedIndex)
	}

	selected := model.SelectedLiveTransaction()
	if selected == nil || selected.Hash != "tx-2" {
		t.Fatalf("expected second transaction selected, got %#v", selected)
	}
}

func TestToggleHelpOverlayUpdatesSnapshot(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.ToggleHelpOverlay()

	if !model.Snapshot().HelpVisible {
		t.Fatal("expected help overlay to become visible")
	}
}

func TestRefreshLiveFeedAcceptsRPCOnlyProfile(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles[0].IndexerURL = ""
	cfg.Profiles[0].RPCEndpoint = "https://rpc.example.com"
	cfg.Profiles[0].BackendMode = "rpc"

	model := NewModel(cfg, "/tmp/config.json", CacheSnapshot{})
	err := model.RefreshLiveFeed(context.Background(), fakeLiveFeedService{
		response: backendclient.LiveFeedSummaryResponse{
			LastIngestedLedger: 9,
		},
	})
	if err != nil {
		t.Fatalf("RefreshLiveFeed() error = %v", err)
	}

	snapshot := model.Snapshot()
	if !snapshot.LiveFeed.Configured {
		t.Fatal("expected rpc-only profile to be considered configured")
	}
	if snapshot.LiveFeed.BackendURL != "https://rpc.example.com" {
		t.Fatalf("expected rpc endpoint label, got %q", snapshot.LiveFeed.BackendURL)
	}
}

func TestCommandPaletteStateTransitions(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	model.OpenCommandPalette("Search / Command", "lookup tx ")
	model.AppendCommandPaletteInput("abc")
	model.TrimCommandPaletteInput()

	snapshot := model.Snapshot()
	if !snapshot.Command.Visible {
		t.Fatal("expected command palette to be visible")
	}
	if snapshot.Command.Input != "lookup tx ab" {
		t.Fatalf("expected command input to be updated, got %q", snapshot.Command.Input)
	}

	model.CloseCommandPalette()
	if model.Snapshot().Command.Visible {
		t.Fatal("expected command palette to be hidden")
	}
}

func TestCommandPaletteInfersSearchResults(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	model.OpenCommandPalette("Search / Command", "GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML")
	snapshot := model.Snapshot()
	if len(snapshot.Command.Results) != 1 {
		t.Fatalf("expected one account result, got %d", len(snapshot.Command.Results))
	}
	if snapshot.Command.Results[0].Kind != "account" || snapshot.Command.Results[0].Command == "" {
		t.Fatalf("expected executable account result, got %#v", snapshot.Command.Results[0])
	}

	model.SetCommandPaletteInput("12345")
	snapshot = model.Snapshot()
	if len(snapshot.Command.Results) != 1 {
		t.Fatalf("expected one ledger result, got %d", len(snapshot.Command.Results))
	}
	if snapshot.Command.Results[0].Kind != "ledger" || !snapshot.Command.Results[0].Enabled {
		t.Fatalf("expected executable ledger result, got %#v", snapshot.Command.Results[0])
	}
}

func TestCommandPaletteMergesBackendSearchResults(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.OpenCommandPalette("Search / Command", "GABC")

	model.MergeCommandPaletteBackendResults([]backendclient.SearchResult{
		{
			Kind:        "account",
			Title:       "Account GABC",
			Description: "balance 10",
			Command:     "lookup account GABC",
		},
		{
			Kind:        "account",
			Title:       "Duplicate",
			Description: "ignored",
			Command:     "lookup account GABC",
		},
	})

	snapshot := model.Snapshot()
	if len(snapshot.Command.Results) != 1 {
		t.Fatalf("expected one deduped backend result, got %d", len(snapshot.Command.Results))
	}
	if snapshot.Command.Results[0].Source != "indexer" {
		t.Fatalf("expected indexer source, got %#v", snapshot.Command.Results[0])
	}
}

func TestCommandPaletteMergesLocalSearchResults(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.OpenCommandPalette("Search / Command", "whale")

	model.MergeCommandPaletteLocalResults([]SearchResult{
		{
			Kind:        "bookmark",
			Title:       "Whale account",
			Description: "account GABC",
			Command:     "lookup account GABC",
			Enabled:     true,
			Source:      "local",
		},
	})

	snapshot := model.Snapshot()
	if len(snapshot.Command.Results) != 1 {
		t.Fatalf("expected one local result, got %d", len(snapshot.Command.Results))
	}
	if snapshot.Command.Results[0].Source != "local" {
		t.Fatalf("expected local source, got %#v", snapshot.Command.Results[0])
	}
}

func TestCommandPaletteBackendMergeKeepsLocalResults(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.OpenCommandPalette("Search / Command", "whale")
	model.MergeCommandPaletteLocalResults([]SearchResult{
		{
			Kind:        "bookmark",
			Title:       "Whale account",
			Description: "account GLOCAL",
			Command:     "lookup account GLOCAL",
			Enabled:     true,
			Source:      "local",
		},
	})

	model.MergeCommandPaletteBackendResults([]backendclient.SearchResult{
		{
			Kind:        "account",
			Title:       "Account GREMOTE",
			Description: "balance 10",
			Command:     "lookup account GREMOTE",
		},
	})

	snapshot := model.Snapshot()
	if len(snapshot.Command.Results) != 2 {
		t.Fatalf("expected local and backend results, got %#v", snapshot.Command.Results)
	}
	seen := map[string]bool{}
	for _, result := range snapshot.Command.Results {
		seen[result.Command] = true
	}
	if !seen["lookup account GLOCAL"] || !seen["lookup account GREMOTE"] {
		t.Fatalf("expected both local and backend commands, got %#v", snapshot.Command.Results)
	}
}

func TestCommandPaletteWorkspaceCompletions(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"bookmark", []string{"bookmark add", "bookmark remove", "bookmark note"}},
		{"bookmark a", []string{"bookmark add"}},
		{"note", []string{"note add", "note remove", "note body"}},
		{"note b", []string{"note body"}},
		{"label", []string{"label add", "label remove", "label delete", "label color"}},
		{"label r", []string{"label remove"}},
		{"open", []string{"open recent", "open bookmarks", "open notes", "open labels", "open views", "open cache"}},
		{"open r", []string{"open recent"}},
		{"view", []string{"view save", "view open", "view delete"}},
	}
	for _, tc := range cases {
		model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
		model.OpenCommandPalette("Search / Command", tc.input)
		snapshot := model.Snapshot()
		for _, want := range tc.want {
			found := false
			for _, r := range snapshot.Command.Results {
				if strings.Contains(r.Title, want) || strings.Contains(r.Command, want) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("input=%q: expected a completion containing %q, got %#v", tc.input, want, snapshot.Command.Results)
			}
		}
	}
}

func TestRankSearchResultsPrefersDirectTargetMatch(t *testing.T) {
	results := RankSearchResults("GABC", []SearchResult{
		{
			Kind:        "note",
			Title:       "Mentions GABC",
			Description: "GABC is referenced here",
			Command:     "lookup account GOTHER",
			Enabled:     true,
			Source:      "local",
		},
		{
			Kind:        "account",
			Title:       "Account GABC",
			Description: "balance 10",
			Command:     "lookup account GABC",
			Enabled:     true,
			Source:      "indexer",
		},
	})

	if len(results) != 2 || results[0].Command != "lookup account GABC" {
		t.Fatalf("expected direct target match first, got %#v", results)
	}
}

func TestFocusNextAvailableCyclesDesktopAreas(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})

	model.FocusNextAvailable(FocusMain, FocusSidebar, FocusStatus)
	if got := model.Snapshot().Focus; got != FocusSidebar {
		t.Fatalf("expected sidebar focus, got %q", got)
	}

	model.FocusNextAvailable(FocusMain, FocusSidebar, FocusStatus)
	if got := model.Snapshot().Focus; got != FocusStatus {
		t.Fatalf("expected status focus, got %q", got)
	}

	model.FocusNextAvailable(FocusMain, FocusSidebar, FocusStatus)
	if got := model.Snapshot().Focus; got != FocusMain {
		t.Fatalf("expected main focus, got %q", got)
	}
}
