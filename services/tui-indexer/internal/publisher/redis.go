package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

const (
	ChannelLedgers      = "tui-indexer:stream:ledgers"
	ChannelTransactions = "tui-indexer:stream:transactions"
)

// RedisPublisher publishes ingestion events to Redis pub/sub channels.
type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(redisURL string) (*RedisPublisher, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisPublisher{client: client}, nil
}

// LedgerSummary is a compact representation published to Redis.
type LedgerSummary struct {
	Sequence         uint32 `json:"sequence"`
	Hash             string `json:"hash"`
	ClosedAt         string `json:"closed_at"`
	TransactionCount int32  `json:"transaction_count"`
	OperationCount   int32  `json:"operation_count"`
	ProtocolVersion  int32  `json:"protocol_version"`
}

// TransactionSummary is a compact representation published to Redis.
type TransactionSummary struct {
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

func (p *RedisPublisher) PublishLedger(ctx context.Context, ledger *store.Ledger) error {
	summary := LedgerSummary{
		Sequence:         ledger.Sequence,
		Hash:             ledger.Hash,
		ClosedAt:         ledger.ClosedAt.Format("2006-01-02T15:04:05Z"),
		TransactionCount: ledger.TransactionCount,
		OperationCount:   ledger.OperationCount,
		ProtocolVersion:  ledger.ProtocolVersion,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshal ledger summary: %w", err)
	}

	if err := p.client.Publish(ctx, ChannelLedgers, data).Err(); err != nil {
		log.Printf("redis publish ledger %d: %v", ledger.Sequence, err)
		return err
	}

	return nil
}

func (p *RedisPublisher) PublishTransactions(ctx context.Context, txs []store.Transaction, ops []store.Operation) error {
	return p.publishTransactionSummaries(ctx, SummariesFromStore(txs, ops))
}

func (p *RedisPublisher) publishTransactionSummaries(ctx context.Context, summaries []TransactionSummary) error {
	if len(summaries) == 0 {
		return nil
	}

	data, err := json.Marshal(summaries)
	if err != nil {
		return fmt.Errorf("marshal transaction summaries: %w", err)
	}

	if err := p.client.Publish(ctx, ChannelTransactions, data).Err(); err != nil {
		log.Printf("redis publish transactions: %v", err)
		return err
	}

	return nil
}

func (p *RedisPublisher) Close() error {
	return p.client.Close()
}
