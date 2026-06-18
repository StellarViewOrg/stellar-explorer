package source

import (
	"context"
	"os"
	"testing"
)

func getRPCClient(t *testing.T) *RPCClient {
	endpoint := os.Getenv("TEST_RPC_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://soroban-testnet.stellar.org"
	}
	return NewRPCClient(endpoint)
}

func TestGetLatestLedger(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	result, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	if result.Sequence == 0 {
		t.Error("expected non-zero sequence")
	}
	if result.ID == "" {
		t.Error("expected non-empty id")
	}
}

func TestGetLedgers(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	// First get latest ledger to know a valid start point
	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	// Request 2 ledgers starting a few behind latest
	start := latest.Sequence - 5
	result, err := client.GetLedgers(ctx, GetLedgersParams{
		StartLedger: start,
		Pagination:  &Pagination{Limit: 2},
	})
	if err != nil {
		t.Fatalf("GetLedgers failed: %v", err)
	}

	if len(result.Ledgers) == 0 {
		t.Fatal("expected at least one ledger")
	}

	first := result.Ledgers[0]
	if first.Sequence != start {
		t.Errorf("expected first ledger sequence %d, got %d", start, first.Sequence)
	}
	if first.Hash == "" {
		t.Error("expected non-empty hash")
	}
	if first.HeaderXDR == "" {
		t.Error("expected non-empty headerXdr")
	}
	if result.Cursor == "" {
		t.Error("expected non-empty cursor for pagination")
	}
}

func TestGetTransactions(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	// Request transactions from a recent ledger range
	start := latest.Sequence - 10
	result, err := client.GetTransactions(ctx, GetTransactionsParams{
		StartLedger: start,
		Pagination:  &Pagination{Limit: 5},
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}

	// Testnet may or may not have transactions in these ledgers,
	// so we just verify the structure is valid
	if result.LatestLedger == 0 {
		t.Error("expected non-zero latestLedger")
	}
	if result.OldestLedger == 0 {
		t.Error("expected non-zero oldestLedger")
	}

	// If there are transactions, validate their structure
	for i, tx := range result.Transactions {
		if tx.EnvelopeXDR == "" {
			t.Errorf("transaction %d: expected non-empty envelopeXdr", i)
		}
		if tx.Ledger == 0 {
			t.Errorf("transaction %d: expected non-zero ledger", i)
		}
		if tx.Status == "" {
			t.Errorf("transaction %d: expected non-empty status", i)
		}
	}
}

func TestGetLedgersPagination(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	start := latest.Sequence - 10

	// First page
	page1, err := client.GetLedgers(ctx, GetLedgersParams{
		StartLedger: start,
		Pagination:  &Pagination{Limit: 3},
	})
	if err != nil {
		t.Fatalf("GetLedgers page 1 failed: %v", err)
	}

	if len(page1.Ledgers) != 3 {
		t.Fatalf("expected 3 ledgers in page 1, got %d", len(page1.Ledgers))
	}

	// Second page using cursor
	page2, err := client.GetLedgers(ctx, GetLedgersParams{
		Pagination: &Pagination{
			Cursor: page1.Cursor,
			Limit:  3,
		},
	})
	if err != nil {
		t.Fatalf("GetLedgers page 2 failed: %v", err)
	}

	if len(page2.Ledgers) == 0 {
		t.Fatal("expected ledgers in page 2")
	}

	// Verify continuity: page2 first ledger should follow page1 last ledger
	lastSeq := page1.Ledgers[len(page1.Ledgers)-1].Sequence
	firstSeq := page2.Ledgers[0].Sequence
	if firstSeq != lastSeq+1 {
		t.Errorf("expected page 2 to start at %d, got %d", lastSeq+1, firstSeq)
	}
}
