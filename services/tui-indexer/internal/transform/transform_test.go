package transform

import (
	"context"
	"os"
	"testing"

	"github.com/stellar/go-stellar-sdk/network"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
)

func getRPCClient(t *testing.T) *source.RPCClient {
	endpoint := os.Getenv("TEST_RPC_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://soroban-testnet.stellar.org"
	}
	return source.NewRPCClient(endpoint)
}

func TestLedgerFromRPC(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	start := latest.Sequence - 3
	result, err := client.GetLedgers(ctx, source.GetLedgersParams{
		StartLedger: start,
		Pagination:  &source.Pagination{Limit: 2},
	})
	if err != nil {
		t.Fatalf("GetLedgers failed: %v", err)
	}

	if len(result.Ledgers) == 0 {
		t.Fatal("expected at least one ledger")
	}

	for i, entry := range result.Ledgers {
		ledger, err := LedgerFromRPC(entry)
		if err != nil {
			t.Fatalf("LedgerFromRPC[%d] failed: %v", i, err)
		}

		if ledger.Sequence == 0 {
			t.Errorf("ledger[%d]: expected non-zero sequence", i)
		}
		if ledger.Hash == "" {
			t.Errorf("ledger[%d]: expected non-empty hash", i)
		}
		if ledger.PrevHash == "" {
			t.Errorf("ledger[%d]: expected non-empty prev_hash", i)
		}
		if ledger.ClosedAt.IsZero() {
			t.Errorf("ledger[%d]: expected non-zero closed_at", i)
		}
		if ledger.ProtocolVersion == 0 {
			t.Errorf("ledger[%d]: expected non-zero protocol_version", i)
		}
		if ledger.BaseFee == 0 {
			t.Errorf("ledger[%d]: expected non-zero base_fee", i)
		}
		if ledger.HeaderXDR == nil {
			t.Errorf("ledger[%d]: expected non-nil header_xdr", i)
		}
	}
}

func TestTransactionFromRPC(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	// Fetch a larger range to increase chances of finding transactions on testnet
	start := latest.Sequence - 50
	result, err := client.GetTransactions(ctx, source.GetTransactionsParams{
		StartLedger: start,
		Pagination:  &source.Pagination{Limit: 100},
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}

	if len(result.Transactions) == 0 {
		t.Skip("no transactions found in recent ledgers on testnet")
	}

	for i, entry := range result.Transactions {
		tx, err := TransactionFromRPC(entry, network.TestNetworkPassphrase)
		if err != nil {
			t.Fatalf("TransactionFromRPC[%d] failed: %v", i, err)
		}

		if tx.Hash == "" {
			t.Errorf("tx[%d]: expected non-empty hash", i)
		}
		if tx.LedgerSequence == 0 {
			t.Errorf("tx[%d]: expected non-zero ledger_sequence", i)
		}
		if tx.Account == "" {
			t.Errorf("tx[%d]: expected non-empty account", i)
		}
		if tx.EnvelopeXDR == "" {
			t.Errorf("tx[%d]: expected non-empty envelope_xdr", i)
		}
		if tx.CreatedAt.IsZero() {
			t.Errorf("tx[%d]: expected non-zero created_at", i)
		}

		// Stop after checking a few
		if i >= 4 {
			break
		}
	}
}

func TestOperationsFromRPC(t *testing.T) {
	client := getRPCClient(t)
	ctx := context.Background()

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		t.Fatalf("GetLatestLedger failed: %v", err)
	}

	start := latest.Sequence - 50
	result, err := client.GetTransactions(ctx, source.GetTransactionsParams{
		StartLedger: start,
		Pagination:  &source.Pagination{Limit: 100},
	})
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}

	if len(result.Transactions) == 0 {
		t.Skip("no transactions found in recent ledgers on testnet")
	}

	totalOps := 0
	for i, entry := range result.Transactions {
		ops, err := OperationsFromRPC(entry, network.TestNetworkPassphrase)
		if err != nil {
			t.Fatalf("OperationsFromRPC[%d] failed: %v", i, err)
		}

		for j, op := range ops {
			if op.TransactionHash == "" {
				t.Errorf("tx[%d].op[%d]: expected non-empty transaction_hash", i, j)
			}
			if op.TypeName == "" {
				t.Errorf("tx[%d].op[%d]: expected non-empty type_name", i, j)
			}
			if op.Details == "" {
				t.Errorf("tx[%d].op[%d]: expected non-empty details", i, j)
			}
			if op.ApplicationOrder == 0 {
				t.Errorf("tx[%d].op[%d]: expected non-zero application_order", i, j)
			}
			totalOps++
		}

		if i >= 4 {
			break
		}
	}

	if totalOps == 0 {
		t.Error("expected at least one operation across transactions")
	}

	t.Logf("Parsed %d operations from %d transactions", totalOps, min(len(result.Transactions), 5))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
