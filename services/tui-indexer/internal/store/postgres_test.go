package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func getTestDB(t *testing.T) *PostgresStore {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgresql://explorer:explorer_dev@localhost:54320/stellar_explorer?sslmode=disable"
	}
	store, err := NewPostgresStore(url)
	if err != nil {
		t.Skipf("Skipping: cannot connect to test database: %v", err)
	}
	return store
}

func TestInsertLedger(t *testing.T) {
	store := getTestDB(t)
	defer store.Close()

	now := time.Now().UTC().Truncate(time.Microsecond)
	ledger := &Ledger{
		Sequence:          99999999,
		Hash:              "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PrevHash:          "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ClosedAt:          now,
		TotalCoins:        1000000000000,
		FeePool:           100000,
		BaseFee:           100,
		BaseReserve:       5000000,
		MaxTxSetSize:      1000,
		ProtocolVersion:   21,
		TransactionCount:  5,
		OperationCount:    10,
		SuccessfulTxCount: 4,
		FailedTxCount:     1,
	}

	ctx := context.Background()

	err := store.InsertLedger(ctx, ledger)
	if err != nil {
		t.Fatalf("InsertLedger failed: %v", err)
	}

	// Insert again — should be idempotent (ON CONFLICT DO NOTHING)
	err = store.InsertLedger(ctx, ledger)
	if err != nil {
		t.Fatalf("Idempotent InsertLedger failed: %v", err)
	}

	// Clean up test data
	_, _ = store.db.ExecContext(ctx, "DELETE FROM ledgers WHERE sequence = 99999999")
}

