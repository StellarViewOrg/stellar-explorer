package horizonbackend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	hProtocol "github.com/stellar/go-stellar-sdk/protocols/horizon"
	"github.com/stellar/go-stellar-sdk/protocols/horizon/base"
	"github.com/stellar/go-stellar-sdk/protocols/horizon/effects"
	hoperations "github.com/stellar/go-stellar-sdk/protocols/horizon/operations"
)

func ledgerSummary(ledger hProtocol.Ledger) backendclient.LedgerSummary {
	failed := int32(0)
	if ledger.FailedTransactionCount != nil {
		failed = *ledger.FailedTransactionCount
	}
	successful := ledger.SuccessfulTransactionCount
	return backendclient.LedgerSummary{
		Sequence:          uint32(ledger.Sequence),
		Hash:              ledger.Hash,
		ClosedAt:          ledger.ClosedAt,
		TransactionCount:  successful + failed,
		OperationCount:    int32(ledger.OperationCount),
		SuccessfulTxCount: successful,
		FailedTxCount:     failed,
	}
}

func transactionSummary(tx hProtocol.Transaction) backendclient.TransactionSummary {
	return backendclient.TransactionSummary{
		Hash:             tx.Hash,
		LedgerSequence:   uint32(tx.Ledger),
		ApplicationOrder: 0,
		Account:          tx.Account,
		OperationCount:   tx.OperationCount,
		Status:           transactionStatus(tx.Successful),
		IsSoroban:        strings.TrimSpace(tx.ResultMetaXdr) != "",
		CreatedAt:        tx.LedgerCloseTime,
	}
}

func transactionDetail(tx hProtocol.Transaction) *backendclient.TransactionDetail {
	memoType, memoText, memoHash := memoFields(tx.MemoType, tx.Memo, []byte(tx.MemoBytes))
	var accountMuxed *string
	if strings.TrimSpace(tx.AccountMuxed) != "" {
		value := strings.TrimSpace(tx.AccountMuxed)
		accountMuxed = &value
	}
	var accountMuxedID *int64
	if tx.AccountMuxedID > 0 {
		value := int64(tx.AccountMuxedID)
		accountMuxedID = &value
	}
	var resultMeta *string
	if strings.TrimSpace(tx.ResultMetaXdr) != "" {
		value := strings.TrimSpace(tx.ResultMetaXdr)
		resultMeta = &value
	}
	var feeMeta *string
	if strings.TrimSpace(tx.FeeMetaXdr) != "" {
		value := strings.TrimSpace(tx.FeeMetaXdr)
		feeMeta = &value
	}

	return &backendclient.TransactionDetail{
		Hash:             tx.Hash,
		LedgerSequence:   uint32(tx.Ledger),
		ApplicationOrder: 0,
		Account:          tx.Account,
		AccountMuxed:     accountMuxed,
		AccountMuxedID:   accountMuxedID,
		AccountSequence:  tx.AccountSequence,
		FeeCharged:       tx.FeeCharged,
		MaxFee:           tx.MaxFee,
		OperationCount:   tx.OperationCount,
		MemoType:         memoType,
		MemoText:         memoText,
		MemoHash:         memoHash,
		Status:           transactionStatus(tx.Successful),
		IsSoroban:        strings.TrimSpace(tx.ResultMetaXdr) != "",
		EnvelopeXDR:      tx.EnvelopeXdr,
		ResultXDR:        tx.ResultXdr,
		ResultMetaXDR:    resultMeta,
		FeeMetaXDR:       feeMeta,
		CreatedAt:        tx.LedgerCloseTime,
	}
}

func operationSummary(op hoperations.Operation) backendclient.OperationSummary {
	base := op.GetBase()
	summary := backendclient.OperationSummary{
		TransactionHash:  base.TransactionHash,
		ApplicationOrder: operationIndex(base.ID),
		Type:             int16(base.TypeI),
		TypeName:         strings.TrimSpace(base.GetType()),
		Details:          operationDetailsJSON(base),
		CreatedAt:        base.LedgerCloseTime,
	}
	if strings.TrimSpace(base.SourceAccount) != "" {
		value := strings.TrimSpace(base.SourceAccount)
		summary.SourceAccount = &value
	}
	if strings.TrimSpace(base.SourceAccountMuxed) != "" {
		value := strings.TrimSpace(base.SourceAccountMuxed)
		summary.SourceAccountMuxed = &value
	}
	if base.SourceAccountMuxedID > 0 {
		value := int64(base.SourceAccountMuxedID)
		summary.SourceMuxedID = &value
	}
	populateOperationFields(&summary, op)
	return summary
}

