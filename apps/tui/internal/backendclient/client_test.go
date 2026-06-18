package backendclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newTestClient(t *testing.T, fn roundTripFunc) *Client {
	t.Helper()

	client, err := New(
		"http://tui-indexer.test",
		WithHTTPClient(&http.Client{Transport: fn}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return client
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewRejectsInvalidBaseURL(t *testing.T) {
	t.Parallel()

	if _, err := New(""); err == nil {
		t.Fatal("expected error for empty base URL")
	}

	if _, err := New("localhost:8081"); err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/healthz" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{"status":"ok","database":"ok","rpc":"disabled","last_ingested_ledger":99}`), nil
	})

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if health.Status != "ok" {
		t.Fatalf("expected status ok, got %q", health.Status)
	}

	if health.LastIngestedLedger != 99 {
		t.Fatalf("expected last ingested ledger 99, got %d", health.LastIngestedLedger)
	}
}

func TestLiveFeedSummary(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/feed/live/summary" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"last_ingested_ledger":1234,
			"latest_ledger":{
				"sequence":1234,
				"hash":"ledger-hash",
				"closed_at":"`+now+`",
				"transaction_count":2,
				"operation_count":5,
				"successful_tx_count":2,
				"failed_tx_count":0
			},
			"recent_transactions":[
				{
					"hash":"tx-1",
					"ledger_sequence":1234,
					"application_order":1,
					"account":"GABC",
					"operation_count":2,
					"status":1,
					"is_soroban":true,
					"created_at":"`+now+`"
				}
			]
		}`), nil
	})

	summary, err := client.LiveFeedSummary(context.Background())
	if err != nil {
		t.Fatalf("LiveFeedSummary() error = %v", err)
	}

	if summary.LastIngestedLedger != 1234 {
		t.Fatalf("expected last ingested ledger 1234, got %d", summary.LastIngestedLedger)
	}

	if summary.LatestLedger == nil || summary.LatestLedger.Hash != "ledger-hash" {
		t.Fatalf("unexpected latest ledger: %+v", summary.LatestLedger)
	}

	if len(summary.RecentTransactions) != 1 || summary.RecentTransactions[0].Hash != "tx-1" {
		t.Fatalf("unexpected recent transactions: %+v", summary.RecentTransactions)
	}
}

func TestSearch(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/search" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "GABC/123" {
			t.Fatalf("unexpected query %q", r.URL.RawQuery)
		}
		if r.URL.Query().Get("limit") != "8" {
			t.Fatalf("unexpected limit %q", r.URL.RawQuery)
		}

		return jsonResponse(http.StatusOK, `{"results":[{"kind":"account","title":"Account GABC","description":"balance 10","command":"lookup account GABC"}]}`), nil
	})

	response, err := client.Search(context.Background(), "GABC/123", 8)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(response.Results) != 1 || response.Results[0].Command != "lookup account GABC" {
		t.Fatalf("unexpected search results: %+v", response.Results)
	}
}

func TestLedger(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/ledgers/55" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"ledger":{
				"sequence":55,
				"hash":"ledger-hash",
				"closed_at":"`+now+`",
				"transaction_count":2,
				"operation_count":3,
				"successful_tx_count":2,
				"failed_tx_count":0
			},
			"transactions":[
				{"hash":"tx-ledger-1","ledger_sequence":55,"application_order":1,"account":"GABC","operation_count":2,"status":1,"is_soroban":false,"created_at":"`+now+`"}
			]
		}`), nil
	})

	response, err := client.Ledger(context.Background(), 55)
	if err != nil {
		t.Fatalf("Ledger() error = %v", err)
	}
	if response.Ledger == nil || response.Ledger.Sequence != 55 {
		t.Fatalf("unexpected ledger: %+v", response.Ledger)
	}
	if len(response.Transactions) != 1 || response.Transactions[0].Hash != "tx-ledger-1" {
		t.Fatalf("unexpected ledger transactions: %+v", response.Transactions)
	}
}

func TestLedgerTransactions(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/ledgers/55/transactions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "3" {
			t.Fatalf("unexpected query %q", r.URL.RawQuery)
		}

		return jsonResponse(http.StatusOK, `{
			"transactions":[
				{"hash":"tx-ledger-list-1","ledger_sequence":55,"application_order":1,"account":"GABC","operation_count":2,"status":1,"is_soroban":false,"created_at":"`+now+`"}
			]
		}`), nil
	})

	response, err := client.LedgerTransactions(context.Background(), 55, 3, 0)
	if err != nil {
		t.Fatalf("LedgerTransactions() error = %v", err)
	}
	if len(response) != 1 || response[0].Hash != "tx-ledger-list-1" {
		t.Fatalf("unexpected ledger transaction list: %+v", response)
	}
}

