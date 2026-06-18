package horizonbackend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/stellar/go-stellar-sdk/clients/horizonclient"
	hProtocol "github.com/stellar/go-stellar-sdk/protocols/horizon"
)

const defaultTimeout = 10 * time.Second

// Backend adapts Horizon API calls into the TUI lookupBackend shape.
type Backend struct {
	profile config.Profile
	client  *horizonclient.Client
}

// Option customizes backend construction.
type Option func(*Backend)

// WithHorizonClient injects a custom Horizon client (used in tests).
func WithHorizonClient(client *horizonclient.Client) Option {
	return func(b *Backend) {
		if client != nil {
			b.client = client
		}
	}
}

// New constructs a Horizon backend for the given profile.
func New(profile config.Profile, opts ...Option) (*Backend, error) {
	horizonURL := config.ResolveHorizonURL(profile)
	if strings.TrimSpace(horizonURL) == "" {
		return nil, errors.New("horizon_url is not configured for the current profile")
	}

	backend := &Backend{
		profile: profile,
		client: &horizonclient.Client{
			HorizonURL: horizonURL,
			HTTP: &http.Client{
				Timeout: defaultTimeout,
			},
		},
	}
	for _, opt := range opts {
		opt(backend)
	}
	return backend, nil
}

// Label returns the configured Horizon endpoint for UI source reporting.
func (b *Backend) Label() string {
	if b == nil {
		return ""
	}
	return config.ResolveHorizonURL(b.profile)
}

func (b *Backend) LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	ledgers, err := b.client.Ledgers(horizonclient.LedgerRequest{
		Order: horizonclient.OrderDesc,
		Limit: 1,
	})
	if err != nil {
		return backendclient.LiveFeedSummaryResponse{}, wrapHorizonError(err)
	}
	if len(ledgers.Embedded.Records) == 0 {
		return backendclient.LiveFeedSummaryResponse{}, errors.New("no ledgers returned from horizon")
	}

	latest := ledgers.Embedded.Records[0]
	summary := ledgerSummary(latest)

	txPage, err := b.client.Transactions(horizonclient.TransactionRequest{
		Order: horizonclient.OrderDesc,
		Limit: 10,
	})
	if err != nil {
		return backendclient.LiveFeedSummaryResponse{}, wrapHorizonError(err)
	}

	response := backendclient.LiveFeedSummaryResponse{
		LastIngestedLedger: uint32(latest.Sequence),
		LatestLedger:       &summary,
		RecentTransactions: transactionSummaries(txPage.Embedded.Records),
	}
	return response, nil
}

func (b *Backend) Ledger(ctx context.Context, sequence uint32) (backendclient.LedgerLookupResponse, error) {
	if sequence == 0 {
		return backendclient.LedgerLookupResponse{}, errors.New("ledger sequence is required")
	}

	ledger, err := b.client.LedgerDetail(sequence)
	if err != nil {
		return backendclient.LedgerLookupResponse{}, wrapHorizonError(err)
	}

	txPage, err := b.client.Transactions(horizonclient.TransactionRequest{
		ForLedger: uint(sequence),
		Order:     horizonclient.OrderAsc,
		Limit:     50,
	})
	if err != nil {
		return backendclient.LedgerLookupResponse{}, wrapHorizonError(err)
	}

	summary := ledgerSummary(ledger)
	return backendclient.LedgerLookupResponse{
		Ledger:       &summary,
		Transactions: transactionSummaries(txPage.Embedded.Records),
	}, nil
}

func (b *Backend) Ledgers(ctx context.Context, limit int, before uint32) ([]backendclient.LedgerSummary, error) {
	limit = normalizeLimit(limit)
	request := horizonclient.LedgerRequest{
		Order: horizonclient.OrderDesc,
		Limit: uint(limit),
	}
	if before > 0 {
		request.Cursor = fmt.Sprintf("%d", before)
	}

	page, err := b.client.Ledgers(request)
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	summaries := make([]backendclient.LedgerSummary, 0, len(page.Embedded.Records))
	for _, ledger := range page.Embedded.Records {
		if before > 0 && uint32(ledger.Sequence) >= before {
			continue
		}
		summaries = append(summaries, ledgerSummary(ledger))
	}
	return summaries, nil
}

