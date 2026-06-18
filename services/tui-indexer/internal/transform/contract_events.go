package transform

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

// ContractEventsFromTransaction extracts all contract events from a transaction's
// ResultMetaXDR (base64-encoded TransactionMeta). Handles V3 and V4 meta formats.
func ContractEventsFromTransaction(entry source.TransactionEntry, networkPassphrase string) ([]store.ContractEvent, error) {
	if entry.ResultMetaXDR == "" {
		return nil, nil
	}

	var meta xdr.TransactionMeta
	if err := xdr.SafeUnmarshalBase64(entry.ResultMetaXDR, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal transaction meta: %w", err)
	}

	var rawEvents []xdr.ContractEvent
	switch meta.V {
	case 3:
		if meta.V3 != nil && meta.V3.SorobanMeta != nil {
			rawEvents = meta.V3.SorobanMeta.Events
		}
	case 4:
		if meta.V4 != nil {
			for _, op := range meta.V4.Operations {
				rawEvents = append(rawEvents, op.Events...)
			}
		}
	default:
		// V0, V1, V2 don't have Soroban contract events
		return nil, nil
	}

	if len(rawEvents) == 0 {
		return nil, nil
	}

	txHash, err := computeTxHash(entry.EnvelopeXDR, networkPassphrase)
	if err != nil {
		return nil, fmt.Errorf("compute tx hash: %w", err)
	}

	result := make([]store.ContractEvent, 0, len(rawEvents))
	for _, ce := range rawEvents {
		event, err := contractEventFromXDR(ce, txHash, entry.Ledger, entry.CreatedAt)
		if err != nil {
			// Skip events we can't parse but don't fail the whole transaction
			continue
		}
		result = append(result, *event)
	}
	return result, nil
}

func contractEventFromXDR(ce xdr.ContractEvent, txHash string, ledgerSeq uint32, unixTime int64) (*store.ContractEvent, error) {
	if ce.ContractId == nil {
		return nil, fmt.Errorf("contract event has no contract ID")
	}

	contractID, err := strkey.Encode(strkey.VersionByteContract, ce.ContractId[:])
	if err != nil {
		return nil, fmt.Errorf("encode contract ID: %w", err)
	}

	if ce.Body.V != 0 {
		return nil, fmt.Errorf("unsupported contract event body version: %d", ce.Body.V)
	}

	body := ce.Body.V0
	topics := body.Topics
	value := body.Data

	// Encode topics as a JSON array of base64 XDR strings
	eb := xdr.NewEncodingBuffer()
	topicXDRs := make([]string, len(topics))
	for i, t := range topics {
		b, err := eb.MarshalBinary(t)
		if err == nil {
			topicXDRs[i] = base64.StdEncoding.EncodeToString(b)
		}
	}
	topicsXDRJSON, _ := json.Marshal(topicXDRs)

	// Encode value XDR
	valueB, err := eb.MarshalBinary(value)
	if err != nil {
		return nil, fmt.Errorf("marshal event value: %w", err)
	}
	valueXDR := base64.StdEncoding.EncodeToString(valueB)

	// Decode topics to human-readable strings for topic_1..4 columns
	decodedTopics := make([]string, len(topics))
	for i, t := range topics {
		decodedTopics[i] = scValToString(t)
	}
	decodedTopicsJSON, _ := json.Marshal(decodedTopics)
	decodedTopicsStr := string(decodedTopicsJSON)

	// Decode value to a JSON string
	decodedValueStr := fmt.Sprintf("%q", scValToString(value))

	event := &store.ContractEvent{
		ContractID:      contractID,
		TransactionHash: txHash,
		LedgerSequence:  ledgerSeq,
		Type:            int16(ce.Type), // XDR enum: System=0, Contract=1, Diagnostic=2
		TopicsXDR:       string(topicsXDRJSON),
		ValueXDR:        valueXDR,
		TopicsDecoded:   &decodedTopicsStr,
		ValueDecoded:    &decodedValueStr,
		CreatedAt:       time.Unix(unixTime, 0).UTC(),
	}

	// Populate indexed topic columns (up to 4)
	if len(decodedTopics) > 0 {
		s := decodedTopics[0]
		event.Topic1 = &s
	}
	if len(decodedTopics) > 1 {
		s := decodedTopics[1]
		event.Topic2 = &s
	}
	if len(decodedTopics) > 2 {
		s := decodedTopics[2]
		event.Topic3 = &s
	}
	if len(decodedTopics) > 3 {
		s := decodedTopics[3]
		event.Topic4 = &s
	}

	return event, nil
}

// computeTxHash parses the envelope XDR and computes the transaction hash.
func computeTxHash(envelopeXDR string, networkPassphrase string) (string, error) {
	var envelope xdr.TransactionEnvelope
	if err := xdr.SafeUnmarshalBase64(envelopeXDR, &envelope); err != nil {
		return "", fmt.Errorf("unmarshal envelope: %w", err)
	}
	return computeTransactionHash(envelope, networkPassphrase)
}

// scValToString converts an XDR ScVal to a human-readable string for indexing.
func scValToString(v xdr.ScVal) string {
	return sordecode.ScValToString(v)
}