func effectSummary(effect effects.Effect, transactionHash string) backendclient.EffectSummary {
	account := strings.TrimSpace(effect.GetAccount())
	return backendclient.EffectSummary{
		TransactionHash: transactionHash,
		Type:            effectTypeI(effect),
		TypeName:        strings.TrimSpace(effect.GetType()),
		Account:         account,
		Details:         effectDetailsJSON(effect),
		CreatedAt:       effectCreatedAt(effect),
	}
}

func accountDetail(account hProtocol.Account) *backendclient.AccountDetail {
	nativeBalance, buying, selling := nativeBalanceFields(account)
	var homeDomain *string
	if strings.TrimSpace(account.HomeDomain) != "" {
		value := strings.TrimSpace(account.HomeDomain)
		homeDomain = &value
	}
	var sponsor *string
	if strings.TrimSpace(account.Sponsor) != "" {
		value := strings.TrimSpace(account.Sponsor)
		sponsor = &value
	}
	var inflationDest *string
	if strings.TrimSpace(account.InflationDestination) != "" {
		value := strings.TrimSpace(account.InflationDestination)
		inflationDest = &value
	}
	thresholds := fmt.Sprintf(`{"low":%d,"med":%d,"high":%d}`,
		account.Thresholds.LowThreshold,
		account.Thresholds.MedThreshold,
		account.Thresholds.HighThreshold,
	)
	updatedAt := time.Time{}
	if account.LastModifiedTime != nil {
		updatedAt = account.LastModifiedTime.UTC()
	}

	return &backendclient.AccountDetail{
		ID:                 account.AccountID,
		Sequence:           account.Sequence,
		SequenceLedger:     optionalUint32(account.SequenceLedger),
		SequenceTime:       parseSequenceTime(account.SequenceTime),
		Balance:            nativeBalance,
		BuyingLiabilities:  buying,
		SellingLiabilities: selling,
		NumSubentries:      account.SubentryCount,
		HomeDomain:         homeDomain,
		Flags:              accountFlags(account.Flags),
		InflationDest:      inflationDest,
		Thresholds:         &thresholds,
		LastModifiedLedger: int64(account.LastModifiedLedger),
		Sponsor:            sponsor,
		NumSponsored:       int32(account.NumSponsored),
		NumSponsoring:      int32(account.NumSponsoring),
		UpdatedAt:          updatedAt,
	}
}

func trustlineSummaries(account hProtocol.Account) []backendclient.TrustlineSummary {
	lines := make([]backendclient.TrustlineSummary, 0)
	for _, balance := range account.Balances {
		if balance.Asset.Type == "native" {
			continue
		}
		code, issuer := assetCodeIssuer(balance.Asset)
		if code == "" {
			continue
		}
		var sponsor *string
		if strings.TrimSpace(balance.Sponsor) != "" {
			value := strings.TrimSpace(balance.Sponsor)
			sponsor = &value
		}
		updatedAt := trustlineUpdatedAt(account, balance)
		lines = append(lines, backendclient.TrustlineSummary{
			AssetType:          assetTypeCode(balance.Asset.Type),
			AssetCode:          code,
			AssetIssuer:        issuer,
			Balance:            balance.Balance,
			LimitAmount:        balance.Limit,
			BuyingLiabilities:  balance.BuyingLiabilities,
			SellingLiabilities: balance.SellingLiabilities,
			Flags:              balanceFlags(balance),
			LastModifiedLedger: int64(balance.LastModifiedLedger),
			Sponsor:            sponsor,
			UpdatedAt:          updatedAt,
		})
	}
	return lines
}

func signerSummaries(account hProtocol.Account) []backendclient.AccountSignerSummary {
	signers := make([]backendclient.AccountSignerSummary, 0, len(account.Signers))
	for _, signer := range account.Signers {
		signers = append(signers, backendclient.AccountSignerSummary{
			SignerKey:          signer.Key,
			Weight:             int32(signer.Weight),
			Type:               signer.Type,
			LastModifiedLedger: int64(account.LastModifiedLedger),
		})
	}
	return signers
}

