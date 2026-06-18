package rpcbackend

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/soroban"
	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// Backend adapts direct Stellar RPC calls into the TUI lookupBackend shape.
type Backend struct {
	profile  config.Profile
	client   *rpcclient.Client
	enricher *soroban.Enricher
}

// New constructs a direct RPC backend for the given profile.
func New(profile config.Profile) (*Backend, error) {
	if strings.TrimSpace(profile.RPCEndpoint) == "" {
		return nil, errors.New("rpc_endpoint is not configured for the current profile")
	}

	client := rpcclient.New(profile.RPCEndpoint)
	return &Backend{
		profile:  profile,
		client:   client,
		enricher: soroban.NewEnricher(client),
	}, nil
}

// EnrichTransaction applies in-client Soroban decoding to a transaction lookup payload.
func (b *Backend) EnrichTransaction(ctx context.Context, response *backendclient.TransactionLookupResponse) error {
	if b == nil || b.enricher == nil {
		return nil
	}
	return b.enricher.EnrichTransaction(ctx, response)
}

// EnrichContract applies in-client Soroban decoding to a contract lookup payload.
func (b *Backend) EnrichContract(ctx context.Context, response *backendclient.ContractLookupResponse) error {
	if b == nil || b.enricher == nil {
		return nil
	}
	return b.enricher.EnrichContract(ctx, response)
}

// Label returns the configured RPC endpoint for UI source reporting.
func (b *Backend) Label() string {
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.profile.RPCEndpoint)
}

func (b *Backend) Search(ctx context.Context, query string, limit int) (backendclient.SearchResponse, error) {
	return backendclient.SearchResponse{Results: []backendclient.SearchResult{}}, nil
}

func (b *Backend) LiveFeedSummary(ctx context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	latest, err := b.client.GetLatestLedger(ctx)
	if err != nil {
		return backendclient.LiveFeedSummaryResponse{}, err
	}

	txs, err := b.client.GetTransactions(ctx, rpcclient.GetTransactionsParams{
		StartLedger: latest.Sequence,
		Pagination: &rpcclient.Pagination{
			Limit: 10,
		},
	})
	if err != nil {
		return backendclient.LiveFeedSummaryResponse{}, err
	}

	response := backendclient.LiveFeedSummaryResponse{
		LastIngestedLedger: latest.Sequence,
		LatestLedger: &backendclient.LedgerSummary{
			Sequence: latest.Sequence,
			Hash:     latest.ID,
			ClosedAt: parseRPCTime(latest.CloseTime),
		},
		RecentTransactions: make([]backendclient.TransactionSummary, 0, len(txs.Transactions)),
	}

	for _, tx := range txs.Transactions {
		response.RecentTransactions = append(response.RecentTransactions, backendclient.TransactionSummary{
			Hash:             tx.TxHash,
			LedgerSequence:   tx.Ledger,
			ApplicationOrder: tx.ApplicationOrder,
			Account:          "rpc-source-unavailable",
			Status:           rpcStatusCode(tx.Status),
			IsSoroban:        false,
			CreatedAt:        time.Unix(tx.CreatedAt, 0).UTC(),
		})
	}

	return response, nil
}

func (b *Backend) Ledger(ctx context.Context, sequence uint32) (backendclient.LedgerLookupResponse, error) {
	if sequence == 0 {
		return backendclient.LedgerLookupResponse{}, errors.New("ledger sequence is required")
	}

	ledgers, err := b.client.GetLedgers(ctx, rpcclient.GetLedgersParams{
		StartLedger: sequence,
		Pagination:  &rpcclient.Pagination{Limit: 1},
	})
	if err != nil {
		return backendclient.LedgerLookupResponse{}, err
	}
	if len(ledgers.Ledgers) == 0 || ledgers.Ledgers[0].Sequence != sequence {
		return backendclient.LedgerLookupResponse{}, errors.New("ledger not found in RPC retention window")
	}

	ledger := ledgers.Ledgers[0]
	return backendclient.LedgerLookupResponse{
		Ledger: &backendclient.LedgerSummary{
			Sequence:          ledger.Sequence,
			Hash:              ledger.ID,
			ClosedAt:          parseRPCTime(ledger.CloseTime),
			TransactionCount:  ledger.TransactionCount,
			OperationCount:    ledger.OperationCount,
			SuccessfulTxCount: ledger.SuccessfulTransactionCount,
			FailedTxCount:     ledger.FailedTransactionCount,
		},
	}, nil
}

