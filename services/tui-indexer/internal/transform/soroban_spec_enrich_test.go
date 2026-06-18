package transform

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func TestEnrichSorobanOperations_DecodesHelloWorldArguments(t *testing.T) {
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

	to := xdr.ScString("World")
	arg := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &to}
	envelope := buildInvokeEnvelope(t, contractID, "hello", []xdr.ScVal{arg})
	functionName := "hello"
	ops := []store.Operation{{
		TransactionHash:  "tx-hello",
		ApplicationOrder: 1,
		TypeName:         "invoke_host_function",
		ContractID:       &contractID,
		FunctionName:     &functionName,
		Details:          `{"function_name":"hello","contract_id":"` + contractID + `","arguments":["arg1=World"]}`,
	}}

	txEntry := source.TransactionEntry{EnvelopeXDR: envelope}
	if err := EnrichSorobanOperations(context.Background(), loader, ops, txEntry); err != nil {
		t.Fatalf("enrich operations: %v", err)
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
	first, ok := decodedArgs[0].(map[string]interface{})
	if !ok || first["name"] != "to" || first["value"] != "World" {
		t.Fatalf("unexpected decoded arg: %#v", decodedArgs[0])
	}
}

func TestEnrichContractEvents_UsesSpecWhenAvailable(t *testing.T) {
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

	events := []store.ContractEvent{{
		ContractID:      contractID,
		TransactionHash: "tx-hello-event",
		LedgerSequence:  10,
		Type:            1,
		TopicsXDR:       `["` + topicXDR + `"]`,
		ValueXDR:        valueXDR,
	}}
	if err := EnrichContractEvents(context.Background(), loader, events); err != nil {
		t.Fatalf("enrich events: %v", err)
	}
	if events[0].ValueDecoded == nil || *events[0].ValueDecoded == "" {
		t.Fatal("expected enriched value_decoded payload")
	}
}

func loadHelloWorldRegistry(t *testing.T) *sordecode.SpecRegistry {
	t.Helper()
	return loadRegistryFromFixture(t, "hello_world.wasm")
}

func loadRegistryFromFixture(t *testing.T, fixture string) *sordecode.SpecRegistry {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	wasmPath := filepath.Join(filepath.Dir(file), "..", "sordecode", "testdata", fixture)
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("read wasm fixture: %v", err)
	}
	specBytes, err := sordecode.ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		t.Fatalf("extract spec: %v", err)
	}
	entries, err := sordecode.DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	return sordecode.NewRegistry(entries)
}

func buildInvokeEnvelope(t *testing.T, contractID, functionName string, args []xdr.ScVal) string {
	t.Helper()
	contractBytes, err := strkey.Decode(strkey.VersionByteContract, contractID)
	if err != nil {
		t.Fatalf("contract id: %v", err)
	}
	var contractIDBytes xdr.ContractId
	copy(contractIDBytes[:], contractBytes)
	address := xdr.ScAddress{
		Type:       xdr.ScAddressTypeScAddressTypeContract,
		ContractId: &contractIDBytes,
	}
	hostFn, err := xdr.NewHostFunction(xdr.HostFunctionTypeHostFunctionTypeInvokeContract, xdr.InvokeContractArgs{
		ContractAddress: address,
		FunctionName:    xdr.ScSymbol(functionName),
		Args:            args,
	})
	if err != nil {
		t.Fatalf("host function: %v", err)
	}
	invoke := xdr.InvokeHostFunctionOp{HostFunction: hostFn}
	op := xdr.Operation{
		Body: xdr.OperationBody{
			Type:                 xdr.OperationTypeInvokeHostFunction,
			InvokeHostFunctionOp: &invoke,
		},
	}
	accountID := xdr.MustAddress("GBRPYHIL2CI3FNQ4BXLFMNDLFJUNPU2HY3ZMFSHONUCEOASW7QC7OX2H")
	ed25519, ok := accountID.GetEd25519()
	if !ok {
		t.Fatal("expected ed25519 account")
	}
	tx := xdr.Transaction{
		SourceAccount: xdr.MuxedAccount{
			Type:    xdr.CryptoKeyTypeKeyTypeEd25519,
			Ed25519: &ed25519,
		},
		Fee:        100,
		SeqNum:     1,
		Cond:       xdr.Preconditions{Type: xdr.PreconditionTypePrecondNone},
		Memo:       xdr.Memo{Type: xdr.MemoTypeMemoNone},
		Operations: []xdr.Operation{op},
		Ext:        xdr.TransactionExt{V: 0},
	}
	envelope := xdr.TransactionEnvelope{
		Type: xdr.EnvelopeTypeEnvelopeTypeTx,
		V1: &xdr.TransactionV1Envelope{
			Tx:         tx,
			Signatures: nil,
		},
	}
	encoded, err := xdr.MarshalBase64(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return encoded
}

func bytes32(seed string) []byte {
	raw := make([]byte, 32)
	copy(raw, []byte(seed))
	return raw
}

func emptyVec() **xdr.ScVec {
	empty := xdr.ScVec{}
	vec := &empty
	return &vec
}
