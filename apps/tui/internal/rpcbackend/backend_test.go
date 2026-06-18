package rpcbackend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcclient"
	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestAccountUsesLedgerEntriesRPC(t *testing.T) {
	accountID := xdr.MustAddress("GC3C4AKRBQLHOJ45U4XG35ESVWRDECWO5XLDGYADO6DPR3L7KIDVUMML")
	key, err := accountID.LedgerKey()
	if err != nil {
		t.Fatalf("LedgerKey() error = %v", err)
	}
	keyXDR, err := key.MarshalBinaryBase64()
	if err != nil {
		t.Fatalf("MarshalBinaryBase64() error = %v", err)
	}

	otherSigner := xdr.MustSigner("GAOQJGUAB7NI7K7I62ORBXMN3J4SSWQUQ7FOEPSDJ322W2HMCNWPHXFB")
	sponsor := xdr.MustAddress("GCO26ZSBD63TKYX45H2C7D2WOFWOUSG5BMTNC3BG4QMXM3PAYI6WHKVZ")
	signerSponsor := xdr.SponsorshipDescriptor(&sponsor)
	entry := xdr.AccountEntry{
		AccountId:     accountID,
		Balance:       xdr.Int64(125000000),
		SeqNum:        xdr.SequenceNumber(77),
		NumSubEntries: xdr.Uint32(2),
		Flags:         xdr.Uint32(3),
		HomeDomain:    xdr.String32("stellar.org"),
		Thresholds:    xdr.Thresholds{1, 1, 2, 3},
		Signers: []xdr.Signer{
			{Key: otherSigner, Weight: 2},
		},
		Ext: xdr.AccountEntryExt{
			V: 1,
			V1: &xdr.AccountEntryExtensionV1{
				Liabilities: xdr.Liabilities{
					Buying:  2500000,
					Selling: 1000000,
				},
				Ext: xdr.AccountEntryExtensionV1Ext{
					V: 2,
					V2: &xdr.AccountEntryExtensionV2{
						NumSponsored:        1,
						NumSponsoring:       2,
						SignerSponsoringIDs: []xdr.SponsorshipDescriptor{signerSponsor},
						Ext: xdr.AccountEntryExtensionV2Ext{
							V: 3,
							V3: &xdr.AccountEntryExtensionV3{
								SeqLedger: 321,
								SeqTime:   xdr.TimePoint(1714000000),
							},
						},
					},
				},
			},
		},
	}

	ledgerData, err := xdr.NewLedgerEntryData(xdr.LedgerEntryTypeAccount, entry)
	if err != nil {
		t.Fatalf("NewLedgerEntryData() error = %v", err)
	}
	dataXDR, err := xdr.MarshalBase64(ledgerData)
	if err != nil {
		t.Fatalf("MarshalBase64() error = %v", err)
	}

	backend, err := New(config.Profile{
		Name:        "rpc-only",
		Network:     "testnet",
		RPCEndpoint: "http://rpc.test",
		BackendMode: "rpc",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	backend.client = rpcclient.NewWithHTTPClient("http://rpc.test", newJSONRPCClient(t, func(req map[string]any) map[string]any {
		assertLedgerEntriesRequest(t, req, keyXDR)
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"entries": []map[string]any{
					{
						"key":                   keyXDR,
						"xdr":                   dataXDR,
						"lastModifiedLedgerSeq": 1234,
					},
				},
				"latestLedger": 1234,
			},
		}
	}))

	response, err := backend.Account(context.Background(), accountID.Address())
	if err != nil {
		t.Fatalf("Account() error = %v", err)
	}

	if response.Account == nil {
		t.Fatal("expected account payload")
	}
	if response.Account.ID != accountID.Address() {
		t.Fatalf("expected account id %q, got %q", accountID.Address(), response.Account.ID)
	}
	if response.Account.Balance != "12.5000000" {
		t.Fatalf("expected formatted balance, got %q", response.Account.Balance)
	}
	if response.Account.HomeDomain == nil || *response.Account.HomeDomain != "stellar.org" {
		t.Fatalf("expected home domain, got %#v", response.Account.HomeDomain)
	}
	if len(response.Signers) != 2 {
		t.Fatalf("expected master signer plus one signer, got %d", len(response.Signers))
	}
	if response.Signers[0].Type != "master" {
		t.Fatalf("expected first signer to be master, got %q", response.Signers[0].Type)
	}
	if response.Account.SequenceLedger == nil || *response.Account.SequenceLedger != 321 {
		t.Fatalf("expected sequence ledger 321, got %#v", response.Account.SequenceLedger)
	}
	expectedSeqTime := time.Unix(1714000000, 0).UTC()
	if response.Account.SequenceTime == nil || !response.Account.SequenceTime.Equal(expectedSeqTime) {
		t.Fatalf("expected sequence time %s, got %#v", expectedSeqTime, response.Account.SequenceTime)
	}
	if response.Account.BuyingLiabilities != "0.2500000" || response.Account.SellingLiabilities != "0.1000000" {
		t.Fatalf("expected liabilities to render from entry ext, got buying=%q selling=%q", response.Account.BuyingLiabilities, response.Account.SellingLiabilities)
	}
	if response.Account.NumSponsored != 1 || response.Account.NumSponsoring != 2 {
		t.Fatalf("expected sponsorship counts 1/2, got %d/%d", response.Account.NumSponsored, response.Account.NumSponsoring)
	}
	if response.Signers[1].Sponsor == nil || *response.Signers[1].Sponsor != sponsor.Address() {
		t.Fatalf("expected signer sponsor %q, got %#v", sponsor.Address(), response.Signers[1].Sponsor)
	}
}

