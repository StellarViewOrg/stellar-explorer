package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func lookupDisplayLimit(lookup app.LookupSnapshot) int {
	if lookup.ExpandAll {
		return 4096
	}
	return 72
}

func renderContractWorkspaceTabBar(lookup app.LookupSnapshot, width int) string {
	if !app.LookupSupportsContractWorkspaceTabs(lookup) {
		return ""
	}
	tabs := app.AvailableContractWorkspaceTabs(lookup.Contract)
	if len(tabs) <= 1 {
		return ""
	}
	active := lookup.ContractTab
	if active == "" {
		active = app.ContractWorkspaceTabOverview
	}
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := contractWorkspaceTabLabel(tab)
		if tab == active {
			parts = append(parts, selectedRowStyle.Render("["+label+"]"))
			continue
		}
		parts = append(parts, mutedStyle.Render(label))
	}
	line := strings.Join(parts, "  ")
	return truncate("Tabs: "+line+"  (tab/shift+tab)", width)
}

func contractWorkspaceTabLabel(tab app.ContractWorkspaceTab) string {
	switch tab {
	case app.ContractWorkspaceTabSpec:
		return "spec"
	case app.ContractWorkspaceTabStorage:
		return "storage"
	case app.ContractWorkspaceTabEvents:
		return "events"
	case app.ContractWorkspaceTabInvocations:
		return "invocations"
	case app.ContractWorkspaceTabTransactions:
		return "transactions"
	default:
		return "overview"
	}
}

func renderLookupWorkspaceModes(lookup app.LookupSnapshot, width int) string {
	if lookup.State != app.ViewStateReady || lookup.Explorer != nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if app.LookupSupportsDecodeShortcuts(lookup) {
		mode := string(lookup.DecodeMode)
		if strings.TrimSpace(mode) == "" {
			mode = string(app.ContractDecodeModeDecoded)
		}
		parts = append(parts, "decode="+mode)
	}
	if lookup.ExpandAll {
		parts = append(parts, "expanded")
	}
	if lookup.VisualMode {
		parts = append(parts, lipgloss.NewStyle().Bold(true).Render("visual"))
	}
	if len(parts) == 0 {
		return ""
	}
	return truncate("Workspace: "+strings.Join(parts, "  ")+"  (c raw  d decoded  e expand  v visual)", width)
}

func lookupContractSummarySections(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return []lookupSection{{Title: "Empty", Body: "No contract payload available.", Muted: true}}
	}
	contract := lookup.Contract.Contract
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Contract:", Body: contract.ContractID, Emph: true},
		{Title: "Type", Body: fmt.Sprintf("%d", contract.ContractType)},
		{Title: "Label", Body: derefString(contract.Label)},
		{Title: "WASM", Body: derefString(contract.WasmHash)},
		{Title: "Storage", Body: fmt.Sprintf("%d entries", contract.StorageEntryCount)},
		{Title: "Invocations", Body: fmt.Sprintf("%d", contract.InvocationCount)},
		{Title: "Events", Body: fmt.Sprintf("%d", contract.EventCount)},
		{Title: "SEP-41/50", Body: fmt.Sprintf("%t / %t", contract.IsSep41Token, contract.IsSep50NFT)},
	}
	if contract.CreatorAccount != nil && strings.TrimSpace(*contract.CreatorAccount) != "" {
		lines = append(lines, lookupSection{Title: "Creator", Body: *contract.CreatorAccount, Command: "lookup account " + *contract.CreatorAccount, Hint: "enter open creator account"})
	}
	return lines
}

func lookupContractOverviewSections(lookup app.LookupSnapshot) []lookupSection {
	lines := lookupContractSummarySections(lookup)
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return lines
	}
	contract := lookup.Contract.Contract
	lines = append(lines, sectionHeader("Workspace"))
	if contract.StorageEntryCount > 0 || len(lookup.Contract.Storage) > 0 {
		lines = append(lines, lookupSection{Title: "Storage", Body: fmt.Sprintf("%d entries", contract.StorageEntryCount), Command: "open storage", Hint: "enter open storage explorer"})
	}
	if contract.EventCount > 0 || len(lookup.Contract.RecentEvents) > 0 {
		lines = append(lines, lookupSection{Title: "Events", Body: fmt.Sprintf("%d events", contract.EventCount), Command: "open events", Hint: "enter open events explorer"})
	}
	if contract.InvocationCount > 0 {
		lines = append(lines, lookupSection{Title: "Invocations", Body: fmt.Sprintf("%d rows", contract.InvocationCount), Command: "open invocations", Hint: "enter open invocations explorer"})
	}
	if len(lookup.Contract.RecentTransactions) > 0 {
		lines = append(lines, lookupSection{Title: "Transactions", Body: fmt.Sprintf("%d rows", len(lookup.Contract.RecentTransactions)), Command: "open txs", Hint: "enter open transaction explorer"})
	}
	if lookup.Contract.Spec != nil && lookup.Contract.Spec.Available {
		lines = append(lines, lookupSection{Title: "Spec", Body: fmt.Sprintf("%d methods", lookup.Contract.Spec.FunctionCount), Command: "open decode decoded", Hint: "tab to spec pane or use open decode"})
	}
	lines = append(lines, lookupContractTokenMetadataSections(lookup.Contract.Contract)...)
	return lines
}