func TestInsertTransactionBatch(t *testing.T) {
	store := getTestDB(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	txs := []Transaction{
		{
			Hash:             "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
			LedgerSequence:   88888888,
			ApplicationOrder: 1,
			Account:          "GABC",
			AccountSequence:  100,
			FeeCharged:       100,
			MaxFee:           200,
			OperationCount:   1,
			MemoType:         0,
			Status:           1,
			IsSoroban:        false,
			EnvelopeXDR:      "AAAA",
			ResultXDR:        "BBBB",
			CreatedAt:        now,
		},
	}

	err := store.InsertTransactionBatch(ctx, txs)
	if err != nil {
		t.Fatalf("InsertTransactionBatch failed: %v", err)
	}

	// Insert again — idempotent
	err = store.InsertTransactionBatch(ctx, txs)
	if err != nil {
		t.Fatalf("Idempotent InsertTransactionBatch failed: %v", err)
	}

	// Empty batch should be no-op
	err = store.InsertTransactionBatch(ctx, nil)
	if err != nil {
		t.Fatalf("Empty InsertTransactionBatch failed: %v", err)
	}

	// Clean up
	_, _ = store.db.ExecContext(ctx, "DELETE FROM transactions WHERE hash = 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc'")
}

func TestInsertOperationBatch(t *testing.T) {
	store := getTestDB(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	typeName := "payment"

	ops := []Operation{
		{
			TransactionID:    0,
			TransactionHash:  "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
			ApplicationOrder: 1,
			Type:             1,
			TypeName:         typeName,
			Details:          `{"type": "payment"}`,
			CreatedAt:        now,
		},
	}

	err := store.InsertOperationBatch(ctx, ops)
	if err != nil {
		t.Fatalf("InsertOperationBatch failed: %v", err)
	}

	// Clean up
	_, _ = store.db.ExecContext(ctx, "DELETE FROM operations WHERE transaction_hash = 'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd'")
}

func TestIngestionState(t *testing.T) {
	store := getTestDB(t)
	defer store.Close()

	ctx := context.Background()

	// Clean up first to ensure a fresh state
	_, _ = store.db.ExecContext(ctx, "DELETE FROM ingestion_state WHERE key = 'last_ingested_ledger'")

	err := store.SetLastIngestedLedger(ctx, 12345)
	if err != nil {
		t.Fatalf("SetLastIngestedLedger failed: %v", err)
	}

	seq, err := store.GetLastIngestedLedger(ctx)
	if err != nil {
		t.Fatalf("GetLastIngestedLedger failed: %v", err)
	}
	if seq != 12345 {
		t.Errorf("expected 12345, got %d", seq)
	}

	// Verify forward-only: setting a lower value should not regress
	err = store.SetLastIngestedLedger(ctx, 100)
	if err != nil {
		t.Fatalf("SetLastIngestedLedger (lower) failed: %v", err)
	}
	seq, err = store.GetLastIngestedLedger(ctx)
	if err != nil {
		t.Fatalf("GetLastIngestedLedger after lower set failed: %v", err)
	}
	if seq != 12345 {
		t.Errorf("cursor should not regress: expected 12345, got %d", seq)
	}

	// Clean up
	_, _ = store.db.ExecContext(ctx, "DELETE FROM ingestion_state WHERE key = 'last_ingested_ledger'")
}

func TestExplorerReadQueries(t *testing.T) {
	store := getTestDB(t)
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	hash := "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	accountID := "GACCOUNTREADQUERYTEST"
	issuer := "GISSUERREADQUERYTEST"
	contractID := "CCONTRACTREADQUERYTEST"

	_, _ = store.db.ExecContext(ctx, "DELETE FROM contract_events WHERE contract_id = $1", contractID)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM operations WHERE transaction_hash = $1", hash)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM transactions WHERE hash = $1", hash)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM account_signers WHERE account_id = $1", accountID)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM trustlines WHERE account_id = $1", accountID)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1", accountID)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM assets WHERE asset_code = 'USDC' AND asset_issuer = $1", issuer)
	_, _ = store.db.ExecContext(ctx, "DELETE FROM contracts WHERE contract_id = $1", contractID)
	defer func() {
		_, _ = store.db.ExecContext(ctx, "DELETE FROM contract_events WHERE contract_id = $1", contractID)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM operations WHERE transaction_hash = $1", hash)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM transactions WHERE hash = $1", hash)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM account_signers WHERE account_id = $1", accountID)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM trustlines WHERE account_id = $1", accountID)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1", accountID)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM assets WHERE asset_code = 'USDC' AND asset_issuer = $1", issuer)
		_, _ = store.db.ExecContext(ctx, "DELETE FROM contracts WHERE contract_id = $1", contractID)
	}()

	_, err := store.db.ExecContext(ctx, `
		INSERT INTO accounts (
			id, sequence, balance, buying_liabilities, selling_liabilities, num_subentries,
			flags, last_modified_ledger, num_sponsored, num_sponsoring, updated_at
		) VALUES ($1, 99, 12.5, 0, 0, 1, 0, 70000001, 0, 0, $2)`,
		accountID, now,
	)
	if err != nil {
		t.Fatalf("insert account: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO trustlines (
			account_id, asset_type, asset_code, asset_issuer, balance, limit_amount,
			buying_liabilities, selling_liabilities, flags, last_modified_ledger, updated_at
		) VALUES ($1, 1, 'USDC', $2, 100, 1000, 0, 0, 0, 70000001, $3)`,
		accountID, issuer, now,
	)
	if err != nil {
		t.Fatalf("insert trustline: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO account_signers (account_id, signer_key, weight, type, last_modified_ledger)
		VALUES ($1, 'GSIGNERREADQUERYTEST', 1, 'ed25519_public_key', 70000001)`,
		accountID,
	)
	if err != nil {
		t.Fatalf("insert signer: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO assets (
			asset_type, asset_code, asset_issuer, num_accounts, total_supply, flags,
			auth_required, auth_revocable, auth_immutable, clawback_enabled, updated_at
		) VALUES (1, 'USDC', $1, 1, 1000, 0, false, false, false, false, $2)`,
		issuer, now,
	)
	if err != nil {
		t.Fatalf("insert asset: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO contracts (
			contract_id, created_ledger, created_at, last_modified_ledger, contract_type,
			is_sep41_token, is_sep50_nft, storage_entry_count, event_count, invocation_count, updated_at
		) VALUES ($1, 70000001, $2, 70000001, 0, false, false, 1, 1, 1, $2)`,
		contractID, now,
	)
	if err != nil {
		t.Fatalf("insert contract: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO transactions (
			hash, ledger_sequence, application_order, account, account_sequence, fee_charged,
			max_fee, operation_count, memo_type, status, is_soroban, envelope_xdr, result_xdr, created_at
		) VALUES ($1, 70000001, 1, $2, 1, 100, 100, 1, 0, 1, false, 'AAAA', 'BBBB', $3)`,
		hash, accountID, now,
	)
	if err != nil {
		t.Fatalf("insert transaction: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO operations (
			transaction_id, transaction_hash, application_order, type, type_name,
			source_account, asset_code, asset_issuer, amount, destination, contract_id, details, created_at
		) VALUES (1, $1, 1, 1, 'payment', $2, 'USDC', $3, 10, 'GDESTREADQUERYTEST', $4, '{}'::jsonb, $5)`,
		hash, accountID, issuer, contractID, now,
	)
	if err != nil {
		t.Fatalf("insert operation: %v", err)
	}

	_, err = store.db.ExecContext(ctx, `
		INSERT INTO contract_events (
			contract_id, transaction_hash, ledger_sequence, type, topic_1, topics_xdr, value_xdr, created_at
		) VALUES ($1, $2, 70000001, 0, 'transfer', 'AAAA', 'BBBB', $3)`,
		contractID, hash, now,
	)
	if err != nil {
		t.Fatalf("insert contract event: %v", err)
	}

	accountTxs, err := store.ListTransactionSummariesByAccount(ctx, accountID, 5)
	if err != nil || len(accountTxs) != 1 || accountTxs[0].Hash != hash {
		t.Fatalf("ListTransactionSummariesByAccount() = %#v, %v", accountTxs, err)
	}

	accountOps, err := store.ListOperationSummariesByAccount(ctx, accountID, 5, 0)
	if err != nil || len(accountOps) != 1 || accountOps[0].TypeName != "payment" {
		t.Fatalf("ListOperationSummariesByAccount() = %#v, %v", accountOps, err)
	}

	accountTimeline, err := store.ListAccountTimeline(ctx, accountID, 5, 0, "")
	if err != nil || len(accountTimeline) < 2 || accountTimeline[0].Command == "" {
		t.Fatalf("ListAccountTimeline() = %#v, %v", accountTimeline, err)
	}

	holders, err := store.ListAssetHoldersByCodeIssuer(ctx, "USDC", issuer, 5, 0)
	if err != nil || len(holders) != 1 || holders[0].AccountID != accountID {
		t.Fatalf("ListAssetHoldersByCodeIssuer() = %#v, %v", holders, err)
	}

	assetTxs, err := store.ListTransactionSummariesByAsset(ctx, "USDC", issuer, 5)
	if err != nil || len(assetTxs) != 1 || assetTxs[0].Hash != hash {
		t.Fatalf("ListTransactionSummariesByAsset() = %#v, %v", assetTxs, err)
	}

	assetTimeline, err := store.ListAssetTimeline(ctx, "USDC", issuer, 5, 0, "")
	if err != nil || len(assetTimeline) < 2 || assetTimeline[0].Command == "" {
		t.Fatalf("ListAssetTimeline() = %#v, %v", assetTimeline, err)
	}

	contractTxs, err := store.ListTransactionSummariesByContract(ctx, contractID, 5)
	if err != nil || len(contractTxs) != 1 || contractTxs[0].Hash != hash {
		t.Fatalf("ListTransactionSummariesByContract() = %#v, %v", contractTxs, err)
	}

	events, err := store.ListContractEventSummariesByContractID(ctx, contractID, 5, 0)
	if err != nil || len(events) != 1 || events[0].TransactionHash != hash {
		t.Fatalf("ListContractEventSummariesByContractID() = %#v, %v", events, err)
	}

	contractTimeline, err := store.ListContractTimeline(ctx, contractID, 5, 0, "")
	if err != nil || len(contractTimeline) < 2 || contractTimeline[0].Command == "" {
		t.Fatalf("ListContractTimeline() = %#v, %v", contractTimeline, err)
	}
}

