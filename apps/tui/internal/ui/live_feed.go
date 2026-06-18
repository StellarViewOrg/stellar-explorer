package ui

import (
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

type liveFeedCopyField int

const (
	liveFeedCopyHash liveFeedCopyField = iota
	liveFeedCopyLedger
	liveFeedCopyAccount
)

func (f liveFeedCopyField) next() liveFeedCopyField {
	return liveFeedCopyField((int(f) + 1) % 3)
}

func (f liveFeedCopyField) prev() liveFeedCopyField {
	return liveFeedCopyField((int(f) + 2) % 3)
}

func (f liveFeedCopyField) label() string {
	switch f {
	case liveFeedCopyLedger:
		return "ledger"
	case liveFeedCopyAccount:
		return "account"
	default:
		return "hash"
	}
}

func liveFeedView(snapshot app.Snapshot, width, height int, offset int, selected int, copyField liveFeedCopyField) []string {
	if !snapshot.LiveFeed.Configured {
		return renderPanelState("State: configuration required", "Set rpc_endpoint or indexer_url in the active profile.", panelStateWarn, width)
	}

	if snapshot.LiveFeed.State == app.ViewStateError || !snapshot.LiveFeed.Available {
		lines := []string{
			keyValue("Source", valueOrFallback(snapshot.LiveFeed.BackendURL, "not configured"), width),
			"",
		}
		return append(lines, renderPanelState("State: backend unavailable", valueOrFallback(snapshot.LiveFeed.Error, "unknown backend error"), panelStateError, width)...)
	}

	lines := []string{
		statusInfoStyle.Render("State: " + string(snapshot.LiveFeed.State)),
		fmt.Sprintf("Last ingested ledger: %d", snapshot.LiveFeed.LastIngestedLedger),
		fmt.Sprintf("Mode: %s  Filter: %s  Scrollback: %d/%d  New: %d", liveFeedModeLabel(snapshot), snapshot.LiveFeed.Filter, snapshot.LiveFeed.ScrollbackCount, snapshot.LiveFeed.MaxScrollback, snapshot.LiveFeed.LastUpdateCount),
		liveFeedReplayHint(snapshot),
		fmt.Sprintf("Copy field: %s (left/right to change, y to copy)", copyField.label()),
		mutedStyle.Render("p pause/resume  t cycle class  [/] replay scrollback  r refresh  enter open selected tx"),
	}
	lines = append(lines, renderSourceSummary(snapshot.LiveFeed.Source, width)...)
	if snapshot.LiveFeed.LatestLedger != nil {
		ledger := snapshot.LiveFeed.LatestLedger
		lines = append(lines,
			keyValue("Latest", fmt.Sprintf("%d  %s", ledger.Sequence, renderTimestamp(ledger.ClosedAt)), width),
			keyValue("Load", fmt.Sprintf("%d tx / %d ops", ledger.TransactionCount, ledger.OperationCount), width),
		)
	}
	lines = append(lines, "", tableHeaderStyle.Render("Live transactions"))

	if snapshot.LiveFeed.State == app.ViewStateEmpty || len(snapshot.LiveFeed.RecentTransactions) == 0 {
		return append(lines, mutedStyle.Render("No recent transactions yet. Refresh or wait for the next ledger."))
	}

	tableHeight := height - len(lines)
	if tableHeight < 2 {
		tableHeight = 2
	}
	return append(lines, renderLiveTable(snapshot, width, tableHeight, offset, selected, copyField)...)
}

func liveFeedModeLabel(snapshot app.Snapshot) string {
	if snapshot.LiveFeed.Paused {
		return "paused"
	}
	switch snapshot.LiveFeed.SourceMode {
	case app.LiveFeedSourceStream:
		return "stream"
	case app.LiveFeedSourceDegraded:
		return "degraded"
	default:
		return "poll"
	}
}

func liveFeedReplayHint(snapshot app.Snapshot) string {
	if !snapshot.LiveFeed.Paused {
		return mutedStyle.Render("Replay: pause the feed, then use [ and ] to walk retained scrollback.")
	}
	return fmt.Sprintf("Replay offset: %d (older [ / newer ])", snapshot.LiveFeed.ReplayOffset)
}

func renderLiveTable(snapshot app.Snapshot, width, height int, offset int, selected int, copyField liveFeedCopyField) []string {
	hashWidth := clamp(width/4, 15, 15)
	ledgerWidth := 9
	opsWidth := 4
	sorobanWidth := 5
	accountWidth := width - hashWidth - ledgerWidth - opsWidth - sorobanWidth - 10
	if accountWidth < 8 {
		accountWidth = 8
	}

	header := tableHeaderStyle.Render(fmt.Sprintf(
		"  %-*s  %-*s  %-*s  %-*s  %s",
		hashWidth, liveFeedColumnLabel("Hash", copyField == liveFeedCopyHash),
		ledgerWidth, liveFeedColumnLabel("Ledger", copyField == liveFeedCopyLedger),
		opsWidth, "Ops",
		sorobanWidth, "Soro",
		liveFeedColumnLabel("Account", copyField == liveFeedCopyAccount),
	))
	lines := []string{header}

	visibleRows := liveFeedVisibleRows(height)
	count := len(snapshot.LiveFeed.RecentTransactions)
	offset = clamp(offset, 0, max(0, count-visibleRows))
	limit := min(count, offset+visibleRows)
	for index := offset; index < limit; index++ {
		tx := snapshot.LiveFeed.RecentTransactions[index]
		row := fmt.Sprintf(
			"  [%d] %-*s  %-*d  %-*d  %-*t  %s",
			index+1,
			hashWidth, truncate(tx.Hash, hashWidth),
			ledgerWidth, tx.LedgerSequence,
			opsWidth, tx.OperationCount,
			sorobanWidth, tx.IsSoroban,
			truncate(tx.Account, accountWidth),
		)
		if selected == index {
			row = ">" + strings.TrimPrefix(row, " ")
			row = selectedRowStyle.Width(width).Render(row)
		}
		lines = append(lines, row)
	}
	return lines
}

func liveFeedColumnLabel(label string, active bool) string {
	if active {
		return "[" + label + "]"
	}
	return label
}

func liveFeedVisibleRows(height int) int {
	return max(1, height-1)
}
