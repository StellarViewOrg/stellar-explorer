package main

import (
	"context"
	"errors"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func TestHybridLookupBackendFallsBackToRPCForUnsupportedLedgerEndpoint(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			err:   backendclient.HTTPError{StatusCode: 404, Message: "ledger endpoint not found"},
		},
		testLookupBackend{
			label: "http://rpc.test",
			ledger: backendclient.LedgerLookupResponse{
				Ledger: &backendclient.LedgerSummary{Sequence: 12345, Hash: "ledger-hash"},
			},
		},
	)

	response, err := backend.Ledger(context.Background(), 12345)
	if err != nil {
		t.Fatalf("Ledger() error = %v", err)
	}
	if response.Ledger == nil || response.Ledger.Sequence != 12345 {
		t.Fatalf("expected fallback ledger payload, got %#v", response)
	}
	if backend.Label() != "http://rpc.test (rpc fallback)" {
		t.Fatalf("expected rpc fallback label, got %q", backend.Label())
	}
	meta := backend.SourceMetadata()
	if !meta.FallbackUsed || meta.Actual != "rpc" || !meta.Degraded {
		t.Fatalf("expected rpc fallback metadata, got %#v", meta)
	}
	if meta.Policy != "single-source: indexer -> rpc; no field merge" {
		t.Fatalf("expected explicit single-source policy, got %q", meta.Policy)
	}
}

func TestHybridLookupBackendDoesNotFallbackForBadRequest(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			err:   backendclient.HTTPError{StatusCode: 400, Message: "bad query"},
		},
		testLookupBackend{
			label: "http://rpc.test",
			ledger: backendclient.LedgerLookupResponse{
				Ledger: &backendclient.LedgerSummary{Sequence: 12345},
			},
		},
	)

	_, err := backend.Ledger(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected primary bad request error")
	}
	if backend.Label() != "http://indexer.test (hybrid)" {
		t.Fatalf("expected primary label to remain active, got %q", backend.Label())
	}
	meta := backend.SourceMetadata()
	if !meta.Degraded || meta.Actual != "indexer" {
		t.Fatalf("expected degraded primary metadata, got %#v", meta)
	}
}

type testLookupBackend struct {
	label       string
	liveFeed    backendclient.LiveFeedSummaryResponse
	ledger      backendclient.LedgerLookupResponse
	transaction backendclient.TransactionLookupResponse
	account     backendclient.AccountLookupResponse
	asset       backendclient.AssetLookupResponse
	contract    backendclient.ContractLookupResponse
	err         error
}

func (b testLookupBackend) Label() string {
	return b.label
}

func (b testLookupBackend) Search(context.Context, string, int) (backendclient.SearchResponse, error) {
	if b.err != nil {
		return backendclient.SearchResponse{}, b.err
	}
	return backendclient.SearchResponse{}, nil
}

func (b testLookupBackend) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	if b.err != nil {
		return backendclient.LiveFeedSummaryResponse{}, b.err
	}
	return b.liveFeed, nil
}

func (b testLookupBackend) Ledger(context.Context, uint32) (backendclient.LedgerLookupResponse, error) {
	if b.err != nil {
		return backendclient.LedgerLookupResponse{}, b.err
	}
	return b.ledger, nil
}

func (b testLookupBackend) Transaction(context.Context, string) (backendclient.TransactionLookupResponse, error) {
	if b.err != nil {
		return backendclient.TransactionLookupResponse{}, b.err
	}
	return b.transaction, nil
}

func (b testLookupBackend) Account(context.Context, string) (backendclient.AccountLookupResponse, error) {
	if b.err != nil {
		return backendclient.AccountLookupResponse{}, b.err
	}
	return b.account, nil
}

func (b testLookupBackend) Asset(context.Context, string, string) (backendclient.AssetLookupResponse, error) {
	if b.err != nil {
		return backendclient.AssetLookupResponse{}, b.err
	}
	return b.asset, nil
}

func (b testLookupBackend) Contract(context.Context, string) (backendclient.ContractLookupResponse, error) {
	if b.err != nil {
		return backendclient.ContractLookupResponse{}, b.err
	}
	return b.contract, nil
}