func (b *Backend) Ledgers(ctx context.Context, limit int, before uint32) ([]backendclient.LedgerSummary, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	latest, err := b.client.GetLatestLedger(ctx)
	if err != nil {
		return nil, err
	}

	target := latest.Sequence
	if before > 0 {
		if before <= 1 {
			return []backendclient.LedgerSummary{}, nil
		}
		target = before - 1
	}

	start := target
	if start > uint32(limit) {
		start = target - uint32(limit) + 1
	}

	ledgers, err := b.client.GetLedgers(ctx, rpcclient.GetLedgersParams{
		StartLedger: start,
		Pagination:  &rpcclient.Pagination{Limit: limit},
	})
	if err != nil {
		return nil, err
	}

	summaries := make([]backendclient.LedgerSummary, 0, len(ledgers.Ledgers))
	for index := len(ledgers.Ledgers) - 1; index >= 0; index-- {
		ledger := ledgers.Ledgers[index]
		summaries = append(summaries, backendclient.LedgerSummary{
			Sequence:          ledger.Sequence,
			Hash:              ledger.ID,
			ClosedAt:          parseRPCTime(ledger.CloseTime),
			TransactionCount:  ledger.TransactionCount,
			OperationCount:    ledger.OperationCount,
			SuccessfulTxCount: ledger.SuccessfulTransactionCount,
			FailedTxCount:     ledger.FailedTransactionCount,
		})
	}
	return summaries, nil
}

func (b *Backend) LedgerTransactions(ctx context.Context, sequence uint32, limit int, offset int) ([]backendclient.TransactionSummary, error) {
	response, err := b.Ledger(ctx, sequence)
	if err != nil {
		return nil, err
	}
	if offset <= 0 {
		return response.Transactions, nil
	}
	if offset >= len(response.Transactions) {
		return []backendclient.TransactionSummary{}, nil
	}
	return response.Transactions[offset:], nil
}

func (b *Backend) Transaction(ctx context.Context, hash string) (backendclient.TransactionLookupResponse, error) {
	result, err := b.client.GetTransaction(ctx, strings.TrimSpace(hash))
	if err != nil {
		return backendclient.TransactionLookupResponse{}, err
	}

	if strings.EqualFold(result.Status, "NOT_FOUND") {
		return backendclient.TransactionLookupResponse{}, errors.New("transaction not found in RPC retention window")
	}

	response := backendclient.TransactionLookupResponse{
		Transaction: &backendclient.TransactionDetail{
			Hash:             result.TxHash,
			LedgerSequence:   result.Ledger,
			ApplicationOrder: result.ApplicationOrder,
			Account:          "rpc-source-unavailable",
			AccountSequence:  0,
			FeeCharged:       0,
			MaxFee:           0,
			OperationCount:   0,
			MemoType:         0,
			Status:           rpcStatusCode(result.Status),
			IsSoroban:        len(result.DiagnosticEventsXDR) > 0,
			EnvelopeXDR:      result.EnvelopeXDR,
			ResultXDR:        result.ResultXDR,
			ResultMetaXDR:    optionalString(result.ResultMetaXDR),
			CreatedAt:        time.Unix(result.CreatedAt, 0).UTC(),
		},
	}
	_ = b.enricher.EnrichTransaction(ctx, &response)
	return response, nil
}

