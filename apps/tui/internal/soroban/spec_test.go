package soroban

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/sordecode"
)

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "sordecode", "testdata", name)
}

func loadWASMFixture(t *testing.T, name string) []byte {
	t.Helper()
	bytes, err := os.ReadFile(testdataPath(t, name))
	if err != nil {
		t.Fatalf("read wasm fixture %s: %v", name, err)
	}
	return bytes
}

func specJSONFromWASM(t *testing.T, name string) string {
	t.Helper()
	wasmBytes := loadWASMFixture(t, name)
	specBytes, err := sordecode.ExtractSpecFromWASM(wasmBytes)
	if err != nil {
		t.Fatalf("extract spec from %s: %v", name, err)
	}
	entries, err := sordecode.DecodeEntries(specBytes)
	if err != nil {
		t.Fatalf("decode entries from %s: %v", name, err)
	}
	encoded, err := json.Marshal(sordecode.EntriesToJSON(entries))
	if err != nil {
		t.Fatalf("marshal spec json from %s: %v", name, err)
	}
	return string(encoded)
}

func TestBuildContractSpec_HelloWorld(t *testing.T) {
	spec := BuildContractSpec("CHELLO", specJSONFromWASM(t, "hello_world.wasm"))
	if spec == nil {
		t.Fatal("expected contract spec")
	}
	if !spec.Available || spec.DecodeStatus != "decoded" {
		t.Fatalf("unexpected decode status: available=%v status=%q", spec.Available, spec.DecodeStatus)
	}
	if spec.FunctionCount != 1 || len(spec.Functions) != 1 {
		t.Fatalf("expected one function, got %#v", spec.Functions)
	}
	if spec.Functions[0].Name != "hello" {
		t.Fatalf("expected hello function, got %#v", spec.Functions[0])
	}
	if len(spec.Functions[0].Inputs) != 1 || spec.Functions[0].Inputs[0].Name != "to" {
		t.Fatalf("unexpected hello inputs: %#v", spec.Functions[0].Inputs)
	}
}

func TestBuildContractSpec_MissingSpec(t *testing.T) {
	spec := BuildContractSpec("CMISSING", "")
	if spec == nil {
		t.Fatal("expected contract spec shell")
	}
	if spec.Available || spec.DecodeStatus != "missing" {
		t.Fatalf("expected missing spec, got available=%v status=%q", spec.Available, spec.DecodeStatus)
	}
}
