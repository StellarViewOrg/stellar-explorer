package app

import (
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestApplyLiveFeedStreamUpdateDedupesAndOrders(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	now := time.Unix(1715000000, 0).UTC()

	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-1", LedgerSequence: 10, CreatedAt: now},
		},
	})
	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-2", LedgerSequence: 11, CreatedAt: now.Add(time.Minute)},
			{Hash: "tx-1", LedgerSequence: 10, CreatedAt: now},
		},
	})

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 2 {
		t.Fatalf("expected 2 transactions, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-2" {
		t.Fatalf("expected newest transaction first, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.SourceMode != LiveFeedSourceStream {
		t.Fatalf("expected stream source mode, got %q", snapshot.LiveFeed.SourceMode)
	}
}

func TestApplyLiveFeedStreamUpdateBuffersWhilePaused(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.SetLiveFeedPaused(true)

	added := model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{{Hash: "tx-paused", LedgerSequence: 3}},
	})
	if added != 0 {
		t.Fatalf("expected paused stream to buffer updates, got %d added", added)
	}
	if len(model.Snapshot().LiveFeed.RecentTransactions) != 0 {
		t.Fatal("expected paused stream to preserve existing view")
	}

	model.SetLiveFeedPaused(false)
	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 1 || snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-paused" {
		t.Fatalf("expected buffered stream update after resume, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
}

func TestApplyLiveFeedStreamUpdatePreservesSelection(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	_ = model.SetScreen(ScreenLiveFeed)

	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-new", LedgerSequence: 12},
			{Hash: "tx-old", LedgerSequence: 11},
		},
	})
	model.MoveSelection(1)

	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-newer", LedgerSequence: 13},
		},
	})

	selected := model.SelectedLiveTransaction()
	if selected == nil || selected.Hash != "tx-old" {
		t.Fatalf("expected selection to remain on tx-old, got %#v", selected)
	}
}

func TestLiveFeedAccountFilter(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-a", Account: "GABC"},
			{Hash: "tx-b", Account: "GDEF"},
		},
	})

	if err := model.SetLiveFeedFilterCommand([]string{"filter", "account", "GABC"}); err != nil {
		t.Fatalf("SetLiveFeedFilterCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 1 || snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-a" {
		t.Fatalf("expected account-filtered feed, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.ScrollbackCount != 2 {
		t.Fatalf("expected full scrollback to remain available, got %d", snapshot.LiveFeed.ScrollbackCount)
	}
}

func TestLiveFeedContractFilter(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-a", PrimaryContractID: "CCONTRACT"},
			{Hash: "tx-b", PrimaryContractID: "COTHER"},
		},
	})

	if err := model.SetLiveFeedFilterCommand([]string{"filter", "contract", "CCONTRACT"}); err != nil {
		t.Fatalf("SetLiveFeedFilterCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 1 || snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-a" {
		t.Fatalf("expected contract-filtered feed, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
}

func TestLiveFeedOperationFilter(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-a", PrimaryOperationType: "invoke_host_function"},
			{Hash: "tx-b", PrimaryOperationType: "payment"},
		},
	})

	if err := model.SetLiveFeedFilterCommand([]string{"filter", "operation", "invoke_host_function"}); err != nil {
		t.Fatalf("SetLiveFeedFilterCommand() error = %v", err)
	}

	snapshot := model.Snapshot()
	if len(snapshot.LiveFeed.RecentTransactions) != 1 || snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-a" {
		t.Fatalf("expected operation-filtered feed, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
}

func TestShiftLiveFeedReplayWhilePaused(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	now := time.Unix(1715000000, 0).UTC()
	for index := 0; index < 3; index++ {
		model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-" + string(rune('a'+index)), LedgerSequence: uint32(10 + index), CreatedAt: now.Add(time.Duration(index) * time.Minute)},
			},
		})
	}

	model.SetLiveFeedPaused(true)
	model.ShiftLiveFeedReplay(1)

	snapshot := model.Snapshot()
	if snapshot.LiveFeed.ReplayOffset != 1 {
		t.Fatalf("expected replay offset 1, got %d", snapshot.LiveFeed.ReplayOffset)
	}
	if len(snapshot.LiveFeed.RecentTransactions) != 2 {
		t.Fatalf("expected replay window to drop one row, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
	if snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-b" {
		t.Fatalf("expected replay to start at older visible row, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
}

func TestSetLiveFeedPausedResetsReplayOffset(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.ApplyLiveFeedStreamUpdate(LiveFeedStreamUpdate{
		Transactions: []backendclient.TransactionSummary{
			{Hash: "tx-1", LedgerSequence: 1},
			{Hash: "tx-2", LedgerSequence: 2},
		},
	})
	model.SetLiveFeedPaused(true)
	model.ShiftLiveFeedReplay(1)
	model.SetLiveFeedPaused(false)

	snapshot := model.Snapshot()
	if snapshot.LiveFeed.ReplayOffset != 0 {
		t.Fatalf("expected replay offset reset, got %d", snapshot.LiveFeed.ReplayOffset)
	}
	if snapshot.LiveFeed.RecentTransactions[0].Hash != "tx-2" {
		t.Fatalf("expected live edge after resume, got %#v", snapshot.LiveFeed.RecentTransactions)
	}
}
