package readapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func TestHandleAccountLookupIncludesActivitySlices(t *testing.T) {
	server := NewServer(fakeReadStore{
		account: &store.ReadAccountDetail{ID: "GACCOUNT", Balance: "12.5000000"},
		trustlines: []store.ReadTrustlineSummary{
			{AssetCode: "USDC", AssetIssuer: "GISSUER", Balance: "5.0000000"},
		},
		signers: []store.ReadAccountSignerSummary{
			{SignerKey: "GSIGNER", Weight: 1, Type: "ed25519_public_key"},
		},
		accountTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-account", LedgerSequence: 123},
		},
		accountOperations: []store.ReadOperationSummary{
			{ApplicationOrder: 1, TypeName: "payment"},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/accounts/GACCOUNT?limit=3", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response accountLookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Account == nil || response.Account.ID != "GACCOUNT" {
		t.Fatalf("expected account payload, got %#v", response.Account)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-account" {
		t.Fatalf("expected account recent transactions, got %#v", response.RecentTransactions)
	}
	if len(response.RecentOperations) != 1 || response.RecentOperations[0].TypeName != "payment" {
		t.Fatalf("expected account recent operations, got %#v", response.RecentOperations)
	}
}

func TestHandleContractLookupIncludesEventAndTransactionSlices(t *testing.T) {
	server := NewServer(fakeReadStore{
		contract: &store.ReadContractDetail{ContractID: "CCONTRACT", InvocationCount: 4},
		contractSpec: &store.ReadContractSpec{
			ContractID:    "CCONTRACT",
			Available:     true,
			DecodeStatus:  "decoded",
			FunctionCount: 1,
			Functions: []store.ReadContractSpecFunction{
				{Name: "balance", Inputs: []store.ReadContractSpecValue{{Name: "id", Type: "address"}}, Outputs: []string{"i128"}},
			},
		},
		contractStorage: []store.ReadContractStorageSummary{
			{ContractID: "CCONTRACT", DurabilityLabel: "persistent", DecodeStatus: "decoded", DisplayKey: "DataKey", DisplayValue: "DataValue"},
		},
		contractTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-contract", LedgerSequence: 456},
		},
		contractEvents: []store.ReadContractEventSummary{
			{ContractID: "CCONTRACT", TransactionHash: "tx-contract", LedgerSequence: 456, Type: 0, DecodeStatus: "raw"},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response contractLookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Contract == nil || response.Contract.ContractID != "CCONTRACT" {
		t.Fatalf("expected contract payload, got %#v", response.Contract)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-contract" {
		t.Fatalf("expected contract recent transactions, got %#v", response.RecentTransactions)
	}
	if len(response.RecentEvents) != 1 || response.RecentEvents[0].TransactionHash != "tx-contract" {
		t.Fatalf("expected contract recent events, got %#v", response.RecentEvents)
	}
	if response.Spec == nil || response.Spec.FunctionCount != 1 {
		t.Fatalf("expected contract spec payload, got %#v", response.Spec)
	}
	if len(response.Storage) != 1 || response.Storage[0].DisplayKey != "DataKey" {
		t.Fatalf("expected contract storage payload, got %#v", response.Storage)
	}
}

func TestHandleLedgerLookupIncludesTransactionSlice(t *testing.T) {
	server := NewServer(fakeReadStore{
		ledger: &store.ReadLedgerSummary{Sequence: 789, Hash: "ledger-hash", ClosedAt: time.Unix(1715000000, 0).UTC()},
		ledgerTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-ledger", LedgerSequence: 789},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/ledgers/789", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response ledgerLookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Ledger == nil || response.Ledger.Sequence != 789 {
		t.Fatalf("expected ledger payload, got %#v", response.Ledger)
	}
	if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-ledger" {
		t.Fatalf("expected ledger transactions, got %#v", response.Transactions)
	}
}

func TestHandleAssetLookupIncludesHolderAndTransactionSlices(t *testing.T) {
	server := NewServer(fakeReadStore{
		asset: &store.ReadAssetDetail{AssetCode: "USDC", AssetIssuer: "GISSUER"},
		assetHolders: []store.ReadAssetHolderSummary{
			{AccountID: "GHOLDER", Balance: "100.0000000"},
		},
		assetTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-asset", LedgerSequence: 321},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/assets/USDC:GISSUER", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response assetLookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Asset == nil || response.Asset.AssetCode != "USDC" {
		t.Fatalf("expected asset payload, got %#v", response.Asset)
	}
	if len(response.TopHolders) != 1 || response.TopHolders[0].AccountID != "GHOLDER" {
		t.Fatalf("expected asset holders, got %#v", response.TopHolders)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-asset" {
		t.Fatalf("expected asset transactions, got %#v", response.RecentTransactions)
	}
}

func TestHandleTransactionLookupIncludesEffects(t *testing.T) {
	now := time.Unix(1715000000, 0).UTC()
	server := NewServer(fakeReadStore{
		transaction: &store.ReadTransactionDetail{Hash: "tx-effects", LedgerSequence: 100, Account: "GSOURCE", Status: 1, CreatedAt: now},
		transactionOperations: []store.ReadOperationSummary{
			{TransactionHash: "tx-effects", ApplicationOrder: 1, TypeName: "payment", CreatedAt: now},
		},
		transactionEffects: []store.ReadEffectSummary{
			{TransactionHash: "tx-effects", TypeName: "account_credited", Account: "GDEST", CreatedAt: now},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions/tx-effects", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response transactionLookupResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Transaction == nil || response.Transaction.Hash != "tx-effects" {
		t.Fatalf("expected transaction payload, got %#v", response.Transaction)
	}
	if len(response.Operations) != 1 || response.Operations[0].TypeName != "payment" {
		t.Fatalf("expected transaction operations, got %#v", response.Operations)
	}
	if len(response.Effects) != 1 || response.Effects[0].Account != "GDEST" {
		t.Fatalf("expected transaction effects, got %#v", response.Effects)
	}
}

func TestHandleContractSubresourceStorageAndInvocations(t *testing.T) {
	now := time.Unix(1715000000, 0).UTC()
	server := NewServer(fakeReadStore{
		contractStorage: []store.ReadContractStorageSummary{
			{ContractID: "CCONTRACT", DisplayKey: "balance", DisplayValue: "100", DurabilityLabel: "persistent"},
		},
		contractInvocations: []store.ReadOperationSummary{
			{TransactionHash: "tx-invoke-1", ApplicationOrder: 1, TypeName: "invoke_host_function", CreatedAt: now},
		},
	}, nil, nil)

	t.Run("storage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/storage?limit=2", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var response contractStorageListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Storage) != 1 || response.Storage[0].DisplayKey != "balance" {
			t.Fatalf("expected contract storage list, got %#v", response.Storage)
		}
	})

	t.Run("invocations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/invocations?limit=2", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var response operationListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Operations) != 1 || response.Operations[0].TransactionHash != "tx-invoke-1" {
			t.Fatalf("expected contract invocation list, got %#v", response.Operations)
		}
	})
}

func TestHandleLedgerTransactionList(t *testing.T) {
	server := NewServer(fakeReadStore{
		ledgerTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-ledger-list", LedgerSequence: 789},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/ledgers/789/transactions?limit=2", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response transactionListResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-ledger-list" {
		t.Fatalf("expected ledger transaction list, got %#v", response.Transactions)
	}
}

func TestHandleTopLevelExplorerLists(t *testing.T) {
	now := time.Unix(1715000000, 0).UTC()
	server := NewServer(fakeReadStore{
		ledgers: []store.ReadLedgerSummary{
			{Sequence: 789, Hash: "ledger-hash", ClosedAt: now},
		},
		accounts: []store.ReadAccountDetail{
			{ID: "GACCOUNT", Balance: "12.5000000"},
		},
		assets: []store.ReadAssetDetail{
			{AssetCode: "USDC", AssetIssuer: "GISSUER"},
		},
		contracts: []store.ReadContractDetail{
			{ContractID: "CCONTRACT", InvocationCount: 4},
		},
	}, nil, nil)

	tests := []struct {
		path string
		want string
	}{
		{path: "/v1/ledgers?limit=2", want: "ledgers"},
		{path: "/v1/accounts?limit=2", want: "accounts"},
		{path: "/v1/assets?limit=2", want: "assets"},
		{path: "/v1/contracts?limit=2", want: "contracts"},
	}

	for _, test := range tests {
		req := httptest.NewRequest(http.MethodGet, test.path, nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", test.path, rec.Code, rec.Body.String())
		}
		var payload map[string]json.RawMessage
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("decode %s response: %v", test.path, err)
		}
		if len(payload[test.want]) == 0 {
			t.Fatalf("%s expected %q payload, got %s", test.path, test.want, rec.Body.String())
		}
	}
}

func TestHandleLedgerListAcceptsBeforeCursor(t *testing.T) {
	server := NewServer(fakeReadStore{
		ledgers: []store.ReadLedgerSummary{
			{Sequence: 788, Hash: "ledger-hash"},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/ledgers?limit=2&before=789", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleAccountSubresourceLists(t *testing.T) {
	server := NewServer(fakeReadStore{
		accountTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-account-list", LedgerSequence: 123},
		},
		accountOperations: []store.ReadOperationSummary{
			{TransactionHash: "tx-account-list", ApplicationOrder: 1, TypeName: "payment"},
		},
		accountTimeline: []store.ReadTimelineItem{
			{Kind: "tx", Title: "Transaction tx-account-list", Command: "lookup tx tx-account-list", OccurredAt: time.Unix(1715000000, 0).UTC()},
		},
	}, nil, nil)

	t.Run("transactions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/GACCOUNT/transactions?limit=4", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response transactionListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-account-list" {
			t.Fatalf("expected account transaction list, got %#v", response.Transactions)
		}
	})

	t.Run("operations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/GACCOUNT/operations?limit=4", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response operationListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Operations) != 1 || response.Operations[0].TypeName != "payment" {
			t.Fatalf("expected account operation list, got %#v", response.Operations)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/accounts/GACCOUNT/timeline?limit=4&offset=2", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response timelineResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Items) != 1 || response.Items[0].Command != "lookup tx tx-account-list" {
			t.Fatalf("expected account timeline list, got %#v", response.Items)
		}
	})
}

func TestHandleAssetSubresourceLists(t *testing.T) {
	server := NewServer(fakeReadStore{
		assetHolders: []store.ReadAssetHolderSummary{
			{AccountID: "GHOLDER", Balance: "100.0000000"},
		},
		assetTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-asset-list", LedgerSequence: 321},
		},
		assetTimeline: []store.ReadTimelineItem{
			{Kind: "holder", Title: "Holder GHOLDER", Command: "lookup account GHOLDER"},
		},
	}, nil, nil)

	t.Run("transactions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/assets/USDC:GISSUER/transactions", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response transactionListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-asset-list" {
			t.Fatalf("expected asset transaction list, got %#v", response.Transactions)
		}
	})

	t.Run("holders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/assets/USDC:GISSUER/holders", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response assetHolderListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Holders) != 1 || response.Holders[0].AccountID != "GHOLDER" {
			t.Fatalf("expected asset holder list, got %#v", response.Holders)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/assets/USDC:GISSUER/timeline?limit=4&offset=2", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response timelineResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Items) != 1 || response.Items[0].Command != "lookup account GHOLDER" {
			t.Fatalf("expected asset timeline list, got %#v", response.Items)
		}
	})
}