func TestTransaction(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/transactions/abc%2F123" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"transaction":{
				"hash":"abc/123",
				"ledger_sequence":55,
				"application_order":1,
				"account":"GABC",
				"account_sequence":99,
				"fee_charged":100,
				"max_fee":200,
				"operation_count":1,
				"memo_type":0,
				"status":1,
				"is_soroban":false,
				"envelope_xdr":"env",
				"result_xdr":"res",
				"created_at":"`+now+`"
			},
			"operations":[
				{
					"application_order":1,
					"type":1,
					"type_name":"payment",
					"details":"{}",
					"created_at":"`+now+`"
				}
			],
			"effects":[
				{
					"transaction_hash":"abc/123",
					"type":0,
					"type_name":"account_credited",
					"account":"GDEST",
					"details":"{}",
					"created_at":"`+now+`"
				}
			]
		}`), nil
	})

	response, err := client.Transaction(context.Background(), "abc/123")
	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	if response.Transaction == nil || response.Transaction.Hash != "abc/123" {
		t.Fatalf("unexpected transaction: %+v", response.Transaction)
	}

	if len(response.Operations) != 1 || response.Operations[0].TypeName != "payment" {
		t.Fatalf("unexpected operations: %+v", response.Operations)
	}
	if len(response.Effects) != 1 || response.Effects[0].TypeName != "account_credited" {
		t.Fatalf("unexpected effects: %+v", response.Effects)
	}
}

func TestTransactionRejectsEmptyHash(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		t.Fatal("unexpected request")
		return nil, nil
	})

	if _, err := client.Transaction(context.Background(), " "); err == nil {
		t.Fatal("expected error for empty transaction hash")
	}
}

func TestAccount(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/accounts/GABC%2F123" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"account":{
				"id":"GABC/123",
				"sequence":7,
				"balance":"100.0",
				"buying_liabilities":"0",
				"selling_liabilities":"0",
				"num_subentries":2,
				"flags":1,
				"last_modified_ledger":9,
				"num_sponsored":0,
				"num_sponsoring":0,
				"updated_at":"`+now+`"
			},
			"trustlines":[{"asset_type":1,"asset_code":"USDC","asset_issuer":"GISS","balance":"5","limit_amount":"100","buying_liabilities":"0","selling_liabilities":"0","flags":0,"last_modified_ledger":10,"updated_at":"`+now+`"}],
			"signers":[{"signer_key":"GABC/123","weight":1,"type":"ed25519","last_modified_ledger":10}],
			"recent_transactions":[{"hash":"tx-1","ledger_sequence":55,"application_order":1,"account":"GABC/123","operation_count":2,"status":1,"is_soroban":false,"created_at":"`+now+`"}],
			"recent_operations":[{"transaction_hash":"tx-1","application_order":1,"type":1,"type_name":"payment","details":"{}","created_at":"`+now+`"}]
		}`), nil
	})

	response, err := client.Account(context.Background(), "GABC/123")
	if err != nil {
		t.Fatalf("Account() error = %v", err)
	}

	if response.Account == nil || response.Account.ID != "GABC/123" {
		t.Fatalf("unexpected account: %+v", response.Account)
	}

	if len(response.Trustlines) != 1 || response.Trustlines[0].AssetCode != "USDC" {
		t.Fatalf("unexpected trustlines: %+v", response.Trustlines)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-1" {
		t.Fatalf("unexpected account transactions: %+v", response.RecentTransactions)
	}
	if len(response.RecentOperations) != 1 || response.RecentOperations[0].TypeName != "payment" {
		t.Fatalf("unexpected account operations: %+v", response.RecentOperations)
	}
	if response.RecentOperations[0].TransactionHash != "tx-1" {
		t.Fatalf("unexpected operation transaction hash: %+v", response.RecentOperations[0])
	}
}

