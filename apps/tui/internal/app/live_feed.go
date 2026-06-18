package app

import (
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

const (
	LiveFeedSourcePoll      = "poll"
	LiveFeedSourceStream    = "stream"
	LiveFeedSourceDegraded  = "degraded"
	LiveFeedFilterAccount   = "account"
	LiveFeedFilterContract  = "contract"
	LiveFeedFilterAsset     = "asset"
	LiveFeedFilterOperation = "operation"
)

// LiveFeedFilterSpec captures class and entity-specific live feed filters.
type LiveFeedFilterSpec struct {
	Class         string
	Account       string
	Contract      string
	Asset         string
	OperationType string
}

func defaultLiveFeedFilterSpec() LiveFeedFilterSpec {
	return LiveFeedFilterSpec{Class: LiveFeedFilterAll}
}

func formatLiveFeedFilter(spec LiveFeedFilterSpec) string {
	parts := make([]string, 0, 5)
	class := normalizeLiveFeedFilter(spec.Class)
	if class != LiveFeedFilterAll {
		parts = append(parts, class)
	}
	if value := strings.TrimSpace(spec.Account); value != "" {
		parts = append(parts, LiveFeedFilterAccount+":"+truncateLiveFeedFilterValue(value))
	}
	if value := strings.TrimSpace(spec.Contract); value != "" {
		parts = append(parts, LiveFeedFilterContract+":"+truncateLiveFeedFilterValue(value))
	}
	if value := strings.TrimSpace(spec.Asset); value != "" {
		parts = append(parts, LiveFeedFilterAsset+":"+truncateLiveFeedFilterValue(value))
	}
	if value := strings.TrimSpace(spec.OperationType); value != "" {
		parts = append(parts, LiveFeedFilterOperation+":"+value)
	}
	if len(parts) == 0 {
		return LiveFeedFilterAll
	}
	return strings.Join(parts, " ")
}

func truncateLiveFeedFilterValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:6] + "..." + value[len(value)-4:]
}

func parseLiveFeedFilterCommand(fields []string) (LiveFeedFilterSpec, error) {
	if len(fields) < 2 {
		return LiveFeedFilterSpec{}, fmt.Errorf("usage: live filter all|soroban|classic|clear|account|contract|asset|operation <value>")
	}

	spec := defaultLiveFeedFilterSpec()
	switch strings.ToLower(strings.TrimSpace(fields[1])) {
	case LiveFeedFilterAll:
		return defaultLiveFeedFilterSpec(), nil
	case LiveFeedFilterSoroban:
		spec.Class = LiveFeedFilterSoroban
		return spec, nil
	case LiveFeedFilterClassic:
		spec.Class = LiveFeedFilterClassic
		return spec, nil
	case "clear":
		return defaultLiveFeedFilterSpec(), nil
	case LiveFeedFilterAccount, LiveFeedFilterContract, LiveFeedFilterAsset, LiveFeedFilterOperation:
		if len(fields) < 3 || strings.TrimSpace(fields[2]) == "" {
			return LiveFeedFilterSpec{}, fmt.Errorf("usage: live filter %s <value>", fields[1])
		}
		value := strings.TrimSpace(strings.Join(fields[2:], " "))
		switch fields[1] {
		case LiveFeedFilterAccount:
			spec.Account = value
		case LiveFeedFilterContract:
			spec.Contract = value
		case LiveFeedFilterAsset:
			spec.Asset = value
		case LiveFeedFilterOperation:
			spec.OperationType = strings.ToLower(value)
		}
		return spec, nil
	default:
		return LiveFeedFilterSpec{}, fmt.Errorf("unknown live feed filter %q", fields[1])
	}
}

func filterLiveFeedTransactions(transactions []backendclient.TransactionSummary, spec LiveFeedFilterSpec) []backendclient.TransactionSummary {
	out := make([]backendclient.TransactionSummary, 0, len(transactions))
	for _, tx := range transactions {
		if matchesLiveFeedFilter(tx, spec) {
			out = append(out, tx)
		}
	}
	return out
}