func (b *Backend) Account(ctx context.Context, id string) (backendclient.AccountLookupResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return backendclient.AccountLookupResponse{}, errors.New("account id is required")
	}

	accountID, err := xdr.AddressToAccountId(id)
	if err != nil {
		return backendclient.AccountLookupResponse{}, fmt.Errorf("parse account id: %w", err)
	}

	key, err := accountID.LedgerKey()
	if err != nil {
		return backendclient.AccountLookupResponse{}, fmt.Errorf("build ledger key: %w", err)
	}

	keyXDR, err := key.MarshalBinaryBase64()
	if err != nil {
		return backendclient.AccountLookupResponse{}, fmt.Errorf("marshal ledger key: %w", err)
	}

	result, err := b.client.GetLedgerEntries(ctx, []string{keyXDR})
	if err != nil {
		return backendclient.AccountLookupResponse{}, err
	}
	if len(result.Entries) == 0 {
		return backendclient.AccountLookupResponse{}, errors.New("account not found")
	}
	if strings.TrimSpace(result.Entries[0].DataXDR) == "" {
		return backendclient.AccountLookupResponse{}, errors.New("account data unavailable")
	}

	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(result.Entries[0].DataXDR, &data); err != nil {
		return backendclient.AccountLookupResponse{}, fmt.Errorf("decode account entry: %w", err)
	}

	accountEntry, ok := data.GetAccount()
	if !ok {
		return backendclient.AccountLookupResponse{}, errors.New("ledger entry is not an account")
	}

	signers := make([]backendclient.AccountSignerSummary, 0, len(accountEntry.Signers)+1)
	signerSponsors := accountEntry.SponsorPerSigner()
	if accountEntry.MasterKeyWeight() > 0 {
		signers = append(signers, backendclient.AccountSignerSummary{
			SignerKey:          accountEntry.AccountId.Address(),
			Weight:             int32(accountEntry.MasterKeyWeight()),
			Type:               "master",
			LastModifiedLedger: int64(result.Entries[0].LastModifiedLedger),
		})
	}
	for _, signer := range accountEntry.Signers {
		var sponsor *string
		if accountSponsor, ok := signerSponsors[signer.Key.Address()]; ok {
			value := accountSponsor.Address()
			sponsor = &value
		}
		signers = append(signers, backendclient.AccountSignerSummary{
			SignerKey:          signer.Key.Address(),
			Weight:             int32(signer.Weight),
			Type:               signerTypeLabel(signer.Key),
			Sponsor:            sponsor,
			LastModifiedLedger: int64(result.Entries[0].LastModifiedLedger),
		})
	}

	var homeDomain *string
	if value := trimNullString(string(accountEntry.HomeDomain)); value != "" {
		homeDomain = &value
	}

	var inflationDest *string
	if accountEntry.InflationDest != nil {
		value := accountEntry.InflationDest.Address()
		inflationDest = &value
	}

	thresholds := fmt.Sprintf(`{"master":%d,"low":%d,"med":%d,"high":%d}`,
		accountEntry.MasterKeyWeight(),
		accountEntry.ThresholdLow(),
		accountEntry.ThresholdMedium(),
		accountEntry.ThresholdHigh(),
	)
	liabilities := accountEntry.Liabilities()
	seqLedger := optionalInt64(int64(accountEntry.SeqLedger()))
	seqTime := optionalUnixTime(int64(accountEntry.SeqTime()))

	return backendclient.AccountLookupResponse{
		Account: &backendclient.AccountDetail{
			ID:                 accountEntry.AccountId.Address(),
			Sequence:           int64(accountEntry.SeqNum),
			SequenceLedger:     seqLedger,
			SequenceTime:       seqTime,
			Balance:            formatStroops(int64(accountEntry.Balance)),
			BuyingLiabilities:  formatStroops(int64(liabilities.Buying)),
			SellingLiabilities: formatStroops(int64(liabilities.Selling)),
			NumSubentries:      int32(accountEntry.NumSubEntries),
			HomeDomain:         homeDomain,
			Flags:              int32(accountEntry.Flags),
			InflationDest:      inflationDest,
			Thresholds:         &thresholds,
			LastModifiedLedger: int64(result.Entries[0].LastModifiedLedger),
			NumSponsored:       int32(accountEntry.NumSponsored()),
			NumSponsoring:      int32(accountEntry.NumSponsoring()),
		},
		Trustlines: nil,
		Signers:    signers,
	}, nil
}

func (b *Backend) Accounts(ctx context.Context, limit int) ([]backendclient.AccountDetail, error) {
	return nil, errors.New("account list is unavailable in direct RPC mode")
}

func (b *Backend) AccountOperations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	return nil, errors.New("account operation list is unavailable in direct RPC mode")
}

func (b *Backend) AccountTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	return nil, errors.New("account timeline is unavailable in direct RPC mode")
}

func (b *Backend) Asset(ctx context.Context, code string, issuer string) (backendclient.AssetLookupResponse, error) {
	return backendclient.AssetLookupResponse{}, errors.New("asset lookup is unavailable in direct RPC mode")
}

func (b *Backend) Assets(ctx context.Context, limit int) ([]backendclient.AssetDetail, error) {
	return nil, errors.New("asset list is unavailable in direct RPC mode")
}

func (b *Backend) AssetHolders(ctx context.Context, code string, issuer string, limit int, offset int) ([]backendclient.AssetHolderSummary, error) {
	return nil, errors.New("asset holder list is unavailable in direct RPC mode")
}

