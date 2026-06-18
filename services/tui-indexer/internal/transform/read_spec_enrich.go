package transform

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

// OperationNeedsSpecReEnrich reports whether a read-time operation should be spec-decoded again.
func OperationNeedsSpecReEnrich(op store.ReadOperationSummary) bool {
	if strings.TrimSpace(op.TypeName) != "invoke_host_function" {
		return false
	}
	details, err := parseOperationDetailsMap(op.Details)
	if err != nil {
		return true
	}
	status := stringFromDetails(details, "spec_decode_status")
	if status == specDecodeStatusDecoded {
		return false
	}
	return status == "" || status == specDecodeStatusNoSpec || status == specDecodeStatusPartial
}

// EventNeedsSpecReEnrich reports whether a read-time contract event should be spec-decoded again.
func EventNeedsSpecReEnrich(event store.ReadContractEventSummary) bool {
	if event.SpecDecodeStatus == specDecodeStatusDecoded {
		return false
	}
	if event.ValueDecoded == nil || strings.TrimSpace(*event.ValueDecoded) == "" {
		return true
	}
	raw := strings.TrimSpace(*event.ValueDecoded)
	if !strings.HasPrefix(raw, "{") {
		return true
	}
	var payload decodedEventPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return true
	}
	if payload.SpecDecodeStatus == specDecodeStatusDecoded {
		return false
	}
	return payload.SpecDecodeStatus == "" ||
		payload.SpecDecodeStatus == specDecodeStatusNoSpec ||
		payload.SpecDecodeStatus == specDecodeStatusPartial ||
		len(payload.Fields) == 0
}

// ReEnrichReadOperationSummaries applies contract spec decoding to indexed operations at read time.
func ReEnrichReadOperationSummaries(
	ctx context.Context,
	loader ContractSpecRegistryLoader,
	ops []store.ReadOperationSummary,
	txByHash map[string]source.TransactionEntry,
) error {
	if loader == nil || len(ops) == 0 {
		return nil
	}

	byHash := make(map[string][]int)
	for index, op := range ops {
		if !OperationNeedsSpecReEnrich(op) {
			continue
		}
		hash := strings.TrimSpace(op.TransactionHash)
		if hash == "" {
			continue
		}
		byHash[hash] = append(byHash[hash], index)
	}
	if len(byHash) == 0 {
		return nil
	}

	for hash, indexes := range byHash {
		txEntry, ok := txByHash[hash]
		if !ok {
			continue
		}
		batch := make([]store.Operation, 0, len(indexes))
		for _, index := range indexes {
			batch = append(batch, readOperationToStoreOperation(ops[index]))
		}
		if err := EnrichSorobanOperations(ctx, loader, batch, txEntry); err != nil {
			return err
		}
		for batchIndex, opIndex := range indexes {
			ops[opIndex].Details = batch[batchIndex].Details
		}
	}

	return nil
}

// ReEnrichReadContractEventSummaries applies contract spec decoding to indexed events at read time.
func ReEnrichReadContractEventSummaries(
	ctx context.Context,
	loader ContractSpecRegistryLoader,
	events []store.ReadContractEventSummary,
) error {
	if loader == nil || len(events) == 0 {
		return nil
	}

	batch := make([]store.ContractEvent, 0)
	indexes := make([]int, 0)
	for index, event := range events {
		if !EventNeedsSpecReEnrich(event) {
			continue
		}
		batch = append(batch, readEventToStoreEvent(event))
		indexes = append(indexes, index)
	}
	if len(batch) == 0 {
		return nil
	}
	if err := EnrichContractEvents(ctx, loader, batch); err != nil {
		return err
	}
	for batchIndex, eventIndex := range indexes {
		if batch[batchIndex].ValueDecoded != nil {
			events[eventIndex].ValueDecoded = batch[batchIndex].ValueDecoded
		}
		if batch[batchIndex].Topic1 != nil {
			events[eventIndex].Topic1 = batch[batchIndex].Topic1
		}
	}
	return nil
}

func readOperationToStoreOperation(op store.ReadOperationSummary) store.Operation {
	return store.Operation{
		TransactionHash:  op.TransactionHash,
		ApplicationOrder: op.ApplicationOrder,
		TypeName:         op.TypeName,
		ContractID:       op.ContractID,
		FunctionName:     op.FunctionName,
		Details:          op.Details,
	}
}

func readEventToStoreEvent(event store.ReadContractEventSummary) store.ContractEvent {
	return store.ContractEvent{
		ContractID:      event.ContractID,
		TransactionHash: event.TransactionHash,
		LedgerSequence:  event.LedgerSequence,
		Type:            event.Type,
		Topic1:          event.Topic1,
		Topic2:          event.Topic2,
		Topic3:          event.Topic3,
		Topic4:          event.Topic4,
		TopicsXDR:       stringOrEmpty(event.TopicsXDR),
		ValueXDR:        stringOrEmpty(event.ValueXDR),
		ValueDecoded:    event.ValueDecoded,
		CreatedAt:       event.CreatedAt,
	}
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func TransactionEntryFromReadDetail(tx *store.ReadTransactionDetail) source.TransactionEntry {
	if tx == nil {
		return source.TransactionEntry{}
	}
	meta := ""
	if tx.ResultMetaXDR != nil {
		meta = strings.TrimSpace(*tx.ResultMetaXDR)
	}
	return source.TransactionEntry{
		EnvelopeXDR:   strings.TrimSpace(tx.EnvelopeXDR),
		ResultMetaXDR: meta,
	}
}
