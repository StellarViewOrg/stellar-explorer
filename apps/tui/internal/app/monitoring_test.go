package app

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestCaptureAndRestoreLiveFeedMonitoringContext(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	_ = model.SetScreen(ScreenLiveFeed)

	model.liveFeedAll = []backendclient.TransactionSummary{
		{Hash: "hash-a", LedgerSequence: 1, Account: "GAAA", IsSoroban: false},
		{Hash: "hash-b", LedgerSequence: 2, Account: "GBBB", IsSoroban: true},
		{Hash: "hash-c", LedgerSequence: 3, Account: "GCCC", IsSoroban: true},
	}
	model.setLiveFeedFilterSpec(LiveFeedFilterSpec{Class: LiveFeedFilterSoroban})
	model.SetLiveFeedPaused(true)
	model.selection.LiveFeedIndex = 0
	model.applyLiveFeedView(0)

	model.CaptureLiveFeedMonitoringContext(0, 4)
	model.UpdateLookupTransaction("hash-b", backendclient.TransactionLookupResponse{}, DefaultSourceMetadata(config.Default().Profiles[0], "transaction"))
	model.SetLookupReturnContext(ScreenLiveFeed, "live feed")

	if !model.ReturnToLiveMonitoring() {
		t.Fatal("expected ReturnToLiveMonitoring to succeed")
	}

	snapshot := model.Snapshot()
	if snapshot.Current != ScreenLiveFeed {
		t.Fatalf("expected live feed screen, got %s", snapshot.Current)
	}
	if snapshot.LiveFeed.Filter != LiveFeedFilterSoroban {
		t.Fatalf("expected soroban filter restored, got %q", snapshot.LiveFeed.Filter)
	}
	if !snapshot.LiveFeed.Paused {
		t.Fatal("expected paused state to be restored")
	}
	if snapshot.Selection.LiveFeedIndex != 0 {
		t.Fatalf("expected selected hash-b at index 0 in filtered view, got %d", snapshot.Selection.LiveFeedIndex)
	}
	if snapshot.Selection.LiveFeedScrollOffset != 4 {
		t.Fatalf("expected scroll offset 4, got %d", snapshot.Selection.LiveFeedScrollOffset)
	}
}

func TestBackFromLiveLookupReturnsToMonitoring(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	_ = model.SetScreen(ScreenLiveFeed)

	model.liveFeedAll = []backendclient.TransactionSummary{
		{Hash: "hash-a", LedgerSequence: 1},
	}
	model.selection.LiveFeedIndex = 0
	model.applyLiveFeedView(0)
	model.CaptureLiveFeedMonitoringContext(0, 2)

	model.UpdateLookupTransaction("hash-a", backendclient.TransactionLookupResponse{}, DefaultSourceMetadata(config.Default().Profiles[0], "transaction"))
	model.SetLookupReturnContext(ScreenLiveFeed, "live feed")

	if !model.Back() {
		t.Fatal("expected back navigation to succeed")
	}
	if model.Snapshot().Current != ScreenLiveFeed {
		t.Fatalf("expected live feed after back, got %s", model.Snapshot().Current)
	}
}

func TestParseLiveFeedFilterValue(t *testing.T) {
	spec, err := ParseLiveFeedFilterValue("soroban account:GABC123")
	if err != nil {
		t.Fatalf("ParseLiveFeedFilterValue() error = %v", err)
	}
	if spec.Class != LiveFeedFilterSoroban {
		t.Fatalf("class = %q, want soroban", spec.Class)
	}
	if spec.Account != "GABC123" {
		t.Fatalf("account = %q, want GABC123", spec.Account)
	}
}