func TestHandleContractSubresourceLists(t *testing.T) {
	server := NewServer(fakeReadStore{
		contractTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-contract-list", LedgerSequence: 456},
		},
		contractEvents: []store.ReadContractEventSummary{
			{TransactionHash: "tx-contract-list", LedgerSequence: 456, Type: 0},
		},
		contractTimeline: []store.ReadTimelineItem{
			{Kind: "event", Title: "Event type 0", Command: "lookup tx tx-contract-list"},
		},
	}, nil, nil)

	t.Run("transactions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/transactions", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response transactionListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-contract-list" {
			t.Fatalf("expected contract transaction list, got %#v", response.Transactions)
		}
	})

	t.Run("events", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/events", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response contractEventListResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Events) != 1 || response.Events[0].TransactionHash != "tx-contract-list" {
			t.Fatalf("expected contract event list, got %#v", response.Events)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/timeline?limit=4&offset=2", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response timelineResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Items) != 1 || response.Items[0].Command != "lookup tx tx-contract-list" {
			t.Fatalf("expected contract timeline list, got %#v", response.Items)
		}
	})

	t.Run("timeline rejects unsupported type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/contracts/CCONTRACT/timeline?type=holder", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestHandleLiveFeedSummary(t *testing.T) {
	now := time.Unix(1715000000, 0).UTC()
	server := NewServer(fakeReadStore{
		lastIngestedLedger: 900,
		ledger:             &store.ReadLedgerSummary{Sequence: 900, Hash: "ledger-live", ClosedAt: now},
		recentTransactions: []store.ReadTransactionSummary{
			{Hash: "tx-live", LedgerSequence: 900, CreatedAt: now},
		},
	}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/feed/live/summary", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response liveFeedSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.LastIngestedLedger != 900 {
		t.Fatalf("expected last ingested ledger 900, got %d", response.LastIngestedLedger)
	}
	if response.LatestLedger == nil || response.LatestLedger.Sequence != 900 {
		t.Fatalf("expected latest ledger payload, got %#v", response.LatestLedger)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-live" {
		t.Fatalf("expected recent transactions, got %#v", response.RecentTransactions)
	}
}

func TestHandleSearchReturnsIndexedResults(t *testing.T) {
	server := NewServer(fakeReadStore{
		searchResults: []store.ReadSearchResult{
			{Kind: "transaction", Title: "Transaction tx-search", Command: "lookup tx tx-search"},
		},
	}, nil, nil)

	t.Run("empty query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/search", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var response searchResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Results) != 0 {
			t.Fatalf("expected empty search results, got %#v", response.Results)
		}
	})

	t.Run("indexed query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/search?q=tx-search&limit=5", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var response searchResponse
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(response.Results) != 1 || response.Results[0].Command != "lookup tx tx-search" {
			t.Fatalf("expected indexed search result, got %#v", response.Results)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/search?q=tx-search&limit=0", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestHandleContractSubresourcePaginationUsesOffset(t *testing.T) {
	now := time.Unix(1715000000, 0).UTC()
	server := NewServer(fakeReadStore{
		contractStorage: []store.ReadContractStorageSummary{
			{ContractID: "CCONTRACT", DisplayKey: "offset-storage", DurabilityLabel: "persistent"},
		},
		contractEvents: []store.ReadContractEventSummary{
			{ContractID: "CCONTRACT", TransactionHash: "tx-event-offset", LedgerSequence: 456, Type: 0},
		},
		contractInvocations: []store.ReadOperationSummary{
			{TransactionHash: "tx-invoke-offset", ApplicationOrder: 2, TypeName: "invoke_host_function", CreatedAt: now},
		},
	}, nil, nil)

	tests := []struct {
		name string
		path string
	}{
		{name: "storage", path: "/v1/contracts/CCONTRACT/storage?limit=1&offset=3"},
		{name: "events", path: "/v1/contracts/CCONTRACT/events?limit=1&offset=3"},
		{name: "invocations", path: "/v1/contracts/CCONTRACT/invocations?limit=1&offset=3"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

type fakeReadStore struct {
	lastIngestedLedger    uint32
	recentTransactions    []store.ReadTransactionSummary
	searchResults         []store.ReadSearchResult
	ledger                *store.ReadLedgerSummary
	transaction           *store.ReadTransactionDetail
	transactionOperations []store.ReadOperationSummary
	transactionEffects    []store.ReadEffectSummary
	account               *store.ReadAccountDetail
	asset                 *store.ReadAssetDetail
	contract              *store.ReadContractDetail
	contractSpec          *store.ReadContractSpec
	contractStorage       []store.ReadContractStorageSummary
	ledgers               []store.ReadLedgerSummary
	accounts              []store.ReadAccountDetail
	assets                []store.ReadAssetDetail
	contracts             []store.ReadContractDetail
	trustlines            []store.ReadTrustlineSummary
	signers               []store.ReadAccountSignerSummary
	ledgerTransactions    []store.ReadTransactionSummary
	accountTransactions   []store.ReadTransactionSummary
	accountOperations     []store.ReadOperationSummary
	accountTimeline       []store.ReadTimelineItem
	assetHolders          []store.ReadAssetHolderSummary
	assetTransactions     []store.ReadTransactionSummary
	assetTimeline         []store.ReadTimelineItem
	contractTransactions  []store.ReadTransactionSummary
	contractEvents        []store.ReadContractEventSummary
	contractInvocations   []store.ReadOperationSummary
	contractTimeline      []store.ReadTimelineItem
}

func (f fakeReadStore) Ping(context.Context) error { return nil }

func (f fakeReadStore) GetLastIngestedLedger(context.Context) (uint32, error) {
	return f.lastIngestedLedger, nil
}

func (f fakeReadStore) GetLatestLedgerSummary(context.Context) (*store.ReadLedgerSummary, error) {
	return f.ledger, nil
}

func (f fakeReadStore) ListLedgerSummaries(context.Context, int, uint32) ([]store.ReadLedgerSummary, error) {
	return f.ledgers, nil
}

func (f fakeReadStore) ListRecentTransactionSummaries(context.Context, int) ([]store.ReadTransactionSummary, error) {
	return f.recentTransactions, nil
}

func (f fakeReadStore) Search(context.Context, string, int) ([]store.ReadSearchResult, error) {
	return f.searchResults, nil
}

func (f fakeReadStore) GetLedgerSummaryBySequence(context.Context, uint32) (*store.ReadLedgerSummary, error) {
	return f.ledger, nil
}

func (f fakeReadStore) ListTransactionSummariesByLedger(context.Context, uint32, int, int) ([]store.ReadTransactionSummary, error) {
	return f.ledgerTransactions, nil
}

func (f fakeReadStore) GetTransactionByHash(context.Context, string) (*store.ReadTransactionDetail, error) {
	if f.transaction == nil {
		return nil, store.ErrNotFound
	}
	return f.transaction, nil
}

func (f fakeReadStore) ListOperationsByTransactionHash(context.Context, string) ([]store.ReadOperationSummary, error) {
	return f.transactionOperations, nil
}

func (f fakeReadStore) ListEffectsByTransactionHash(context.Context, string) ([]store.ReadEffectSummary, error) {
	return f.transactionEffects, nil
}

func (f fakeReadStore) GetAccountByID(context.Context, string) (*store.ReadAccountDetail, error) {
	return f.account, nil
}

func (f fakeReadStore) ListAccountDetails(context.Context, int) ([]store.ReadAccountDetail, error) {
	return f.accounts, nil
}

func (f fakeReadStore) ListTrustlinesByAccountID(context.Context, string) ([]store.ReadTrustlineSummary, error) {
	return f.trustlines, nil
}

func (f fakeReadStore) ListAccountSignersByAccountID(context.Context, string) ([]store.ReadAccountSignerSummary, error) {
	return f.signers, nil
}

func (f fakeReadStore) ListTransactionSummariesByAccount(context.Context, string, int) ([]store.ReadTransactionSummary, error) {
	return f.accountTransactions, nil
}

func (f fakeReadStore) ListOperationSummariesByAccount(context.Context, string, int, int) ([]store.ReadOperationSummary, error) {
	return f.accountOperations, nil
}

func (f fakeReadStore) ListAccountTimeline(context.Context, string, int, int, string) ([]store.ReadTimelineItem, error) {
	return f.accountTimeline, nil
}

func (f fakeReadStore) GetAssetByCodeIssuer(context.Context, string, string) (*store.ReadAssetDetail, error) {
	return f.asset, nil
}

func (f fakeReadStore) ListAssetDetails(context.Context, int) ([]store.ReadAssetDetail, error) {
	return f.assets, nil
}

func (f fakeReadStore) ListAssetHoldersByCodeIssuer(context.Context, string, string, int, int) ([]store.ReadAssetHolderSummary, error) {
	return f.assetHolders, nil
}

func (f fakeReadStore) ListTransactionSummariesByAsset(context.Context, string, string, int) ([]store.ReadTransactionSummary, error) {
	return f.assetTransactions, nil
}

func (f fakeReadStore) ListAssetTimeline(context.Context, string, string, int, int, string) ([]store.ReadTimelineItem, error) {
	return f.assetTimeline, nil
}

func (f fakeReadStore) GetContractByID(context.Context, string) (*store.ReadContractDetail, error) {
	return f.contract, nil
}

func (f fakeReadStore) GetContractSpecByID(context.Context, string) (*store.ReadContractSpec, error) {
	if f.contractSpec == nil {
		return nil, store.ErrNotFound
	}
	return f.contractSpec, nil
}

func (f fakeReadStore) ListContractDetails(context.Context, int) ([]store.ReadContractDetail, error) {
	return f.contracts, nil
}

func (f fakeReadStore) ListTransactionSummariesByContract(context.Context, string, int) ([]store.ReadTransactionSummary, error) {
	return f.contractTransactions, nil
}

func (f fakeReadStore) ListContractEventSummariesByContractID(context.Context, string, int, int) ([]store.ReadContractEventSummary, error) {
	return f.contractEvents, nil
}

func (f fakeReadStore) ListContractStorageByContractID(context.Context, string, int, int) ([]store.ReadContractStorageSummary, error) {
	return f.contractStorage, nil
}

func (f fakeReadStore) ListOperationSummariesByContract(context.Context, string, int, int) ([]store.ReadOperationSummary, error) {
	return f.contractInvocations, nil
}

func (f fakeReadStore) ListContractTimeline(context.Context, string, int, int, string) ([]store.ReadTimelineItem, error) {
	return f.contractTimeline, nil
}
