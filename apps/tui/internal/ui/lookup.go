package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

var (
	breadcrumbSeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	breadcrumbActiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	breadcrumbHistoryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

func formatBreadcrumbQuery(query string) string {
	query = strings.TrimSpace(query)
	if strings.Contains(query, ":") {
		parts := strings.SplitN(query, ":", 2)
		code := parts[0]
		issuer := parts[1]
		if len(issuer) > 8 {
			issuer = issuer[:4] + "..." + issuer[len(issuer)-4:]
		}
		return code + ":" + issuer
	}
	if len(query) > 12 {
		return query[:6] + "..." + query[len(query)-4:]
	}
	return query
}

func renderBreadcrumbs(snapshot app.Snapshot, width int) string {
	if len(snapshot.LookupRoute) == 0 {
		return ""
	}

	parts := make([]string, 0, len(snapshot.LookupRoute))
	for index, step := range snapshot.LookupRoute {
		label := routeStepLabel(step, index == len(snapshot.LookupRoute)-1)
		if index == len(snapshot.LookupRoute)-1 {
			parts = append(parts, breadcrumbActiveStyle.Render(label))
			continue
		}
		parts = append(parts, breadcrumbHistoryStyle.Render(label))
	}

	separator := breadcrumbSeparatorStyle.Render(" › ")
	crumbs := strings.Join(parts, separator)

	if lipgloss.Width(crumbs) > width && len(parts) > 1 {
		if len(parts) > 3 {
			truncatedParts := append([]string{breadcrumbHistoryStyle.Render("...")}, parts[len(parts)-2:]...)
			crumbs = strings.Join(truncatedParts, separator)
		}
	}

	return crumbs
}

type lookupSection struct {
	Title   string
	Body    string
	Copy    string
	Command string
	Hint    string
	Emph    bool
	Muted   bool
	Divider bool
}

func (s lookupSection) ClipboardValue() string {
	if !s.Selectable() {
		return ""
	}
	if strings.TrimSpace(s.Copy) != "" {
		return strings.TrimSpace(s.Copy)
	}
	return strings.TrimSpace(s.Body)
}

func (s lookupSection) Selectable() bool {
	return !s.Divider && !s.Muted
}

func (s lookupSection) Navigable() bool {
	return strings.TrimSpace(s.Command) != ""
}

func sectionIsSelectable(section lookupSection, visualMode bool) bool {
	if !section.Selectable() {
		return false
	}
	if visualMode {
		return section.Navigable()
	}
	return true
}

func lookupView(snapshot app.Snapshot, width, height int, offset int, selectedSection int) []string {
	sections := lookupSections(snapshot.Lookup)
	lines := []string{
		tableHeaderStyle.Render("Commands"),
		"lookup ledger <sequence>",
		"lookup tx <hash>",
		"lookup op <txhash>:<index>",
		"lookup account <id>",
		"lookup asset <code:issuer>",
		"lookup contract <id>",
		"open ledgers|accounts|assets|contracts",
		"open txs|ops|op <n>|holders|events|storage|invocations",
		"open event <n>|open storage <n>|open invocation <n>",
		"open decode raw|decoded",
		"m bookmark palette  , quick bookmark  - remove bookmark",
		"; note palette  . label palette  /bookmark|note|label add|remove",
		"",
	}
	lines = append(lines, renderSourceSummary(snapshot.Lookup.Source, width)...)

	if snapshot.Lookup.Query == "" {
		return append(lines, "", "No lookup loaded yet.")
	}

	if crumbs := renderBreadcrumbs(snapshot, width); crumbs != "" {
		lines = append(lines, "", "Route: "+crumbs)
	} else {
		lines = append(lines, "", "Last query: "+string(snapshot.Lookup.Kind)+" "+snapshot.Lookup.Query)
	}
	if snapshot.Lookup.ReturnContext != nil && strings.TrimSpace(snapshot.Lookup.ReturnContext.Label) != "" {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("Return: b or live return -> %s", snapshot.Lookup.ReturnContext.Label)))
	}
	if tabBar := renderContractWorkspaceTabBar(snapshot.Lookup, width); tabBar != "" {
		lines = append(lines, tabBar)
	}
	if modes := renderLookupWorkspaceModes(snapshot.Lookup, width); modes != "" {
		lines = append(lines, modes)
	} else if snapshot.Lookup.State == app.ViewStateReady && supportsContractDecodeMode(snapshot.Lookup.Kind) {
		mode := string(snapshot.Lookup.DecodeMode)
		if strings.TrimSpace(mode) == "" {
			mode = string(app.ContractDecodeModeDecoded)
		}
		lines = append(lines, "Contract decode: "+mode+"  (c raw  d decoded  or open decode raw|decoded)")
	}
	switch snapshot.Lookup.State {
	case app.ViewStateLoading:
		return append(lines, renderPanelState("State: loading", "Waiting for backend response...", panelStateWarn, width)...)
	case app.ViewStateError:
		return append(lines, renderPanelState("State: error", valueOrFallback(snapshot.Lookup.Error, "unknown lookup error"), panelStateError, width)...)
	case app.ViewStateEmpty:
		return append(lines, renderPanelState("State: empty", "No entity matched the current lookup.", panelStateInfo, width)...)
	}

	lines = append(lines,
		"",
		statusInfoStyle.Render("State: ready"),
		mutedStyle.Render("j/k move  pgup/pgdn scroll  home/end jump  enter follow selection"),
	)
	if hint := selectedLookupActionHint(snapshot, selectedSection); strings.TrimSpace(hint) != "" {
		lines = append(lines, mutedStyle.Render(hint))
	}
	lines = append(lines, renderLookupSections(sections, width, height-len(lines), offset, selectedSection, snapshot.Lookup.VisualMode)...)
	if len(lines) > height {
		return lines[:height]
	}
	return lines
}

