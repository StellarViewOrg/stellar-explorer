package livestream

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseTransactionMessagesBatch(t *testing.T) {
	payload, err := json.Marshal([]TransactionMessage{
		{
			Hash:                 "tx-1",
			LedgerSequence:       10,
			Account:              "GABC",
			OperationCount:       2,
			IsSoroban:            true,
			PrimaryContractID:    "CCONTRACT",
			PrimaryOperationType: "invoke_host_function",
		},
		{Hash: "tx-2", LedgerSequence: 10, Account: "GDEF", OperationCount: 1},
	})
	if err != nil {
		t.Fatalf("marshal batch: %v", err)
	}

	transactions, err := parseTransactionMessages(payload)
	if err != nil {
		t.Fatalf("parseTransactionMessages() error = %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(transactions))
	}
	if transactions[0].Hash != "tx-1" || !transactions[0].IsSoroban {
		t.Fatalf("unexpected first transaction: %#v", transactions[0])
	}
	if transactions[0].PrimaryContractID != "CCONTRACT" {
		t.Fatalf("expected primary contract metadata, got %#v", transactions[0])
	}
	if transactions[0].PrimaryOperationType != "invoke_host_function" {
		t.Fatalf("expected primary operation metadata, got %#v", transactions[0])
	}
}

func TestParseLedgerMessage(t *testing.T) {
	payload, err := json.Marshal(LedgerMessage{
		Sequence:         42,
		Hash:             "ledger-hash",
		ClosedAt:         "2026-06-17T12:00:00Z",
		TransactionCount: 3,
		OperationCount:   5,
	})
	if err != nil {
		t.Fatalf("marshal ledger: %v", err)
	}

	ledger, err := parseLedgerMessage(payload)
	if err != nil {
		t.Fatalf("parseLedgerMessage() error = %v", err)
	}
	if ledger.Sequence != 42 {
		t.Fatalf("expected sequence 42, got %d", ledger.Sequence)
	}
	if ledger.ClosedAt != time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected closed_at: %v", ledger.ClosedAt)
	}
}

func TestParsePubSubMessageRoutesChannels(t *testing.T) {
	txPayload, err := json.Marshal([]TransactionMessage{{Hash: "tx-stream", LedgerSequence: 9}})
	if err != nil {
		t.Fatalf("marshal tx payload: %v", err)
	}

	update, err := parsePubSubMessage(ChannelTransactions, txPayload)
	if err != nil {
		t.Fatalf("parsePubSubMessage() error = %v", err)
	}
	if len(update.Transactions) != 1 || update.Transactions[0].Hash != "tx-stream" {
		t.Fatalf("unexpected transaction update: %#v", update)
	}

	ledgerPayload, err := json.Marshal(LedgerMessage{Sequence: 9, Hash: "ledger", ClosedAt: "2026-06-17T12:00:00Z"})
	if err != nil {
		t.Fatalf("marshal ledger payload: %v", err)
	}
	update, err = parsePubSubMessage(ChannelLedgers, ledgerPayload)
	if err != nil {
		t.Fatalf("parsePubSubMessage() ledger error = %v", err)
	}
	if update.Ledger == nil || update.Ledger.Sequence != 9 {
		t.Fatalf("unexpected ledger update: %#v", update)
	}
}
