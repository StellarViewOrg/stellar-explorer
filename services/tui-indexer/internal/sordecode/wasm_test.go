package sordecode

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadWASMFixture(t *testing.T, name string) []byte {
	t.Helper()
	bytes, err := os.ReadFile(testdataPath(t, name))
	if err != nil {
		t.Fatalf("read wasm fixture %s: %v", name, err)
	}
	return bytes
}

func TestExtractSpecFromWASM_HelloWorld(t *testing.T) {
	wasmBytes := loadWASMFixture(t, "hello_world.wasm")
	specBytes, err := ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		t.Fatalf("extract spec: %v", err)
	}
	entries, err := DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected hello_world spec entries")
	}
	registry := NewRegistry(entries)
	fn, ok := registry.Function("hello")
	if !ok {
		t.Fatal("expected hello function in spec")
	}
	if len(fn.Inputs) != 1 || string(fn.Inputs[0].Name) != "to" {
		t.Fatalf("unexpected hello inputs: %#v", fn.Inputs)
	}
}

func TestExtractSpecFromWASM_Increment(t *testing.T) {
	wasmBytes := loadWASMFixture(t, "increment.wasm")
	specBytes, err := ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		t.Fatalf("extract spec: %v", err)
	}
	entries, err := DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	registry := NewRegistry(entries)
	fn, ok := registry.Function("increment")
	if !ok {
		t.Fatal("expected increment function in spec")
	}
	if len(fn.Inputs) != 0 {
		t.Fatalf("expected increment to have no inputs, got %#v", fn.Inputs)
	}
	if len(fn.Outputs) != 1 || FormatTypeDef(fn.Outputs[0]) != "u32" {
		t.Fatalf("unexpected increment output: %#v", fn.Outputs)
	}
}

func TestExtractSpecFromWASM_TokenHasEvents(t *testing.T) {
	wasmBytes := loadWASMFixture(t, "token.wasm")
	specBytes, err := ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		t.Fatalf("extract spec: %v", err)
	}
	entries, err := DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries: %v", err)
	}
	registry := NewRegistry(entries)
	if registry.EventCount() == 0 {
		t.Fatal("expected token contract events in spec")
	}
	_, ok := registry.Function("transfer")
	if !ok {
		t.Fatal("expected transfer function in token spec")
	}
}
