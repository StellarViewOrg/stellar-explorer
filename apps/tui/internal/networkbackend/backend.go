package networkbackend

import (
	"context"
	"errors"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/horizonbackend"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcbackend"
)

// Backend composes Horizon and RPC into one network-facing lookup backend.
type Backend struct {
	profile  config.Profile
	rpc      *rpcbackend.Backend
	horizon  *horizonbackend.Backend
	lastMeta app.SourceMetadata
}

// New constructs a network backend for the given profile.
func New(profile config.Profile) (*Backend, error) {
	rpc, err := rpcbackend.New(profile)
	if err != nil {
		return nil, err
	}

	var horizon *horizonbackend.Backend
	if config.ResolveHorizonURL(profile) != "" {
		horizon, _ = horizonbackend.New(profile)
	}

	return &Backend{
		profile: profile,
		rpc:     rpc,
		horizon: horizon,
		lastMeta: app.SourceMetadata{
			Mode:      profile.NormalizedBackendMode(),
			Policy:    "network: horizon history + rpc soroban/state",
			Preferred: "horizon",
			Actual:    actualSource(horizon != nil),
			Label:     combinedLabel(profile, horizon != nil),
		},
	}, nil
}

// Label returns the configured network source label.
func (b *Backend) Label() string {
	if b == nil {
		return ""
	}
	if strings.TrimSpace(b.lastMeta.Label) != "" {
		return b.lastMeta.Label
	}
	return combinedLabel(b.profile, b.horizon != nil)
}

// SourceMetadata returns the most recent routing metadata.
func (b *Backend) SourceMetadata() app.SourceMetadata {
	if b == nil {
		return app.SourceMetadata{}
	}
	return b.lastMeta
}

func (b *Backend) Search(ctx context.Context, query string, limit int) (backendclient.SearchResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.Search(ctx, query, limit)
		b.setMeta("search", "horizon", err != nil, errorText(err))
		return response, err
	}
	b.setMeta("search", "rpc", false, "")
	return backendclient.SearchResponse{Results: []backendclient.SearchResult{}}, nil
}

