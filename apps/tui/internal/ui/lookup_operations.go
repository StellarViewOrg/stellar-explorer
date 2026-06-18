package ui

import (
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func lookupOperationDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Operation == nil {
		return []lookupSection{{Title: "Empty", Body: "No operation payload available.", Muted: true}}
	}

	op := lookup.Operation.Operation
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Operation:", Body: fmt.Sprintf("%s #%d", op.TypeName, op.ApplicationOrder), Emph: true},
		{Title: "Type", Body: fmt.Sprintf("%s (%d)", op.TypeName, op.Type)},
		{Title: "Parent Tx", Body: lookup.Operation.ParentTransactionHash, Copy: lookup.Operation.ParentTransactionHash, Command: "lookup tx " + lookup.Operation.ParentTransactionHash, Hint: "enter open parent transaction"},
		{Title: "Created", Body: renderTimestamp(op.CreatedAt)},
	}

	if op.SourceAccount != nil && strings.TrimSpace(*op.SourceAccount) != "" {
		lines = append(lines, relatedEntityRow("Source", strings.TrimSpace(*op.SourceAccount), strings.TrimSpace(*op.SourceAccount), "lookup account "+strings.TrimSpace(*op.SourceAccount), "enter open source account"))
	}
	if op.Destination != nil && strings.TrimSpace(*op.Destination) != "" {
		lines = append(lines, relatedEntityRow("Destination", strings.TrimSpace(*op.Destination), strings.TrimSpace(*op.Destination), "lookup account "+strings.TrimSpace(*op.Destination), "enter open destination account"))
	}
	if op.AssetCode != nil && op.AssetIssuer != nil && strings.TrimSpace(*op.AssetCode) != "" && strings.TrimSpace(*op.AssetIssuer) != "" {
		asset := strings.TrimSpace(*op.AssetCode) + ":" + strings.TrimSpace(*op.AssetIssuer)
		lines = append(lines, relatedEntityRow("Asset", asset, asset, "lookup asset "+asset, "enter open operation asset"))
	}
	if op.Amount != nil && strings.TrimSpace(*op.Amount) != "" {
		lines = append(lines, lookupSection{Title: "Amount", Body: strings.TrimSpace(*op.Amount)})
	}
	if op.ContractID != nil && strings.TrimSpace(*op.ContractID) != "" {
		lines = append(lines, relatedEntityRow("Contract", strings.TrimSpace(*op.ContractID), strings.TrimSpace(*op.ContractID), "lookup contract "+strings.TrimSpace(*op.ContractID), "enter open operation contract"))
	}
	if op.FunctionName != nil && strings.TrimSpace(*op.FunctionName) != "" {
		lines = append(lines, lookupSection{Title: "Function", Body: strings.TrimSpace(*op.FunctionName), Emph: true})
	}
	if details := strings.TrimSpace(op.Details); details != "" && op.TypeName != "invoke_host_function" {
		lines = append(lines, sectionHeader("Details"), lookupSection{Title: "Payload", Body: truncate(details, 72)})
	}
	lines = appendSorobanOperationSections(lines, op, lookup.DecodeMode, lookupDisplayLimit(lookup))

	lines = append(lines, sectionHeader("Relations"))
	lines = append(lines, relatedEntityRow("Parent Transaction", truncate(lookup.Operation.ParentTransactionHash, 24), lookup.Operation.ParentTransactionHash, "lookup tx "+lookup.Operation.ParentTransactionHash, "enter open parent transaction"))
	return lines
}

func appendTransactionEffectsSections(lines []lookupSection, lookup app.LookupSnapshot) []lookupSection {
	if lookup.Transaction == nil {
		return lines
	}

	lines = append(lines, sectionHeader("Effects"))
	if isIndexedLookupSource(lookup.Source) {
		effects := lookup.Transaction.Effects
		if len(effects) == 0 {
			lines = append(lines, lookupSection{Title: "Indexed", Body: "No effects indexed for this transaction.", Muted: true})
			return lines
		}
		for index, effect := range effects {
			body := effect.TypeName
			if details := strings.TrimSpace(effect.Details); details != "" {
				body += "  " + truncate(details, 40)
			}
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Effect %d", index+1),
				body,
				effect.Account,
				"lookup account "+strings.TrimSpace(effect.Account),
				"enter open affected account",
			))
		}
		return lines
	}

	lines = append(lines, lookupSection{Title: "Unavailable", Body: "Effects require indexed backend data.", Muted: true})
	return lines
}

func isIndexedLookupSource(source app.SourceMetadata) bool {
	label := strings.ToLower(strings.TrimSpace(source.Label))
	actual := strings.ToLower(strings.TrimSpace(source.Actual))
	preferred := strings.ToLower(strings.TrimSpace(source.Preferred))
	for _, value := range []string{label, actual, preferred} {
		if strings.Contains(value, "indexer") || strings.Contains(value, "hybrid") {
			return true
		}
	}
	return false
}

func operationOpenCommand(index int) string {
	return fmt.Sprintf("open op %d", index+1)
}

func summarizeOperationRich(op backendclient.OperationSummary) string {
	parts := []string{op.TypeName}
	if op.FunctionName != nil && strings.TrimSpace(*op.FunctionName) != "" {
		parts = append(parts, strings.TrimSpace(*op.FunctionName))
	}
	if op.AssetCode != nil && op.AssetIssuer != nil && strings.TrimSpace(*op.AssetCode) != "" && strings.TrimSpace(*op.AssetIssuer) != "" {
		parts = append(parts, strings.TrimSpace(*op.AssetCode)+":"+truncate(strings.TrimSpace(*op.AssetIssuer), 8))
	}
	if op.Amount != nil && strings.TrimSpace(*op.Amount) != "" {
		parts = append(parts, strings.TrimSpace(*op.Amount))
	}
	if op.Destination != nil && strings.TrimSpace(*op.Destination) != "" {
		parts = append(parts, "to "+truncate(strings.TrimSpace(*op.Destination), 10))
	}
	if details := strings.TrimSpace(op.Details); details != "" && !strings.Contains(strings.Join(parts, " "), truncate(details, 12)) {
		parts = append(parts, truncate(details, 24))
	}
	if strings.TrimSpace(op.TransactionHash) != "" {
		return truncate(truncate(op.TransactionHash, 14)+"  "+strings.Join(parts, "  "), 72)
	}
	return truncate(strings.Join(parts, "  "), 72)
}