func TestHybridLiveFeedSummaryFallsBackToRPC(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			err:   backendclient.HTTPError{StatusCode: 502, Message: "bad gateway"},
		},
		testLookupBackend{
			label: "http://rpc.test",
			liveFeed: backendclient.LiveFeedSummaryResponse{
				LastIngestedLedger: 42,
			},
		},
	)

	response, err := backend.LiveFeedSummary(context.Background())
	if err != nil {
		t.Fatalf("LiveFeedSummary() error = %v", err)
	}
	if response.LastIngestedLedger != 42 {
		t.Fatalf("expected fallback live feed payload, got %#v", response)
	}
	meta := backend.SourceMetadata()
	if !meta.FallbackUsed || meta.Actual != "rpc" {
		t.Fatalf("expected rpc fallback metadata, got %#v", meta)
	}
}

func TestHybridSearchFallsBackToRPC(t *testing.T) {
	backend := newHybridLookupBackend(
		testLookupBackend{
			label: "http://indexer.test",
			err:   errors.New("connection refused"),
		},
		testLookupBackend{label: "http://rpc.test"},
	)

	_, err := backend.Search(context.Background(), "GABC", 5)
	if err != nil {
		t.Fatalf("Search() should succeed via rpc fallback, got %v", err)
	}
	meta := backend.SourceMetadata()
	if meta.Actual != "rpc" || !meta.FallbackUsed {
		t.Fatalf("expected rpc fallback search metadata, got %#v", meta)
	}
}

func TestHybridContractEventsFallsBackToRPC(t *testing.T) {
	backend := newHybridLookupBackend(
		testExplorerListBackend{
			testLookupBackend: testLookupBackend{
				label: "http://indexer.test",
				err:   backendclient.HTTPError{StatusCode: 502, Message: "bad gateway"},
			},
		},
		testExplorerListBackend{
			testLookupBackend: testLookupBackend{label: "http://rpc.test"},
			contractEvents: []backendclient.ContractEventSummary{
				{ContractID: "C123", TransactionHash: "abc"},
			},
		},
	)

	response, err := backend.ContractEvents(context.Background(), "C123", 10, 0)
	if err != nil {
		t.Fatalf("ContractEvents() error = %v", err)
	}
	if len(response) != 1 || response[0].ContractID != "C123" {
		t.Fatalf("expected fallback contract events, got %#v", response)
	}
	meta := backend.SourceMetadata()
	if !meta.FallbackUsed || meta.Actual != "rpc" {
		t.Fatalf("expected rpc fallback metadata, got %#v", meta)
	}
}

type testExplorerListBackend struct {
	testLookupBackend
	contractEvents []backendclient.ContractEventSummary
}

func (b testExplorerListBackend) ContractEvents(context.Context, string, int, int) ([]backendclient.ContractEventSummary, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.contractEvents, nil
}

func (b testExplorerListBackend) ContractStorage(context.Context, string, int, int) ([]backendclient.ContractStorageSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) Ledgers(context.Context, int, uint32) ([]backendclient.LedgerSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) LedgerTransactions(context.Context, uint32, int, int) ([]backendclient.TransactionSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) Accounts(context.Context, int) ([]backendclient.AccountDetail, error) {
	return nil, b.err
}

func (b testExplorerListBackend) AccountOperations(context.Context, string, int, int) ([]backendclient.OperationSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) AccountTimeline(context.Context, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return nil, b.err
}

func (b testExplorerListBackend) Assets(context.Context, int) ([]backendclient.AssetDetail, error) {
	return nil, b.err
}

func (b testExplorerListBackend) AssetHolders(context.Context, string, string, int, int) ([]backendclient.AssetHolderSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) AssetTimeline(context.Context, string, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return nil, b.err
}

func (b testExplorerListBackend) Contracts(context.Context, int) ([]backendclient.ContractDetail, error) {
	return nil, b.err
}

func (b testExplorerListBackend) ContractInvocations(context.Context, string, int, int) ([]backendclient.OperationSummary, error) {
	return nil, b.err
}

func (b testExplorerListBackend) ContractTimeline(context.Context, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return nil, b.err
}

func TestShouldFallbackToRPCForNetworkAndUnsupportedErrors(t *testing.T) {
	tests := []error{
		backendclient.HTTPError{StatusCode: 501, Message: "not implemented"},
		errors.New("connection refused"),
		errors.New("unsupported endpoint"),
	}

	for _, err := range tests {
		if !shouldFallbackToRPC(err) {
			t.Fatalf("expected fallback for %v", err)
		}
	}
}
