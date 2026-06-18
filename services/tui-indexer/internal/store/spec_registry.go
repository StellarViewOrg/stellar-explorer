package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
)

// GetSpecRegistryForContract loads a contract spec registry from indexed contract_code XDR.
func (s *PostgresStore) GetSpecRegistryForContract(ctx context.Context, contractID string) (*sordecode.SpecRegistry, error) {
	contractID = strings.TrimSpace(contractID)
	if contractID == "" {
		return nil, nil
	}
	var specXDR sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT cc.spec_xdr
		FROM contracts c
		LEFT JOIN contract_code cc ON cc.wasm_hash = c.wasm_hash
		WHERE c.contract_id = $1
		LIMIT 1`, contractID,
	).Scan(&specXDR)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !specXDR.Valid || strings.TrimSpace(specXDR.String) == "" {
		return nil, nil
	}
	return sordecode.RegistryFromSpecXDR(specXDR.String)
}

// RepopulateContractEventSummary rebuilds read API event summaries after read-time spec decoding.
func RepopulateContractEventSummary(event *ReadContractEventSummary) {
	populateContractEventSummary(event)
}

// ApplyDecodedEventPayload parses enriched value_decoded JSON into event summary fields.
func ApplyDecodedEventPayload(event *ReadContractEventSummary) {
	applyDecodedEventPayload(event)
}

func applyDecodedEventPayload(event *ReadContractEventSummary) {
	if event == nil || !hasText(event.ValueDecoded) {
		return
	}
	raw := strings.TrimSpace(*event.ValueDecoded)
	if !strings.HasPrefix(raw, "{") {
		return
	}
	var payload struct {
		Text             string                   `json:"text"`
		EventName        string                   `json:"event_name"`
		Fields           []ReadContractEventField `json:"fields"`
		SpecDecodeStatus string                   `json:"spec_decode_status"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return
	}
	if payload.EventName != "" {
		event.EventName = payload.EventName
	}
	if len(payload.Fields) > 0 {
		event.FieldsDecoded = payload.Fields
	}
	if payload.SpecDecodeStatus != "" {
		event.SpecDecodeStatus = payload.SpecDecodeStatus
	}
}
