package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

type hybridLookupBackend struct {
	primary  lookupBackend
	fallback lookupBackend
	lastMeta app.SourceMetadata
}

func newHybridLookupBackend(primary lookupBackend, fallback lookupBackend) *hybridLookupBackend {
	return &hybridLookupBackend{
		primary:  primary,
		fallback: fallback,
		lastMeta: app.SourceMetadata{
			Mode:      "hybrid",
			Preferred: "indexer",
			Actual:    "indexer",
			Policy:    "single-source: indexer -> rpc; no field merge",
			Label:     labelWithMode(primary, "hybrid"),
		},
	}
}

func (b *hybridLookupBackend) Label() string {
	if b == nil {
		return ""
	}
	if strings.TrimSpace(b.lastMeta.Label) != "" {
		return b.lastMeta.Label
	}
	return labelWithMode(b.primary, "hybrid")
}

func (b *hybridLookupBackend) SourceMetadata() app.SourceMetadata {
	if b == nil {
		return app.SourceMetadata{}
	}
	return b.lastMeta
}

func (b *hybridLookupBackend) LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	response, err := b.primary.LiveFeedSummary(ctx)
	if err == nil {
		b.lastMeta = successSourceMetadata("live-feed", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("live-feed", b.primary, err)
		return backendclient.LiveFeedSummaryResponse{}, err
	}
	response, fallbackErr := b.fallback.LiveFeedSummary(ctx)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("live-feed", b.fallback, err, fallbackErr)
		return backendclient.LiveFeedSummaryResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("live-feed", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Search(ctx context.Context, query string, limit int) (backendclient.SearchResponse, error) {
	response, err := b.primary.Search(ctx, query, limit)
	if err == nil {
		b.lastMeta = successSourceMetadata("search", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("search", b.primary, err)
		return backendclient.SearchResponse{}, err
	}
	response, fallbackErr := b.fallback.Search(ctx, query, limit)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("search", b.fallback, err, fallbackErr)
		return backendclient.SearchResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("search", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Ledger(ctx context.Context, sequence uint32) (backendclient.LedgerLookupResponse, error) {
	response, err := b.primary.Ledger(ctx, sequence)
	if err == nil {
		b.lastMeta = successSourceMetadata("ledger", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("ledger", b.primary, err)
		return backendclient.LedgerLookupResponse{}, err
	}
	response, fallbackErr := b.fallback.Ledger(ctx, sequence)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("ledger", b.fallback, err, fallbackErr)
		return backendclient.LedgerLookupResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("ledger", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Ledgers(ctx context.Context, limit int, before uint32) ([]backendclient.LedgerSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer ledger list is not implemented")
		return b.fallbackLedgers(ctx, limit, before, err)
	}
	response, err := primary.Ledgers(ctx, limit, before)
	if err == nil {
		b.lastMeta = successSourceMetadata("ledgers", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("ledgers", b.primary, err)
		return nil, err
	}
	return b.fallbackLedgers(ctx, limit, before, err)
}

func (b *hybridLookupBackend) fallbackLedgers(ctx context.Context, limit int, before uint32, primaryErr error) ([]backendclient.LedgerSummary, error) {
	fallback, ok := b.fallback.(explorerListBackend)
	if !ok {
		b.lastMeta = fallbackFailureMetadata("ledgers", b.fallback, primaryErr, errors.New("rpc ledger list is not implemented"))
		return nil, primaryErr
	}
	response, fallbackErr := fallback.Ledgers(ctx, limit, before)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("ledgers", b.fallback, primaryErr, fallbackErr)
		return nil, joinFallbackError(primaryErr, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("ledgers", b.fallback, primaryErr)
	return response, nil
}

func (b *hybridLookupBackend) fallbackContractEvents(ctx context.Context, id string, limit int, offset int, primaryErr error) ([]backendclient.ContractEventSummary, error) {
	fallback, ok := b.fallback.(explorerListBackend)
	if !ok {
		b.lastMeta = fallbackFailureMetadata("contract-events", b.fallback, primaryErr, errors.New("rpc contract event list is not implemented"))
		return nil, primaryErr
	}
	response, fallbackErr := fallback.ContractEvents(ctx, id, limit, offset)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("contract-events", b.fallback, primaryErr, fallbackErr)
		return nil, joinFallbackError(primaryErr, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("contract-events", b.fallback, primaryErr)
	return response, nil
}

func (b *hybridLookupBackend) fallbackContractStorage(ctx context.Context, id string, limit int, offset int, primaryErr error) ([]backendclient.ContractStorageSummary, error) {
	fallback, ok := b.fallback.(explorerListBackend)
	if !ok {
		b.lastMeta = fallbackFailureMetadata("contract-storage", b.fallback, primaryErr, errors.New("rpc contract storage list is not implemented"))
		return nil, primaryErr
	}
	response, fallbackErr := fallback.ContractStorage(ctx, id, limit, offset)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("contract-storage", b.fallback, primaryErr, fallbackErr)
		return nil, joinFallbackError(primaryErr, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("contract-storage", b.fallback, primaryErr)
	return response, nil
}

func (b *hybridLookupBackend) LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]backendclient.TransactionSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer ledger transaction list is not implemented")
		b.lastMeta = primaryFailureMetadata("ledger-transactions", b.primary, err)
		return nil, err
	}
	response, err := primary.LedgerTransactions(ctx, sequence, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("ledger-transactions", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("ledger-transactions", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) Transaction(ctx context.Context, hash string) (backendclient.TransactionLookupResponse, error) {
	response, err := b.primary.Transaction(ctx, hash)
	if err == nil {
		b.lastMeta = successSourceMetadata("transaction", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("transaction", b.primary, err)
		return backendclient.TransactionLookupResponse{}, err
	}
	response, fallbackErr := b.fallback.Transaction(ctx, hash)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("transaction", b.fallback, err, fallbackErr)
		return backendclient.TransactionLookupResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("transaction", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Accounts(ctx context.Context, limit int) ([]backendclient.AccountDetail, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer account list is not implemented")
		b.lastMeta = primaryFailureMetadata("accounts", b.primary, err)
		return nil, err
	}
	response, err := primary.Accounts(ctx, limit)
	if err == nil {
		b.lastMeta = successSourceMetadata("accounts", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("accounts", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) AccountOperations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer account operation list is not implemented")
		b.lastMeta = primaryFailureMetadata("account-operations", b.primary, err)
		return nil, err
	}
	response, err := primary.AccountOperations(ctx, id, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("account-operations", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("account-operations", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer account timeline is not implemented")
		b.lastMeta = primaryFailureMetadata("account-timeline", b.primary, err)
		return nil, err
	}
	response, err := primary.AccountTimeline(ctx, id, limit, offset, category)
	if err == nil {
		b.lastMeta = successSourceMetadata("account-timeline", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("account-timeline", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) Account(ctx context.Context, id string) (backendclient.AccountLookupResponse, error) {
	response, err := b.primary.Account(ctx, id)
	if err == nil {
		b.lastMeta = successSourceMetadata("account", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("account", b.primary, err)
		return backendclient.AccountLookupResponse{}, err
	}
	response, fallbackErr := b.fallback.Account(ctx, id)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("account", b.fallback, err, fallbackErr)
		return backendclient.AccountLookupResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("account", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Assets(ctx context.Context, limit int) ([]backendclient.AssetDetail, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer asset list is not implemented")
		b.lastMeta = primaryFailureMetadata("assets", b.primary, err)
		return nil, err
	}
	response, err := primary.Assets(ctx, limit)
	if err == nil {
		b.lastMeta = successSourceMetadata("assets", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("assets", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]backendclient.AssetHolderSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer asset holder list is not implemented")
		b.lastMeta = primaryFailureMetadata("asset-holders", b.primary, err)
		return nil, err
	}
	response, err := primary.AssetHolders(ctx, code, issuer, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("asset-holders", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("asset-holders", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer asset timeline is not implemented")
		b.lastMeta = primaryFailureMetadata("asset-timeline", b.primary, err)
		return nil, err
	}
	response, err := primary.AssetTimeline(ctx, code, issuer, limit, offset, category)
	if err == nil {
		b.lastMeta = successSourceMetadata("asset-timeline", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("asset-timeline", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) Asset(ctx context.Context, code string, issuer string) (backendclient.AssetLookupResponse, error) {
	response, err := b.primary.Asset(ctx, code, issuer)
	if err == nil {
		b.lastMeta = successSourceMetadata("asset", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("asset", b.primary, err)
		return backendclient.AssetLookupResponse{}, err
	}
	response, fallbackErr := b.fallback.Asset(ctx, code, issuer)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("asset", b.fallback, err, fallbackErr)
		return backendclient.AssetLookupResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("asset", b.fallback, err)
	return response, nil
}

func (b *hybridLookupBackend) Contracts(ctx context.Context, limit int) ([]backendclient.ContractDetail, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer contract list is not implemented")
		b.lastMeta = primaryFailureMetadata("contracts", b.primary, err)
		return nil, err
	}
	response, err := primary.Contracts(ctx, limit)
	if err == nil {
		b.lastMeta = successSourceMetadata("contracts", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("contracts", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) ContractStorage(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractStorageSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer contract storage list is not implemented")
		return b.fallbackContractStorage(ctx, id, limit, offset, err)
	}
	response, err := primary.ContractStorage(ctx, id, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("contract-storage", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("contract-storage", b.primary, err)
		return nil, err
	}
	return b.fallbackContractStorage(ctx, id, limit, offset, err)
}

func (b *hybridLookupBackend) ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer contract invocation list is not implemented")
		b.lastMeta = primaryFailureMetadata("contract-invocations", b.primary, err)
		return nil, err
	}
	response, err := primary.ContractInvocations(ctx, id, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("contract-invocations", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("contract-invocations", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) ContractEvents(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractEventSummary, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer contract event list is not implemented")
		return b.fallbackContractEvents(ctx, id, limit, offset, err)
	}
	response, err := primary.ContractEvents(ctx, id, limit, offset)
	if err == nil {
		b.lastMeta = successSourceMetadata("contract-events", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("contract-events", b.primary, err)
		return nil, err
	}
	return b.fallbackContractEvents(ctx, id, limit, offset, err)
}

func (b *hybridLookupBackend) ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	primary, ok := b.primary.(explorerListBackend)
	if !ok {
		err := errors.New("indexer contract timeline is not implemented")
		b.lastMeta = primaryFailureMetadata("contract-timeline", b.primary, err)
		return nil, err
	}
	response, err := primary.ContractTimeline(ctx, id, limit, offset, category)
	if err == nil {
		b.lastMeta = successSourceMetadata("contract-timeline", b.primary)
		return response, nil
	}
	b.lastMeta = primaryFailureMetadata("contract-timeline", b.primary, err)
	return nil, err
}

func (b *hybridLookupBackend) Contract(ctx context.Context, id string) (backendclient.ContractLookupResponse, error) {
	response, err := b.primary.Contract(ctx, id)
	if err == nil {
		b.lastMeta = successSourceMetadata("contract", b.primary)
		return response, nil
	}
	if !shouldFallbackToRPC(err) {
		b.lastMeta = primaryFailureMetadata("contract", b.primary, err)
		return backendclient.ContractLookupResponse{}, err
	}
	response, fallbackErr := b.fallback.Contract(ctx, id)
	if fallbackErr != nil {
		b.lastMeta = fallbackFailureMetadata("contract", b.fallback, err, fallbackErr)
		return backendclient.ContractLookupResponse{}, joinFallbackError(err, fallbackErr)
	}
	b.lastMeta = fallbackSuccessMetadata("contract", b.fallback, err)
	return response, nil
}

func shouldFallbackToRPC(err error) bool {
	if err == nil {
		return false
	}

	var httpErr backendclient.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case 404, 405, 501, 502, 503, 504:
			return true
		default:
			return false
		}
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "connection refused") ||
		strings.Contains(message, "no such host") ||
		strings.Contains(message, "unsupported") ||
		strings.Contains(message, "not implemented")
}

func joinFallbackError(primaryErr error, fallbackErr error) error {
	return fmt.Errorf("indexer failed: %v; rpc fallback failed: %w", primaryErr, fallbackErr)
}

func labelWithMode(backend lookupBackend, mode string) string {
	if backend == nil {
		return ""
	}
	label := strings.TrimSpace(backend.Label())
	if label == "" {
		return mode
	}
	return fmt.Sprintf("%s (%s)", label, mode)
}

func successSourceMetadata(operation string, backend lookupBackend) app.SourceMetadata {
	return app.SourceMetadata{
		Mode:      "hybrid",
		Operation: operation,
		Policy:    "single-source: indexer -> rpc; no field merge",
		Preferred: "indexer",
		Actual:    "indexer",
		Label:     labelWithMode(backend, "hybrid"),
	}
}

func primaryFailureMetadata(operation string, backend lookupBackend, err error) app.SourceMetadata {
	return app.SourceMetadata{
		Mode:           "hybrid",
		Operation:      operation,
		Policy:         "single-source: indexer -> rpc; no field merge",
		Preferred:      "indexer",
		Actual:         "indexer",
		Label:          labelWithMode(backend, "hybrid"),
		Degraded:       true,
		DegradedReason: fmt.Sprintf("indexer error: %v", err),
	}
}

func fallbackSuccessMetadata(operation string, backend lookupBackend, primaryErr error) app.SourceMetadata {
	return app.SourceMetadata{
		Mode:           "hybrid",
		Operation:      operation,
		Policy:         "single-source: indexer -> rpc; no field merge",
		Preferred:      "indexer",
		Actual:         "rpc",
		Label:          labelWithMode(backend, "rpc fallback"),
		FallbackUsed:   true,
		Degraded:       true,
		DegradedReason: fmt.Sprintf("indexer unavailable, used rpc fallback: %v", primaryErr),
	}
}

func fallbackFailureMetadata(operation string, backend lookupBackend, primaryErr error, fallbackErr error) app.SourceMetadata {
	return app.SourceMetadata{
		Mode:           "hybrid",
		Operation:      operation,
		Policy:         "single-source: indexer -> rpc; no field merge",
		Preferred:      "indexer",
		Actual:         "rpc",
		Label:          labelWithMode(backend, "rpc fallback"),
		FallbackUsed:   true,
		Degraded:       true,
		DegradedReason: fmt.Sprintf("indexer failed: %v; rpc fallback failed: %v", primaryErr, fallbackErr),
	}
}