func (b *Backend) LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]backendclient.TransactionSummary, error) {
	if sequence == 0 {
		return nil, errors.New("ledger sequence is required")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	fetchLimit := limit + offset
	if fetchLimit > 50 {
		fetchLimit = 50
	}

	page, err := b.client.Transactions(horizonclient.TransactionRequest{
		ForLedger: uint(sequence),
		Order:     horizonclient.OrderAsc,
		Limit:     uint(fetchLimit),
	})
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	transactions := transactionSummaries(page.Embedded.Records)
	return slicePage(transactions, limit, offset), nil
}

func (b *Backend) Transaction(ctx context.Context, hash string) (backendclient.TransactionLookupResponse, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return backendclient.TransactionLookupResponse{}, errors.New("transaction hash is required")
	}

	tx, err := b.client.TransactionDetail(hash)
	if err != nil {
		return backendclient.TransactionLookupResponse{}, wrapHorizonError(err)
	}

	opsPage, err := b.client.Operations(horizonclient.OperationRequest{
		ForTransaction: hash,
		Order:          horizonclient.OrderAsc,
		Limit:          200,
	})
	if err != nil {
		return backendclient.TransactionLookupResponse{}, wrapHorizonError(err)
	}

	effectsPage, err := b.client.Effects(horizonclient.EffectRequest{
		ForTransaction: hash,
		Order:          horizonclient.OrderAsc,
		Limit:          200,
	})
	if err != nil {
		return backendclient.TransactionLookupResponse{}, wrapHorizonError(err)
	}

	operations := make([]backendclient.OperationSummary, 0, len(opsPage.Embedded.Records))
	for _, op := range opsPage.Embedded.Records {
		operations = append(operations, operationSummary(op))
	}

	effectRows := make([]backendclient.EffectSummary, 0, len(effectsPage.Embedded.Records))
	for _, effect := range effectsPage.Embedded.Records {
		effectRows = append(effectRows, effectSummary(effect, hash))
	}

	return backendclient.TransactionLookupResponse{
		Transaction: transactionDetail(tx),
		Operations:  operations,
		Effects:     effectRows,
	}, nil
}

func (b *Backend) Account(ctx context.Context, id string) (backendclient.AccountLookupResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return backendclient.AccountLookupResponse{}, errors.New("account id is required")
	}

	account, err := b.client.AccountDetail(horizonclient.AccountRequest{AccountID: id})
	if err != nil {
		return backendclient.AccountLookupResponse{}, wrapHorizonError(err)
	}

	txPage, err := b.client.Transactions(horizonclient.TransactionRequest{
		ForAccount: id,
		Order:      horizonclient.OrderDesc,
		Limit:      10,
	})
	if err != nil {
		return backendclient.AccountLookupResponse{}, wrapHorizonError(err)
	}

	opsPage, err := b.client.Operations(horizonclient.OperationRequest{
		ForAccount: id,
		Order:      horizonclient.OrderDesc,
		Limit:      10,
	})
	if err != nil {
		return backendclient.AccountLookupResponse{}, wrapHorizonError(err)
	}

	recentOps := make([]backendclient.OperationSummary, 0, len(opsPage.Embedded.Records))
	for _, op := range opsPage.Embedded.Records {
		recentOps = append(recentOps, operationSummary(op))
	}

	return backendclient.AccountLookupResponse{
		Account:            accountDetail(account),
		Trustlines:         trustlineSummaries(account),
		Signers:            signerSummaries(account),
		RecentTransactions: transactionSummaries(txPage.Embedded.Records),
		RecentOperations:   recentOps,
	}, nil
}

func (b *Backend) Accounts(ctx context.Context, limit int) ([]backendclient.AccountDetail, error) {
	page, err := b.client.Accounts(horizonclient.AccountsRequest{
		Order: horizonclient.OrderDesc,
		Limit: uint(normalizeLimit(limit)),
	})
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	accounts := make([]backendclient.AccountDetail, 0, len(page.Embedded.Records))
	for _, account := range page.Embedded.Records {
		accounts = append(accounts, *accountDetail(account))
	}
	return accounts, nil
}

func (b *Backend) AccountOperations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("account id is required")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	fetchLimit := limit + offset
	if fetchLimit > 50 {
		fetchLimit = 50
	}

	page, err := b.client.Operations(horizonclient.OperationRequest{
		ForAccount: id,
		Order:      horizonclient.OrderDesc,
		Limit:      uint(fetchLimit),
	})
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	operations := make([]backendclient.OperationSummary, 0, len(page.Embedded.Records))
	for _, op := range page.Embedded.Records {
		operations = append(operations, operationSummary(op))
	}
	return slicePage(operations, limit, offset), nil
}

func (b *Backend) AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("account id is required")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	category = strings.TrimSpace(category)

	switch category {
	case "op", "ops", "operation", "operations":
		ops, err := b.AccountOperations(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		return timelineFromOperationSummaries(ops), nil
	case "tx", "transaction", "transactions", "":
		fetchLimit := limit + offset
		if fetchLimit > 50 {
			fetchLimit = 50
		}
		page, err := b.client.Transactions(horizonclient.TransactionRequest{
			ForAccount: id,
			Order:      horizonclient.OrderDesc,
			Limit:      uint(fetchLimit),
		})
		if err != nil {
			return nil, wrapHorizonError(err)
		}
		txs := slicePage(transactionSummaries(page.Embedded.Records), limit, offset)
		return timelineFromTransactionSummaries(txs), nil
	default:
		return nil, fmt.Errorf("unsupported timeline category %q", category)
	}
}

func (b *Backend) Asset(ctx context.Context, code string, issuer string) (backendclient.AssetLookupResponse, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return backendclient.AssetLookupResponse{}, errors.New("asset code and issuer are required")
	}

	page, err := b.client.Assets(horizonclient.AssetRequest{
		ForAssetCode:   code,
		ForAssetIssuer: issuer,
		Limit:          1,
	})
	if err != nil {
		return backendclient.AssetLookupResponse{}, wrapHorizonError(err)
	}
	if len(page.Embedded.Records) == 0 {
		return backendclient.AssetLookupResponse{}, errors.New("asset not found")
	}

	holders, err := b.AssetHolders(ctx, code, issuer, 10, 0)
	if err != nil {
		return backendclient.AssetLookupResponse{}, err
	}

	return backendclient.AssetLookupResponse{
		Asset:      assetDetail(page.Embedded.Records[0]),
		TopHolders: holders,
	}, nil
}

