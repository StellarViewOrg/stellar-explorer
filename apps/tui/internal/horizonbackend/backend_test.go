package horizonbackend

import (
	"context"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/stellar/go-stellar-sdk/clients/horizonclient"
	"github.com/stellar/go-stellar-sdk/support/http/httptest"
)

func TestLedgerReturnsTransactions(t *testing.T) {
	hmock := httptest.NewClient()
	client := &horizonclient.Client{
		HorizonURL: "https://localhost/",
		HTTP:       hmock,
	}

	hmock.On("GET", "https://localhost/ledgers/69859").ReturnString(200, testLedgerResponse)
	hmock.On(
		"GET",
		"https://localhost/ledgers/69859/transactions?limit=50&order=asc",
	).ReturnString(200, testEmptyTransactionsPage)

	backend, err := New(config.Profile{
		Name:       "testnet",
		Network:    "testnet",
		HorizonURL: "https://localhost/",
	}, WithHorizonClient(client))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	response, err := backend.Ledger(context.Background(), 69859)
	if err != nil {
		t.Fatalf("Ledger() error = %v", err)
	}
	if response.Ledger == nil || response.Ledger.Sequence != 69859 {
		t.Fatalf("unexpected ledger payload: %#v", response.Ledger)
	}
	if response.Ledger.FailedTxCount != 1 {
		t.Fatalf("failed tx count = %d, want 1", response.Ledger.FailedTxCount)
	}
}

func TestNewRequiresHorizonURL(t *testing.T) {
	_, err := New(config.Profile{
		Name:    "custom",
		Network: "custom-net",
	})
	if err == nil {
		t.Fatal("expected missing horizon url error")
	}
}

const testLedgerResponse = `{
  "id": "71a40c0581d8d7c1158e1d9368024c5f9fd70de17a8d277cdd96781590cc10fb",
  "hash": "71a40c0581d8d7c1158e1d9368024c5f9fd70de17a8d277cdd96781590cc10fb",
  "sequence": 69859,
  "successful_transaction_count": 0,
  "failed_transaction_count": 1,
  "operation_count": 0,
  "closed_at": "2019-03-03T13:38:16Z"
}`

const testEmptyTransactionsPage = `{
  "_embedded": {
    "records": []
  }
}`