func lookupContractTokenMetadataSections(contract *backendclient.ContractDetail) []lookupSection {
	if contract == nil {
		return nil
	}
	var lines []lookupSection
	if contract.TokenSymbol != nil && strings.TrimSpace(*contract.TokenSymbol) != "" {
		lines = append(lines, sectionHeader("Token Metadata"), lookupSection{Title: "Symbol", Body: *contract.TokenSymbol})
	}
	if contract.TokenName != nil && strings.TrimSpace(*contract.TokenName) != "" {
		if !hasSectionHeader(lines, "Token Metadata") {
			lines = append(lines, sectionHeader("Token Metadata"))
		}
		lines = append(lines, lookupSection{Title: "Name", Body: *contract.TokenName})
	}
	if contract.TokenDecimals != nil {
		if !hasSectionHeader(lines, "Token Metadata") {
			lines = append(lines, sectionHeader("Token Metadata"))
		}
		lines = append(lines, lookupSection{Title: "Decimals", Body: fmt.Sprintf("%d", *contract.TokenDecimals)})
	}
	return lines
}

func lookupContractTabSections(lookup app.LookupSnapshot) []lookupSection {
	tab := lookup.ContractTab
	if tab == "" {
		tab = app.ContractWorkspaceTabOverview
	}
	switch tab {
	case app.ContractWorkspaceTabSpec:
		if lookup.Contract == nil {
			return lookupContractOverviewSections(lookup)
		}
		return append(lookupContractSummarySections(lookup), contractSpecSections(lookup.Contract.Spec, lookup.DecodeMode)...)
	case app.ContractWorkspaceTabStorage:
		return lookupContractStorageTabSections(lookup)
	case app.ContractWorkspaceTabEvents:
		return lookupContractEventsTabSections(lookup)
	case app.ContractWorkspaceTabInvocations:
		return lookupContractInvocationsTabSections(lookup)
	case app.ContractWorkspaceTabTransactions:
		return lookupContractTransactionsTabSections(lookup)
	default:
		return lookupContractOverviewSections(lookup)
	}
}

func lookupContractStorageTabSections(lookup app.LookupSnapshot) []lookupSection {
	lines := lookupContractSummarySections(lookup)
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return lines
	}
	contract := lookup.Contract.Contract
	lines = append(lines,
		sectionHeader("Storage Explorer"),
		lookupSection{Title: "Open Storage", Body: fmt.Sprintf("%d rows", contract.StorageEntryCount), Command: "open storage", Hint: "enter open dedicated storage list"},
	)
	lines = append(lines, contractStorageSections(lookup.Contract.Storage, lookup.DecodeMode)...)
	return lines
}

func lookupContractEventsTabSections(lookup app.LookupSnapshot) []lookupSection {
	lines := lookupContractSummarySections(lookup)
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return lines
	}
	contract := lookup.Contract.Contract
	lines = append(lines,
		sectionHeader("Event Explorer"),
		lookupSection{Title: "Open Events", Body: fmt.Sprintf("%d rows", contract.EventCount), Command: "open events", Hint: "enter open dedicated event list"},
		lookupSection{Title: "Timeline: Events", Body: "event activity only", Command: "open timeline type event", Hint: "enter open contract event timeline"},
	)
	if len(lookup.Contract.RecentEvents) > 0 {
		lines = append(lines, sectionHeader(fmt.Sprintf("Recent Events (%d)", len(lookup.Contract.RecentEvents))))
	}
	for index, event := range lookup.Contract.RecentEvents {
		lines = append(lines, relatedEntityRow(
			fmt.Sprintf("Event %d", index+1),
			summarizeContractEvent(event, lookup.DecodeMode),
			contractEventCopyValue(event),
			eventOpenCommand(index),
			"enter open event detail",
		))
		if !lookup.ExpandAll && index >= 11 {
			lines = append(lines, lookupSection{Title: "More Events", Body: fmt.Sprintf("%d additional", len(lookup.Contract.RecentEvents)-index-1), Muted: true})
			break
		}
	}
	return lines
}

func lookupContractInvocationsTabSections(lookup app.LookupSnapshot) []lookupSection {
	lines := lookupContractSummarySections(lookup)
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return lines
	}
	contract := lookup.Contract.Contract
	lines = append(lines,
		sectionHeader("Invocation Explorer"),
		lookupSection{Title: "Open Invocations", Body: fmt.Sprintf("%d rows", contract.InvocationCount), Command: "open invocations", Hint: "enter open dedicated invocation list"},
		lookupSection{Title: "Timeline: Transactions", Body: "invocation activity via transactions", Command: "open timeline type tx", Hint: "enter open contract transaction timeline"},
	)
	return lines
}

func lookupContractTransactionsTabSections(lookup app.LookupSnapshot) []lookupSection {
	lines := lookupContractSummarySections(lookup)
	if lookup.Contract == nil || lookup.Contract.Contract == nil {
		return lines
	}
	lines = append(lines,
		sectionHeader("Explorer"),
		lookupSection{Title: "Open Timeline", Body: "transactions + events", Command: "open timeline", Hint: "enter open contract activity timeline"},
		lookupSection{Title: "Open Transactions", Body: fmt.Sprintf("%d rows", len(lookup.Contract.RecentTransactions)), Command: "open txs", Hint: "enter open dedicated transaction list"},
	)
	if len(lookup.Contract.RecentTransactions) > 0 {
		lines = append(lines, sectionHeader(fmt.Sprintf("Recent Transactions (%d)", len(lookup.Contract.RecentTransactions))))
	}
	for index, tx := range lookup.Contract.RecentTransactions {
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Contract tx %d", index+1),
			Body:    summarizeTransactionSummary(tx),
			Copy:    tx.Hash,
			Command: "lookup tx " + tx.Hash,
			Hint:    "enter open contract transaction",
		})
		if !lookup.ExpandAll && index >= 11 {
			lines = append(lines, lookupSection{Title: "More Transactions", Body: fmt.Sprintf("%d additional", len(lookup.Contract.RecentTransactions)-index-1), Muted: true})
			break
		}
	}
	return lines
}