func TestRankReadSearchResultsIncludesActivityTypes(t *testing.T) {
	results := rankReadSearchResults("payment", []ReadSearchResult{
		{
			Kind:        "contract",
			Title:       "Contract helper",
			Description: "mentions payment",
			Command:     "lookup contract COTHER",
		},
		{
			Kind:        "operation",
			Title:       "Operation payment:transfer",
			Description: "tx abc",
			Command:     "lookup tx abc",
		},
		{
			Kind:        "contract-event",
			Title:       "Contract Event 0",
			Description: "topic transfer",
			Command:     "lookup tx def",
		},
	})

	if len(results) != 3 {
		t.Fatalf("expected three results, got %#v", results)
	}
	if results[0].Kind != "operation" {
		t.Fatalf("expected operation search result first, got %#v", results)
	}
}

func TestPopulateContractSpecSummary(t *testing.T) {
	raw := `[{"kind":"SC_SPEC_ENTRY_FUNCTION_V0","name":"balance","doc":"Get balance","inputs":[{"name":"id","type":"SC_SPEC_TYPE_ADDRESS"}],"outputs":["SC_SPEC_TYPE_I128"]},{"kind":"SC_SPEC_ENTRY_UDT_STRUCT_V0","name":"Allowance"}]`
	spec := ReadContractSpec{ContractID: "CCONTRACT", Raw: &raw}

	populateContractSpecSummary(&spec)

	if !spec.Available || spec.DecodeStatus != "decoded" {
		t.Fatalf("expected decoded spec, got available=%v status=%q", spec.Available, spec.DecodeStatus)
	}
	if spec.FunctionCount != 1 || spec.SchemaCount != 1 {
		t.Fatalf("unexpected counts: functions=%d schemas=%d", spec.FunctionCount, spec.SchemaCount)
	}
	if got := spec.Functions[0].Inputs[0].Name; got != "id" {
		t.Fatalf("unexpected input name: %q", got)
	}
}

