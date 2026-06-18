package pipeline

import (
	"context"
	"testing"

	"github.com/stellar/go-stellar-sdk/network"
)

func TestBackfillProcessLedgerBatch(t *testing.T) {
	rpc, db := getTestDeps(t)
	defer db.Close()

	ctx := context.Background()

	latest, err := rpc.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger: %v", err)
	}

	p := NewBackfillPipeline(rpc, db, network.TestNetworkPassphrase, 3, 1)

	start := latest.Sequence - 4
	count, err := p.processLedgerBatch(ctx, start, 2)
	if err != nil {
		t.Fatalf("processLedgerBatch: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 ledgers, got %d", count)
	}

	t.Logf("processLedgerBatch: processed %d ledgers starting at %d", count, start)

	cleanupTestLedgers(t, db, start, start+1)
}

func TestBackfillSmallRange(t *testing.T) {
	rpc, db := getTestDeps(t)
	defer db.Close()

	ctx := context.Background()

	// Get latest ledger
	latest, err := rpc.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	// Backfill 5 ledgers with 2 workers
	start := latest.Sequence - 6
	end := latest.Sequence - 2
	p := NewBackfillPipeline(rpc, db, network.TestNetworkPassphrase, 3, 2)

	err = p.Run(ctx, start, end)
	if err != nil {
		t.Fatalf("Backfill failed: %v", err)
	}

	t.Logf("Backfill completed: ledgers %d-%d", start, end)

	// Clean up
	for seq := start; seq <= end; seq++ {
		cleanupTestLedgers(t, db, seq)
	}
}