func assetDetail(asset hProtocol.AssetStat) *backendclient.AssetDetail {
	code, issuer := assetCodeIssuer(asset.Asset)
	var sacContract *string
	if strings.TrimSpace(asset.ContractID) != "" {
		value := strings.TrimSpace(asset.ContractID)
		sacContract = &value
	}
	return &backendclient.AssetDetail{
		AssetType:            assetTypeCode(asset.Asset.Type),
		AssetCode:            code,
		AssetIssuer:          issuer,
		NumAccounts:          asset.Accounts.Authorized,
		TotalSupply:          asset.Balances.Authorized,
		NumClaimableBalances: asset.NumClaimableBalances,
		NumLiquidityPools:    asset.NumLiquidityPools,
		NumContracts:         asset.NumContracts,
		Flags:                accountFlags(asset.Flags),
		AuthRequired:         asset.Flags.AuthRequired,
		AuthRevocable:        asset.Flags.AuthRevocable,
		AuthImmutable:        asset.Flags.AuthImmutable,
		ClawbackEnabled:      asset.Flags.AuthClawbackEnabled,
		SACContractID:        sacContract,
		UpdatedAt:            time.Now().UTC(),
	}
}

func assetHolderSummary(account hProtocol.Account, code, issuer string) (backendclient.AssetHolderSummary, bool) {
	for _, balance := range account.Balances {
		assetCode, assetIssuer := assetCodeIssuer(balance.Asset)
		if assetCode != code || assetIssuer != issuer {
			continue
		}
		var sponsor *string
		if strings.TrimSpace(balance.Sponsor) != "" {
			value := strings.TrimSpace(balance.Sponsor)
			sponsor = &value
		}
		updatedAt := trustlineUpdatedAt(account, balance)
		return backendclient.AssetHolderSummary{
			AccountID:          account.AccountID,
			Balance:            balance.Balance,
			LimitAmount:        balance.Limit,
			BuyingLiabilities:  balance.BuyingLiabilities,
			SellingLiabilities: balance.SellingLiabilities,
			LastModifiedLedger: int64(balance.LastModifiedLedger),
			Sponsor:            sponsor,
			UpdatedAt:          updatedAt,
		}, true
	}
	return backendclient.AssetHolderSummary{}, false
}

func timelineFromTransactions(txs []hProtocol.Transaction) []backendclient.TimelineItem {
	items := make([]backendclient.TimelineItem, 0, len(txs))
	for _, tx := range txs {
		items = append(items, backendclient.TimelineItem{
			Kind:        "tx",
			Title:       "Transaction",
			Description: truncate(tx.Hash, 18),
			Command:     "lookup tx " + tx.Hash,
			OccurredAt:  tx.LedgerCloseTime,
		})
	}
	return items
}

func timelineFromOperationSummaries(ops []backendclient.OperationSummary) []backendclient.TimelineItem {
	items := make([]backendclient.TimelineItem, 0, len(ops))
	for _, op := range ops {
		items = append(items, backendclient.TimelineItem{
			Kind:        "op",
			Title:       strings.TrimSpace(op.TypeName),
			Description: truncate(op.TransactionHash, 18),
			Command:     fmt.Sprintf("lookup op %s:%d", op.TransactionHash, op.ApplicationOrder),
			OccurredAt:  op.CreatedAt,
		})
	}
	return items
}

func timelineFromOperations(ops []hoperations.Operation) []backendclient.TimelineItem {
	items := make([]backendclient.TimelineItem, 0, len(ops))
	for _, op := range ops {
		base := op.GetBase()
		items = append(items, backendclient.TimelineItem{
			Kind:        "op",
			Title:       strings.TrimSpace(base.GetType()),
			Description: truncate(base.TransactionHash, 18),
			Command:     fmt.Sprintf("lookup op %s:%d", base.TransactionHash, operationIndex(base.ID)),
			OccurredAt:  base.LedgerCloseTime,
		})
	}
	return items
}

func populateOperationFields(summary *backendclient.OperationSummary, op hoperations.Operation) {
	switch typed := op.(type) {
	case hoperations.Payment:
		code, issuer := assetCodeIssuer(typed.Asset)
		if code != "" {
			summary.AssetCode = &code
		}
		if issuer != "" {
			summary.AssetIssuer = &issuer
		}
		if strings.TrimSpace(typed.Amount) != "" {
			amount := strings.TrimSpace(typed.Amount)
			summary.Amount = &amount
		}
		if strings.TrimSpace(typed.To) != "" {
			dest := strings.TrimSpace(typed.To)
			summary.Destination = &dest
		}
	case hoperations.InvokeHostFunction:
		if strings.TrimSpace(typed.Function) != "" {
			fn := strings.TrimSpace(typed.Function)
			summary.FunctionName = &fn
		}
		if strings.TrimSpace(typed.Address) != "" {
			contract := strings.TrimSpace(typed.Address)
			summary.ContractID = &contract
		}
	}
}