func matchesLiveFeedFilter(tx backendclient.TransactionSummary, spec LiveFeedFilterSpec) bool {
	switch normalizeLiveFeedFilter(spec.Class) {
	case LiveFeedFilterSoroban:
		if !tx.IsSoroban {
			return false
		}
	case LiveFeedFilterClassic:
		if tx.IsSoroban {
			return false
		}
	}

	if account := strings.TrimSpace(spec.Account); account != "" {
		if !strings.EqualFold(strings.TrimSpace(tx.Account), account) {
			return false
		}
	}

	if contract := strings.TrimSpace(spec.Contract); contract != "" {
		if tx.PrimaryContractID == "" || !strings.EqualFold(strings.TrimSpace(tx.PrimaryContractID), contract) {
			return false
		}
	}

	if asset := strings.TrimSpace(spec.Asset); asset != "" {
		code, issuer, ok := strings.Cut(asset, ":")
		if !ok || strings.TrimSpace(code) == "" || strings.TrimSpace(issuer) == "" {
			return false
		}
		if !strings.EqualFold(strings.TrimSpace(tx.PrimaryAssetCode), strings.TrimSpace(code)) {
			return false
		}
		if !strings.EqualFold(strings.TrimSpace(tx.PrimaryAssetIssuer), strings.TrimSpace(issuer)) {
			return false
		}
	}

	if operationType := strings.ToLower(strings.TrimSpace(spec.OperationType)); operationType != "" {
		if tx.PrimaryOperationType == "" || !strings.EqualFold(strings.TrimSpace(tx.PrimaryOperationType), operationType) {
			return false
		}
	}

	return true
}

func nextLiveFeedFilterClass(class string) string {
	switch normalizeLiveFeedFilter(class) {
	case LiveFeedFilterAll:
		return LiveFeedFilterSoroban
	case LiveFeedFilterSoroban:
		return LiveFeedFilterClassic
	default:
		return LiveFeedFilterAll
	}
}

func (m *Model) applyLiveFeedView(added int) {
	selectedHash := ""
	if m.selection.LiveFeedIndex >= 0 && m.selection.LiveFeedIndex < len(m.liveFeed.RecentTransactions) {
		selectedHash = strings.TrimSpace(m.liveFeed.RecentTransactions[m.selection.LiveFeedIndex].Hash)
	}

	filtered := filterLiveFeedTransactions(m.liveFeedAll, m.liveFeedFilterSpec)
	if m.liveFeed.Paused && m.liveFeed.ReplayOffset > 0 {
		start := m.liveFeed.ReplayOffset
		if start >= len(filtered) {
			start = max(0, len(filtered)-1)
			m.liveFeed.ReplayOffset = start
		}
		filtered = filtered[start:]
	}

	m.liveFeed.Filter = formatLiveFeedFilter(m.liveFeedFilterSpec)
	m.liveFeed.RecentTransactions = filtered
	m.liveFeed.Scrollback = append([]backendclient.TransactionSummary(nil), m.liveFeedAll...)
	m.liveFeed.ScrollbackCount = len(m.liveFeedAll)
	m.liveFeed.MaxScrollback = defaultLiveFeedScrollback
	m.liveFeed.LastUpdateCount = added

	if len(m.liveFeed.RecentTransactions) == 0 {
		if len(m.liveFeedAll) > 0 || !m.liveFeed.Available {
			m.liveFeed.State = ViewStateEmpty
		}
	} else {
		m.liveFeed.State = ViewStateReady
	}

	if selectedHash != "" {
		for index, tx := range m.liveFeed.RecentTransactions {
			if strings.TrimSpace(tx.Hash) == selectedHash {
				m.selection.LiveFeedIndex = index
				return
			}
		}
	}
	m.selection.LiveFeedIndex = clampLiveFeedIndex(m.selection.LiveFeedIndex, len(m.liveFeed.RecentTransactions))
}