func TestContractUsesLedgerEntriesRPC(t *testing.T) {
	var contractHash xdr.Hash
	for i := range contractHash {
		contractHash[i] = byte(i + 1)
	}
	contractID := xdr.ContractId(contractHash)
	contractAddress, err := xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, contractID)
	if err != nil {
		t.Fatalf("NewScAddress() error = %v", err)
	}

	var key xdr.LedgerKey
	if err := key.SetContractData(
		contractAddress,
		xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance},
		xdr.ContractDataDurabilityPersistent,
	); err != nil {
		t.Fatalf("SetContractData() error = %v", err)
	}
	keyXDR, err := key.MarshalBinaryBase64()
	if err != nil {
		t.Fatalf("MarshalBinaryBase64() error = %v", err)
	}

	executable, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableWasm, contractHash)
	if err != nil {
		t.Fatalf("NewContractExecutable() error = %v", err)
	}
	nameKey, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol("name"))
	if err != nil {
		t.Fatalf("NewScVal(name key) error = %v", err)
	}
	nameVal, err := xdr.NewScVal(xdr.ScValTypeScvString, xdr.ScString("USD Coin"))
	if err != nil {
		t.Fatalf("NewScVal(name value) error = %v", err)
	}
	symbolKey, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol("symbol"))
	if err != nil {
		t.Fatalf("NewScVal(symbol key) error = %v", err)
	}
	symbolVal, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol("USDC"))
	if err != nil {
		t.Fatalf("NewScVal(symbol value) error = %v", err)
	}
	decimalsKey, err := xdr.NewScVal(xdr.ScValTypeScvSymbol, xdr.ScSymbol("decimals"))
	if err != nil {
		t.Fatalf("NewScVal(decimals key) error = %v", err)
	}
	decimalsVal, err := xdr.NewScVal(xdr.ScValTypeScvU32, xdr.Uint32(7))
	if err != nil {
		t.Fatalf("NewScVal(decimals value) error = %v", err)
	}
	storage := xdr.ScMap{
		{Key: nameKey, Val: nameVal},
		{Key: symbolKey, Val: symbolVal},
		{Key: decimalsKey, Val: decimalsVal},
	}
	entry := xdr.ContractDataEntry{
		Contract:   contractAddress,
		Key:        xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance},
		Durability: xdr.ContractDataDurabilityPersistent,
		Val: xdr.ScVal{
			Type: xdr.ScValTypeScvContractInstance,
			Instance: &xdr.ScContractInstance{
				Executable: executable,
				Storage:    &storage,
			},
		},
	}

	ledgerData, err := xdr.NewLedgerEntryData(xdr.LedgerEntryTypeContractData, entry)
	if err != nil {
		t.Fatalf("NewLedgerEntryData() error = %v", err)
	}
	dataXDR, err := xdr.MarshalBase64(ledgerData)
	if err != nil {
		t.Fatalf("MarshalBase64() error = %v", err)
	}

	backend, err := New(config.Profile{
		Name:        "rpc-only",
		Network:     "testnet",
		RPCEndpoint: "http://rpc.test",
		BackendMode: "rpc",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	backend.client = rpcclient.NewWithHTTPClient("http://rpc.test", newJSONRPCClient(t, func(req map[string]any) map[string]any {
		assertLedgerEntriesRequest(t, req, keyXDR)
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"entries": []map[string]any{
					{
						"key":                   keyXDR,
						"xdr":                   dataXDR,
						"lastModifiedLedgerSeq": 2222,
					},
				},
				"latestLedger": 2222,
			},
		}
	}))

	contractAddressStr := strkey.MustEncode(strkey.VersionByteContract, contractID[:])
	response, err := backend.Contract(context.Background(), contractAddressStr)
	if err != nil {
		t.Fatalf("Contract() error = %v", err)
	}

	if response.Contract == nil {
		t.Fatal("expected contract payload")
	}
	if response.Contract.ContractID != contractAddressStr {
		t.Fatalf("expected contract id %q, got %q", contractAddressStr, response.Contract.ContractID)
	}
	if response.Contract.ContractType != 0 {
		t.Fatalf("expected wasm contract type, got %d", response.Contract.ContractType)
	}
	if response.Contract.LastModifiedLedger != 2222 {
		t.Fatalf("expected last modified ledger 2222, got %d", response.Contract.LastModifiedLedger)
	}
	expectedHash := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	if response.Contract.WasmHash == nil || *response.Contract.WasmHash != expectedHash {
		t.Fatalf("expected wasm hash %q, got %#v", expectedHash, response.Contract.WasmHash)
	}
	if response.Contract.StorageEntryCount != 3 {
		t.Fatalf("expected storage entry count 3, got %d", response.Contract.StorageEntryCount)
	}
	if !response.Contract.IsSep41Token {
		t.Fatal("expected contract to be classified as token from storage metadata")
	}
	if response.Contract.TokenName == nil || *response.Contract.TokenName != "USD Coin" {
		t.Fatalf("expected token name %q, got %#v", "USD Coin", response.Contract.TokenName)
	}
	if response.Contract.TokenSymbol == nil || *response.Contract.TokenSymbol != "USDC" {
		t.Fatalf("expected token symbol %q, got %#v", "USDC", response.Contract.TokenSymbol)
	}
	if response.Contract.TokenDecimals == nil || *response.Contract.TokenDecimals != 7 {
		t.Fatalf("expected token decimals 7, got %#v", response.Contract.TokenDecimals)
	}
	if response.Contract.Label == nil || *response.Contract.Label != "USDC" {
		t.Fatalf("expected label %q, got %#v", "USDC", response.Contract.Label)
	}
}