func operationDetailsJSON(base hoperations.Base) string {
	payload := map[string]any{
		"type": strings.TrimSpace(base.Type),
		"id":   base.ID,
	}
	if strings.TrimSpace(base.TransactionHash) != "" {
		payload["transaction_hash"] = base.TransactionHash
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"type":%q}`, strings.TrimSpace(base.Type))
	}
	return string(encoded)
}

func effectDetailsJSON(effect effects.Effect) string {
	payload := map[string]any{
		"type": strings.TrimSpace(effect.GetType()),
		"id":   effect.GetID(),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"type":%q}`, strings.TrimSpace(effect.GetType()))
	}
	return string(encoded)
}

func memoFields(memoType string, memo string, memoBytes []byte) (int16, *string, *string) {
	switch strings.TrimSpace(memoType) {
	case "text":
		if strings.TrimSpace(memo) != "" {
			value := strings.TrimSpace(memo)
			return 1, &value, nil
		}
	case "hash":
		if len(memoBytes) > 0 {
			value := fmt.Sprintf("%x", memoBytes)
			return 2, nil, &value
		}
	case "return":
		return 3, nil, nil
	case "id":
		if strings.TrimSpace(memo) != "" {
			text := strings.TrimSpace(memo)
			return 4, &text, nil
		}
	}
	return 0, nil, nil
}

func transactionStatus(successful bool) int16 {
	if successful {
		return 1
	}
	return 0
}

func assetCodeIssuer(asset base.Asset) (string, string) {
	switch asset.Type {
	case "native":
		return "XLM", ""
	case "credit_alphanum4", "credit_alphanum12":
		return strings.TrimSpace(asset.Code), strings.TrimSpace(asset.Issuer)
	default:
		return strings.TrimSpace(asset.Code), strings.TrimSpace(asset.Issuer)
	}
}

func assetTypeCode(assetType string) int16 {
	switch assetType {
	case "native":
		return 0
	case "credit_alphanum4":
		return 1
	case "credit_alphanum12":
		return 2
	default:
		return 1
	}
}

func nativeBalanceFields(account hProtocol.Account) (balance string, buying string, selling string) {
	for _, entry := range account.Balances {
		if entry.Asset.Type != "native" {
			continue
		}
		return entry.Balance, entry.BuyingLiabilities, entry.SellingLiabilities
	}
	if value, err := account.GetNativeBalance(); err == nil {
		return value, "0", "0"
	}
	return "0", "0", "0"
}

func accountFlags(flags hProtocol.AccountFlags) int32 {
	var value int32
	if flags.AuthRequired {
		value |= 1
	}
	if flags.AuthRevocable {
		value |= 2
	}
	if flags.AuthImmutable {
		value |= 4
	}
	if flags.AuthClawbackEnabled {
		value |= 8
	}
	return value
}

func balanceFlags(balance hProtocol.Balance) int32 {
	var value int32
	if balance.IsAuthorized != nil && *balance.IsAuthorized {
		value |= 1
	}
	if balance.IsAuthorizedToMaintainLiabilities != nil && *balance.IsAuthorizedToMaintainLiabilities {
		value |= 2
	}
	if balance.IsClawbackEnabled != nil && *balance.IsClawbackEnabled {
		value |= 8
	}
	return value
}

func operationIndex(id string) int32 {
	parts := strings.Split(strings.TrimSpace(id), "-")
	if len(parts) == 0 {
		return 0
	}
	parsed, err := strconv.ParseInt(parts[len(parts)-1], 10, 32)
	if err != nil {
		return 0
	}
	return int32(parsed)
}

func effectTypeI(effect effects.Effect) int16 {
	if base, ok := effect.(interface{ GetBase() effects.Base }); ok {
		return int16(base.GetBase().TypeI)
	}
	return 0
}

func effectCreatedAt(effect effects.Effect) time.Time {
	if base, ok := effect.(interface{ GetBase() effects.Base }); ok {
		return base.GetBase().LedgerCloseTime
	}
	return time.Time{}
}

func trustlineUpdatedAt(account hProtocol.Account, balance hProtocol.Balance) time.Time {
	if account.LastModifiedTime != nil {
		return account.LastModifiedTime.UTC()
	}
	return time.Time{}
}

func optionalUint32(value uint32) *int64 {
	if value == 0 {
		return nil
	}
	parsed := int64(value)
	return &parsed
}

func parseSequenceTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed == 0 {
		return nil
	}
	result := time.Unix(parsed, 0).UTC()
	return &result
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