func lookupSections(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Explorer != nil {
		return lookupExplorerSections(lookup)
	}
	var sections []lookupSection
	switch lookup.Kind {
	case app.LookupLedger:
		sections = lookupLedgerDetails(lookup)
	case app.LookupTransaction:
		sections = lookupTransactionDetails(lookup)
	case app.LookupOperation:
		sections = lookupOperationDetails(lookup)
	case app.LookupAccount:
		sections = lookupAccountDetails(lookup)
	case app.LookupAsset:
		sections = lookupAssetDetails(lookup)
	case app.LookupContract:
		sections = lookupContractDetails(lookup)
	case app.LookupEvent:
		sections = lookupEventDetails(lookup)
	case app.LookupStorage:
		sections = lookupStorageEntryDetails(lookup)
	default:
		return nil
	}
	return appendLookupMetadataSections(sections, lookup.Metadata)
}

func appendLookupMetadataSections(lines []lookupSection, metadata app.LookupMetadataSnapshot) []lookupSection {
	if len(metadata.Labels) == 0 && len(metadata.Bookmarks) == 0 && len(metadata.Notes) == 0 && metadata.Cached == nil {
		return lines
	}
	lines = append(lines, sectionHeader("Local Investigation"))
	if metadata.Cached != nil {
		body := "cached"
		if !metadata.Cached.UpdatedAt.IsZero() {
			body = "cached " + renderTimestamp(metadata.Cached.UpdatedAt)
		}
		if strings.TrimSpace(metadata.Cached.SourceLabel) != "" {
			body += "  " + strings.TrimSpace(metadata.Cached.SourceLabel)
		}
		lines = append(lines, lookupSection{
			Title:   "Revisit",
			Body:    truncate(body, 72),
			Command: "open cache",
			Hint:    "enter reload cached payload",
		})
	}
	for _, label := range metadata.Labels {
		body := strings.TrimSpace(label.Name)
		if strings.TrimSpace(label.Color) != "" {
			body += "  " + strings.TrimSpace(label.Color)
		}
		lines = append(lines, lookupSection{Title: "Label", Body: body})
	}
	for _, bookmark := range metadata.Bookmarks {
		body := strings.TrimSpace(bookmark.Title)
		if strings.TrimSpace(bookmark.Notes) != "" {
			body += "  " + truncate(strings.TrimSpace(bookmark.Notes), 48)
		}
		lines = append(lines, lookupSection{Title: "Bookmark", Body: body})
	}
	for _, note := range metadata.Notes {
		body := strings.TrimSpace(note.Title)
		if strings.TrimSpace(note.Body) != "" {
			body += "  " + truncate(strings.TrimSpace(note.Body), 48)
		}
		lines = append(lines, lookupSection{Title: "Note", Body: body})
	}
	return lines
}

