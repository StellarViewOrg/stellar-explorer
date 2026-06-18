package app

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestAvailableContractWorkspaceTabs(t *testing.T) {
	response := &backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{
			ContractID:        "CCONTRACT",
			StorageEntryCount: 3,
			InvocationCount:   2,
			EventCount:        4,
		},
		Spec: &backendclient.ContractSpec{Available: true},
		Storage: []backendclient.ContractStorageSummary{
			{ContractID: "CCONTRACT"},
		},
		RecentEvents: []backendclient.ContractEventSummary{
			{TransactionHash: "tx-1"},
		},
		RecentTransactions: []backendclient.TransactionSummary{
			{Hash: "tx-1"},
		},
	}

	tabs := AvailableContractWorkspaceTabs(response)
	if len(tabs) != 6 {
		t.Fatalf("expected six tabs, got %#v", tabs)
	}
	if tabs[0] != ContractWorkspaceTabOverview || tabs[1] != ContractWorkspaceTabSpec {
		t.Fatalf("unexpected tab order: %#v", tabs)
	}
}

func TestCycleContractWorkspaceTab(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupContract("CCONTRACT", backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{
			ContractID:        "CCONTRACT",
			StorageEntryCount: 1,
			EventCount:        1,
		},
		Spec: &backendclient.ContractSpec{Available: true},
	})

	if err := model.CycleContractWorkspaceTab(1); err != nil {
		t.Fatalf("cycle tab: %v", err)
	}
	if model.Snapshot().Lookup.ContractTab != ContractWorkspaceTabSpec {
		t.Fatalf("expected spec tab, got %q", model.Snapshot().Lookup.ContractTab)
	}
}

func TestSetContractDecodeModeForOperation(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.lookup = LookupSnapshot{
		Kind:  LookupOperation,
		State: ViewStateReady,
		Operation: &backendclient.OperationLookupSnapshot{
			Operation: backendclient.OperationSummary{TypeName: "invoke_host_function"},
		},
	}
	if err := model.SetContractDecodeMode(ContractDecodeModeRaw); err != nil {
		t.Fatalf("set decode mode: %v", err)
	}
	if model.Snapshot().Lookup.DecodeMode != ContractDecodeModeRaw {
		t.Fatalf("expected raw decode mode, got %q", model.Snapshot().Lookup.DecodeMode)
	}
}

func TestToggleLookupVisualMode(t *testing.T) {
	model := NewModel(config.Default(), "/tmp/config.json", CacheSnapshot{})
	model.UpdateLookupContract("CCONTRACT", backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT"},
	})
	if err := model.ToggleLookupVisualMode(); err != nil {
		t.Fatalf("toggle visual mode: %v", err)
	}
	if !model.Snapshot().Lookup.VisualMode {
		t.Fatal("expected visual mode enabled")
	}
}
