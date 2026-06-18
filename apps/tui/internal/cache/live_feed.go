package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// UpsertLiveTransactions stores recent live-feed rows for one profile.
func (s *Store) UpsertLiveTransactions(ctx context.Context, transactions []LiveTransaction) error {
	if s == nil || s.db == nil {
		return errors.New("cache: store is not initialized")
	}
	if len(transactions) == 0 {
		return nil
	}

	const query = `
INSERT INTO live_transactions (
    profile_id, hash, ledger_sequence, application_order, account,
    operation_count, status, is_soroban, created_at, cached_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(profile_id, hash) DO UPDATE SET
    ledger_sequence = excluded.ledger_sequence,
    application_order = excluded.application_order,
    account = excluded.account,
    operation_count = excluded.operation_count,
    status = excluded.status,
    is_soroban = excluded.is_soroban,
    created_at = excluded.created_at,
    cached_at = excluded.cached_at;`

	now := time.Now().UTC()
	for _, tx := range transactions {
		if tx.ProfileID == "" || tx.Hash == "" {
			continue
		}
		cachedAt := tx.CachedAt
		if cachedAt.IsZero() {
			cachedAt = now
		}
		if _, err := s.db.ExecContext(
			ctx,
			query,
			tx.ProfileID,
			tx.Hash,
			int64(tx.LedgerSequence),
			int64(tx.ApplicationOrder),
			tx.Account,
			int64(tx.OperationCount),
			int64(tx.Status),
			boolToInt(tx.IsSoroban),
			tx.CreatedAt,
			cachedAt,
		); err != nil {
			return fmt.Errorf("cache: upsert live transaction %q: %w", tx.Hash, err)
		}
	}

	return nil
}

// ListLiveTransactions returns recent cached live-feed rows for one profile.
func (s *Store) ListLiveTransactions(ctx context.Context, profileID string, limit int) ([]LiveTransaction, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("cache: store is not initialized")
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT profile_id, hash, ledger_sequence, application_order, account,
       operation_count, status, is_soroban, created_at, cached_at
FROM live_transactions
WHERE profile_id = ?
ORDER BY ledger_sequence DESC, application_order DESC, cached_at DESC
LIMIT ?;`, profileID, limit)
	if err != nil {
		return nil, fmt.Errorf("cache: list live transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]LiveTransaction, 0, limit)
	for rows.Next() {
		var tx LiveTransaction
		var ledgerSequence int64
		var applicationOrder int64
		var operationCount int64
		var status int64
		var isSoroban int64
		if err := rows.Scan(
			&tx.ProfileID,
			&tx.Hash,
			&ledgerSequence,
			&applicationOrder,
			&tx.Account,
			&operationCount,
			&status,
			&isSoroban,
			&tx.CreatedAt,
			&tx.CachedAt,
		); err != nil {
			return nil, fmt.Errorf("cache: scan live transaction: %w", err)
		}
		tx.LedgerSequence = uint32(ledgerSequence)
		tx.ApplicationOrder = int32(applicationOrder)
		tx.OperationCount = int32(operationCount)
		tx.Status = int16(status)
		tx.IsSoroban = isSoroban != 0
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cache: iterate live transactions: %w", err)
	}

	return transactions, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