func lookupExplorerSections(lookup app.LookupSnapshot) []lookupSection {
	explorer := lookup.Explorer
	if explorer == nil {
		return nil
	}
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Explorer"),
	}
	if header := explorerContextHeader(lookup); header.Title != "" {
		lines = append(lines, header)
	}
	lines = append(lines, lookupSection{Title: "View", Body: valueOrFallback(strings.TrimSpace(explorer.Title), "Transactions"), Emph: true})
	if strings.TrimSpace(explorer.BackCommand) != "" || strings.TrimSpace(explorer.ParentLabel) != "" {
		lines = append(lines, lookupSection{
			Title:   "Back",
			Body:    valueOrFallback(strings.TrimSpace(explorer.ParentLabel), "detail"),
			Command: valueOrFallback(strings.TrimSpace(explorer.BackCommand), "open detail"),
			Hint:    "enter return to parent detail",
		})
	}
	switch explorer.Kind {
	case app.LookupExplorerTransactions:
		lines = append(lines, sectionHeader(fmt.Sprintf("Transactions (%d)", len(explorer.Transactions))))
		for index, tx := range explorer.Transactions {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Tx %d", index+1),
				summarizeTransactionSummary(tx),
				tx.Hash,
				"lookup tx "+tx.Hash,
				"enter open transaction detail",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerOperations:
		lines = append(lines, sectionHeader(fmt.Sprintf("Operations (%d)", len(explorer.Operations))))
		for index, op := range explorer.Operations {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Op %d", index+1),
				summarizeOperationRich(op),
				op.TransactionHash,
				operationOpenCommand(index),
				"enter open operation detail",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerHolders:
		lines = append(lines, sectionHeader(fmt.Sprintf("Holders (%d)", len(explorer.Holders))))
		for index, holder := range explorer.Holders {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Holder %d", index+1),
				summarizeAssetHolder(holder),
				holder.AccountID,
				"lookup account "+holder.AccountID,
				"enter open holder account",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerEvents:
		lines = append(lines, sectionHeader(fmt.Sprintf("Events (%d)", len(explorer.Events))))
		for index, event := range explorer.Events {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Event %d", index+1),
				summarizeContractEvent(event, lookup.DecodeMode),
				contractEventCopyValue(event),
				eventOpenCommand(index),
				"enter open event detail",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerStorage:
		lines = append(lines, sectionHeader(fmt.Sprintf("Storage (%d)", len(explorer.Storage))))
		for index, entry := range explorer.Storage {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Storage %d", index+1),
				summarizeContractStorage(entry, lookup.DecodeMode),
				contractStorageCopyValue(entry),
				storageOpenCommand(index),
				"enter open storage detail",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerInvocations:
		lines = append(lines, sectionHeader(fmt.Sprintf("Invocations (%d)", len(explorer.Operations))))
		for index, op := range explorer.Operations {
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("Invocation %d", index+1),
				summarizeOperationRich(op),
				op.TransactionHash,
				invocationOpenCommand(index),
				"enter open invocation detail",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	case app.LookupExplorerResults:
		lines = append(lines, sectionHeader(fmt.Sprintf("Results (%d)", len(explorer.Results))))
		for index, result := range explorer.Results {
			copyValue := strings.TrimSpace(result.Description)
			if copyValue == "" {
				copyValue = strings.TrimSpace(result.Command)
			}
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("%s %d", titleOrDefault(result.Kind, "Entity"), index+1),
				summarizeSearchResult(result),
				copyValue,
				result.Command,
				"enter open entity detail",
			))
		}
	case app.LookupExplorerTimeline:
		lines = append(lines, sectionHeader(fmt.Sprintf("Timeline (%d)", len(explorer.Results))))
		for index, result := range explorer.Results {
			copyValue := strings.TrimSpace(result.Description)
			if copyValue == "" {
				copyValue = strings.TrimSpace(result.Command)
			}
			lines = append(lines, relatedEntityRow(
				fmt.Sprintf("%s %d", titleOrDefault(result.Kind, "Item"), index+1),
				summarizeSearchResult(result),
				copyValue,
				result.Command,
				"enter open timeline item",
			))
		}
		lines = appendExplorerNextPage(lines, explorer.NextCommand)
	}
	return lines
}

func appendExplorerNextPage(lines []lookupSection, command string) []lookupSection {
	command = strings.TrimSpace(command)
	if command == "" {
		return lines
	}
	return append(lines, lookupSection{
		Title:   "Next Page",
		Body:    "load more rows",
		Command: command,
		Hint:    "enter load next page",
	})
}

func lookupLedgerDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Ledger == nil || lookup.Ledger.Ledger == nil {
		return []lookupSection{{Title: "Empty", Body: "No ledger payload available.", Muted: true}}
	}
	ledger := lookup.Ledger.Ledger
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Ledger:", Body: fmt.Sprintf("%d", ledger.Sequence), Emph: true},
		{Title: "Hash", Body: ledger.Hash},
		{Title: "Closed", Body: renderTimestamp(ledger.ClosedAt)},
		{Title: "Counts", Body: fmt.Sprintf("%d tx / %d ops", ledger.TransactionCount, ledger.OperationCount)},
		{Title: "Results", Body: fmt.Sprintf("%d successful / %d failed", ledger.SuccessfulTxCount, ledger.FailedTxCount)},
		sectionHeader("Context"),
		{Title: "Previous", Body: fmt.Sprintf("%d", max(1, int(ledger.Sequence)-1)), Command: fmt.Sprintf("lookup ledger %d", max(1, int(ledger.Sequence)-1)), Hint: "enter open previous ledger"},
		{Title: "Next", Body: fmt.Sprintf("%d", ledger.Sequence+1), Command: fmt.Sprintf("lookup ledger %d", ledger.Sequence+1), Hint: "enter try next ledger"},
	}
	if len(lookup.Ledger.Transactions) > 0 {
		lines = append(lines,
			sectionHeader("Explorer"),
			lookupSection{Title: "Open Transactions", Body: fmt.Sprintf("%d rows", len(lookup.Ledger.Transactions)), Command: "open txs", Hint: "enter open dedicated transaction list"},
			sectionHeader(fmt.Sprintf("Transactions (%d)", len(lookup.Ledger.Transactions))),
		)
	}
	for index, tx := range lookup.Ledger.Transactions {
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Tx %d", index+1),
			Body:    summarizeTransactionSummary(tx),
			Copy:    tx.Hash,
			Command: "lookup tx " + tx.Hash,
			Hint:    "enter open ledger transaction",
		})
	}
	return lines
}

func lookupTransactionDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Transaction == nil || lookup.Transaction.Transaction == nil {
		return []lookupSection{{Title: "Empty", Body: "No transaction payload available.", Muted: true}}
	}
	tx := lookup.Transaction.Transaction
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Transaction:", Body: tx.Hash, Emph: true},
		{Title: "Status", Body: fmt.Sprintf("%d", tx.Status)},
		{Title: "Ledger", Body: fmt.Sprintf("%d", tx.LedgerSequence), Command: fmt.Sprintf("lookup ledger %d", tx.LedgerSequence), Hint: "enter open parent ledger"},
		{Title: "Ops", Body: fmt.Sprintf("%d", tx.OperationCount)},
		{Title: "Source", Body: tx.Account, Command: "lookup account " + tx.Account, Hint: "enter open source account"},
		sectionHeader("Navigation"),
		{Title: "Created", Body: renderTimestamp(tx.CreatedAt)},
		{Title: "Soroban", Body: fmt.Sprintf("%t", tx.IsSoroban)},
		sectionHeader("Execution"),
		{Title: "Source Seq", Body: fmt.Sprintf("%d", tx.AccountSequence)},
		{Title: "Fees", Body: fmt.Sprintf("charged %d / max %d", tx.FeeCharged, tx.MaxFee)},
		{Title: "Operations", Body: fmt.Sprintf("%d", len(lookup.Transaction.Operations))},
	}
	if tx.AccountMuxed != nil && strings.TrimSpace(*tx.AccountMuxed) != "" {
		lines = append(lines, lookupSection{Title: "Muxed Source", Body: *tx.AccountMuxed})
	}
	if tx.AccountMuxedID != nil {
		lines = append(lines, lookupSection{Title: "Muxed ID", Body: fmt.Sprintf("%d", *tx.AccountMuxedID)})
	}
	if memo := summarizeTransactionMemo(tx); memo != "" {
		lines = append(lines, lookupSection{Title: "Memo", Body: memo})
	}
	if tx.IsSoroban && tx.SorobanResources != nil && strings.TrimSpace(*tx.SorobanResources) != "" {
		lines = append(lines, lookupSection{Title: "Resources", Body: truncate(strings.TrimSpace(*tx.SorobanResources), 72)})
	}
	if len(lookup.Transaction.Operations) > 0 {
		lines = append(lines,
			sectionHeader("Explorer"),
			lookupSection{Title: "Open Operations", Body: fmt.Sprintf("%d rows", len(lookup.Transaction.Operations)), Command: "open ops", Hint: "enter open dedicated operation list"},
			sectionHeader(fmt.Sprintf("Operations (%d)", len(lookup.Transaction.Operations))),
		)
	}
	for index, op := range lookup.Transaction.Operations {
		lines = append(lines, relatedEntityRow(
			fmt.Sprintf("Op %d", index+1),
			summarizeOperationRich(op),
			op.TransactionHash,
			operationOpenCommand(index),
			"enter open operation detail",
		))
	}
	return appendTransactionEffectsSections(lines, lookup)
}

func lookupAccountDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Account == nil || lookup.Account.Account == nil {
		return []lookupSection{{Title: "Empty", Body: "No account payload available.", Muted: true}}
	}
	account := lookup.Account.Account
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Account:", Body: account.ID, Emph: true},
		{Title: "Sequence", Body: fmt.Sprintf("%d", account.Sequence)},
		{Title: "Seq Ledger", Body: optionalInt64Label(account.SequenceLedger)},
		{Title: "Seq Time", Body: optionalTimeLabel(account.SequenceTime)},
		{Title: "Balance", Body: account.Balance},
		{Title: "Liabilities", Body: fmt.Sprintf("buy %s / sell %s", account.BuyingLiabilities, account.SellingLiabilities)},
		{Title: "Subentries", Body: fmt.Sprintf("%d", account.NumSubentries)},
		{Title: "Flags", Body: fmt.Sprintf("%d", account.Flags)},
		{Title: "Trustlines:", Body: fmt.Sprintf("%d | Signers: %d", len(lookup.Account.Trustlines), len(lookup.Account.Signers))},
	}
	if account.HomeDomain != nil && strings.TrimSpace(*account.HomeDomain) != "" {
		lines = append(lines, lookupSection{Title: "Home", Body: *account.HomeDomain})
	}
	if account.Thresholds != nil && strings.TrimSpace(*account.Thresholds) != "" {
		lines = append(lines, lookupSection{Title: "Thresholds", Body: *account.Thresholds})
	}
	if account.Sponsor != nil && strings.TrimSpace(*account.Sponsor) != "" {
		lines = append(lines, lookupSection{Title: "Sponsor", Body: *account.Sponsor, Command: "lookup account " + *account.Sponsor, Hint: "enter open sponsor account"})
	}
	if len(lookup.Account.RecentTransactions) > 0 {
		lines = append(lines,
			sectionHeader("Explorer"),
			lookupSection{Title: "Open Timeline", Body: "transactions + operations", Command: "open timeline", Hint: "enter open account activity timeline"},
			lookupSection{Title: "Timeline: Transactions", Body: "transaction activity only", Command: "open timeline type tx", Hint: "enter open account transaction timeline"},
			lookupSection{Title: "Open Transactions", Body: fmt.Sprintf("%d rows", len(lookup.Account.RecentTransactions)), Command: "open txs", Hint: "enter open dedicated transaction list"},
			sectionHeader(fmt.Sprintf("Recent Transactions (%d)", len(lookup.Account.RecentTransactions))),
		)
	}
	for index, tx := range lookup.Account.RecentTransactions {
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Account tx %d", index+1),
			Body:    summarizeTransactionSummary(tx),
			Copy:    tx.Hash,
			Command: "lookup tx " + tx.Hash,
			Hint:    "enter open related transaction",
		})
	}
	if len(lookup.Account.RecentOperations) > 0 {
		if len(lookup.Account.RecentTransactions) == 0 {
			lines = append(lines,
				sectionHeader("Explorer"),
				lookupSection{Title: "Open Timeline", Body: "transactions + operations", Command: "open timeline", Hint: "enter open account activity timeline"},
			)
		}
		lines = append(lines, sectionHeader("Operation Explorer"), lookupSection{Title: "Timeline: Operations", Body: "operation activity only", Command: "open timeline type op", Hint: "enter open account operation timeline"}, lookupSection{Title: "Open Operations", Body: fmt.Sprintf("%d rows", len(lookup.Account.RecentOperations)), Command: "open ops", Hint: "enter open dedicated operation list"}, sectionHeader(fmt.Sprintf("Recent Operations (%d)", len(lookup.Account.RecentOperations))))
	}
	for index, op := range lookup.Account.RecentOperations {
		order := int(op.ApplicationOrder)
		if order < 1 {
			order = index + 1
		}
		lines = append(lines, relatedEntityRow(
			fmt.Sprintf("Account op %d", index+1),
			summarizeOperationRich(op),
			op.TransactionHash,
			fmt.Sprintf("lookup op %s:%d", strings.TrimSpace(op.TransactionHash), order),
			"enter open operation detail",
		))
	}
	if len(lookup.Account.Trustlines) > 0 || len(lookup.Account.Signers) > 0 {
		lines = append(lines, sectionHeader("Relations"))
	}
	for index, trustline := range lookup.Account.Trustlines {
		code := strings.TrimSpace(trustline.AssetCode)
		issuer := strings.TrimSpace(trustline.AssetIssuer)
		if code == "" || issuer == "" {
			continue
		}
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Trustline %d", index+1),
			Body:    fmt.Sprintf("%s:%s  balance %s", code, issuer, trustline.Balance),
			Command: fmt.Sprintf("lookup asset %s:%s", code, issuer),
			Hint:    "enter open trusted asset",
		})
	}
	for index, signer := range lookup.Account.Signers {
		body := fmt.Sprintf("%s  weight %d", signer.Type, signer.Weight)
		if signer.Sponsor != nil && strings.TrimSpace(*signer.Sponsor) != "" {
			body = fmt.Sprintf("%s  sponsor %s", body, truncate(*signer.Sponsor, 18))
		}
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Signer %d", index+1),
			Body:    truncate(body, 72),
			Command: signerLookupCommand(signer),
			Hint:    signerLookupHint(signer),
		})
	}
	return lines
}