func (b *Backend) LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.LiveFeedSummary(ctx)
		if err == nil {
			b.setMeta("live-feed", "horizon", false, "")
			return response, nil
		}
		response, rpcErr := b.rpc.LiveFeedSummary(ctx)
		if rpcErr != nil {
			b.setMeta("live-feed", "rpc", true, err.Error())
			return backendclient.LiveFeedSummaryResponse{}, rpcErr
		}
		b.setMeta("live-feed", "rpc", true, err.Error())
		return response, nil
	}
	response, err := b.rpc.LiveFeedSummary(ctx)
	b.setMeta("live-feed", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Ledger(ctx context.Context, sequence uint32) (backendclient.LedgerLookupResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.Ledger(ctx, sequence)
		b.setMeta("ledger", "horizon", err != nil, errorText(err))
		return response, err
	}
	response, err := b.rpc.Ledger(ctx, sequence)
	b.setMeta("ledger", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Ledgers(ctx context.Context, limit int, before uint32) ([]backendclient.LedgerSummary, error) {
	if b.horizon != nil {
		response, err := b.horizon.Ledgers(ctx, limit, before)
		b.setMeta("ledgers", "horizon", err != nil, errorText(err))
		return response, err
	}
	response, err := b.rpc.Ledgers(ctx, limit, before)
	b.setMeta("ledgers", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]backendclient.TransactionSummary, error) {
	if b.horizon != nil {
		response, err := b.horizon.LedgerTransactions(ctx, sequence, limit, offset)
		b.setMeta("ledger-transactions", "horizon", err != nil, errorText(err))
		return response, err
	}
	response, err := b.rpc.LedgerTransactions(ctx, sequence, limit, offset)
	b.setMeta("ledger-transactions", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Transaction(ctx context.Context, hash string) (backendclient.TransactionLookupResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.Transaction(ctx, hash)
		if err != nil {
			fallback, rpcErr := b.rpc.Transaction(ctx, hash)
			if rpcErr != nil {
				b.setMeta("transaction", "rpc", true, err.Error())
				return backendclient.TransactionLookupResponse{}, rpcErr
			}
			b.setMeta("transaction", "rpc", true, err.Error())
			return fallback, nil
		}
		rpcResponse, rpcErr := b.rpc.Transaction(ctx, hash)
		if rpcErr == nil {
			response = mergeTransactionLookup(response, rpcResponse)
			_ = b.rpc.EnrichTransaction(ctx, &response)
			b.setMeta("transaction", "horizon+rpc", false, "")
			return response, nil
		}
		_ = b.rpc.EnrichTransaction(ctx, &response)
		b.setMeta("transaction", "horizon", false, "")
		return response, nil
	}
	response, err := b.rpc.Transaction(ctx, hash)
	b.setMeta("transaction", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Account(ctx context.Context, id string) (backendclient.AccountLookupResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.Account(ctx, id)
		if err != nil {
			fallback, rpcErr := b.rpc.Account(ctx, id)
			if rpcErr != nil {
				b.setMeta("account", "rpc", true, err.Error())
				return backendclient.AccountLookupResponse{}, rpcErr
			}
			b.setMeta("account", "rpc", true, err.Error())
			return fallback, nil
		}
		rpcResponse, rpcErr := b.rpc.Account(ctx, id)
		if rpcErr == nil {
			response = mergeAccountLookup(response, rpcResponse)
			b.setMeta("account", "horizon+rpc", false, "")
			return response, nil
		}
		b.setMeta("account", "horizon", false, "")
		return response, nil
	}
	response, err := b.rpc.Account(ctx, id)
	b.setMeta("account", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Accounts(ctx context.Context, limit int) ([]backendclient.AccountDetail, error) {
	if b.horizon != nil {
		response, err := b.horizon.Accounts(ctx, limit)
		b.setMeta("accounts", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("account list requires horizon")
}

func (b *Backend) AccountOperations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	if b.horizon != nil {
		response, err := b.horizon.AccountOperations(ctx, id, limit, offset)
		b.setMeta("account-operations", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("account operation list requires horizon")
}

func (b *Backend) AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	if b.horizon != nil {
		response, err := b.horizon.AccountTimeline(ctx, id, limit, offset, category)
		b.setMeta("account-timeline", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("account timeline requires horizon")
}

func (b *Backend) Asset(ctx context.Context, code string, issuer string) (backendclient.AssetLookupResponse, error) {
	if b.horizon != nil {
		response, err := b.horizon.Asset(ctx, code, issuer)
		b.setMeta("asset", "horizon", err != nil, errorText(err))
		return response, err
	}
	return backendclient.AssetLookupResponse{}, errors.New("asset lookup requires horizon")
}

func (b *Backend) Assets(ctx context.Context, limit int) ([]backendclient.AssetDetail, error) {
	if b.horizon != nil {
		response, err := b.horizon.Assets(ctx, limit)
		b.setMeta("assets", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("asset list requires horizon")
}

func (b *Backend) AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]backendclient.AssetHolderSummary, error) {
	if b.horizon != nil {
		response, err := b.horizon.AssetHolders(ctx, code, issuer, limit, offset)
		b.setMeta("asset-holders", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("asset holder list requires horizon")
}

func (b *Backend) AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	if b.horizon != nil {
		response, err := b.horizon.AssetTimeline(ctx, code, issuer, limit, offset, category)
		b.setMeta("asset-timeline", "horizon", err != nil, errorText(err))
		return response, err
	}
	return nil, errors.New("asset timeline requires horizon")
}

func (b *Backend) Contract(ctx context.Context, id string) (backendclient.ContractLookupResponse, error) {
	response, err := b.rpc.Contract(ctx, id)
	b.setMeta("contract", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) Contracts(ctx context.Context, limit int) ([]backendclient.ContractDetail, error) {
	return nil, errors.New("contract list is unavailable in network mode")
}

func (b *Backend) ContractEvents(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractEventSummary, error) {
	response, err := b.rpc.ContractEvents(ctx, id, limit, offset)
	b.setMeta("contract-events", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) ContractStorage(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractStorageSummary, error) {
	response, err := b.rpc.ContractStorage(ctx, id, limit, offset)
	b.setMeta("contract-storage", "rpc", err != nil, errorText(err))
	return response, err
}

func (b *Backend) ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	return nil, errors.New("contract invocation list requires tui-indexer")
}

func (b *Backend) ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	return nil, errors.New("contract timeline requires tui-indexer")
}

func mergeTransactionLookup(horizon backendclient.TransactionLookupResponse, rpc backendclient.TransactionLookupResponse) backendclient.TransactionLookupResponse {
	if horizon.Transaction == nil || rpc.Transaction == nil {
		return horizon
	}
	if strings.TrimSpace(horizon.Transaction.EnvelopeXDR) == "" && strings.TrimSpace(rpc.Transaction.EnvelopeXDR) != "" {
		horizon.Transaction.EnvelopeXDR = rpc.Transaction.EnvelopeXDR
	}
	if strings.TrimSpace(horizon.Transaction.ResultXDR) == "" && strings.TrimSpace(rpc.Transaction.ResultXDR) != "" {
		horizon.Transaction.ResultXDR = rpc.Transaction.ResultXDR
	}
	if rpc.Transaction.ResultMetaXDR != nil && strings.TrimSpace(*rpc.Transaction.ResultMetaXDR) != "" {
		horizon.Transaction.ResultMetaXDR = rpc.Transaction.ResultMetaXDR
		horizon.Transaction.IsSoroban = horizon.Transaction.IsSoroban || rpc.Transaction.IsSoroban
	}
	return horizon
}

func mergeAccountLookup(horizon backendclient.AccountLookupResponse, rpc backendclient.AccountLookupResponse) backendclient.AccountLookupResponse {
	if horizon.Account == nil || rpc.Account == nil {
		return horizon
	}
	if len(horizon.Signers) == 0 && len(rpc.Signers) > 0 {
		horizon.Signers = rpc.Signers
	}
	return horizon
}

func (b *Backend) setMeta(operation string, actual string, degraded bool, reason string) {
	b.lastMeta = app.SourceMetadata{
		Mode:           b.profile.NormalizedBackendMode(),
		Operation:      operation,
		Policy:         "network: horizon history + rpc soroban/state",
		Preferred:      "horizon",
		Actual:         actual,
		Label:          combinedLabel(b.profile, b.horizon != nil),
		Degraded:       degraded,
		DegradedReason: reason,
	}
}

func combinedLabel(profile config.Profile, hasHorizon bool) string {
	rpcLabel := strings.TrimSpace(profile.RPCEndpoint)
	horizonLabel := config.ResolveHorizonURL(profile)
	switch {
	case hasHorizon && horizonLabel != "" && rpcLabel != "":
		return horizonLabel + " + " + rpcLabel
	case hasHorizon && horizonLabel != "":
		return horizonLabel
	default:
		return rpcLabel
	}
}

func actualSource(hasHorizon bool) string {
	if hasHorizon {
		return "horizon+rpc"
	}
	return "rpc"
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