func TestAccountSubresourceLists(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	t.Run("transactions", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/accounts/GABC/transactions" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "4" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"transactions":[
					{"hash":"tx-account-list","ledger_sequence":55,"application_order":1,"account":"GABC","operation_count":2,"status":1,"is_soroban":false,"created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AccountTransactions(context.Background(), "GABC", 4)
		if err != nil {
			t.Fatalf("AccountTransactions() error = %v", err)
		}
		if len(response) != 1 || response[0].Hash != "tx-account-list" {
			t.Fatalf("unexpected account transaction list: %+v", response)
		}
	})

	t.Run("operations", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/accounts/GABC/operations" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "4" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"operations":[
					{"transaction_hash":"tx-account-list","application_order":1,"type":1,"type_name":"payment","details":"{}","created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AccountOperations(context.Background(), "GABC", 4, 0)
		if err != nil {
			t.Fatalf("AccountOperations() error = %v", err)
		}
		if len(response) != 1 || response[0].TypeName != "payment" {
			t.Fatalf("unexpected account operation list: %+v", response)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/accounts/GABC/timeline" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "4" || r.URL.Query().Get("offset") != "2" || r.URL.Query().Get("type") != "op" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"items":[
					{"kind":"tx","title":"Transaction tx-account-list","description":"ledger 123","command":"lookup tx tx-account-list","occurred_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AccountTimeline(context.Background(), "GABC", 4, 2, "op")
		if err != nil {
			t.Fatalf("AccountTimeline() error = %v", err)
		}
		if len(response) != 1 || response[0].Command != "lookup tx tx-account-list" {
			t.Fatalf("unexpected account timeline: %+v", response)
		}
	})
}

func TestAsset(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/assets/USDC:GISS" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"asset":{
				"asset_type":1,
				"asset_code":"USDC",
				"asset_issuer":"GISS",
				"num_accounts":100,
				"total_supply":"1000.0000000",
				"num_claimable_balances":2,
				"num_liquidity_pools":3,
				"num_contracts":1,
				"flags":0,
				"auth_required":false,
				"auth_revocable":true,
				"auth_immutable":false,
				"clawback_enabled":false,
				"home_domain":"example.com",
				"updated_at":"`+now+`"
			},
			"top_holders":[{"account_id":"GHOLDER","balance":"100.0","limit_amount":"1000.0","buying_liabilities":"0","selling_liabilities":"0","last_modified_ledger":55,"updated_at":"`+now+`"}],
			"recent_transactions":[{"hash":"tx-asset-1","ledger_sequence":56,"application_order":1,"account":"GHOLDER","operation_count":1,"status":1,"is_soroban":false,"created_at":"`+now+`"}]
		}`), nil
	})

	response, err := client.Asset(context.Background(), "USDC", "GISS")
	if err != nil {
		t.Fatalf("Asset() error = %v", err)
	}
	if response.Asset == nil || response.Asset.AssetCode != "USDC" || response.Asset.NumAccounts != 100 {
		t.Fatalf("unexpected asset: %+v", response.Asset)
	}
	if len(response.TopHolders) != 1 || response.TopHolders[0].AccountID != "GHOLDER" {
		t.Fatalf("unexpected top holders: %+v", response.TopHolders)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-asset-1" {
		t.Fatalf("unexpected asset transactions: %+v", response.RecentTransactions)
	}
}

func TestAssetSubresourceLists(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	t.Run("transactions", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/assets/USDC:GISS/transactions" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "6" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"transactions":[
					{"hash":"tx-asset-list","ledger_sequence":55,"application_order":1,"account":"GABC","operation_count":2,"status":1,"is_soroban":false,"created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AssetTransactions(context.Background(), "USDC", "GISS", 6)
		if err != nil {
			t.Fatalf("AssetTransactions() error = %v", err)
		}
		if len(response) != 1 || response[0].Hash != "tx-asset-list" {
			t.Fatalf("unexpected asset transaction list: %+v", response)
		}
	})

	t.Run("holders", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/assets/USDC:GISS/holders" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "6" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"holders":[
					{"account_id":"GHOLDER","balance":"10","limit_amount":"100","buying_liabilities":"0","selling_liabilities":"0","last_modified_ledger":10,"updated_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AssetHolders(context.Background(), "USDC", "GISS", 6, 0)
		if err != nil {
			t.Fatalf("AssetHolders() error = %v", err)
		}
		if len(response) != 1 || response[0].AccountID != "GHOLDER" {
			t.Fatalf("unexpected asset holder list: %+v", response)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/assets/USDC:GISS/timeline" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "6" || r.URL.Query().Get("offset") != "2" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"items":[
					{"kind":"holder","title":"Holder GHOLDER","description":"balance 10","command":"lookup account GHOLDER","occurred_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.AssetTimeline(context.Background(), "USDC", "GISS", 6, 2, "")
		if err != nil {
			t.Fatalf("AssetTimeline() error = %v", err)
		}
		if len(response) != 1 || response[0].Command != "lookup account GHOLDER" {
			t.Fatalf("unexpected asset timeline: %+v", response)
		}
	})
}

func TestContract(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/contracts/CABC%2F123" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		return jsonResponse(http.StatusOK, `{
			"contract":{
				"contract_id":"CABC/123",
				"created_ledger":11,
				"created_at":"`+now+`",
				"last_modified_ledger":12,
				"contract_type":2,
				"is_sep41_token":true,
				"is_sep50_nft":false,
				"storage_entry_count":4,
				"event_count":8,
				"invocation_count":3,
				"updated_at":"`+now+`"
			},
			"recent_transactions":[{"hash":"tx-contract-1","ledger_sequence":57,"application_order":1,"account":"GABC","operation_count":1,"status":1,"is_soroban":true,"created_at":"`+now+`"}],
			"recent_events":[{"transaction_hash":"tx-contract-1","ledger_sequence":57,"type":0,"created_at":"`+now+`"}]
		}`), nil
	})

	response, err := client.Contract(context.Background(), "CABC/123")
	if err != nil {
		t.Fatalf("Contract() error = %v", err)
	}

	if response.Contract == nil || response.Contract.ContractID != "CABC/123" {
		t.Fatalf("unexpected contract: %+v", response.Contract)
	}
	if len(response.RecentTransactions) != 1 || response.RecentTransactions[0].Hash != "tx-contract-1" {
		t.Fatalf("unexpected contract transactions: %+v", response.RecentTransactions)
	}
	if len(response.RecentEvents) != 1 || response.RecentEvents[0].TransactionHash != "tx-contract-1" {
		t.Fatalf("unexpected contract events: %+v", response.RecentEvents)
	}
}

func TestContractSubresourceLists(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	t.Run("transactions", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/contracts/CCONTRACT/transactions" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "5" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"transactions":[
					{"hash":"tx-contract-list","ledger_sequence":88,"application_order":1,"account":"GABC","operation_count":1,"status":1,"is_soroban":true,"created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.ContractTransactions(context.Background(), "CCONTRACT", 5)
		if err != nil {
			t.Fatalf("ContractTransactions() error = %v", err)
		}
		if len(response) != 1 || response[0].Hash != "tx-contract-list" {
			t.Fatalf("unexpected contract transaction list: %+v", response)
		}
	})

	t.Run("storage", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/contracts/CCONTRACT/storage" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "5" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"storage":[
					{"contract_id":"CCONTRACT","display_key":"balance","display_value":"100","durability_label":"persistent","key_xdr":"key","value_xdr":"value","durability":1,"last_modified_ledger":10,"updated_at":"`+now+`","decode_status":"decoded"}
				]
			}`), nil
		})

		response, err := client.ContractStorage(context.Background(), "CCONTRACT", 5, 0)
		if err != nil {
			t.Fatalf("ContractStorage() error = %v", err)
		}
		if len(response) != 1 || response[0].DisplayKey != "balance" {
			t.Fatalf("unexpected contract storage list: %+v", response)
		}
	})

	t.Run("invocations", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/contracts/CCONTRACT/invocations" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}

			return jsonResponse(http.StatusOK, `{
				"operations":[
					{"transaction_hash":"tx-invoke-1","application_order":1,"type":24,"type_name":"invoke_host_function","details":"{}","created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.ContractInvocations(context.Background(), "CCONTRACT", 5, 0)
		if err != nil {
			t.Fatalf("ContractInvocations() error = %v", err)
		}
		if len(response) != 1 || response[0].TypeName != "invoke_host_function" {
			t.Fatalf("unexpected contract invocation list: %+v", response)
		}
	})

	t.Run("events", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/contracts/CCONTRACT/events" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "5" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"events":[
					{"transaction_hash":"tx-contract-list","ledger_sequence":88,"type":0,"created_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.ContractEvents(context.Background(), "CCONTRACT", 5, 0)
		if err != nil {
			t.Fatalf("ContractEvents() error = %v", err)
		}
		if len(response) != 1 || response[0].TransactionHash != "tx-contract-list" {
			t.Fatalf("unexpected contract event list: %+v", response)
		}
	})

	t.Run("timeline", func(t *testing.T) {
		client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/contracts/CCONTRACT/timeline" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.URL.Query().Get("limit") != "5" || r.URL.Query().Get("offset") != "2" {
				t.Fatalf("unexpected query %q", r.URL.RawQuery)
			}

			return jsonResponse(http.StatusOK, `{
				"items":[
					{"kind":"event","title":"Event type 0","description":"ledger 88","command":"lookup tx tx-contract-list","occurred_at":"`+now+`"}
				]
			}`), nil
		})

		response, err := client.ContractTimeline(context.Background(), "CCONTRACT", 5, 2, "")
		if err != nil {
			t.Fatalf("ContractTimeline() error = %v", err)
		}
		if len(response) != 1 || response[0].Command != "lookup tx tx-contract-list" {
			t.Fatalf("unexpected contract timeline: %+v", response)
		}
	})
}

func TestHTTPError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusNotFound, `{"error":"transaction not found"}`), nil
	})

	_, err := client.Transaction(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected HTTP error")
	}

	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}

	if httpErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", httpErr.StatusCode)
	}

	if httpErr.Message != "transaction not found" {
		t.Fatalf("expected message %q, got %q", "transaction not found", httpErr.Message)
	}
}