func lookupAssetDetails(lookup app.LookupSnapshot) []lookupSection {
	if lookup.Asset == nil || lookup.Asset.Asset == nil {
		return []lookupSection{{Title: "Empty", Body: "No asset payload available.", Muted: true}}
	}
	asset := lookup.Asset.Asset
	lines := []lookupSection{
		{Divider: true},
		sectionHeader("Summary"),
		{Title: "Asset:", Body: asset.AssetCode + ":" + asset.AssetIssuer, Emph: true},
		{Title: "Issuer", Body: asset.AssetIssuer, Command: "lookup account " + asset.AssetIssuer, Hint: "enter open issuer account"},
		{Title: "Accounts", Body: fmt.Sprintf("%d", asset.NumAccounts)},
		{Title: "Supply", Body: asset.TotalSupply},
		{Title: "Pools", Body: fmt.Sprintf("%d", asset.NumLiquidityPools)},
		{Title: "Contracts", Body: fmt.Sprintf("%d", asset.NumContracts)},
		{Title: "Auth", Body: fmt.Sprintf("required=%t revocable=%t immutable=%t clawback=%t", asset.AuthRequired, asset.AuthRevocable, asset.AuthImmutable, asset.ClawbackEnabled)},
	}
	if asset.HomeDomain != nil && strings.TrimSpace(*asset.HomeDomain) != "" {
		lines = append(lines, lookupSection{Title: "Home", Body: *asset.HomeDomain})
	}
	if asset.SACContractID != nil && strings.TrimSpace(*asset.SACContractID) != "" {
		lines = append(lines, lookupSection{Title: "SAC", Body: *asset.SACContractID, Command: "lookup contract " + *asset.SACContractID, Hint: "enter open SAC contract"})
	}
	if len(lookup.Asset.RecentTransactions) > 0 {
		lines = append(lines,
			sectionHeader("Explorer"),
			lookupSection{Title: "Open Timeline", Body: "transactions + holders", Command: "open timeline", Hint: "enter open asset activity timeline"},
			lookupSection{Title: "Timeline: Transactions", Body: "transaction activity only", Command: "open timeline type tx", Hint: "enter open asset transaction timeline"},
			lookupSection{Title: "Open Transactions", Body: fmt.Sprintf("%d rows", len(lookup.Asset.RecentTransactions)), Command: "open txs", Hint: "enter open dedicated transaction list"},
			sectionHeader(fmt.Sprintf("Recent Transactions (%d)", len(lookup.Asset.RecentTransactions))),
		)
	}
	for index, tx := range lookup.Asset.RecentTransactions {
		lines = append(lines, lookupSection{
			Title:   fmt.Sprintf("Asset tx %d", index+1),
			Body:    summarizeTransactionSummary(tx),
			Copy:    tx.Hash,
			Command: "lookup tx " + tx.Hash,
			Hint:    "enter open asset transaction",
		})
	}
	if len(lookup.Asset.TopHolders) > 0 {
		if len(lookup.Asset.RecentTransactions) == 0 {
			lines = append(lines,
				sectionHeader("Explorer"),
				lookupSection{Title: "Open Timeline", Body: "transactions + holders", Command: "open timeline", Hint: "enter open asset activity timeline"},
			)
		}
		lines = append(lines, sectionHeader("Holder Explorer"), lookupSection{Title: "Timeline: Holders", Body: "holder activity only", Command: "open timeline type holder", Hint: "enter open asset holder timeline"}, lookupSection{Title: "Open Holders", Body: fmt.Sprintf("%d rows", len(lookup.Asset.TopHolders)), Command: "open holders", Hint: "enter open dedicated holder list"}, sectionHeader(fmt.Sprintf("Top Holders (%d)", len(lookup.Asset.TopHolders))))
	}
	for index, holder := range lookup.Asset.TopHolders {
		lines = append(lines, relatedEntityRow(
			fmt.Sprintf("Holder %d", index+1),
			summarizeAssetHolder(holder),
			holder.AccountID,
			"lookup account "+holder.AccountID,
			"enter open holder account",
		))
	}
	return lines
}

func lookupContractDetails(lookup app.LookupSnapshot) []lookupSection {
	return lookupContractTabSections(lookup)
}

func contractStorageSections(entries []backendclient.ContractStorageSummary, mode app.ContractDecodeMode) []lookupSection {
	if len(entries) == 0 {
		return nil
	}
	lines := []lookupSection{sectionHeader(fmt.Sprintf("Storage Entries (%d)", len(entries)))}
	for index, entry := range entries {
		if index >= 8 {
			lines = append(lines, lookupSection{Title: "More Storage", Body: fmt.Sprintf("%d additional", len(entries)-index), Muted: true})
			break
		}
		lines = append(lines, relatedEntityRow(
			fmt.Sprintf("Storage %d", index+1),
			summarizeContractStorage(entry, mode),
			contractStorageCopyValue(entry),
			storageOpenCommand(index),
			"enter open storage detail",
		))
	}
	return lines
}

