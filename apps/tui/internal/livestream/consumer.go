package livestream

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxReconnectBackoff     = 30 * time.Second
	initialReconnectBackoff = time.Second
)

// Consumer subscribes to tui-indexer Redis pub/sub channels.
type Consumer struct {
	client *redis.Client
}

// NewConsumer connects to Redis and verifies connectivity.
func NewConsumer(redisURL string) (*Consumer, error) {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil, fmt.Errorf("redis url is required")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Consumer{client: client}, nil
}

// Subscribe listens for ledger and transaction stream events until the context is canceled.
// It reconnects with exponential backoff when the subscription drops.
func (c *Consumer) Subscribe(ctx context.Context) (<-chan Update, <-chan error) {
	updates := make(chan Update, 32)
	errs := make(chan error, 1)

	go func() {
		defer close(updates)
		defer close(errs)
		defer c.client.Close()

		backoff := initialReconnectBackoff
		for {
			if ctx.Err() != nil {
				return
			}

			stopped, err := c.runSubscription(ctx, updates)
			if stopped {
				return
			}
			if err != nil {
				select {
				case errs <- err:
				default:
				}
			}

			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			if backoff < maxReconnectBackoff {
				backoff *= 2
				if backoff > maxReconnectBackoff {
					backoff = maxReconnectBackoff
				}
			}
		}
	}()

	return updates, errs
}

func (c *Consumer) runSubscription(ctx context.Context, updates chan<- Update) (stopped bool, err error) {
	pubsub := c.client.Subscribe(ctx, ChannelLedgers, ChannelTransactions)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		if ctx.Err() != nil {
			return true, nil
		}
		return false, fmt.Errorf("redis subscribe: %w", err)
	}

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return true, nil
		case message, ok := <-ch:
			if !ok {
				return false, fmt.Errorf("redis subscription closed")
			}
			if message == nil {
				continue
			}

			update, err := parsePubSubMessage(message.Channel, []byte(message.Payload))
			if err != nil {
				return false, err
			}
			if update.Ledger == nil && len(update.Transactions) == 0 {
				continue
			}
			select {
			case updates <- update:
			case <-ctx.Done():
				return true, nil
			}
		}
	}
}
