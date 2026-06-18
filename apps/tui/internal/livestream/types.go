package livestream

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

const (
	ChannelLedgers      = "tui-indexer:stream:ledgers"
	ChannelTransactions = "tui-indexer:stream:transactions"
)

// LedgerMessage mirrors the compact ledger payload published by tui-indexer.
type LedgerMessage struct {
	Sequence         uint32 `json:"sequence"`
	Hash             string `json:"hash"`
	ClosedAt         string `json:"closed_at"`
	TransactionCount int32  `json:"transaction_count"`
	OperationCount   int32  `json:"operation_count"`
	ProtocolVersion  int32  `json:"protocol_version"`
}

// TransactionMessage mirrors the compact transaction payload published by tui-indexer.
type TransactionMessage struct {
	Hash                 string `json:"hash"`
	LedgerSequence       uint32 `json:"ledger_sequence"`
	Account              string `json:"account"`
	OperationCount       int32  `json:"operation_count"`
	Status               int16  `json:"status"`
	IsSoroban            bool   `json:"is_soroban"`
	PrimaryContractID    string `json:"primary_contract_id,omitempty"`
	PrimaryAssetCode     string `json:"primary_asset_code,omitempty"`
	PrimaryAssetIssuer   string `json:"primary_asset_issuer,omitempty"`
	PrimaryOperationType string `json:"primary_operation_type,omitempty"`
}

// Update is a normalized live-stream payload for the TUI app model.
type Update struct {
	Ledger       *backendclient.LedgerSummary
	Transactions []backendclient.TransactionSummary
}

func parseLedgerMessage(payload []byte) (*backendclient.LedgerSummary, error) {
	var message LedgerMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		return nil, err
	}
	closedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(message.ClosedAt))
	if err != nil {
		closedAt = time.Time{}
	}
	return &backendclient.LedgerSummary{
		Sequence:         message.Sequence,
		Hash:             strings.TrimSpace(message.Hash),
		ClosedAt:         closedAt,
		TransactionCount: message.TransactionCount,
		OperationCount:   message.OperationCount,
	}, nil
}

func parseTransactionMessages(payload []byte) ([]backendclient.TransactionSummary, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var batch []TransactionMessage
	if err := json.Unmarshal(payload, &batch); err == nil && len(batch) > 0 {
		return convertTransactionMessages(batch), nil
	}

	var single TransactionMessage
	if err := json.Unmarshal(payload, &single); err != nil {
		return nil, err
	}
	if strings.TrimSpace(single.Hash) == "" {
		return nil, nil
	}
	return convertTransactionMessages([]TransactionMessage{single}), nil
}

func parsePubSubMessage(channel string, payload []byte) (Update, error) {
	switch strings.TrimSpace(channel) {
	case ChannelLedgers:
		ledger, err := parseLedgerMessage(payload)
		if err != nil {
			return Update{}, fmt.Errorf("parse ledger stream payload: %w", err)
		}
		return Update{Ledger: ledger}, nil
	case ChannelTransactions:
		transactions, err := parseTransactionMessages(payload)
		if err != nil {
			return Update{}, fmt.Errorf("parse transaction stream payload: %w", err)
		}
		return Update{Transactions: transactions}, nil
	default:
		return Update{}, fmt.Errorf("unknown stream channel %q", channel)
	}
}

func convertTransactionMessages(messages []TransactionMessage) []backendclient.TransactionSummary {
	out := make([]backendclient.TransactionSummary, 0, len(messages))
	for _, message := range messages {
		hash := strings.TrimSpace(message.Hash)
		if hash == "" {
			continue
		}
		out = append(out, backendclient.TransactionSummary{
			Hash:                 hash,
			LedgerSequence:       message.LedgerSequence,
			Account:              strings.TrimSpace(message.Account),
			OperationCount:       message.OperationCount,
			Status:               message.Status,
			IsSoroban:            message.IsSoroban,
			PrimaryContractID:    strings.TrimSpace(message.PrimaryContractID),
			PrimaryAssetCode:     strings.TrimSpace(message.PrimaryAssetCode),
			PrimaryAssetIssuer:   strings.TrimSpace(message.PrimaryAssetIssuer),
			PrimaryOperationType: strings.TrimSpace(message.PrimaryOperationType),
		})
	}
	return out
}
