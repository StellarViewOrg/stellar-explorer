package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

type panelStateTone int

const (
	panelStateInfo panelStateTone = iota
	panelStateWarn
	panelStateError
)

func keyValue(key, value string, width int) string {
	if strings.HasSuffix(key, ":") {
		return truncate(key+" "+value, width)
	}
	keyWidth := 12
	valueWidth := max(4, width-keyWidth-2)
	return fmt.Sprintf("%-*s  %s", keyWidth, key, truncate(value, valueWidth))
}

func activeDataSource(snapshot app.Snapshot) string {
	if snapshot.LiveFeed.Source.Label != "" {
		return snapshot.LiveFeed.Source.Label
	}
	if snapshot.Lookup.Source.Label != "" {
		return snapshot.Lookup.Source.Label
	}
	if snapshot.LiveFeed.BackendURL != "" {
		return snapshot.LiveFeed.BackendURL
	}
	if snapshot.Lookup.BackendURL != "" {
		return snapshot.Lookup.BackendURL
	}
	if snapshot.Profile.IndexerURL != "" {
		return snapshot.Profile.IndexerURL
	}
	if snapshot.Profile.RPCEndpoint != "" {
		return snapshot.Profile.RPCEndpoint
	}
	return "not configured"
}

func renderSourceSummary(source app.SourceMetadata, width int) []string {
	if source.Policy == "" && source.Label == "" && source.Actual == "" {
		return nil
	}
	lines := []string{
		keyValue("Policy", valueOrFallback(source.Policy, "unspecified"), width),
		keyValue("Source", valueOrFallback(source.Actual, source.Preferred), width),
	}
	if source.Label != "" {
		lines = append(lines, keyValue("Resolved", source.Label, width))
	}
	if source.FallbackUsed {
		lines = append(lines, statusWarnStyle.Render(truncate("Fallback active", width)))
	}
	if source.Degraded && strings.TrimSpace(source.DegradedReason) != "" {
		lines = append(lines, statusWarnStyle.Render(truncate("Degraded: "+source.DegradedReason, width)))
	}
	return lines
}

func valueOrFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func fitPair(left, right string, width int) string {
	if lipgloss.Width(left)+lipgloss.Width(right) >= width {
		available := width - lipgloss.Width(left)
		if available < 8 {
			return truncate(left, width)
		}
		return left + truncate(right, available)
	}
	return left + strings.Repeat(" ", width-lipgloss.Width(left)-lipgloss.Width(right)) + right
}

func panelContentWidth(width int) int {
	return max(1, width-4)
}

func truncate(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= limit {
		return value
	}
	runes := []rune(value)
	if limit <= 3 {
		if len(runes) < limit {
			return value
		}
		return string(runes[:limit])
	}
	maxRunes := min(len(runes), limit-3)
	return string(runes[:maxRunes]) + "..."
}

func renderTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return "unknown"
	}
	return ts.UTC().Format(time.RFC3339)
}

func clamp(value, minValue, maxValue int) int {
	return min(max(value, minValue), maxValue)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderPanelState(label, detail string, tone panelStateTone, width int) []string {
	style := statusInfoStyle
	switch tone {
	case panelStateWarn:
		style = statusWarnStyle
	case panelStateError:
		style = statusErrorStyle
	}

	lines := []string{style.Render(label)}
	if trimmed := strings.TrimSpace(detail); trimmed != "" {
		lines = append(lines, truncate(trimmed, width))
	}
	return lines
}