func TestLedgerUsesGetLedgersRPC(t *testing.T) {
	backend, err := New(config.Profile{
		Name:        "rpc-only",
		Network:     "testnet",
		RPCEndpoint: "http://rpc.test",
		BackendMode: "rpc",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	backend.client = rpcclient.NewWithHTTPClient("http://rpc.test", newJSONRPCClient(t, func(req map[string]any) map[string]any {
		method, _ := req["method"].(string)
		if method != "getLedgers" {
			t.Fatalf("expected getLedgers method, got %q", method)
		}
		params, ok := req["params"].(map[string]any)
		if !ok {
			t.Fatalf("expected params object, got %#v", req["params"])
		}
		if params["startLedger"] != float64(12345) {
			t.Fatalf("expected startLedger 12345, got %#v", params["startLedger"])
		}

		return map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"ledgers": []map[string]any{
					{
						"id":                         "ledger-hash",
						"sequence":                   12345,
						"closeTime":                  "2026-04-20T12:00:00Z",
						"transactionCount":           2,
						"operationCount":             5,
						"successfulTransactionCount": 1,
						"failedTransactionCount":     1,
					},
				},
			},
		}
	}))

	response, err := backend.Ledger(context.Background(), 12345)
	if err != nil {
		t.Fatalf("Ledger() error = %v", err)
	}
	if response.Ledger == nil {
		t.Fatal("expected ledger payload")
	}
	if response.Ledger.Sequence != 12345 {
		t.Fatalf("expected sequence 12345, got %d", response.Ledger.Sequence)
	}
	if response.Ledger.TransactionCount != 2 || response.Ledger.OperationCount != 5 {
		t.Fatalf("expected tx/op counts, got %#v", response.Ledger)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newJSONRPCClient(t *testing.T, responder func(map[string]any) map[string]any) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var payload map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}

			responsePayload := responder(payload)
			body, err := json.Marshal(responsePayload)
			if err != nil {
				t.Fatalf("marshal response: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	}
}

func assertLedgerEntriesRequest(t *testing.T, req map[string]any, expectedKey string) {
	t.Helper()

	method, _ := req["method"].(string)
	if method != "getLedgerEntries" {
		t.Fatalf("expected getLedgerEntries method, got %q", method)
	}

	params, ok := req["params"].(map[string]any)
	if !ok {
		t.Fatalf("expected params object, got %#v", req["params"])
	}
	keys, ok := params["keys"].([]any)
	if !ok {
		t.Fatalf("expected keys array, got %#v", params["keys"])
	}
	if len(keys) != 1 || keys[0] != expectedKey {
		t.Fatalf("unexpected keys: %#v", keys)
	}
}