func (b *Backend) Assets(ctx context.Context, limit int) ([]backendclient.AssetDetail, error) {
	page, err := b.client.Assets(horizonclient.AssetRequest{
		Order: horizonclient.OrderDesc,
		Limit: uint(normalizeLimit(limit)),
	})
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	assets := make([]backendclient.AssetDetail, 0, len(page.Embedded.Records))
	for _, asset := range page.Embedded.Records {
		assets = append(assets, *assetDetail(asset))
	}
	return assets, nil
}

func (b *Backend) AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]backendclient.AssetHolderSummary, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return nil, errors.New("asset code and issuer are required")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	fetchLimit := limit + offset
	if fetchLimit > 50 {
		fetchLimit = 50
	}

	page, err := b.client.Accounts(horizonclient.AccountsRequest{
		Asset: fmt.Sprintf("%s:%s", code, issuer),
		Order: horizonclient.OrderDesc,
		Limit: uint(fetchLimit),
	})
	if err != nil {
		return nil, wrapHorizonError(err)
	}

	holders := make([]backendclient.AssetHolderSummary, 0, len(page.Embedded.Records))
	for _, account := range page.Embedded.Records {
		if holder, ok := assetHolderSummary(account, code, issuer); ok {
			holders = append(holders, holder)
		}
	}
	return slicePage(holders, limit, offset), nil
}

func (b *Backend) AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	code = strings.TrimSpace(code)
	issuer = strings.TrimSpace(issuer)
	if code == "" || issuer == "" {
		return nil, errors.New("asset code and issuer are required")
	}

	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)
	category = strings.TrimSpace(category)

	switch category {
	case "holder", "holders", "":
		holders, err := b.AssetHolders(ctx, code, issuer, limit, offset)
		if err != nil {
			return nil, err
		}
		return timelineFromHolders(holders), nil
	default:
		return nil, fmt.Errorf("unsupported asset timeline category %q", category)
	}
}

func (b *Backend) Contract(ctx context.Context, id string) (backendclient.ContractLookupResponse, error) {
	return backendclient.ContractLookupResponse{}, errors.New("contract lookup is unavailable in horizon-only mode")
}

func (b *Backend) Contracts(ctx context.Context, limit int) ([]backendclient.ContractDetail, error) {
	return nil, errors.New("contract list is unavailable in horizon-only mode")
}

func (b *Backend) ContractEvents(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractEventSummary, error) {
	return nil, errors.New("contract event list is unavailable in horizon-only mode")
}

func (b *Backend) ContractStorage(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractStorageSummary, error) {
	return nil, errors.New("contract storage list is unavailable in horizon-only mode")
}

func (b *Backend) ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	return nil, errors.New("contract invocation list is unavailable in horizon-only mode")
}

func (b *Backend) ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	return nil, errors.New("contract timeline is unavailable in horizon-only mode")
}

func transactionSummaries(txs []hProtocol.Transaction) []backendclient.TransactionSummary {
	summaries := make([]backendclient.TransactionSummary, 0, len(txs))
	for _, tx := range txs {
		summaries = append(summaries, transactionSummary(tx))
	}
	return summaries
}

func timelineFromTransactionSummaries(txs []backendclient.TransactionSummary) []backendclient.TimelineItem {
	items := make([]backendclient.TimelineItem, 0, len(txs))
	for _, tx := range txs {
		items = append(items, backendclient.TimelineItem{
			Kind:        "tx",
			Title:       "Transaction",
			Description: truncate(tx.Hash, 18),
			Command:     "lookup tx " + tx.Hash,
			OccurredAt:  tx.CreatedAt,
		})
	}
	return items
}

func timelineFromHolders(holders []backendclient.AssetHolderSummary) []backendclient.TimelineItem {
	items := make([]backendclient.TimelineItem, 0, len(holders))
	for _, holder := range holders {
		items = append(items, backendclient.TimelineItem{
			Kind:        "holder",
			Title:       "Holder",
			Description: truncate(holder.AccountID, 18),
			Command:     "lookup account " + holder.AccountID,
			OccurredAt:  holder.UpdatedAt,
		})
	}
	return items
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func slicePage[T any](values []T, limit int, offset int) []T {
	if offset >= len(values) {
		return nil
	}
	end := offset + limit
	if end > len(values) {
		end = len(values)
	}
	return append([]T(nil), values[offset:end]...)
}

func wrapHorizonError(err error) error {
	if err == nil {
		return nil
	}
	var horizonErr *horizonclient.Error
	if errors.As(err, &horizonErr) && horizonErr.Problem.Detail != "" {
		return errors.New(strings.TrimSpace(horizonErr.Problem.Detail))
	}
	return err
}
