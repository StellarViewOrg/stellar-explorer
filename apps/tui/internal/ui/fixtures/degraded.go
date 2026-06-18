package fixtures

import (
	"context"
	"errors"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

// CacheFallbackTransactionModel returns a lookup model restored from local cache.
func CacheFallbackTransactionModel() *app.Model {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{
		Enabled:   true,
		Available: true,
		Schema:    7,
	})
	model.UpdateLookupTransaction(
		"tx-cache-degraded",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-cache-degraded",
				LedgerSequence: 42,
				Account:        "GSOURCE",
				OperationCount: 1,
				Status:         1,
				CreatedAt:      time.Unix(1700000000, 0).UTC(),
			},
		},
		app.SourceMetadata{
			Mode:           config.BackendModeHybrid,
			Operation:      "transaction",
			Preferred:      "indexer",
			Actual:         "cache",
			Label:          "local cache",
			FallbackUsed:   true,
			Degraded:       true,
			DegradedReason: "backend unavailable; restored from local cache",
		},
	)
	return model
}

// RPCEffectsUnavailableModel returns a transaction lookup without indexed effects.
func RPCEffectsUnavailableModel() *app.Model {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-rpc-only",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-rpc-only",
				LedgerSequence: 77,
				Account:        "GSOURCE",
				OperationCount: 1,
				Status:         1,
			},
		},
		app.SourceMetadata{
			Mode:      config.BackendModeRPC,
			Operation: "transaction",
			Preferred: "rpc",
			Actual:    "rpc",
			Label:     "https://soroban-testnet.stellar.org",
		},
	)
	return model
}

// HybridFallbackLedgerModel returns a ledger lookup resolved via RPC fallback.
func HybridFallbackLedgerModel() *app.Model {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"12345",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{
				Sequence:         12345,
				Hash:             "ledger-fallback",
				TransactionCount: 1,
				OperationCount:   2,
			},
		},
		app.SourceMetadata{
			Mode:           config.BackendModeHybrid,
			Operation:      "ledger",
			Preferred:      "indexer",
			Actual:         "rpc",
			Label:          "https://soroban-testnet.stellar.org (rpc fallback)",
			FallbackUsed:   true,
			Degraded:       true,
			DegradedReason: "indexer unavailable",
			Policy:         "single-source: indexer -> rpc; no field merge",
		},
	)
	return model
}

type failingLiveFeedService struct{}

func (failingLiveFeedService) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return backendclient.LiveFeedSummaryResponse{}, errors.New("connection refused")
}

// LiveFeedUnavailableModel returns a live feed in backend error state.
func LiveFeedUnavailableModel() *app.Model {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	_ = model.SetScreen(app.ScreenLiveFeed)
	_ = model.RefreshLiveFeed(context.Background(), failingLiveFeedService{}, app.DefaultSourceMetadata(config.Default().Profiles[0], "live-feed"))
	return model
}

// SorobanRawDecodeModel returns a contract lookup in raw decode mode.
func SorobanRawDecodeModel() *app.Model {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{
				ContractID:        "CCONTRACT",
				ContractType:      1,
				StorageEntryCount: 1,
				EventCount:        1,
			},
			Spec: &backendclient.ContractSpec{
				ContractID:    "CCONTRACT",
				Available:     true,
				DecodeStatus:  "raw",
				FunctionCount: 1,
			},
			Storage: []backendclient.ContractStorageSummary{
				{
					ContractID:      "CCONTRACT",
					DisplayKey:      "balance",
					DisplayValue:    "100",
					DurabilityLabel: "persistent",
					KeyXDR:          "raw-key-xdr",
					ValueXDR:        "raw-value-xdr",
				},
			},
		},
		app.DefaultSourceMetadata(config.Default().Profiles[0], "contract"),
	)
	model.SetContractDecodeMode(app.ContractDecodeModeRaw)
	return model
}