func (b *Backend) AssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	return nil, errors.New("asset timeline is unavailable in direct RPC mode")
}

func (b *Backend) Contract(ctx context.Context, id string) (backendclient.ContractLookupResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return backendclient.ContractLookupResponse{}, errors.New("contract id is required")
	}

	contractAddress, err := contractAddressFromString(id)
	if err != nil {
		return backendclient.ContractLookupResponse{}, fmt.Errorf("parse contract id: %w", err)
	}

	var key xdr.LedgerKey
	if err := key.SetContractData(
		contractAddress,
		xdr.ScVal{Type: xdr.ScValTypeScvLedgerKeyContractInstance},
		xdr.ContractDataDurabilityPersistent,
	); err != nil {
		return backendclient.ContractLookupResponse{}, fmt.Errorf("build contract ledger key: %w", err)
	}

	keyXDR, err := key.MarshalBinaryBase64()
	if err != nil {
		return backendclient.ContractLookupResponse{}, fmt.Errorf("marshal ledger key: %w", err)
	}

	result, err := b.client.GetLedgerEntries(ctx, []string{keyXDR})
	if err != nil {
		return backendclient.ContractLookupResponse{}, err
	}
	if len(result.Entries) == 0 {
		return backendclient.ContractLookupResponse{}, errors.New("contract not found")
	}
	if strings.TrimSpace(result.Entries[0].DataXDR) == "" {
		return backendclient.ContractLookupResponse{}, errors.New("contract data unavailable")
	}

	var data xdr.LedgerEntryData
	if err := xdr.SafeUnmarshalBase64(result.Entries[0].DataXDR, &data); err != nil {
		return backendclient.ContractLookupResponse{}, fmt.Errorf("decode contract entry: %w", err)
	}

	contractEntry, ok := data.GetContractData()
	if !ok {
		return backendclient.ContractLookupResponse{}, errors.New("ledger entry is not contract data")
	}
	if contractEntry.Key.Type != xdr.ScValTypeScvLedgerKeyContractInstance {
		return backendclient.ContractLookupResponse{}, errors.New("ledger entry is not the contract instance")
	}

	instance, ok := contractEntry.Val.GetInstance()
	if !ok {
		return backendclient.ContractLookupResponse{}, errors.New("contract instance payload unavailable")
	}

	contractType := int16(2)
	var wasmHash *string
	var tokenName *string
	var tokenSymbol *string
	var tokenDecimals *int32
	var label *string
	isSep41Token := false
	storageEntryCount := int32(0)
	if instance.Storage != nil {
		storageEntryCount = int32(len(*instance.Storage))
		tokenName, tokenSymbol, tokenDecimals = extractContractMetadata(instance.Storage)
		if tokenName != nil || tokenSymbol != nil || tokenDecimals != nil {
			isSep41Token = true
		}
	}
	switch instance.Executable.Type {
	case xdr.ContractExecutableTypeContractExecutableWasm:
		contractType = 0
		if hash, ok := instance.Executable.GetWasmHash(); ok {
			value := fmt.Sprintf("%x", hash[:])
			wasmHash = &value
		}
	case xdr.ContractExecutableTypeContractExecutableStellarAsset:
		contractType = 1
		isSep41Token = true
		value := "Stellar Asset Contract"
		label = &value
	}
	if label == nil {
		if tokenSymbol != nil && strings.TrimSpace(*tokenSymbol) != "" {
			value := strings.TrimSpace(*tokenSymbol)
			label = &value
		} else if tokenName != nil && strings.TrimSpace(*tokenName) != "" {
			value := strings.TrimSpace(*tokenName)
			label = &value
		}
	}

	response := backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{
			ContractID:         id,
			WasmHash:           wasmHash,
			CreatedLedger:      0,
			CreatedAt:          time.Time{},
			LastModifiedLedger: int64(result.Entries[0].LastModifiedLedger),
			ContractType:       contractType,
			IsSep41Token:       isSep41Token,
			IsSep50NFT:         false,
			TokenName:          tokenName,
			TokenSymbol:        tokenSymbol,
			TokenDecimals:      tokenDecimals,
			StorageEntryCount:  storageEntryCount,
			EventCount:         0,
			InvocationCount:    0,
			Label:              label,
			UpdatedAt:          time.Time{},
		},
	}
	_ = b.enricher.EnrichContract(ctx, &response)
	return response, nil
}

func (b *Backend) Contracts(ctx context.Context, limit int) ([]backendclient.ContractDetail, error) {
	return nil, errors.New("contract list is unavailable in direct RPC mode")
}

