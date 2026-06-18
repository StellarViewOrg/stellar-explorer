package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestRenderContractWorkspaceTabBar(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract("CCONTRACT", backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{
			ContractID:        "CCONTRACT",
			StorageEntryCount: 2,
			EventCount:        1,
		},
		Spec: &backendclient.ContractSpec{Available: true},
	})

	bar := renderContractWorkspaceTabBar(model.Snapshot().Lookup, 120)
	if !strings.Contains(bar, "[overview]") {
		t.Fatalf("expected active overview tab, got %q", bar)
	}
	if !strings.Contains(bar, "spec") || !strings.Contains(bar, "storage") {
		t.Fatalf("expected available tabs in bar, got %q", bar)
	}
}

func TestLookupContractTabSectionsFilterByActiveTab(t *testing.T) {
	lookup := app.LookupSnapshot{
		Kind:        app.LookupContract,
		State:       app.ViewStateReady,
		ContractTab: app.ContractWorkspaceTabSpec,
		Contract: &backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT"},
			Spec: &backendclient.ContractSpec{
				Available:     true,
				DecodeStatus:  "decoded",
				FunctionCount: 1,
				Functions: []backendclient.ContractSpecFunction{
					{Name: "hello"},
				},
			},
		},
	}

	sections := lookupContractTabSections(lookup)
	foundMethod := false
	for _, section := range sections {
		if section.Title == "Method 1" {
			foundMethod = true
		}
	}
	if !foundMethod {
		t.Fatalf("expected spec tab to render methods, got %#v", sections)
	}
}

func TestLookupModelTabKeyCyclesContractWorkspace(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract("CCONTRACT", backendclient.ContractLookupResponse{
		Contract: &backendclient.ContractDetail{
			ContractID:        "CCONTRACT",
			StorageEntryCount: 1,
		},
		Spec: &backendclient.ContractSpec{Available: true},
	})

	lookupModel := NewLookupModel(model.Snapshot(), 120, 20)
	updated, cmd := lookupModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("expected tab action command")
	}
	action, ok := cmd().(ActionMsg)
	if !ok || action.Kind != ActionCycleContractTab || action.Delta != 1 {
		t.Fatalf("unexpected tab action: %#v", cmd())
	}
	_ = updated
}

func TestVisualModeLimitsSelectionToNavigableRows(t *testing.T) {
	sections := []lookupSection{
		{Title: "Static", Body: "value"},
		relatedEntityRow("Contract", "CCONTRACT", "CCONTRACT", "lookup contract CCONTRACT", "enter"),
	}
	if nextSelectableLookupSection(sections, 0, 1, true) != 1 {
		t.Fatal("expected visual mode to reach navigable row")
	}
	if nextSelectableLookupSection(sections, 1, 1, true) != 1 {
		t.Fatal("expected visual mode to stay on last navigable row")
	}
}