func contractSpecSections(spec *backendclient.ContractSpec, mode app.ContractDecodeMode) []lookupSection {
	if spec == nil {
		return nil
	}
	lines := []lookupSection{
		sectionHeader("Contract Spec"),
		{Title: "Decode", Body: spec.DecodeStatus},
	}
	if !spec.Available {
		lines = append(lines, lookupSection{Title: "Fallback", Body: "contract spec unavailable; raw contract metadata only", Muted: true})
		return lines
	}
	lines = append(lines, lookupSection{Title: "Methods", Body: fmt.Sprintf("%d", spec.FunctionCount)})
	lines = append(lines, lookupSection{Title: "Schemas", Body: fmt.Sprintf("%d", spec.SchemaCount)})
	if spec.EventCount > 0 {
		lines = append(lines, lookupSection{Title: "Events", Body: fmt.Sprintf("%d", spec.EventCount)})
	}
	if spec.SpecXDR != nil && strings.TrimSpace(*spec.SpecXDR) != "" {
		lines = append(lines, lookupSection{Title: "Spec XDR", Body: truncate(strings.TrimSpace(*spec.SpecXDR), 72), Copy: strings.TrimSpace(*spec.SpecXDR), Hint: "copy spec xdr"})
	}
	if mode == app.ContractDecodeModeRaw {
		if spec.Raw != nil && strings.TrimSpace(*spec.Raw) != "" {
			lines = append(lines, lookupSection{Title: "Raw Spec", Body: truncate(strings.TrimSpace(*spec.Raw), 72), Copy: strings.TrimSpace(*spec.Raw), Hint: "copy raw contract spec"})
		}
		return lines
	}
	for index, fn := range spec.Functions {
		if index >= 8 {
			lines = append(lines, lookupSection{Title: "More Methods", Body: fmt.Sprintf("%d additional", len(spec.Functions)-index), Muted: true})
			break
		}
		lines = append(lines, lookupSection{
			Title: fmt.Sprintf("Method %d", index+1),
			Body:  summarizeContractFunction(fn),
		})
	}
	if len(spec.Events) > 0 {
		lines = append(lines, sectionHeader("Spec Events"))
	}
	for index, event := range spec.Events {
		if index >= 6 {
			lines = append(lines, lookupSection{Title: "More Events", Body: fmt.Sprintf("%d additional", len(spec.Events)-index), Muted: true})
			break
		}
		body := event.Name
		if len(event.Params) > 0 {
			body = fmt.Sprintf("%s (%d params)", event.Name, len(event.Params))
		}
		lines = append(lines, lookupSection{Title: fmt.Sprintf("Event %d", index+1), Body: body})
	}
	if len(spec.Schemas) > 0 {
		lines = append(lines, sectionHeader("Schema Types"))
	}
	for index, schema := range spec.Schemas {
		if index >= 6 {
			lines = append(lines, lookupSection{Title: "More Schemas", Body: fmt.Sprintf("%d additional", len(spec.Schemas)-index), Muted: true})
			break
		}
		lines = append(lines, lookupSection{
			Title: schema.Name,
			Body:  schema.Kind,
		})
	}
	if spec.DecodeStatus != "decoded" && spec.Raw != nil && strings.TrimSpace(*spec.Raw) != "" {
		lines = append(lines, lookupSection{Title: "Raw Fallback", Body: truncate(strings.TrimSpace(*spec.Raw), 72), Copy: strings.TrimSpace(*spec.Raw), Hint: "copy raw contract spec"})
	}
	return lines
}

func supportsContractDecodeMode(kind app.LookupKind) bool {
	switch kind {
	case app.LookupContract, app.LookupEvent, app.LookupStorage, app.LookupOperation:
		return true
	default:
		return false
	}
}

func sectionHeader(label string) lookupSection {
	return lookupSection{Body: label, Muted: true}
}

func hasSectionHeader(lines []lookupSection, label string) bool {
	for _, line := range lines {
		if line.Muted && !line.Divider && line.Title == "" && line.Body == label {
			return true
		}
	}
	return false
}

func optionalInt64Label(value *int64) string {
	if value == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d", *value)
}

func optionalTimeLabel(value *time.Time) string {
	if value == nil {
		return "unknown"
	}
	return renderTimestamp(*value)
}

func derefString(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "unknown"
	}
	return *value
}

func signerLookupCommand(signer backendclient.AccountSignerSummary) string {
	if signer.Sponsor != nil && strings.TrimSpace(*signer.Sponsor) != "" {
		return "lookup account " + strings.TrimSpace(*signer.Sponsor)
	}
	return ""
}

func signerLookupHint(signer backendclient.AccountSignerSummary) string {
	if signer.Sponsor != nil && strings.TrimSpace(*signer.Sponsor) != "" {
		return "enter open signer sponsor"
	}
	return ""
}

func summarizeTransactionSummary(tx backendclient.TransactionSummary) string {
	status := fmt.Sprintf("status %d", tx.Status)
	if tx.IsSoroban {
		status += " soroban"
	}
	account := truncate(strings.TrimSpace(tx.Account), 16)
	if account == "" {
		account = "unknown-account"
	}
	when := shortTimestamp(tx.CreatedAt)
	if when == "" {
		return truncate(fmt.Sprintf("%s  %s  %d ops  %s", truncate(tx.Hash, 14), account, tx.OperationCount, status), 72)
	}
	return truncate(fmt.Sprintf("%s  %s  %d ops  %s  %s", truncate(tx.Hash, 14), account, tx.OperationCount, status, when), 72)
}

func summarizeOperation(op backendclient.OperationSummary) string {
	return summarizeOperationRich(op)
}

