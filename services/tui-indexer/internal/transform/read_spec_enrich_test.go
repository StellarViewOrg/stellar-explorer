package transform

import (
	"context"
	"testing"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func TestReEnrichReadOperationSummaries_DecodesPreviouslyUnavailableSpec(t *testing.T) {
	registry := loadHelloWorldRegistry(t)
	contractID, err := strkey.Encode(strkey.VersionByteContract, bytes32("hello-world-contract-id-00000000"))
	if err != nil {
		t.Fatalf("encode contract id: %v", err)
	}
	loader := StaticSpecRegistryLoader{
		Registries: map[string]*sordecode.SpecRegistry{
			contractID: registry,
		},
	}

	fn := "hello"
	ops := []store.ReadOperationSummary{{
		TransactionHash:  "tx-hello",
		ApplicationOrder: 1,
		TypeName:         "invoke_host_function",
		ContractID:       &contractID,
		FunctionName:     &fn,
		Details:          `{"function_name":"hello","contract_id":"` + contractID + `","arguments":["arg1=World"],"spec_decode_status":"spec_unavailable"}`,
	}}

	to := xdr.ScString("World")
	arg := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &to}
	envelope := buildInvokeEnvelope(t, contractID, "hello", []xdr.ScVal{arg})
	txByHash := map[string]source.TransactionEntry{
		"tx-hello": {EnvelopeXDR: envelope},
	}

	if err := ReEnrichReadOperationSummaries(context.Background(), loader, ops, txByHash); err != nil {
		t.Fatalf("re-enrich operations: %v", err)
	}

	details, err := parseOperationDetailsMap(ops[0].Details)
	if err != nil {
		t.Fatalf("parse details: %v", err)
	}
	if details["spec_decode_status"] != specDecodeStatusDecoded {
		t.Fatalf("unexpected spec decode status: %#v", details["spec_decode_status"])
	}
	decodedArgs, ok := details["arguments_decoded"].([]interface{})
	if !ok || len(decodedArgs) != 1 {
		t.Fatalf("expected decoded args, got %#v", details["arguments_decoded"])
	}
}

func TestReEnrichReadContractEventSummaries_DecodesPreviouslyUnavailableSpec(t *testing.T) {
	registry := loadHelloWorldRegistry(t)
	contractID, err := strkey.Encode(strkey.VersionByteContract, bytes32("hello-world-contract-id-00000001"))
	if err != nil {
		t.Fatalf("encode contract id: %v", err)
	}
	loader := StaticSpecRegistryLoader{
		Registries: map[string]*sordecode.SpecRegistry{
			contractID: registry,
		},
	}

	topic := xdr.ScVal{Type: xdr.ScValTypeScvSymbol, Sym: scSymbolPtr("hello")}
	value := xdr.ScVal{Type: xdr.ScValTypeScvVec, Vec: emptyVec()}
	topicXDR, err := xdr.MarshalBase64(topic)
	if err != nil {
		t.Fatalf("marshal topic: %v", err)
	}
	valueXDR, err := xdr.MarshalBase64(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	text := "[]"
	topicsJSON := `["` + topicXDR + `"]`
	events := []store.ReadContractEventSummary{{
		ContractID:      contractID,
		TransactionHash: "tx-hello-event",
		LedgerSequence:  10,
		Type:            1,
		TopicsXDR:       &topicsJSON,
		ValueXDR:        &valueXDR,
		ValueDecoded:    &text,
	}}

	if err := ReEnrichReadContractEventSummaries(context.Background(), loader, events); err != nil {
		t.Fatalf("re-enrich events: %v", err)
	}
	if events[0].ValueDecoded == nil || !stringsHasPrefix(*events[0].ValueDecoded, "{") {
		t.Fatalf("expected enriched event payload, got %#v", events[0].ValueDecoded)
	}
}

func TestOperationNeedsSpecReEnrichSkipsDecodedOperations(t *testing.T) {
	op := store.ReadOperationSummary{
		TypeName: "invoke_host_function",
		Details:  `{"spec_decode_status":"decoded","arguments_decoded":[{"name":"to","value":"World"}]}`,
	}
	if OperationNeedsSpecReEnrich(op) {
		t.Fatal("expected decoded operation to skip re-enrich")
	}
}

func stringsHasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}
