package publisher

import (
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

// SummariesFromStore builds compact Redis summaries with primary operation metadata.
func SummariesFromStore(txs []store.Transaction, ops []store.Operation) []TransactionSummary {
	if len(txs) == 0 {
		return nil
	}

	opsByTx := make(map[string][]store.Operation, len(txs))
	for _, op := range ops {
		hash := strings.TrimSpace(op.TransactionHash)
		if hash == "" {
			continue
		}
		opsByTx[hash] = append(opsByTx[hash], op)
	}

	summaries := make([]TransactionSummary, 0, len(txs))
	for _, tx := range txs {
		contractID, assetCode, assetIssuer, opType := primaryOperationMetadata(opsByTx[tx.Hash])
		summaries = append(summaries, TransactionSummary{
			Hash:                 tx.Hash,
			LedgerSequence:       tx.LedgerSequence,
			Account:              tx.Account,
			OperationCount:       tx.OperationCount,
			Status:               tx.Status,
			IsSoroban:            tx.IsSoroban,
			PrimaryContractID:    contractID,
			PrimaryAssetCode:     assetCode,
			PrimaryAssetIssuer:   assetIssuer,
			PrimaryOperationType: opType,
		})
	}
	return summaries
}

func primaryOperationMetadata(ops []store.Operation) (contractID, assetCode, assetIssuer, opType string) {
	if len(ops) == 0 {
		return "", "", "", ""
	}

	primary := ops[0]
	for _, op := range ops[1:] {
		if op.ApplicationOrder < primary.ApplicationOrder {
			primary = op
		}
	}

	opType = strings.TrimSpace(primary.TypeName)
	if primary.ContractID != nil {
		contractID = strings.TrimSpace(*primary.ContractID)
	}
	if primary.AssetCode != nil {
		assetCode = strings.TrimSpace(*primary.AssetCode)
	}
	if primary.AssetIssuer != nil {
		assetIssuer = strings.TrimSpace(*primary.AssetIssuer)
	}
	return contractID, assetCode, assetIssuer, opType
}