func summarizeTransactionMemo(tx *backendclient.TransactionDetail) string {
	switch tx.MemoType {
	case 0:
		return "none"
	case 1:
		if tx.MemoText == nil || strings.TrimSpace(*tx.MemoText) == "" {
			return "text"
		}
		return "text  " + truncate(strings.TrimSpace(*tx.MemoText), 56)
	case 2:
		return "id"
	case 3:
		if tx.MemoHash == nil || strings.TrimSpace(*tx.MemoHash) == "" {
			return "hash"
		}
		return "hash  " + truncate(strings.TrimSpace(*tx.MemoHash), 52)
	case 4:
		if tx.MemoHash == nil || strings.TrimSpace(*tx.MemoHash) == "" {
			return "return"
		}
		return "return  " + truncate(strings.TrimSpace(*tx.MemoHash), 50)
	default:
		return fmt.Sprintf("type %d", tx.MemoType)
	}
}

func summarizeContractEvent(event backendclient.ContractEventSummary, mode app.ContractDecodeMode) string {
	if mode == app.ContractDecodeModeRaw {
		parts := []string{truncate(event.TransactionHash, 14), fmt.Sprintf("ledger %d", event.LedgerSequence), fmt.Sprintf("type %d", event.Type), "raw"}
		if event.TopicsXDR != nil && strings.TrimSpace(*event.TopicsXDR) != "" {
			parts = append(parts, "topics "+truncate(strings.TrimSpace(*event.TopicsXDR), 18))
		}
		if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
			parts = append(parts, "value "+truncate(strings.TrimSpace(*event.ValueXDR), 18))
		}
		return truncate(strings.Join(parts, "  "), 72)
	}
	if strings.TrimSpace(event.Summary) != "" {
		return truncate(event.Summary, 72)
	}
	parts := []string{truncate(event.TransactionHash, 14), fmt.Sprintf("ledger %d", event.LedgerSequence), fmt.Sprintf("type %d", event.Type)}
	if event.Topic1 != nil && strings.TrimSpace(*event.Topic1) != "" {
		parts = append(parts, "topic "+strings.TrimSpace(*event.Topic1))
	}
	if event.ValueDecoded != nil && strings.TrimSpace(*event.ValueDecoded) != "" {
		parts = append(parts, "value "+truncate(strings.TrimSpace(*event.ValueDecoded), 20))
	} else if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
		parts = append(parts, "raw value "+truncate(strings.TrimSpace(*event.ValueXDR), 20))
	}
	if strings.TrimSpace(event.DecodeStatus) != "" {
		parts = append(parts, strings.TrimSpace(event.DecodeStatus))
	}
	if when := shortTimestamp(event.CreatedAt); when != "" {
		parts = append(parts, when)
	}
	return truncate(strings.Join(parts, "  "), 72)
}

func contractEventCopyValue(event backendclient.ContractEventSummary) string {
	if event.ValueDecoded != nil && strings.TrimSpace(*event.ValueDecoded) != "" {
		return strings.TrimSpace(*event.ValueDecoded)
	}
	if event.ValueXDR != nil && strings.TrimSpace(*event.ValueXDR) != "" {
		return strings.TrimSpace(*event.ValueXDR)
	}
	return strings.TrimSpace(event.TransactionHash)
}

func summarizeContractStorage(entry backendclient.ContractStorageSummary, mode app.ContractDecodeMode) string {
	parts := []string{}
	if strings.TrimSpace(entry.DurabilityLabel) != "" {
		parts = append(parts, strings.TrimSpace(entry.DurabilityLabel))
	}
	if strings.TrimSpace(entry.DecodeStatus) != "" {
		parts = append(parts, strings.TrimSpace(entry.DecodeStatus))
	}
	if mode == app.ContractDecodeModeRaw {
		if strings.TrimSpace(entry.KeyXDR) != "" {
			parts = append(parts, "key "+truncate(strings.TrimSpace(entry.KeyXDR), 22))
		}
		if strings.TrimSpace(entry.ValueXDR) != "" {
			parts = append(parts, "value "+truncate(strings.TrimSpace(entry.ValueXDR), 24))
		}
	} else if strings.TrimSpace(entry.DisplayKey) != "" {
		parts = append(parts, "key "+truncate(strings.TrimSpace(entry.DisplayKey), 22))
		if strings.TrimSpace(entry.DisplayValue) != "" {
			parts = append(parts, "value "+truncate(strings.TrimSpace(entry.DisplayValue), 24))
		}
	}
	if strings.TrimSpace(entry.ExpirationProximity) != "" {
		parts = append(parts, entry.ExpirationProximity)
	}
	if len(parts) == 0 {
		return "storage entry"
	}
	return truncate(strings.Join(parts, "  "), 72)
}

func contractStorageCopyValue(entry backendclient.ContractStorageSummary) string {
	if entry.ValueDecoded != nil && strings.TrimSpace(*entry.ValueDecoded) != "" {
		return strings.TrimSpace(*entry.ValueDecoded)
	}
	return strings.TrimSpace(entry.ValueXDR)
}

func summarizeContractFunction(fn backendclient.ContractSpecFunction) string {
	name := strings.TrimSpace(fn.Name)
	if name == "" {
		name = "unnamed"
	}
	inputs := make([]string, 0, len(fn.Inputs))
	for _, input := range fn.Inputs {
		label := strings.TrimSpace(input.Type)
		if strings.TrimSpace(input.Name) != "" {
			label = strings.TrimSpace(input.Name) + ":" + label
		}
		if strings.TrimSpace(label) != "" {
			inputs = append(inputs, label)
		}
	}
	signature := name + "(" + strings.Join(inputs, ", ") + ")"
	if len(fn.Outputs) > 0 {
		signature += " -> " + strings.Join(fn.Outputs, ", ")
	}
	if fn.Doc != nil && strings.TrimSpace(*fn.Doc) != "" {
		signature += "  " + strings.TrimSpace(*fn.Doc)
	}
	return truncate(signature, 72)
}

