package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) InsertLedger(ctx context.Context, ledger *Ledger) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ledgers (sequence, hash, prev_hash, closed_at, total_coins, fee_pool,
			base_fee, base_reserve, max_tx_set_size, protocol_version,
			transaction_count, operation_count, successful_tx_count, failed_tx_count,
			tx_set_operation_count, header_xdr)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (sequence, closed_at) DO NOTHING`,
		ledger.Sequence, ledger.Hash, ledger.PrevHash, ledger.ClosedAt,
		ledger.TotalCoins, ledger.FeePool, ledger.BaseFee, ledger.BaseReserve,
		ledger.MaxTxSetSize, ledger.ProtocolVersion, ledger.TransactionCount,
		ledger.OperationCount, ledger.SuccessfulTxCount, ledger.FailedTxCount,
		ledger.TxSetOperationCount, ledger.HeaderXDR,
	)
	return err
}

func (s *PostgresStore) InsertTransactionBatch(ctx context.Context, txs []Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.PrepareContext(ctx, `
		INSERT INTO transactions (hash, ledger_sequence, application_order, account,
			account_muxed, account_muxed_id, account_sequence, fee_charged, max_fee, operation_count,
			memo_type, memo_text, memo_hash, status, is_soroban, soroban_resources,
			envelope_xdr, result_xdr, result_meta_xdr, fee_meta_xdr, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range txs {
		_, err := stmt.ExecContext(ctx,
			t.Hash, t.LedgerSequence, t.ApplicationOrder, t.Account,
			t.AccountMuxed, t.AccountMuxedID, t.AccountSequence, t.FeeCharged, t.MaxFee, t.OperationCount,
			t.MemoType, t.MemoText, t.MemoHash, t.Status, t.IsSoroban, t.SorobanResources,
			t.EnvelopeXDR, t.ResultXDR, t.ResultMetaXDR, t.FeeMetaXDR, t.CreatedAt)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func (s *PostgresStore) InsertOperationBatch(ctx context.Context, ops []Operation) error {
	if len(ops) == 0 {
		return nil
	}
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.PrepareContext(ctx, `
		INSERT INTO operations (transaction_id, transaction_hash, application_order,
			type, type_name, source_account, source_account_muxed, source_muxed_id,
			asset_code, asset_issuer, amount,
			destination, destination_muxed, destination_muxed_id,
			contract_id, function_name, details, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, o := range ops {
		_, err := stmt.ExecContext(ctx,
			o.TransactionID, o.TransactionHash, o.ApplicationOrder,
			o.Type, o.TypeName, o.SourceAccount, o.SourceAccountMuxed, o.SourceMuxedID,
			o.AssetCode, o.AssetIssuer, o.Amount,
			o.Destination, o.DestinationMuxed, o.DestinationMuxedID,
			o.ContractID, o.FunctionName, o.Details, o.CreatedAt)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func (s *PostgresStore) InsertTokenEventBatch(ctx context.Context, events []TokenEvent) error {
	if len(events) == 0 {
		return nil
	}
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.PrepareContext(ctx, `
		INSERT INTO token_events (
			event_type, event_type_name,
			from_address, from_muxed, to_address, to_muxed, to_muxed_id,
			asset_type, asset_code, asset_issuer, asset_contract_id,
			amount, amount_formatted,
			transaction_hash, ledger_sequence, operation_index, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.ExecContext(ctx,
			e.EventType, e.EventTypeName,
			e.FromAddress, e.FromMuxed, e.ToAddress, e.ToMuxed, e.ToMuxedID,
			e.AssetType, e.AssetCode, e.AssetIssuer, e.AssetContractID,
			e.Amount, e.AmountFormatted,
			e.TransactionHash, e.LedgerSequence, e.OperationIndex, e.CreatedAt)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func (s *PostgresStore) InsertContractEventBatch(ctx context.Context, events []ContractEvent) error {
	if len(events) == 0 {
		return nil
	}
	dbTx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.PrepareContext(ctx, `
		INSERT INTO contract_events (
			contract_id, transaction_hash, ledger_sequence, type,
			topic_1, topic_2, topic_3, topic_4,
			topics_xdr, value_xdr, topics_decoded, value_decoded, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		_, err := stmt.ExecContext(ctx,
			e.ContractID, e.TransactionHash, e.LedgerSequence, e.Type,
			e.Topic1, e.Topic2, e.Topic3, e.Topic4,
			e.TopicsXDR, e.ValueXDR, e.TopicsDecoded, e.ValueDecoded, e.CreatedAt)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func (s *PostgresStore) UpsertContract(ctx context.Context, c *Contract) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contracts (
			contract_id, wasm_hash, creator_account, created_ledger, created_at,
			last_modified_ledger, contract_type, is_sep41_token, is_sep50_nft,
			token_name, token_symbol, token_decimals, contract_spec, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
		ON CONFLICT (contract_id) DO UPDATE SET
			wasm_hash            = EXCLUDED.wasm_hash,
			last_modified_ledger = EXCLUDED.last_modified_ledger,
			contract_type        = EXCLUDED.contract_type,
			is_sep41_token       = EXCLUDED.is_sep41_token,
			is_sep50_nft         = EXCLUDED.is_sep50_nft,
			token_name           = COALESCE(EXCLUDED.token_name, contracts.token_name),
			token_symbol         = COALESCE(EXCLUDED.token_symbol, contracts.token_symbol),
			token_decimals       = COALESCE(EXCLUDED.token_decimals, contracts.token_decimals),
			contract_spec        = COALESCE(EXCLUDED.contract_spec, contracts.contract_spec),
			updated_at           = NOW()`,
		c.ContractID, c.WasmHash, c.CreatorAccount, c.CreatedLedger, c.CreatedAt,
		c.LastModifiedLedger, c.ContractType, c.IsSep41Token, c.IsSep50NFT,
		c.TokenName, c.TokenSymbol, c.TokenDecimals, c.ContractSpec,
	)
	return err
}

func (s *PostgresStore) UpsertContractCode(ctx context.Context, code *ContractCode) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contract_code (
			wasm_hash, wasm_bytecode, wasm_size, spec_xdr, spec_parsed, created_ledger, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (wasm_hash) DO UPDATE SET
			spec_xdr    = COALESCE(EXCLUDED.spec_xdr, contract_code.spec_xdr),
			spec_parsed = COALESCE(EXCLUDED.spec_parsed, contract_code.spec_parsed),
			contract_count = contract_code.contract_count + 1`,
		code.WasmHash, code.WasmBytecode, code.WasmSize,
		code.SpecXDR, code.SpecParsed, code.CreatedLedger, code.CreatedAt,
	)
	return err
}

// GetLastIngestedLedger returns the last processed ledger sequence.
func (s *PostgresStore) GetLastIngestedLedger(ctx context.Context) (uint32, error) {
	var seq uint32
	err := s.db.QueryRowContext(ctx,
		"SELECT COALESCE(value::int, 0) FROM ingestion_state WHERE key = 'last_ingested_ledger'").
		Scan(&seq)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return seq, err
}

// SetLastIngestedLedger updates the ingestion cursor (only advances forward).
func (s *PostgresStore) SetLastIngestedLedger(ctx context.Context, seq uint32) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ingestion_state (key, value, updated_at)
		VALUES ('last_ingested_ledger', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
		WHERE ingestion_state.value::bigint < $1::bigint`,
		fmt.Sprintf("%d", seq))
	return err
}

func (s *PostgresStore) FindMissingLedgerSequences(ctx context.Context, start uint32, end uint32, limit int) ([]uint32, error) {
	if start == 0 || end < start {
		return nil, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `
		WITH expected AS (
			SELECT generate_series($1::bigint, $2::bigint) AS sequence
		)
		SELECT expected.sequence
		FROM expected
		LEFT JOIN ledgers ON ledgers.sequence = expected.sequence
		WHERE ledgers.sequence IS NULL
		ORDER BY expected.sequence ASC
		LIMIT $3`, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sequences := make([]uint32, 0, limit)
	for rows.Next() {
		var sequence uint32
		if err := rows.Scan(&sequence); err != nil {
			return nil, err
		}
		sequences = append(sequences, sequence)
	}

	return sequences, rows.Err()
}

// CleanupTestData removes data for a specific ledger sequence (used by tests).
func (s *PostgresStore) CleanupTestData(ctx context.Context, sequence uint32) {
	_, _ = s.db.ExecContext(ctx, "DELETE FROM operations WHERE transaction_hash IN (SELECT hash FROM transactions WHERE ledger_sequence = $1)", sequence)
	_, _ = s.db.ExecContext(ctx, "DELETE FROM transactions WHERE ledger_sequence = $1", sequence)
	_, _ = s.db.ExecContext(ctx, "DELETE FROM ledgers WHERE sequence = $1", sequence)
}

// QueryRow executes a query that returns at most one row, exposing the
// underlying *sql.DB.QueryRowContext for ad-hoc queries (e.g. in tests).
func (s *PostgresStore) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) GetLatestLedgerSummary(ctx context.Context) (*ReadLedgerSummary, error) {
	var ledger ReadLedgerSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT sequence, hash, closed_at, transaction_count, operation_count, successful_tx_count, failed_tx_count
		FROM ledgers
		ORDER BY sequence DESC
		LIMIT 1`,
	).Scan(
		&ledger.Sequence,
		&ledger.Hash,
		&ledger.ClosedAt,
		&ledger.TransactionCount,
		&ledger.OperationCount,
		&ledger.SuccessfulTxCount,
		&ledger.FailedTxCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ledger, nil
}

func (s *PostgresStore) GetLedgerSummaryBySequence(ctx context.Context, sequence uint32) (*ReadLedgerSummary, error) {
	var ledger ReadLedgerSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT sequence, hash, closed_at, transaction_count, operation_count, successful_tx_count, failed_tx_count
		FROM ledgers
		WHERE sequence = $1
		LIMIT 1`, sequence,
	).Scan(
		&ledger.Sequence,
		&ledger.Hash,
		&ledger.ClosedAt,
		&ledger.TransactionCount,
		&ledger.OperationCount,
		&ledger.SuccessfulTxCount,
		&ledger.FailedTxCount,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ledger, nil
}

func (s *PostgresStore) ListLedgerSummaries(ctx context.Context, limit int, before uint32) ([]ReadLedgerSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT sequence, hash, closed_at, transaction_count, operation_count, successful_tx_count, failed_tx_count
		FROM ledgers
		WHERE ($2 = 0 OR sequence < $2)
		ORDER BY sequence DESC
		LIMIT $1`, limit, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ledgers := make([]ReadLedgerSummary, 0, limit)
	for rows.Next() {
		var ledger ReadLedgerSummary
		if err := rows.Scan(
			&ledger.Sequence,
			&ledger.Hash,
			&ledger.ClosedAt,
			&ledger.TransactionCount,
			&ledger.OperationCount,
			&ledger.SuccessfulTxCount,
			&ledger.FailedTxCount,
		); err != nil {
			return nil, err
		}
		ledgers = append(ledgers, ledger)
	}

	return ledgers, rows.Err()
}

const readTransactionSummaryFrom = `
FROM transactions t
LEFT JOIN LATERAL (
	SELECT contract_id, asset_code, asset_issuer, type_name
	FROM operations
	WHERE transaction_hash = t.hash
	ORDER BY application_order ASC
	LIMIT 1
) primary_op ON true`

const readTransactionSummarySelect = `
SELECT t.hash, t.ledger_sequence, t.application_order, t.account,
       t.operation_count, t.status, t.is_soroban, t.created_at,
       COALESCE(primary_op.contract_id, ''),
       COALESCE(primary_op.asset_code, ''),
       COALESCE(primary_op.asset_issuer, ''),
       COALESCE(primary_op.type_name, '') ` + readTransactionSummaryFrom

func (s *PostgresStore) ListRecentTransactionSummaries(ctx context.Context, limit int) ([]ReadTransactionSummary, error) {
	rows, err := s.db.QueryContext(ctx, readTransactionSummarySelect+`
		ORDER BY t.ledger_sequence DESC, t.application_order DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]ReadTransactionSummary, 0, limit)
	for rows.Next() {
		var tx ReadTransactionSummary
		if err := scanReadTransactionSummary(rows, &tx); err != nil {
			return nil, err
		}
		summaries = append(summaries, tx)
	}

	return summaries, rows.Err()
}

func (s *PostgresStore) ListTransactionSummariesByLedger(ctx context.Context, sequence uint32, limit int, offset int) ([]ReadTransactionSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, readTransactionSummarySelect+`
		WHERE t.ledger_sequence = $1
		ORDER BY t.application_order ASC, t.created_at ASC
		LIMIT $2 OFFSET $3`, sequence, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]ReadTransactionSummary, 0, limit)
	for rows.Next() {
		var tx ReadTransactionSummary
		if err := scanReadTransactionSummary(rows, &tx); err != nil {
			return nil, err
		}
		summaries = append(summaries, tx)
	}

	return summaries, rows.Err()
}

