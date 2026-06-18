package soroban

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/sordecode"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// EventsFromRPC fetches and decodes recent contract events for one contract.
func EventsFromRPC(ctx context.Context, client *rpcclient.Client, loader *SpecLoader, contractID string, limit int) ([]backendclient.ContractEventSummary, error) {
	if client == nil {
		return nil, fmt.Errorf("rpc client is required")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	latest, err := client.GetLatestLedger(ctx)
	if err != nil {
		return nil, err
	}
	startLedger := uint32(1)
	if latest.Sequence > 17_280 {
		startLedger = latest.Sequence - 17_280
	}

	result, err := client.GetEvents(ctx, rpcclient.GetEventsParams{
		StartLedger: startLedger,
		Filters: []rpcclient.GetEventsFilter{
			{Type: "contract", ContractIds: []string{contractID}},
		},
		Pagination: &rpcclient.Pagination{Limit: limit},
	})
	if err != nil {
		return nil, err
	}

	registry, _ := loader.Registry(ctx, contractID)
	events := make([]backendclient.ContractEventSummary, 0, len(result.Events))
	for _, entry := range result.Events {
		if strings.TrimSpace(entry.ContractID) != contractID {
			continue
		}
		summary := rpcEventToSummary(entry)
		enrichEventSummary(registry, &summary)
		events = append(events, summary)
	}
	return events, nil
}

func rpcEventToSummary(entry rpcclient.EventEntry) backendclient.ContractEventSummary {
	createdAt := time.Time{}
	if parsed, err := time.Parse(time.RFC3339, entry.LedgerClosedAt); err == nil {
		createdAt = parsed.UTC()
	}

	topicsXDR, _ := json.Marshal(entry.Topic)
	topicsXDRStr := string(topicsXDR)
	valueXDR := strings.TrimSpace(entry.Value)

	summary := backendclient.ContractEventSummary{
		ContractID:      entry.ContractID,
		TransactionHash: entry.TxHash,
		LedgerSequence:  entry.Ledger,
		TopicsXDR:       &topicsXDRStr,
		ValueXDR:        &valueXDR,
		CreatedAt:       createdAt,
	}

	topics := decodeTopicVals(entry.Topic)
	if len(topics) > 0 {
		decodedTopics := make([]string, 0, len(topics))
		for _, topic := range topics {
			decodedTopics = append(decodedTopics, sordecode.ScValToString(topic))
		}
		summary.Topics = decodedTopics
		if len(decodedTopics) > 0 {
			summary.Topic1 = &decodedTopics[0]
		}
		if len(decodedTopics) > 1 {
			summary.Topic2 = &decodedTopics[1]
		}
		if len(decodedTopics) > 2 {
			summary.Topic3 = &decodedTopics[2]
		}
		if len(decodedTopics) > 3 {
			summary.Topic4 = &decodedTopics[3]
		}
	}
	if value, ok := decodeValueVal(valueXDR); ok {
		text := sordecode.ScValToString(value)
		summary.ValueDecoded = &text
	}
	return summary
}

func enrichEventSummary(registry *sordecode.SpecRegistry, summary *backendclient.ContractEventSummary) {
	if summary == nil {
		return
	}
	topics := decodeTopicValsFromSummary(*summary)
	value, ok := decodeValueVal(stringPtr(summary.ValueXDR))
	if !ok {
		summary.DecodeStatus = "raw"
		return
	}

	if registry == nil {
		summary.DecodeStatus = "raw"
		if summary.ValueDecoded == nil {
			text := sordecode.ScValToString(value)
			summary.ValueDecoded = &text
		}
		return
	}

	decoded, err := registry.DecodeEvent(topics, value)
	if err != nil {
		summary.DecodeStatus = "partial"
		return
	}

	payload := decodedEventPayload{
		Text:             sordecode.ScValToString(value),
		EventName:        decoded.Name,
		Fields:           decoded.Fields,
		SpecDecodeStatus: specDecodeStatusForEvent(decoded),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		summary.DecodeStatus = "partial"
		return
	}
	encodedStr := string(encoded)
	summary.ValueDecoded = &encodedStr
	summary.EventName = decoded.Name
	summary.SpecDecodeStatus = payload.SpecDecodeStatus
	summary.FieldsDecoded = fieldsToBackend(decoded.Fields)
	summary.DecodeStatus = payload.SpecDecodeStatus
	if decoded.Name != "" {
		summary.Topic1 = &decoded.Name
	}
}

func specDecodeStatusForEvent(decoded sordecode.DecodedEvent) string {
	if decoded.Name == "" || len(decoded.Fields) == 0 {
		return "raw"
	}
	for _, field := range decoded.Fields {
		if field.Value == "" || field.Value == "<missing>" {
			return specDecodeStatusPartial
		}
	}
	return specDecodeStatusDecoded
}

func fieldsToBackend(fields []sordecode.NamedField) []backendclient.ContractEventField {
	rows := make([]backendclient.ContractEventField, 0, len(fields))
	for _, field := range fields {
		rows = append(rows, backendclient.ContractEventField{
			Name:     field.Name,
			Type:     field.Type,
			Location: field.Location,
			Value:    field.Value,
		})
	}
	return rows
}

func decodeTopicVals(topicXDRs []string) []xdr.ScVal {
	topics := make([]xdr.ScVal, 0, len(topicXDRs))
	for _, topicXDR := range topicXDRs {
		var topic xdr.ScVal
		if err := xdr.SafeUnmarshalBase64(strings.TrimSpace(topicXDR), &topic); err != nil {
			continue
		}
		topics = append(topics, topic)
	}
	return topics
}

func decodeTopicValsFromSummary(summary backendclient.ContractEventSummary) []xdr.ScVal {
	if summary.TopicsXDR != nil && strings.TrimSpace(*summary.TopicsXDR) != "" {
		var topicXDRs []string
		if err := json.Unmarshal([]byte(*summary.TopicsXDR), &topicXDRs); err == nil {
			return decodeTopicVals(topicXDRs)
		}
	}
	return nil
}

func decodeValueVal(valueXDR string) (xdr.ScVal, bool) {
	valueXDR = strings.TrimSpace(valueXDR)
	if valueXDR == "" {
		return xdr.ScVal{}, false
	}
	var value xdr.ScVal
	if err := xdr.SafeUnmarshalBase64(valueXDR, &value); err != nil {
		return xdr.ScVal{}, false
	}
	return value, true
}

func encodeScValBase64(value xdr.ScVal) string {
	eb := xdr.NewEncodingBuffer()
	b, err := eb.MarshalBinary(value)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