func summarizeAssetHolder(holder backendclient.AssetHolderSummary) string {
	parts := []string{truncate(holder.AccountID, 18), "balance " + holder.Balance}
	if strings.TrimSpace(holder.LimitAmount) != "" {
		parts = append(parts, "limit "+holder.LimitAmount)
	}
	if holder.Sponsor != nil && strings.TrimSpace(*holder.Sponsor) != "" {
		parts = append(parts, "sponsor "+truncate(strings.TrimSpace(*holder.Sponsor), 12))
	}
	return truncate(strings.Join(parts, "  "), 72)
}

func summarizeSearchResult(result app.SearchResult) string {
	parts := []string{}
	if strings.TrimSpace(result.Title) != "" {
		parts = append(parts, strings.TrimSpace(result.Title))
	}
	if strings.TrimSpace(result.Description) != "" {
		parts = append(parts, strings.TrimSpace(result.Description))
	}
	if strings.TrimSpace(result.Source) != "" {
		parts = append(parts, strings.TrimSpace(result.Source))
	}
	return truncate(strings.Join(parts, "  "), 72)
}

func titleOrDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	runes := []rune(value)
	if len(runes) == 0 {
		return fallback
	}
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func shortTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format("2006-01-02 15:04")
}

func lookupOperationCommand(op backendclient.OperationSummary) (string, string) {
	if op.ContractID != nil && strings.TrimSpace(*op.ContractID) != "" {
		return "lookup contract " + strings.TrimSpace(*op.ContractID), "enter open operation contract"
	}
	if op.Destination != nil && strings.TrimSpace(*op.Destination) != "" {
		return "lookup account " + strings.TrimSpace(*op.Destination), "enter open destination account"
	}
	if op.AssetCode != nil && op.AssetIssuer != nil && strings.TrimSpace(*op.AssetCode) != "" && strings.TrimSpace(*op.AssetIssuer) != "" {
		return fmt.Sprintf("lookup asset %s:%s", strings.TrimSpace(*op.AssetCode), strings.TrimSpace(*op.AssetIssuer)), "enter open operation asset"
	}
	if op.SourceAccount != nil && strings.TrimSpace(*op.SourceAccount) != "" {
		return "lookup account " + strings.TrimSpace(*op.SourceAccount), "enter open source account"
	}
	if strings.TrimSpace(op.TransactionHash) != "" {
		return "lookup tx " + strings.TrimSpace(op.TransactionHash), "enter open operation transaction"
	}
	return "", ""
}

func renderLookupSections(sections []lookupSection, width, height, offset, selected int, visualMode bool) []string {
	if len(sections) == 0 {
		return nil
	}
	if height < 1 {
		height = 1
	}

	offset = clamp(offset, 0, max(0, len(sections)-height))
	end := min(len(sections), offset+height)
	lines := make([]string, 0, end-offset)

	for index := offset; index < end; index++ {
		section := sections[index]
		line := ""
		switch {
		case section.Divider:
			line = mutedStyle.Render(strings.Repeat("─", max(6, min(width, 32))))
		case section.Title == "":
			line = truncate(section.Body, width)
		default:
			line = keyValue(section.Title, section.Body, width)
		}
		if section.Emph {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		if section.Muted && !section.Divider && section.Title == "" {
			line = tableHeaderStyle.Render(line)
		}
		if section.Muted {
			line = mutedStyle.Render(line)
		}
		if index == selected && sectionIsSelectable(section, visualMode) {
			line = selectedRowStyle.Width(width).Render(line)
		}
		lines = append(lines, line)
	}
	return lines
}

func selectedLookupClipboardValue(snapshot app.Snapshot, selected int) string {
	sections := lookupSections(snapshot.Lookup)
	if selected < 0 || selected >= len(sections) {
		return ""
	}
	return sections[selected].ClipboardValue()
}

func selectedLookupActionCommand(snapshot app.Snapshot, selected int) string {
	sections := lookupSections(snapshot.Lookup)
	if selected < 0 || selected >= len(sections) {
		return ""
	}
	return strings.TrimSpace(sections[selected].Command)
}

func selectedLookupActionHint(snapshot app.Snapshot, selected int) string {
	sections := lookupSections(snapshot.Lookup)
	if selected < 0 || selected >= len(sections) {
		return ""
	}
	section := sections[selected]
	if strings.TrimSpace(section.Hint) != "" {
		return section.Hint
	}
	if strings.TrimSpace(section.Command) != "" {
		return "enter follow selected entity"
	}
	return ""
}

func firstSelectableLookupSection(sections []lookupSection, visualMode bool) int {
	for index, section := range sections {
		if sectionIsSelectable(section, visualMode) {
			return index
		}
	}
	return 0
}

func lastSelectableLookupSection(sections []lookupSection, visualMode bool) int {
	for index := len(sections) - 1; index >= 0; index-- {
		if sectionIsSelectable(sections[index], visualMode) {
			return index
		}
	}
	return 0
}

func nextSelectableLookupSection(sections []lookupSection, current, delta int, visualMode bool) int {
	if len(sections) == 0 {
		return 0
	}
	if delta == 0 {
		if current >= 0 && current < len(sections) && sectionIsSelectable(sections[current], visualMode) {
			return current
		}
		return firstSelectableLookupSection(sections, visualMode)
	}

	index := clamp(current+delta, 0, len(sections)-1)
	for index >= 0 && index < len(sections) {
		if sectionIsSelectable(sections[index], visualMode) {
			return index
		}
		index += delta
	}
	if delta > 0 {
		return lastSelectableLookupSection(sections, visualMode)
	}
	return firstSelectableLookupSection(sections, visualMode)
}
