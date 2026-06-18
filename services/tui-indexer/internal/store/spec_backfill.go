package store

import (
	"context"
	"strings"
	"time"
)

// BackfillOperationDetails persists spec-decoded operation details discovered at read time.
func (s *PostgresStore) BackfillOperationDetails(
	ctx context.Context,
	transactionHash string,
	applicationOrder int32,
	createdAt time.Time,
	details string,
) error {
	transactionHash = strings.TrimSpace(transactionHash)
	details = strings.TrimSpace(details)
	if transactionHash == "" || details == "" || applicationOrder <= 0 || createdAt.IsZero() {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE operations
		SET details = $4::jsonb
		WHERE transaction_hash = $1
		  AND application_order = $2
		  AND created_at = $3`,
		transactionHash,
		applicationOrder,
		createdAt,
		details,
	)
	return err
}

// BackfillContractEventSpecDecode persists spec-decoded contract event payloads discovered at read time.
func (s *PostgresStore) BackfillContractEventSpecDecode(ctx context.Context, event ReadContractEventSummary) error {
	contractID := strings.TrimSpace(event.ContractID)
	transactionHash := strings.TrimSpace(event.TransactionHash)
	if contractID == "" || transactionHash == "" || event.CreatedAt.IsZero() {
		return nil
	}
	valueDecoded := ""
	if event.ValueDecoded != nil {
		valueDecoded = strings.TrimSpace(*event.ValueDecoded)
	}
	if valueDecoded == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE contract_events
		SET value_decoded = $5::jsonb,
		    topic_1 = COALESCE($6, topic_1)
		WHERE contract_id = $1
		  AND transaction_hash = $2
		  AND ledger_sequence = $3
		  AND created_at = $4`,
		contractID,
		transactionHash,
		event.LedgerSequence,
		event.CreatedAt,
		valueDecoded,
		event.Topic1,
	)
	return err
}