// ApplyLiveFeedStreamUpdate merges stream payloads into the retained scrollback.
func (m *Model) ApplyLiveFeedStreamUpdate(update LiveFeedStreamUpdate) int {
	if m.liveFeed.Paused {
		m.liveFeedPendingStream = append(m.liveFeedPendingStream, update)
		return 0
	}

	added := 0
	if len(update.Transactions) > 0 {
		added = m.mergeLiveFeedTransactions(update.Transactions)
	}
	if update.Ledger != nil {
		m.liveFeed.LatestLedger = update.Ledger
		if update.Ledger.Sequence > m.liveFeed.LastIngestedLedger {
			m.liveFeed.LastIngestedLedger = update.Ledger.Sequence
		}
	}

	if added > 0 || update.Ledger != nil {
		m.liveFeed.Available = true
		m.liveFeed.Error = ""
		m.applyLiveFeedView(added)
		m.liveFeed.SourceMode = LiveFeedSourceStream
		if added > 0 {
			m.status = Status{
				Level:   StatusInfo,
				Message: fmt.Sprintf("Live stream received %d new transaction(s).", added),
			}
		}
	}

	return added
}

func (m *Model) flushPendingLiveFeedStreamUpdates() {
	if len(m.liveFeedPendingStream) == 0 {
		return
	}

	pending := append([]LiveFeedStreamUpdate(nil), m.liveFeedPendingStream...)
	m.liveFeedPendingStream = nil
	for _, update := range pending {
		_ = m.ApplyLiveFeedStreamUpdate(update)
	}
}

// LiveFeedStreamUpdate is the app-facing stream payload shape.
type LiveFeedStreamUpdate struct {
	Ledger       *backendclient.LedgerSummary
	Transactions []backendclient.TransactionSummary
}

// SetLiveFeedSourceMode records how live feed rows are being ingested.
func (m *Model) SetLiveFeedSourceMode(mode string) {
	switch strings.TrimSpace(mode) {
	case LiveFeedSourcePoll, LiveFeedSourceStream, LiveFeedSourceDegraded:
		m.liveFeed.SourceMode = mode
	default:
		m.liveFeed.SourceMode = LiveFeedSourcePoll
	}
}

// ShiftLiveFeedReplay moves the paused replay window through retained scrollback.
func (m *Model) ShiftLiveFeedReplay(delta int) {
	if !m.liveFeed.Paused || delta == 0 {
		return
	}

	filtered := filterLiveFeedTransactions(m.liveFeedAll, m.liveFeedFilterSpec)
	maxOffset := max(0, len(filtered)-1)
	next := m.liveFeed.ReplayOffset + delta
	if next < 0 {
		next = 0
	}
	if next > maxOffset {
		next = maxOffset
	}
	if next == m.liveFeed.ReplayOffset {
		return
	}

	m.liveFeed.ReplayOffset = next
	m.applyLiveFeedView(0)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Live feed replay offset: %d.", m.liveFeed.ReplayOffset),
	}
}

func (m *Model) setLiveFeedFilterSpec(spec LiveFeedFilterSpec) {
	spec.Class = normalizeLiveFeedFilter(spec.Class)
	m.liveFeedFilterSpec = spec
	m.liveFeed.ReplayOffset = 0
	m.applyLiveFeedView(0)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Live feed filter: %s.", formatLiveFeedFilter(spec)),
	}
}

// ParseLiveFeedFilterValue converts a formatted live filter string into a filter spec.
func ParseLiveFeedFilterValue(value string) (LiveFeedFilterSpec, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == LiveFeedFilterAll {
		return defaultLiveFeedFilterSpec(), nil
	}

	spec := defaultLiveFeedFilterSpec()
	fields := strings.Fields(value)
	for _, field := range fields {
		switch strings.ToLower(field) {
		case LiveFeedFilterSoroban:
			spec.Class = LiveFeedFilterSoroban
		case LiveFeedFilterClassic:
			spec.Class = LiveFeedFilterClassic
		default:
			key, raw, ok := strings.Cut(field, ":")
			if !ok {
				return LiveFeedFilterSpec{}, fmt.Errorf("unknown live feed filter token %q", field)
			}
			switch strings.ToLower(strings.TrimSpace(key)) {
			case LiveFeedFilterAccount:
				spec.Account = strings.TrimSpace(raw)
			case LiveFeedFilterContract:
				spec.Contract = strings.TrimSpace(raw)
			case LiveFeedFilterAsset:
				spec.Asset = strings.TrimSpace(raw)
			case LiveFeedFilterOperation:
				spec.OperationType = strings.ToLower(strings.TrimSpace(raw))
			default:
				return LiveFeedFilterSpec{}, fmt.Errorf("unknown live feed filter token %q", field)
			}
		}
	}
	return spec, nil
}