func (s *PostgresStore) ListTransactionSummariesByAccount(ctx context.Context, accountID string, limit int) ([]ReadTransactionSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT ON (t.hash) t.hash, t.ledger_sequence, t.application_order, t.account,
		       t.operation_count, t.status, t.is_soroban, t.created_at,
		       COALESCE(primary_op.contract_id, ''),
		       COALESCE(primary_op.asset_code, ''),
		       COALESCE(primary_op.asset_issuer, ''),
		       COALESCE(primary_op.type_name, '')
		`+readTransactionSummaryFrom+`
		LEFT JOIN operations o ON o.transaction_hash = t.hash
		WHERE t.account = $1 OR o.source_account = $1 OR o.destination = $1
		ORDER BY t.hash, t.created_at DESC, t.application_order DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]ReadTransactionSummary, 0, limit)
	for rows.Next() {
		var tx ReadTransactionSummary
		if err := scanReadTransactionSummary(rows, &tx); err != nil {
			return nil, err
		}
		summaries = append(summaries, tx)
	}

	return trimAndSortTransactionSummaries(summaries, limit), rows.Err()
}

func (s *PostgresStore) ListTransactionSummariesByAsset(ctx context.Context, code string, issuer string, limit int) ([]ReadTransactionSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT ON (t.hash) t.hash, t.ledger_sequence, t.application_order, t.account,
		       t.operation_count, t.status, t.is_soroban, t.created_at,
		       COALESCE(primary_op.contract_id, ''),
		       COALESCE(primary_op.asset_code, ''),
		       COALESCE(primary_op.asset_issuer, ''),
		       COALESCE(primary_op.type_name, '')
		FROM operations o
		JOIN transactions t ON t.hash = o.transaction_hash
		LEFT JOIN LATERAL (
			SELECT contract_id, asset_code, asset_issuer, type_name
			FROM operations
			WHERE transaction_hash = t.hash
			ORDER BY application_order ASC
			LIMIT 1
		) primary_op ON true
		WHERE o.asset_code = $1 AND o.asset_issuer = $2
		ORDER BY t.hash, t.created_at DESC, t.application_order DESC`, code, issuer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]ReadTransactionSummary, 0, limit)
	for rows.Next() {
		var tx ReadTransactionSummary
		if err := scanReadTransactionSummary(rows, &tx); err != nil {
			return nil, err
		}
		summaries = append(summaries, tx)
	}

	return trimAndSortTransactionSummaries(summaries, limit), rows.Err()
}

func (s *PostgresStore) ListTransactionSummariesByContract(ctx context.Context, contractID string, limit int) ([]ReadTransactionSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT ON (t.hash) t.hash, t.ledger_sequence, t.application_order, t.account,
		       t.operation_count, t.status, t.is_soroban, t.created_at,
		       COALESCE(primary_op.contract_id, ''),
		       COALESCE(primary_op.asset_code, ''),
		       COALESCE(primary_op.asset_issuer, ''),
		       COALESCE(primary_op.type_name, '')
		FROM operations o
		JOIN transactions t ON t.hash = o.transaction_hash
		LEFT JOIN LATERAL (
			SELECT contract_id, asset_code, asset_issuer, type_name
			FROM operations
			WHERE transaction_hash = t.hash
			ORDER BY application_order ASC
			LIMIT 1
		) primary_op ON true
		WHERE o.contract_id = $1
		ORDER BY t.hash, t.created_at DESC, t.application_order DESC`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]ReadTransactionSummary, 0, limit)
	for rows.Next() {
		var tx ReadTransactionSummary
		if err := scanReadTransactionSummary(rows, &tx); err != nil {
			return nil, err
		}
		summaries = append(summaries, tx)
	}

	return trimAndSortTransactionSummaries(summaries, limit), rows.Err()
}

