package publisher

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func getTestRedis(t *testing.T) *RedisPublisher {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = "redis://localhost:63790"
	}
	pub, err := NewRedisPublisher(url)
	if err != nil {
		t.Skipf("Skipping: cannot connect to test Redis: %v", err)
	}
	return pub
}

func TestPublishLedger(t *testing.T) {
	pub := getTestRedis(t)
	defer pub.Close()

	ctx := context.Background()

	// Subscribe to the channel
	opts, _ := redis.ParseURL("redis://localhost:63790")
	sub := redis.NewClient(opts)
	defer sub.Close()

	pubsub := sub.Subscribe(ctx, ChannelLedgers)
	defer pubsub.Close()

	// Wait for subscription to be ready
	_, err := pubsub.Receive(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Publish a ledger
	ledger := &store.Ledger{
		Sequence:         12345,
		Hash:             "testhash",
		ClosedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		TransactionCount: 5,
		OperationCount:   10,
		ProtocolVersion:  21,
	}

	err = pub.PublishLedger(ctx, ledger)
	if err != nil {
		t.Fatalf("PublishLedger failed: %v", err)
	}

	// Receive the message
	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	var summary LedgerSummary
	if err := json.Unmarshal([]byte(msg.Payload), &summary); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if summary.Sequence != 12345 {
		t.Errorf("expected sequence 12345, got %d", summary.Sequence)
	}
	if summary.TransactionCount != 5 {
		t.Errorf("expected tx count 5, got %d", summary.TransactionCount)
	}
}

func TestPublishTransactions(t *testing.T) {
	pub := getTestRedis(t)
	defer pub.Close()

	ctx := context.Background()

	// Subscribe
	opts, _ := redis.ParseURL("redis://localhost:63790")
	sub := redis.NewClient(opts)
	defer sub.Close()

	pubsub := sub.Subscribe(ctx, ChannelTransactions)
	defer pubsub.Close()

	_, err := pubsub.Receive(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	txs := []store.Transaction{
		{
			Hash:           "txhash1",
			LedgerSequence: 100,
			Account:        "GABC",
			OperationCount: 2,
			Status:         1,
			IsSoroban:      true,
		},
	}

	err = pub.PublishTransactions(ctx, txs, []store.Operation{
		{
			TransactionHash:  "txhash1",
			ApplicationOrder: 1,
			TypeName:         "invoke_host_function",
			ContractID:       strPtr("CABC"),
		},
	})
	if err != nil {
		t.Fatalf("PublishTransactions failed: %v", err)
	}

	msg, err := pubsub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage failed: %v", err)
	}

	var summaries []TransactionSummary
	if err := json.Unmarshal([]byte(msg.Payload), &summaries); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(summaries))
	}
	if summaries[0].Hash != "txhash1" {
		t.Errorf("expected hash txhash1, got %s", summaries[0].Hash)
	}
	if !summaries[0].IsSoroban {
		t.Error("expected is_soroban=true")
	}
	if summaries[0].PrimaryContractID != "CABC" {
		t.Errorf("expected primary contract CABC, got %q", summaries[0].PrimaryContractID)
	}
	if summaries[0].PrimaryOperationType != "invoke_host_function" {
		t.Errorf("expected invoke_host_function, got %q", summaries[0].PrimaryOperationType)
	}
}

func strPtr(value string) *string {
	return &value
}
