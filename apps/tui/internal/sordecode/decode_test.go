package sordecode

import (
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestDecodeInvocation_HelloWorld(t *testing.T) {
	registry := loadRegistryFromWASM(t, "hello_world.wasm")
	to := xdr.ScString("World")
	arg := xdr.ScVal{Type: xdr.ScValTypeScvString, Str: &to}
	decoded, err := registry.DecodeInvocation("hello", []xdr.ScVal{arg})
	if err != nil {
		t.Fatalf("decode invocation: %v", err)
	}
	if len(decoded.Params) != 1 {
		t.Fatalf("expected one param, got %#v", decoded.Params)
	}
	if decoded.Params[0].Name != "to" || decoded.Params[0].Type != "string" || decoded.Params[0].Value != "World" {
		t.Fatalf("unexpected decoded param: %#v", decoded.Params[0])
	}
}

func TestDecodeInvocation_UnknownFunctionFallback(t *testing.T) {
	registry := loadRegistryFromWASM(t, "hello_world.wasm")
	arg := xdr.ScVal{Type: xdr.ScValTypeScvU32}
	value := xdr.Uint32(7)
	arg.U32 = &value
	decoded, err := registry.DecodeInvocation("missing", []xdr.ScVal{arg})
	if err != nil {
		t.Fatalf("decode invocation: %v", err)
	}
	if decoded.Params[0].Name != "arg1" || decoded.Params[0].Value != "7" {
		t.Fatalf("unexpected fallback param: %#v", decoded.Params[0])
	}
}

func TestEntriesToJSON_IncludesEventFields(t *testing.T) {
	entries := loadEntriesFromWASM(t, "token.wasm")
	jsonEntries := EntriesToJSON(entries)
	foundTransfer := false
	foundEvent := false
	for _, entry := range jsonEntries {
		if entry["name"] == "transfer" {
			foundTransfer = true
			inputs, ok := entry["inputs"].([]map[string]interface{})
			if !ok || len(inputs) < 3 {
				t.Fatalf("expected transfer inputs in json spec: %#v", entry["inputs"])
			}
		}
		if _, ok := entry["prefix_topics"]; ok {
			foundEvent = true
			params, ok := entry["params"].([]map[string]interface{})
			if !ok || len(params) == 0 {
				t.Fatalf("expected event params in json spec: %#v", entry)
			}
		}
	}
	if !foundTransfer {
		t.Fatal("expected transfer function in json spec")
	}
	if !foundEvent {
		t.Fatal("expected event entry in json spec")
	}
}

func loadRegistryFromWASM(t *testing.T, fixture string) *SpecRegistry {
	t.Helper()
	return NewRegistry(loadEntriesFromWASM(t, fixture))
}

func loadEntriesFromWASM(t *testing.T, fixture string) []xdr.ScSpecEntry {
	t.Helper()
	specBytes, err := ExtractSpecFromWASM(loadWASMFixture(t, fixture))
	if err != nil {
		t.Fatalf("extract spec: %v", err)
	}
	entries, err := DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	return entries
}