func (b *Backend) ContractEvents(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractEventSummary, error) {
	if b.enricher == nil {
		return nil, errors.New("contract event list is unavailable in direct RPC mode")
	}
	return b.enricher.ContractEvents(ctx, strings.TrimSpace(id), limit, offset)
}

func (b *Backend) ContractStorage(ctx context.Context, id string, limit int, offset int) ([]backendclient.ContractStorageSummary, error) {
	if b.enricher == nil {
		return nil, errors.New("contract storage list is unavailable in direct RPC mode")
	}
	return b.enricher.ContractStorage(ctx, strings.TrimSpace(id), limit, offset)
}

func (b *Backend) ContractInvocations(ctx context.Context, id string, limit int, offset int) ([]backendclient.OperationSummary, error) {
	return nil, errors.New("contract invocation list is unavailable in direct RPC mode")
}

func (b *Backend) ContractTimeline(ctx context.Context, id string, limit int, offset int, category string) ([]backendclient.TimelineItem, error) {
	return nil, errors.New("contract timeline is unavailable in direct RPC mode")
}

func contractAddressFromString(value string) (xdr.ScAddress, error) {
	raw, err := strkey.Decode(strkey.VersionByteContract, value)
	if err != nil {
		return xdr.ScAddress{}, err
	}
	if len(raw) != 32 {
		return xdr.ScAddress{}, fmt.Errorf("unexpected contract id length: %d", len(raw))
	}

	var contractID xdr.ContractId
	copy(contractID[:], raw)
	return xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, contractID)
}

func parseRPCTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func rpcStatusCode(status string) int16 {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SUCCESS":
		return 1
	case "FAILED":
		return 0
	default:
		return -1
	}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func optionalUnixTime(value int64) *time.Time {
	if value == 0 {
		return nil
	}
	ts := time.Unix(value, 0).UTC()
	return &ts
}

func extractContractMetadata(storage *xdr.ScMap) (*string, *string, *int32) {
	if storage == nil {
		return nil, nil, nil
	}

	var tokenName *string
	var tokenSymbol *string
	var tokenDecimals *int32
	for _, entry := range *storage {
		keyName := normalizeContractStorageKey(entry.Key)
		switch keyName {
		case "name":
			if value := scValString(entry.Val); value != nil {
				tokenName = value
			}
		case "symbol":
			if value := scValString(entry.Val); value != nil {
				tokenSymbol = value
			}
		case "decimal", "decimals":
			if value := scValInt32(entry.Val); value != nil {
				tokenDecimals = value
			}
		}
	}

	return tokenName, tokenSymbol, tokenDecimals
}

func normalizeContractStorageKey(value xdr.ScVal) string {
	if str, ok := value.GetStr(); ok {
		return strings.ToLower(strings.TrimSpace(string(str)))
	}
	if sym, ok := value.GetSym(); ok {
		return strings.ToLower(strings.TrimSpace(string(sym)))
	}
	return ""
}

func scValString(value xdr.ScVal) *string {
	if str, ok := value.GetStr(); ok {
		result := strings.TrimSpace(string(str))
		if result != "" {
			return &result
		}
	}
	if sym, ok := value.GetSym(); ok {
		result := strings.TrimSpace(string(sym))
		if result != "" {
			return &result
		}
	}
	return nil
}

func scValInt32(value xdr.ScVal) *int32 {
	if v, ok := value.GetU32(); ok {
		result := int32(v)
		return &result
	}
	if v, ok := value.GetI32(); ok {
		result := int32(v)
		return &result
	}
	return nil
}

func formatStroops(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	whole := value / 10000000
	fraction := value % 10000000
	return fmt.Sprintf("%s%d.%07d", sign, whole, fraction)
}

func trimNullString(value string) string {
	return strings.TrimRight(value, "\x00 ")
}

func signerTypeLabel(key xdr.SignerKey) string {
	switch key.Type {
	case xdr.SignerKeyTypeSignerKeyTypeEd25519:
		return "ed25519"
	case xdr.SignerKeyTypeSignerKeyTypeHashX:
		return "hash_x"
	case xdr.SignerKeyTypeSignerKeyTypePreAuthTx:
		return "pre_auth_tx"
	case xdr.SignerKeyTypeSignerKeyTypeEd25519SignedPayload:
		return "signed_payload"
	default:
		return "signer"
	}
}
