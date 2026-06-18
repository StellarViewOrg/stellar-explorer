package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

// ContractWorkspaceTab identifies the active contract investigation pane.
type ContractWorkspaceTab string

const (
	ContractWorkspaceTabOverview     ContractWorkspaceTab = "overview"
	ContractWorkspaceTabSpec         ContractWorkspaceTab = "spec"
	ContractWorkspaceTabStorage      ContractWorkspaceTab = "storage"
	ContractWorkspaceTabEvents       ContractWorkspaceTab = "events"
	ContractWorkspaceTabInvocations  ContractWorkspaceTab = "invocations"
	ContractWorkspaceTabTransactions ContractWorkspaceTab = "transactions"
)

// AvailableContractWorkspaceTabs returns the tabs that have data for the current contract lookup.
func AvailableContractWorkspaceTabs(response *backendclient.ContractLookupResponse) []ContractWorkspaceTab {
	if response == nil || response.Contract == nil {
		return []ContractWorkspaceTab{ContractWorkspaceTabOverview}
	}
	contract := response.Contract
	tabs := []ContractWorkspaceTab{ContractWorkspaceTabOverview}
	if response.Spec != nil && response.Spec.Available {
		tabs = append(tabs, ContractWorkspaceTabSpec)
	}
	if len(response.Storage) > 0 || contract.StorageEntryCount > 0 {
		tabs = append(tabs, ContractWorkspaceTabStorage)
	}
	if len(response.RecentEvents) > 0 || contract.EventCount > 0 {
		tabs = append(tabs, ContractWorkspaceTabEvents)
	}
	if contract.InvocationCount > 0 {
		tabs = append(tabs, ContractWorkspaceTabInvocations)
	}
	if len(response.RecentTransactions) > 0 {
		tabs = append(tabs, ContractWorkspaceTabTransactions)
	}
	return tabs
}

// LookupSupportsContractWorkspaceTabs reports whether tab cycling is available in the lookup view.
func LookupSupportsContractWorkspaceTabs(lookup LookupSnapshot) bool {
	return lookup.State == ViewStateReady &&
		lookup.Kind == LookupContract &&
		lookup.Explorer == nil &&
		lookup.Contract != nil &&
		lookup.Contract.Contract != nil
}

// LookupSupportsVisualMode reports whether visual navigation mode is available.
func LookupSupportsVisualMode(lookup LookupSnapshot) bool {
	if lookup.State != ViewStateReady || lookup.Explorer != nil {
		return false
	}
	switch lookup.Kind {
	case LookupContract, LookupEvent, LookupStorage, LookupOperation, LookupTransaction, LookupAccount, LookupAsset, LookupLedger:
		return true
	default:
		return false
	}
}

// LookupSupportsDecodeShortcuts reports whether c/d decode shortcuts apply to the lookup view.
func LookupSupportsDecodeShortcuts(lookup LookupSnapshot) bool {
	if lookup.State != ViewStateReady || lookup.Explorer != nil {
		return false
	}
	switch lookup.Kind {
	case LookupContract, LookupEvent, LookupStorage:
		return true
	case LookupOperation:
		return lookup.Operation != nil && strings.TrimSpace(lookup.Operation.Operation.TypeName) == "invoke_host_function"
	default:
		return false
	}
}

// SupportsContractWorkspaceTabs reports whether tab cycling is available in the current lookup view.
func (m *Model) SupportsContractWorkspaceTabs() bool {
	return LookupSupportsContractWorkspaceTabs(m.lookup)
}

// SupportsLookupVisualMode reports whether visual navigation mode is available.
func (m *Model) SupportsLookupVisualMode() bool {
	return LookupSupportsVisualMode(m.lookup)
}

// SupportsLookupDecodeShortcuts reports whether c/d decode shortcuts apply to the current lookup.
func (m *Model) SupportsLookupDecodeShortcuts() bool {
	return LookupSupportsDecodeShortcuts(m.lookup)
}

// CycleContractWorkspaceTab advances the active contract workspace tab.
func (m *Model) CycleContractWorkspaceTab(delta int) error {
	if !m.SupportsContractWorkspaceTabs() {
		return errors.New("contract tabs are only available from contract detail views")
	}
	if delta == 0 {
		return nil
	}
	tabs := AvailableContractWorkspaceTabs(m.lookup.Contract)
	if len(tabs) <= 1 {
		return nil
	}
	current := m.lookup.ContractTab
	if current == "" {
		current = ContractWorkspaceTabOverview
	}
	index := 0
	for i, tab := range tabs {
		if tab == current {
			index = i
			break
		}
	}
	next := index + delta
	for next < 0 {
		next += len(tabs)
	}
	next %= len(tabs)
	m.lookup.ContractTab = tabs[next]
	resetLookupSelection(&m.lookup)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Contract tab: %s.", tabs[next]),
	}
	return nil
}

// ToggleLookupExpandAll toggles expanded rendering for long Soroban and entity fields.
func (m *Model) ToggleLookupExpandAll() error {
	if m.lookup.State != ViewStateReady || m.lookup.Explorer != nil {
		return errors.New("expand mode is only available from entity detail views")
	}
	m.lookup.ExpandAll = !m.lookup.ExpandAll
	state := "compact"
	if m.lookup.ExpandAll {
		state = "expanded"
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Lookup display: %s.", state),
	}
	return nil
}

// ToggleLookupVisualMode toggles keyboard navigation across traversable entity rows only.
func (m *Model) ToggleLookupVisualMode() error {
	if !m.SupportsLookupVisualMode() {
		return errors.New("visual mode is only available from entity detail views")
	}
	m.lookup.VisualMode = !m.lookup.VisualMode
	if m.lookup.VisualMode {
		resetLookupSelection(&m.lookup)
	}
	state := "off"
	if m.lookup.VisualMode {
		state = "on"
	}
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Visual mode: %s.", state),
	}
	return nil
}