func (s *PostgresStore) GetTransactionByHash(ctx context.Context, hash string) (*ReadTransactionDetail, error) {
	var tx ReadTransactionDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT hash, ledger_sequence, application_order, account, account_muxed, account_muxed_id,
		       account_sequence, fee_charged, max_fee, operation_count, memo_type, memo_text,
		       memo_hash, status, is_soroban, soroban_resources, envelope_xdr, result_xdr,
		       result_meta_xdr, fee_meta_xdr, created_at
		FROM transactions
		WHERE hash = $1
		LIMIT 1`, hash,
	).Scan(
		&tx.Hash,
		&tx.LedgerSequence,
		&tx.ApplicationOrder,
		&tx.Account,
		&tx.AccountMuxed,
		&tx.AccountMuxedID,
		&tx.AccountSequence,
		&tx.FeeCharged,
		&tx.MaxFee,
		&tx.OperationCount,
		&tx.MemoType,
		&tx.MemoText,
		&tx.MemoHash,
		&tx.Status,
		&tx.IsSoroban,
		&tx.SorobanResources,
		&tx.EnvelopeXDR,
		&tx.ResultXDR,
		&tx.ResultMetaXDR,
		&tx.FeeMetaXDR,
		&tx.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (s *PostgresStore) ListOperationsByTransactionHash(ctx context.Context, hash string) ([]ReadOperationSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT transaction_hash, application_order, type, type_name, source_account, source_account_muxed, source_muxed_id,
		       asset_code, asset_issuer, amount, destination, destination_muxed, destination_muxed_id,
		       contract_id, function_name, details, created_at
		FROM operations
		WHERE transaction_hash = $1
		ORDER BY application_order ASC`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ops := make([]ReadOperationSummary, 0)
	for rows.Next() {
		var op ReadOperationSummary
		if err := rows.Scan(
			&op.TransactionHash,
			&op.ApplicationOrder,
			&op.Type,
			&op.TypeName,
			&op.SourceAccount,
			&op.SourceAccountMuxed,
			&op.SourceMuxedID,
			&op.AssetCode,
			&op.AssetIssuer,
			&op.Amount,
			&op.Destination,
			&op.DestinationMuxed,
			&op.DestinationMuxedID,
			&op.ContractID,
			&op.FunctionName,
			&op.Details,
			&op.CreatedAt,
		); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

func (s *PostgresStore) ListEffectsByTransactionHash(ctx context.Context, hash string) ([]ReadEffectSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT transaction_hash, type, type_name, account, details, created_at
		FROM effects
		WHERE transaction_hash = $1
		ORDER BY id ASC`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	effects := make([]ReadEffectSummary, 0)
	for rows.Next() {
		var effect ReadEffectSummary
		if err := rows.Scan(
			&effect.TransactionHash,
			&effect.Type,
			&effect.TypeName,
			&effect.Account,
			&effect.Details,
			&effect.CreatedAt,
		); err != nil {
			return nil, err
		}
		effects = append(effects, effect)
	}

	return effects, rows.Err()
}

func (s *PostgresStore) ListOperationSummariesByContract(ctx context.Context, contractID string, limit int, offset int) ([]ReadOperationSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT transaction_hash, application_order, type, type_name, source_account, source_account_muxed, source_muxed_id,
		       asset_code, asset_issuer, amount, destination, destination_muxed, destination_muxed_id,
		       contract_id, function_name, details, created_at
		FROM operations
		WHERE contract_id = $1 AND type_name = 'invoke_host_function'
		ORDER BY created_at DESC, application_order DESC
		LIMIT $2 OFFSET $3`, contractID, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ops := make([]ReadOperationSummary, 0, limit)
	for rows.Next() {
		var op ReadOperationSummary
		if err := rows.Scan(
			&op.TransactionHash,
			&op.ApplicationOrder,
			&op.Type,
			&op.TypeName,
			&op.SourceAccount,
			&op.SourceAccountMuxed,
			&op.SourceMuxedID,
			&op.AssetCode,
			&op.AssetIssuer,
			&op.Amount,
			&op.Destination,
			&op.DestinationMuxed,
			&op.DestinationMuxedID,
			&op.ContractID,
			&op.FunctionName,
			&op.Details,
			&op.CreatedAt,
		); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

func (s *PostgresStore) ListOperationSummariesByAccount(ctx context.Context, accountID string, limit int, offset int) ([]ReadOperationSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT transaction_hash, application_order, type, type_name, source_account, source_account_muxed, source_muxed_id,
		       asset_code, asset_issuer, amount, destination, destination_muxed, destination_muxed_id,
		       contract_id, function_name, details, created_at
		FROM operations
		WHERE source_account = $1 OR destination = $1
		ORDER BY created_at DESC, application_order DESC
		LIMIT $2 OFFSET $3`, accountID, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ops := make([]ReadOperationSummary, 0, limit)
	for rows.Next() {
		var op ReadOperationSummary
		if err := rows.Scan(
			&op.TransactionHash,
			&op.ApplicationOrder,
			&op.Type,
			&op.TypeName,
			&op.SourceAccount,
			&op.SourceAccountMuxed,
			&op.SourceMuxedID,
			&op.AssetCode,
			&op.AssetIssuer,
			&op.Amount,
			&op.Destination,
			&op.DestinationMuxed,
			&op.DestinationMuxedID,
			&op.ContractID,
			&op.FunctionName,
			&op.Details,
			&op.CreatedAt,
		); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

func (s *PostgresStore) ListAccountTimeline(ctx context.Context, accountID string, limit int, offset int, category string) ([]ReadTimelineItem, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		WITH related_transactions AS (
			SELECT DISTINCT ON (t.hash)
			       'tx'::text AS kind,
			       'Transaction ' || left(t.hash, 16) AS title,
			       'ledger ' || t.ledger_sequence::text || '  ops ' || t.operation_count::text AS description,
			       'lookup tx ' || t.hash AS command,
			       t.created_at AS occurred_at,
			       0 AS item_order
			FROM transactions t
			LEFT JOIN operations o ON o.transaction_hash = t.hash
			WHERE t.account = $1 OR o.source_account = $1 OR o.destination = $1
			ORDER BY t.hash, t.created_at DESC, t.application_order DESC
		),
		related_operations AS (
			SELECT 'op'::text AS kind,
			       'Operation ' || COALESCE(NULLIF(type_name, ''), 'unknown') AS title,
			       'tx ' || left(transaction_hash, 16) AS description,
			       CASE
			         WHEN contract_id IS NOT NULL AND contract_id <> '' THEN 'lookup contract ' || contract_id
			         WHEN destination IS NOT NULL AND destination <> '' THEN 'lookup account ' || destination
			         WHEN asset_code IS NOT NULL AND asset_code <> '' AND asset_issuer IS NOT NULL AND asset_issuer <> '' THEN 'lookup asset ' || asset_code || ':' || asset_issuer
			         WHEN source_account IS NOT NULL AND source_account <> '' THEN 'lookup account ' || source_account
			         ELSE 'lookup tx ' || transaction_hash
			       END AS command,
			       created_at AS occurred_at,
			       application_order AS item_order
			FROM operations
			WHERE source_account = $1 OR destination = $1
		)
		SELECT kind, title, description, command, occurred_at
		FROM (
			SELECT * FROM related_transactions
			UNION ALL
			SELECT * FROM related_operations
		) timeline
		WHERE ($4 = '' OR kind = $4)
		ORDER BY occurred_at DESC, item_order ASC
		LIMIT $2 OFFSET $3`, accountID, limit, normalizeOffset(offset), strings.TrimSpace(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReadTimelineItem, 0, limit)
	for rows.Next() {
		var item ReadTimelineItem
		if err := rows.Scan(
			&item.Kind,
			&item.Title,
			&item.Description,
			&item.Command,
			&item.OccurredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *PostgresStore) ListAssetTimeline(ctx context.Context, code string, issuer string, limit int, offset int, category string) ([]ReadTimelineItem, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		WITH related_transactions AS (
			SELECT DISTINCT ON (t.hash)
			       'tx'::text AS kind,
			       'Transaction ' || left(t.hash, 16) AS title,
			       'ledger ' || t.ledger_sequence::text || '  ops ' || t.operation_count::text AS description,
			       'lookup tx ' || t.hash AS command,
			       t.created_at AS occurred_at,
			       0 AS item_order
			FROM operations o
			JOIN transactions t ON t.hash = o.transaction_hash
			WHERE o.asset_code = $1 AND o.asset_issuer = $2
			ORDER BY t.hash, t.created_at DESC, t.application_order DESC
		),
		related_holders AS (
			SELECT 'holder'::text AS kind,
			       'Holder ' || left(account_id, 16) AS title,
			       'balance ' || balance::text AS description,
			       'lookup account ' || account_id AS command,
			       updated_at AS occurred_at,
			       1 AS item_order
			FROM trustlines
			WHERE asset_code = $1 AND asset_issuer = $2
		)
		SELECT kind, title, description, command, occurred_at
		FROM (
			SELECT * FROM related_transactions
			UNION ALL
			SELECT * FROM related_holders
		) timeline
		WHERE ($5 = '' OR kind = $5)
		ORDER BY occurred_at DESC, item_order ASC
		LIMIT $3 OFFSET $4`, code, issuer, limit, normalizeOffset(offset), strings.TrimSpace(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReadTimelineItem, 0, limit)
	for rows.Next() {
		var item ReadTimelineItem
		if err := rows.Scan(
			&item.Kind,
			&item.Title,
			&item.Description,
			&item.Command,
			&item.OccurredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *PostgresStore) ListContractTimeline(ctx context.Context, contractID string, limit int, offset int, category string) ([]ReadTimelineItem, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		WITH related_transactions AS (
			SELECT DISTINCT ON (t.hash)
			       'tx'::text AS kind,
			       'Transaction ' || left(t.hash, 16) AS title,
			       'ledger ' || t.ledger_sequence::text || '  ops ' || t.operation_count::text AS description,
			       'lookup tx ' || t.hash AS command,
			       t.created_at AS occurred_at,
			       0 AS item_order
			FROM operations o
			JOIN transactions t ON t.hash = o.transaction_hash
			WHERE o.contract_id = $1
			ORDER BY t.hash, t.created_at DESC, t.application_order DESC
		),
		related_events AS (
			SELECT 'event'::text AS kind,
			       'Event type ' || type::text AS title,
			       'ledger ' || ledger_sequence::text || '  tx ' || left(transaction_hash, 16) AS description,
			       'lookup tx ' || transaction_hash AS command,
			       created_at AS occurred_at,
			       1 AS item_order
			FROM contract_events
			WHERE contract_id = $1
		)
		SELECT kind, title, description, command, occurred_at
		FROM (
			SELECT * FROM related_transactions
			UNION ALL
			SELECT * FROM related_events
		) timeline
		WHERE ($4 = '' OR kind = $4)
		ORDER BY occurred_at DESC, item_order ASC
		LIMIT $2 OFFSET $3`, contractID, limit, normalizeOffset(offset), strings.TrimSpace(category))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReadTimelineItem, 0, limit)
	for rows.Next() {
		var item ReadTimelineItem
		if err := rows.Scan(
			&item.Kind,
			&item.Title,
			&item.Description,
			&item.Command,
			&item.OccurredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *PostgresStore) GetAccountByID(ctx context.Context, id string) (*ReadAccountDetail, error) {
	var account ReadAccountDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT id, sequence, balance::text, buying_liabilities::text, selling_liabilities::text,
		       num_subentries, home_domain, flags, inflation_dest, thresholds::text,
		       last_modified_ledger, sponsor, num_sponsored, num_sponsoring,
		       data_entries::text, updated_at
		FROM accounts
		WHERE id = $1
		LIMIT 1`, id,
	).Scan(
		&account.ID,
		&account.Sequence,
		&account.Balance,
		&account.BuyingLiabilities,
		&account.SellingLiabilities,
		&account.NumSubentries,
		&account.HomeDomain,
		&account.Flags,
		&account.InflationDest,
		&account.Thresholds,
		&account.LastModifiedLedger,
		&account.Sponsor,
		&account.NumSponsored,
		&account.NumSponsoring,
		&account.DataEntries,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *PostgresStore) ListAccountDetails(ctx context.Context, limit int) ([]ReadAccountDetail, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, sequence, balance::text, buying_liabilities::text, selling_liabilities::text,
		       num_subentries, home_domain, flags, inflation_dest, thresholds::text,
		       last_modified_ledger, sponsor, num_sponsored, num_sponsoring,
		       data_entries::text, updated_at
		FROM accounts
		ORDER BY updated_at DESC, last_modified_ledger DESC, id ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]ReadAccountDetail, 0, limit)
	for rows.Next() {
		var account ReadAccountDetail
		if err := scanReadAccountDetail(rows, &account); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (s *PostgresStore) ListTrustlinesByAccountID(ctx context.Context, accountID string) ([]ReadTrustlineSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT asset_type, asset_code, asset_issuer, balance::text, limit_amount::text,
		       buying_liabilities::text, selling_liabilities::text, flags,
		       last_modified_ledger, sponsor, updated_at
		FROM trustlines
		WHERE account_id = $1
		ORDER BY balance DESC, asset_code ASC, asset_issuer ASC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trustlines := make([]ReadTrustlineSummary, 0)
	for rows.Next() {
		var trustline ReadTrustlineSummary
		if err := rows.Scan(
			&trustline.AssetType,
			&trustline.AssetCode,
			&trustline.AssetIssuer,
			&trustline.Balance,
			&trustline.LimitAmount,
			&trustline.BuyingLiabilities,
			&trustline.SellingLiabilities,
			&trustline.Flags,
			&trustline.LastModifiedLedger,
			&trustline.Sponsor,
			&trustline.UpdatedAt,
		); err != nil {
			return nil, err
		}
		trustlines = append(trustlines, trustline)
	}

	return trustlines, rows.Err()
}

func (s *PostgresStore) ListAssetHoldersByCodeIssuer(ctx context.Context, code string, issuer string, limit int, offset int) ([]ReadAssetHolderSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT account_id, balance::text, limit_amount::text, buying_liabilities::text,
		       selling_liabilities::text, last_modified_ledger, sponsor, updated_at
		FROM trustlines
		WHERE asset_code = $1 AND asset_issuer = $2
		ORDER BY balance DESC, updated_at DESC
		LIMIT $3 OFFSET $4`, code, issuer, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holders := make([]ReadAssetHolderSummary, 0, limit)
	for rows.Next() {
		var holder ReadAssetHolderSummary
		if err := rows.Scan(
			&holder.AccountID,
			&holder.Balance,
			&holder.LimitAmount,
			&holder.BuyingLiabilities,
			&holder.SellingLiabilities,
			&holder.LastModifiedLedger,
			&holder.Sponsor,
			&holder.UpdatedAt,
		); err != nil {
			return nil, err
		}
		holders = append(holders, holder)
	}

	return holders, rows.Err()
}

func (s *PostgresStore) ListAccountSignersByAccountID(ctx context.Context, accountID string) ([]ReadAccountSignerSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT signer_key, weight, type, sponsor, last_modified_ledger
		FROM account_signers
		WHERE account_id = $1
		ORDER BY weight DESC, signer_key ASC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	signers := make([]ReadAccountSignerSummary, 0)
	for rows.Next() {
		var signer ReadAccountSignerSummary
		if err := rows.Scan(
			&signer.SignerKey,
			&signer.Weight,
			&signer.Type,
			&signer.Sponsor,
			&signer.LastModifiedLedger,
		); err != nil {
			return nil, err
		}
		signers = append(signers, signer)
	}

	return signers, rows.Err()
}

func (s *PostgresStore) GetContractByID(ctx context.Context, contractID string) (*ReadContractDetail, error) {
	var contract ReadContractDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT contract_id, wasm_hash, creator_account, created_ledger, created_at,
		       last_modified_ledger, contract_type, is_sep41_token, is_sep50_nft,
		       token_name, token_symbol, token_decimals, contract_spec::text,
		       storage_entry_count, event_count, invocation_count, label, updated_at
		FROM contracts
		WHERE contract_id = $1
		LIMIT 1`, contractID,
	).Scan(
		&contract.ContractID,
		&contract.WasmHash,
		&contract.CreatorAccount,
		&contract.CreatedLedger,
		&contract.CreatedAt,
		&contract.LastModifiedLedger,
		&contract.ContractType,
		&contract.IsSep41Token,
		&contract.IsSep50NFT,
		&contract.TokenName,
		&contract.TokenSymbol,
		&contract.TokenDecimals,
		&contract.ContractSpec,
		&contract.StorageEntryCount,
		&contract.EventCount,
		&contract.InvocationCount,
		&contract.Label,
		&contract.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (s *PostgresStore) ListContractDetails(ctx context.Context, limit int) ([]ReadContractDetail, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT contract_id, wasm_hash, creator_account, created_ledger, created_at,
		       last_modified_ledger, contract_type, is_sep41_token, is_sep50_nft,
		       token_name, token_symbol, token_decimals, contract_spec::text,
		       storage_entry_count, event_count, invocation_count, label, updated_at
		FROM contracts
		ORDER BY updated_at DESC, invocation_count DESC, event_count DESC, contract_id ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contracts := make([]ReadContractDetail, 0, limit)
	for rows.Next() {
		var contract ReadContractDetail
		if err := scanReadContractDetail(rows, &contract); err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}

	return contracts, rows.Err()
}

func (s *PostgresStore) GetContractSpecByID(ctx context.Context, contractID string) (*ReadContractSpec, error) {
	var spec ReadContractSpec
	var raw *string
	err := s.db.QueryRowContext(ctx, `
		SELECT c.contract_id, c.wasm_hash, COALESCE(c.contract_spec::text, cc.spec_parsed::text),
		       cc.spec_xdr, c.updated_at
		FROM contracts c
		LEFT JOIN contract_code cc ON cc.wasm_hash = c.wasm_hash
		WHERE c.contract_id = $1
		LIMIT 1`, contractID,
	).Scan(
		&spec.ContractID,
		&spec.WasmHash,
		&raw,
		&spec.SpecXDR,
		&spec.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	spec.Raw = raw
	populateContractSpecSummary(&spec)
	return &spec, nil
}

func (s *PostgresStore) ListContractStorageByContractID(ctx context.Context, contractID string, limit int, offset int) ([]ReadContractStorageSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT contract_id, key_decoded::text, value_decoded::text, key_xdr, value_xdr,
		       durability, ttl_ledger, last_modified_ledger, updated_at
		FROM contract_storage
		WHERE contract_id = $1
		ORDER BY updated_at DESC, last_modified_ledger DESC, key_xdr ASC
		LIMIT $2 OFFSET $3`, contractID, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]ReadContractStorageSummary, 0, limit)
	for rows.Next() {
		var entry ReadContractStorageSummary
		if err := rows.Scan(
			&entry.ContractID,
			&entry.KeyDecoded,
			&entry.ValueDecoded,
			&entry.KeyXDR,
			&entry.ValueXDR,
			&entry.Durability,
			&entry.TTLLedger,
			&entry.LastModifiedLedger,
			&entry.UpdatedAt,
		); err != nil {
			return nil, err
		}
		populateContractStorageSummary(&entry)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (s *PostgresStore) ListContractEventSummariesByContractID(ctx context.Context, contractID string, limit int, offset int) ([]ReadContractEventSummary, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT contract_id, transaction_hash, ledger_sequence, type, topic_1, topic_2, topic_3, topic_4,
		       topics_xdr, value_xdr, topics_decoded::text, value_decoded::text, created_at
		FROM contract_events
		WHERE contract_id = $1
		ORDER BY created_at DESC, ledger_sequence DESC
		LIMIT $2 OFFSET $3`, contractID, limit, normalizeOffset(offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]ReadContractEventSummary, 0, limit)
	for rows.Next() {
		var event ReadContractEventSummary
		if err := rows.Scan(
			&event.ContractID,
			&event.TransactionHash,
			&event.LedgerSequence,
			&event.Type,
			&event.Topic1,
			&event.Topic2,
			&event.Topic3,
			&event.Topic4,
			&event.TopicsXDR,
			&event.ValueXDR,
			&event.TopicsDecoded,
			&event.ValueDecoded,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		populateContractEventSummary(&event)
		events = append(events, event)
	}
	return events, rows.Err()
}

func populateContractSpecSummary(spec *ReadContractSpec) {
	if spec == nil {
		return
	}
	spec.Functions = []ReadContractSpecFunction{}
	spec.Schemas = []ReadContractSpecSchema{}
	spec.Events = []ReadContractSpecEvent{}
	spec.DecodeStatus = "missing"
	spec.Available = false
	if spec.Raw == nil || strings.TrimSpace(*spec.Raw) == "" || strings.TrimSpace(*spec.Raw) == "null" {
		return
	}

	spec.Available = true
	spec.DecodeStatus = "raw"
	var entries []map[string]interface{}
	if err := json.Unmarshal([]byte(*spec.Raw), &entries); err != nil {
		return
	}

	for _, entry := range entries {
		kind := strings.ToLower(stringFromMap(entry, "kind"))
		if strings.Contains(kind, "function") {
			fn := ReadContractSpecFunction{
				Name:    stringFromMap(entry, "name"),
				Doc:     optionalStringFromMap(entry, "doc"),
				Inputs:  contractSpecInputs(entry["inputs"]),
				Outputs: contractSpecOutputs(entry["outputs"]),
			}
			spec.Functions = append(spec.Functions, fn)
			continue
		}
		if strings.Contains(kind, "event") {
			spec.Events = append(spec.Events, contractSpecEventFromMap(entry))
			continue
		}

		name := stringFromMap(entry, "name")
		if name == "" {
			continue
		}
		raw := marshalCompactJSON(entry)
		spec.Schemas = append(spec.Schemas, ReadContractSpecSchema{
			Kind: kind,
			Name: name,
			Raw:  raw,
		})
	}

	spec.FunctionCount = len(spec.Functions)
	spec.SchemaCount = len(spec.Schemas)
	spec.EventCount = len(spec.Events)
	spec.DecodeStatus = "decoded"
}

func populateContractEventSummary(event *ReadContractEventSummary) {
	if event == nil {
		return
	}
	event.Topics = compactStringPointers(event.Topic1, event.Topic2, event.Topic3, event.Topic4)
	applyDecodedEventPayload(event)
	switch {
	case len(event.FieldsDecoded) > 0:
		event.DecodeStatus = "decoded"
		if event.SpecDecodeStatus == "" {
			event.SpecDecodeStatus = "decoded"
		}
	case hasText(event.TopicsDecoded) && hasText(event.ValueDecoded):
		event.DecodeStatus = "decoded"
	case hasText(event.TopicsDecoded) || hasText(event.ValueDecoded):
		event.DecodeStatus = "partial"
	default:
		event.DecodeStatus = "raw"
	}
	event.Summary = contractEventSummaryText(*event)
}

func populateContractStorageSummary(entry *ReadContractStorageSummary) {
	if entry == nil {
		return
	}
	entry.DurabilityLabel = contractDurabilityLabel(entry.Durability)
	switch {
	case hasText(entry.KeyDecoded) && hasText(entry.ValueDecoded):
		entry.DecodeStatus = "decoded"
	case hasText(entry.KeyDecoded) || hasText(entry.ValueDecoded):
		entry.DecodeStatus = "partial"
	default:
		entry.DecodeStatus = "raw"
	}
	entry.DisplayKey = displayDecodedOrXDR(entry.KeyDecoded, entry.KeyXDR)
	entry.DisplayValue = displayDecodedOrXDR(entry.ValueDecoded, entry.ValueXDR)
	if entry.TTLLedger != nil {
		entry.ExpirationProximity = fmt.Sprintf("ttl ledger %d", *entry.TTLLedger)
	}
}

func contractDurabilityLabel(value int16) string {
	switch value {
	case 0:
		return "temporary"
	case 1:
		return "persistent"
	case 2:
		return "instance"
	default:
		return fmt.Sprintf("durability %d", value)
	}
}

func displayDecodedOrXDR(decoded *string, raw string) string {
	if hasText(decoded) {
		return truncateSearchText(strings.TrimSpace(*decoded), 80)
	}
	return truncateSearchText(strings.TrimSpace(raw), 80)
}

func contractSpecInputs(value interface{}) []ReadContractSpecValue {
	items, ok := value.([]interface{})
	if !ok {
		return []ReadContractSpecValue{}
	}
	inputs := make([]ReadContractSpecValue, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		inputs = append(inputs, ReadContractSpecValue{
			Name: stringFromMap(entry, "name"),
			Type: contractSpecTypeLabel(entry["type"]),
		})
	}
	return inputs
}

func contractSpecOutputs(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok {
		return []string{}
	}
	outputs := make([]string, 0, len(items))
	for _, item := range items {
		if text := contractSpecTypeLabel(item); text != "" {
			outputs = append(outputs, text)
		}
	}
	return outputs
}

func contractSpecTypeLabel(value interface{}) string {
	switch typed := value.(type) {
	case map[string]interface{}:
		if label := stringFromMap(typed, "type"); label != "" {
			return label
		}
		return strings.TrimSpace(fmt.Sprint(typed))
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func stringFromMap(entry map[string]interface{}, key string) string {
	value, ok := entry[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func optionalStringFromMap(entry map[string]interface{}, key string) *string {
	value := stringFromMap(entry, key)
	if value == "" {
		return nil
	}
	return &value
}

func marshalCompactJSON(value interface{}) *string {
	data, err := json.Marshal(value)
	if err != nil || len(data) == 0 {
		return nil
	}
	text := string(data)
	return &text
}

func compactStringPointers(values ...*string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			out = append(out, strings.TrimSpace(*value))
		}
	}
	return out
}

func hasText(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != "" && strings.TrimSpace(*value) != "null"
}

func contractEventSummaryText(event ReadContractEventSummary) string {
	parts := []string{fmt.Sprintf("ledger %d", event.LedgerSequence), fmt.Sprintf("type %d", event.Type)}
	if event.EventName != "" {
		parts = append(parts, "event "+event.EventName)
	} else if len(event.Topics) > 0 {
		parts = append(parts, "topic "+event.Topics[0])
	}
	if len(event.FieldsDecoded) > 0 {
		parts = append(parts, "fields "+fmt.Sprintf("%d", len(event.FieldsDecoded)))
	} else if hasText(event.ValueDecoded) {
		parts = append(parts, "value "+truncateSearchText(strings.TrimSpace(*event.ValueDecoded), 48))
	}
	if event.SpecDecodeStatus != "" {
		parts = append(parts, event.SpecDecodeStatus)
	} else {
		parts = append(parts, event.DecodeStatus)
	}
	return strings.Join(parts, "  ")
}

func contractSpecEventFromMap(entry map[string]interface{}) ReadContractSpecEvent {
	event := ReadContractSpecEvent{
		Name:       stringFromMap(entry, "name"),
		Doc:        optionalStringFromMap(entry, "doc"),
		DataFormat: stringFromMap(entry, "data_format"),
	}
	if rawTopics, ok := entry["prefix_topics"].([]interface{}); ok {
		for _, rawTopic := range rawTopics {
			if text := strings.TrimSpace(fmt.Sprint(rawTopic)); text != "" {
				event.PrefixTopics = append(event.PrefixTopics, text)
			}
		}
	}
	if rawParams, ok := entry["params"].([]interface{}); ok {
		for _, rawParam := range rawParams {
			paramMap, ok := rawParam.(map[string]interface{})
			if !ok {
				continue
			}
			event.Params = append(event.Params, ReadContractSpecValue{
				Name: stringFromMap(paramMap, "name"),
				Type: contractSpecTypeLabel(paramMap["type"]),
			})
		}
	}
	return event
}

func (s *PostgresStore) Search(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	results := make([]ReadSearchResult, 0, limit*8)
	if sequence, ok := parseSearchLedgerSequence(query); ok {
		ledgerResults, err := s.searchLedgers(ctx, sequence, limit)
		if err != nil {
			return nil, err
		}
		results = append(results, ledgerResults...)
	}

	txResults, err := s.searchTransactions(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, txResults...)

	accountResults, err := s.searchAccounts(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, accountResults...)

	assetResults, err := s.searchAssets(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, assetResults...)

	contractResults, err := s.searchContracts(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, contractResults...)

	operationResults, err := s.searchOperations(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, operationResults...)

	tokenEventResults, err := s.searchTokenEvents(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, tokenEventResults...)

	contractEventResults, err := s.searchContractEvents(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results = append(results, contractEventResults...)

	results = rankReadSearchResults(query, results)
	if len(results) > limit {
		return results[:limit], nil
	}
	return results, nil
}

func (s *PostgresStore) GetAssetByCodeIssuer(ctx context.Context, code string, issuer string) (*ReadAssetDetail, error) {
	var asset ReadAssetDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT asset_type, asset_code, asset_issuer, num_accounts, total_supply::text,
		       num_claimable_balances, num_liquidity_pools, num_contracts, flags,
		       auth_required, auth_revocable, auth_immutable, clawback_enabled,
		       home_domain, toml_name, toml_description, sac_contract_id, updated_at
		FROM assets
		WHERE asset_code = $1 AND asset_issuer = $2
		LIMIT 1`, code, issuer,
	).Scan(
		&asset.AssetType,
		&asset.AssetCode,
		&asset.AssetIssuer,
		&asset.NumAccounts,
		&asset.TotalSupply,
		&asset.NumClaimableBalances,
		&asset.NumLiquidityPools,
		&asset.NumContracts,
		&asset.Flags,
		&asset.AuthRequired,
		&asset.AuthRevocable,
		&asset.AuthImmutable,
		&asset.ClawbackEnabled,
		&asset.HomeDomain,
		&asset.TomlName,
		&asset.TomlDescription,
		&asset.SACContractID,
		&asset.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (s *PostgresStore) ListAssetDetails(ctx context.Context, limit int) ([]ReadAssetDetail, error) {
	limit = normalizeLimit(limit)
	rows, err := s.db.QueryContext(ctx, `
		SELECT asset_type, asset_code, asset_issuer, num_accounts, total_supply::text,
		       num_claimable_balances, num_liquidity_pools, num_contracts, flags,
		       auth_required, auth_revocable, auth_immutable, clawback_enabled,
		       home_domain, toml_name, toml_description, sac_contract_id, updated_at
		FROM assets
		ORDER BY updated_at DESC, num_accounts DESC, asset_code ASC, asset_issuer ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]ReadAssetDetail, 0, limit)
	for rows.Next() {
		var asset ReadAssetDetail
		if err := scanReadAssetDetail(rows, &asset); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, rows.Err()
}

func (s *PostgresStore) searchLedgers(ctx context.Context, sequence uint32, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT sequence, hash, transaction_count, operation_count
		FROM ledgers
		WHERE sequence = $1 OR sequence::text LIKE $2
		ORDER BY CASE WHEN sequence = $1 THEN 0 ELSE 1 END, sequence DESC
		LIMIT $3`, sequence, fmt.Sprintf("%d%%", sequence), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var sequence uint32
		var hash string
		var txCount int32
		var opCount int32
		if err := rows.Scan(&sequence, &hash, &txCount, &opCount); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "ledger",
			Title:       fmt.Sprintf("Ledger %d", sequence),
			Description: fmt.Sprintf("%d tx, %d ops, hash %s", txCount, opCount, truncateSearchText(hash, 12)),
			Command:     fmt.Sprintf("lookup ledger %d", sequence),
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchTransactions(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT hash, ledger_sequence, operation_count, is_soroban
		FROM transactions
		WHERE hash ILIKE $1
		ORDER BY ledger_sequence DESC, application_order DESC
		LIMIT $2`, query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var hash string
		var ledger uint32
		var opCount int32
		var isSoroban bool
		if err := rows.Scan(&hash, &ledger, &opCount, &isSoroban); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "transaction",
			Title:       "Transaction " + truncateSearchText(hash, 16),
			Description: fmt.Sprintf("ledger %d, %d ops, soroban=%t", ledger, opCount, isSoroban),
			Command:     "lookup tx " + hash,
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchAccounts(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, balance::text, num_subentries
		FROM accounts
		WHERE id ILIKE $1
		ORDER BY updated_at DESC
		LIMIT $2`, query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var id string
		var balance string
		var subentries int32
		if err := rows.Scan(&id, &balance, &subentries); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "account",
			Title:       "Account " + truncateSearchText(id, 16),
			Description: fmt.Sprintf("balance %s, %d subentries", balance, subentries),
			Command:     "lookup account " + id,
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchAssets(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT asset_code, asset_issuer, num_accounts, total_supply::text, COALESCE(home_domain, '')
		FROM assets
		WHERE asset_code ILIKE $1
		   OR asset_issuer ILIKE $2
		   OR home_domain ILIKE $2
		   OR toml_name ILIKE $2
		   OR sac_contract_id ILIKE $2
		ORDER BY
		   CASE WHEN asset_code ILIKE $1 THEN 0 ELSE 1 END,
		   num_accounts DESC,
		   updated_at DESC
		LIMIT $3`, query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var code string
		var issuer string
		var accounts int32
		var supply string
		var domain string
		if err := rows.Scan(&code, &issuer, &accounts, &supply, &domain); err != nil {
			return nil, err
		}
		description := fmt.Sprintf("%d accounts, supply %s", accounts, supply)
		if domain != "" {
			description = fmt.Sprintf("%s, %s", description, domain)
		}
		results = append(results, ReadSearchResult{
			Kind:        "asset",
			Title:       "Asset " + code,
			Description: description,
			Command:     fmt.Sprintf("lookup asset %s:%s", code, issuer),
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchContracts(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT contract_id, COALESCE(token_symbol, ''), invocation_count, event_count
		FROM contracts
		WHERE contract_id ILIKE $1 OR token_symbol ILIKE $2 OR token_name ILIKE $2
		ORDER BY updated_at DESC
		LIMIT $3`, query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var id string
		var symbol string
		var invocations int64
		var events int64
		if err := rows.Scan(&id, &symbol, &invocations, &events); err != nil {
			return nil, err
		}
		title := "Contract " + truncateSearchText(id, 16)
		if symbol != "" {
			title = fmt.Sprintf("Contract %s", symbol)
		}
		results = append(results, ReadSearchResult{
			Kind:        "contract",
			Title:       title,
			Description: fmt.Sprintf("%s, %d invocations, %d events", truncateSearchText(id, 12), invocations, events),
			Command:     "lookup contract " + id,
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchOperations(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT transaction_hash, application_order, type_name,
		       COALESCE(source_account, ''), COALESCE(destination, ''),
		       COALESCE(asset_code, ''), COALESCE(asset_issuer, ''),
		       COALESCE(amount::text, ''), COALESCE(contract_id, ''), COALESCE(function_name, '')
		FROM operations
		WHERE type_name ILIKE $1
		   OR function_name ILIKE $2
		   OR source_account ILIKE $2
		   OR destination ILIKE $2
		   OR asset_code ILIKE $1
		   OR asset_issuer ILIKE $2
		   OR contract_id ILIKE $2
		   OR transaction_hash ILIKE $2
		ORDER BY
		   CASE
		     WHEN type_name ILIKE $1 THEN 0
		     WHEN function_name ILIKE $1 THEN 1
		     WHEN asset_code ILIKE $1 THEN 2
		     ELSE 3
		   END,
		   created_at DESC,
		   application_order ASC
		LIMIT $3`, query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var txHash string
		var order int32
		var typeName string
		var source string
		var destination string
		var assetCode string
		var assetIssuer string
		var amount string
		var contractID string
		var functionName string
		if err := rows.Scan(&txHash, &order, &typeName, &source, &destination, &assetCode, &assetIssuer, &amount, &contractID, &functionName); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "operation",
			Title:       operationSearchTitle(typeName, functionName),
			Description: operationSearchDescription(txHash, order, source, destination, assetCode, assetIssuer, amount, contractID),
			Command:     "lookup tx " + txHash,
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchTokenEvents(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT event_type_name, COALESCE(from_address, ''), COALESCE(to_address, ''),
		       COALESCE(asset_code, ''), COALESCE(asset_issuer, ''), COALESCE(asset_contract_id, ''),
		       COALESCE(amount_formatted::text, amount::text), transaction_hash, ledger_sequence
		FROM token_events
		WHERE event_type_name ILIKE $1
		   OR from_address ILIKE $2
		   OR to_address ILIKE $2
		   OR asset_code ILIKE $1
		   OR asset_issuer ILIKE $2
		   OR asset_contract_id ILIKE $2
		   OR transaction_hash ILIKE $2
		ORDER BY
		   CASE
		     WHEN event_type_name ILIKE $1 THEN 0
		     WHEN asset_code ILIKE $1 THEN 1
		     ELSE 2
		   END,
		   created_at DESC
		LIMIT $3`, query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var eventType string
		var from string
		var to string
		var assetCode string
		var assetIssuer string
		var assetContractID string
		var amount string
		var txHash string
		var ledger uint32
		if err := rows.Scan(&eventType, &from, &to, &assetCode, &assetIssuer, &assetContractID, &amount, &txHash, &ledger); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "token-event",
			Title:       "Token " + valueOrSearchFallback(eventType, "event"),
			Description: tokenEventSearchDescription(ledger, from, to, assetCode, assetIssuer, assetContractID, amount),
			Command:     "lookup tx " + txHash,
		})
	}
	return results, rows.Err()
}

func (s *PostgresStore) searchContractEvents(ctx context.Context, query string, limit int) ([]ReadSearchResult, error) {
	if limit <= 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT contract_id, transaction_hash, ledger_sequence, type,
		       COALESCE(topic_1, ''), COALESCE(topic_2, ''), COALESCE(topic_3, ''), COALESCE(topic_4, '')
		FROM contract_events
		WHERE contract_id ILIKE $1
		   OR transaction_hash ILIKE $1
		   OR topic_1 ILIKE $2
		   OR topic_2 ILIKE $2
		   OR topic_3 ILIKE $2
		   OR topic_4 ILIKE $2
		   OR topics_decoded::text ILIKE $2
		   OR value_decoded::text ILIKE $2
		ORDER BY
		   CASE
		     WHEN contract_id ILIKE $1 THEN 0
		     WHEN transaction_hash ILIKE $1 THEN 1
		     WHEN topic_1 ILIKE $2 THEN 2
		     ELSE 3
		   END,
		   created_at DESC
		LIMIT $3`, query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ReadSearchResult, 0, limit)
	for rows.Next() {
		var contractID string
		var txHash string
		var ledger uint32
		var eventType int16
		var topic1 string
		var topic2 string
		var topic3 string
		var topic4 string
		if err := rows.Scan(&contractID, &txHash, &ledger, &eventType, &topic1, &topic2, &topic3, &topic4); err != nil {
			return nil, err
		}
		results = append(results, ReadSearchResult{
			Kind:        "contract-event",
			Title:       fmt.Sprintf("Contract Event %d", eventType),
			Description: contractEventSearchDescription(ledger, contractID, topic1, topic2, topic3, topic4),
			Command:     "lookup tx " + txHash,
		})
	}
	return results, rows.Err()
}

func rankReadSearchResults(query string, results []ReadSearchResult) []ReadSearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	ranked := append([]ReadSearchResult(nil), results...)
	sort.SliceStable(ranked, func(i, j int) bool {
		left := readSearchResultScore(query, ranked[i])
		right := readSearchResultScore(query, ranked[j])
		if left == right {
			return i < j
		}
		return left > right
	})
	return dedupeReadSearchResults(ranked)
}

func readSearchResultScore(query string, result ReadSearchResult) int {
	score := 0
	kind := strings.ToLower(strings.TrimSpace(result.Kind))
	title := strings.ToLower(strings.TrimSpace(result.Title))
	description := strings.ToLower(strings.TrimSpace(result.Description))
	command := strings.ToLower(strings.TrimSpace(result.Command))
	target := readSearchCommandTarget(command)

	switch kind {
	case "ledger":
		score += 10
	case "transaction":
		score += 8
	case "operation":
		score += 8
	case "token-event":
		score += 7
	case "contract-event":
		score += 7
	case "account":
		score += 7
	case "asset":
		score += 6
	case "contract":
		score += 6
	}
	if query == "" {
		return score
	}

	if target == query {
		score += 140
	} else if strings.HasPrefix(target, query) {
		score += 90
	} else if strings.Contains(target, query) {
		score += 45
	}
	if title == query || strings.TrimPrefix(title, kind+" ") == query {
		score += 100
	} else if strings.HasPrefix(title, query) || strings.HasPrefix(strings.TrimPrefix(title, kind+" "), query) {
		score += 70
	} else if strings.Contains(title, query) {
		score += 35
	}
	if strings.Contains(description, query) {
		score += 15
	}
	return score
}

func readSearchCommandTarget(command string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(command)))
	if len(fields) < 3 || fields[0] != "lookup" {
		return ""
	}
	return strings.Join(fields[2:], " ")
}

func dedupeReadSearchResults(results []ReadSearchResult) []ReadSearchResult {
	seen := make(map[string]struct{}, len(results))
	deduped := make([]ReadSearchResult, 0, len(results))
	for _, result := range results {
		key := strings.TrimSpace(result.Command)
		if key == "" {
			key = strings.TrimSpace(result.Kind) + ":" + strings.TrimSpace(result.Title) + ":" + strings.TrimSpace(result.Description)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, result)
	}
	return deduped
}

func operationSearchTitle(typeName string, functionName string) string {
	typeName = valueOrSearchFallback(typeName, "operation")
	functionName = strings.TrimSpace(functionName)
	if functionName != "" {
		return fmt.Sprintf("Operation %s:%s", typeName, functionName)
	}
	return "Operation " + typeName
}

func operationSearchDescription(txHash string, order int32, source string, destination string, assetCode string, assetIssuer string, amount string, contractID string) string {
	parts := []string{
		fmt.Sprintf("tx %s", truncateSearchText(txHash, 12)),
		fmt.Sprintf("op %d", order),
	}
	if assetCode != "" && assetIssuer != "" {
		parts = append(parts, fmt.Sprintf("asset %s:%s", assetCode, truncateSearchText(assetIssuer, 8)))
	}
	if amount != "" {
		parts = append(parts, "amount "+amount)
	}
	if destination != "" {
		parts = append(parts, "to "+truncateSearchText(destination, 12))
	} else if contractID != "" {
		parts = append(parts, "contract "+truncateSearchText(contractID, 12))
	} else if source != "" {
		parts = append(parts, "source "+truncateSearchText(source, 12))
	}
	return strings.Join(parts, ", ")
}

func tokenEventSearchDescription(ledger uint32, from string, to string, assetCode string, assetIssuer string, assetContractID string, amount string) string {
	parts := []string{fmt.Sprintf("ledger %d", ledger)}
	if assetCode != "" && assetIssuer != "" {
		parts = append(parts, fmt.Sprintf("asset %s:%s", assetCode, truncateSearchText(assetIssuer, 8)))
	} else if assetContractID != "" {
		parts = append(parts, "asset contract "+truncateSearchText(assetContractID, 12))
	}
	if amount != "" {
		parts = append(parts, "amount "+amount)
	}
	if from != "" {
		parts = append(parts, "from "+truncateSearchText(from, 12))
	}
	if to != "" {
		parts = append(parts, "to "+truncateSearchText(to, 12))
	}
	return strings.Join(parts, ", ")
}

func contractEventSearchDescription(ledger uint32, contractID string, topics ...string) string {
	parts := []string{
		fmt.Sprintf("ledger %d", ledger),
		"contract " + truncateSearchText(contractID, 12),
	}
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic == "" {
			continue
		}
		parts = append(parts, "topic "+truncateSearchText(topic, 16))
		break
	}
	return strings.Join(parts, ", ")
}

func valueOrSearchFallback(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func parseSearchLedgerSequence(query string) (uint32, bool) {
	if query == "" {
		return 0, false
	}
	sequence, err := strconv.ParseUint(query, 10, 32)
	return uint32(sequence), err == nil && sequence > 0
}

func scanReadTransactionSummary(scanner interface {
	Scan(dest ...interface{}) error
}, tx *ReadTransactionSummary) error {
	return scanner.Scan(
		&tx.Hash,
		&tx.LedgerSequence,
		&tx.ApplicationOrder,
		&tx.Account,
		&tx.OperationCount,
		&tx.Status,
		&tx.IsSoroban,
		&tx.CreatedAt,
		&tx.PrimaryContractID,
		&tx.PrimaryAssetCode,
		&tx.PrimaryAssetIssuer,
		&tx.PrimaryOperationType,
	)
}

func scanReadAccountDetail(scanner interface {
	Scan(dest ...interface{}) error
}, account *ReadAccountDetail) error {
	return scanner.Scan(
		&account.ID,
		&account.Sequence,
		&account.Balance,
		&account.BuyingLiabilities,
		&account.SellingLiabilities,
		&account.NumSubentries,
		&account.HomeDomain,
		&account.Flags,
		&account.InflationDest,
		&account.Thresholds,
		&account.LastModifiedLedger,
		&account.Sponsor,
		&account.NumSponsored,
		&account.NumSponsoring,
		&account.DataEntries,
		&account.UpdatedAt,
	)
}

func scanReadContractDetail(scanner interface {
	Scan(dest ...interface{}) error
}, contract *ReadContractDetail) error {
	return scanner.Scan(
		&contract.ContractID,
		&contract.WasmHash,
		&contract.CreatorAccount,
		&contract.CreatedLedger,
		&contract.CreatedAt,
		&contract.LastModifiedLedger,
		&contract.ContractType,
		&contract.IsSep41Token,
		&contract.IsSep50NFT,
		&contract.TokenName,
		&contract.TokenSymbol,
		&contract.TokenDecimals,
		&contract.ContractSpec,
		&contract.StorageEntryCount,
		&contract.EventCount,
		&contract.InvocationCount,
		&contract.Label,
		&contract.UpdatedAt,
	)
}

func scanReadAssetDetail(scanner interface {
	Scan(dest ...interface{}) error
}, asset *ReadAssetDetail) error {
	return scanner.Scan(
		&asset.AssetType,
		&asset.AssetCode,
		&asset.AssetIssuer,
		&asset.NumAccounts,
		&asset.TotalSupply,
		&asset.NumClaimableBalances,
		&asset.NumLiquidityPools,
		&asset.NumContracts,
		&asset.Flags,
		&asset.AuthRequired,
		&asset.AuthRevocable,
		&asset.AuthImmutable,
		&asset.ClawbackEnabled,
		&asset.HomeDomain,
		&asset.TomlName,
		&asset.TomlDescription,
		&asset.SACContractID,
		&asset.UpdatedAt,
	)
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 10
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func trimAndSortTransactionSummaries(summaries []ReadTransactionSummary, limit int) []ReadTransactionSummary {
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].LedgerSequence != summaries[j].LedgerSequence {
			return summaries[i].LedgerSequence > summaries[j].LedgerSequence
		}
		if !summaries[i].CreatedAt.Equal(summaries[j].CreatedAt) {
			return summaries[i].CreatedAt.After(summaries[j].CreatedAt)
		}
		return summaries[i].ApplicationOrder > summaries[j].ApplicationOrder
	})
	if normalized := normalizeLimit(limit); len(summaries) > normalized {
		return summaries[:normalized]
	}
	return summaries
}

func truncateSearchText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func (s *PostgresStore) DB() *sql.DB {
	return s.db
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