func TestPopulateContractEventSummary(t *testing.T) {
	topic := "transfer"
	value := `{"amount":"100"}`
	event := ReadContractEventSummary{
		TransactionHash: "tx",
		LedgerSequence:  42,
		Type:            0,
		Topic1:          &topic,
		ValueDecoded:    &value,
	}

	populateContractEventSummary(&event)

	if event.DecodeStatus != "partial" {
		t.Fatalf("expected partial decode status, got %q", event.DecodeStatus)
	}
	if len(event.Topics) != 1 || event.Topics[0] != "transfer" {
		t.Fatalf("unexpected topics: %#v", event.Topics)
	}
	if event.Summary == "" {
		t.Fatalf("expected event summary")
	}
}

func TestPopulateContractEventSummary_SpecEnrichedPayload(t *testing.T) {
	topic := "transfer"
	value := `{"text":"100","event_name":"transfer","fields":[{"name":"amount","type":"i128","location":"value","value":"100"}],"spec_decode_status":"decoded"}`
	event := ReadContractEventSummary{
		TransactionHash: "tx",
		LedgerSequence:  42,
		Type:            1,
		Topic1:          &topic,
		ValueDecoded:    &value,
	}

	populateContractEventSummary(&event)

	if event.DecodeStatus != "decoded" {
		t.Fatalf("expected decoded status, got %q", event.DecodeStatus)
	}
	if event.EventName != "transfer" {
		t.Fatalf("unexpected event name: %q", event.EventName)
	}
	if len(event.FieldsDecoded) != 1 || event.FieldsDecoded[0].Name != "amount" {
		t.Fatalf("unexpected fields: %#v", event.FieldsDecoded)
	}
	if event.SpecDecodeStatus != "decoded" {
		t.Fatalf("unexpected spec decode status: %q", event.SpecDecodeStatus)
	}
}

func TestPopulateContractStorageSummary(t *testing.T) {
	key := `{"symbol":"DataKey"}`
	entry := ReadContractStorageSummary{
		KeyDecoded:         &key,
		KeyXDR:             "raw-key",
		ValueXDR:           "raw-value",
		Durability:         1,
		LastModifiedLedger: 77,
	}

	populateContractStorageSummary(&entry)

	if entry.DurabilityLabel != "persistent" {
		t.Fatalf("unexpected durability label: %q", entry.DurabilityLabel)
	}
	if entry.DecodeStatus != "partial" {
		t.Fatalf("unexpected decode status: %q", entry.DecodeStatus)
	}
	if entry.DisplayKey == "" || entry.DisplayValue == "" {
		t.Fatalf("expected display fields, got key=%q value=%q", entry.DisplayKey, entry.DisplayValue)
	}
}
