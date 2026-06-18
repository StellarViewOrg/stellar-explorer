package app

import (
	"fmt"
	"strings"
)

// LiveFeedMonitoringContext stores live-feed position and filter state so a user
// can drill into a transaction and return without losing monitoring context.
type LiveFeedMonitoringContext struct {
	Active        bool
	SelectedIndex int
	ScrollOffset  int
	SelectedHash  string
	Paused        bool
	ReplayOffset  int
	FilterSpec    LiveFeedFilterSpec
}

// LookupReturnContext records where a lookup view should return on back navigation.
type LookupReturnContext struct {
	Screen Screen
	Label  string
}

// CaptureLiveFeedMonitoringContext snapshots the current live-feed monitoring state.
func (m *Model) CaptureLiveFeedMonitoringContext(selectedIndex, scrollOffset int) {
	if m.current != ScreenLiveFeed {
		return
	}

	selectedHash := ""
	if selectedIndex >= 0 && selectedIndex < len(m.liveFeed.RecentTransactions) {
		selectedHash = strings.TrimSpace(m.liveFeed.RecentTransactions[selectedIndex].Hash)
	}

	m.liveFeedMonitoring = LiveFeedMonitoringContext{
		Active:        true,
		SelectedIndex: selectedIndex,
		ScrollOffset:  scrollOffset,
		SelectedHash:  selectedHash,
		Paused:        m.liveFeed.Paused,
		ReplayOffset:  m.liveFeed.ReplayOffset,
		FilterSpec:    m.liveFeedFilterSpec,
	}
}

// RestoreLiveFeedMonitoringContext reapplies a previously captured monitoring context.
func (m *Model) RestoreLiveFeedMonitoringContext() bool {
	if !m.liveFeedMonitoring.Active {
		return false
	}

	context := m.liveFeedMonitoring
	m.liveFeedMonitoring = LiveFeedMonitoringContext{}

	m.liveFeedFilterSpec = context.FilterSpec
	m.liveFeed.Paused = context.Paused
	m.liveFeed.ReplayOffset = context.ReplayOffset
	m.applyLiveFeedView(0)

	if context.SelectedHash != "" {
		for index, tx := range m.liveFeed.RecentTransactions {
			if strings.EqualFold(strings.TrimSpace(tx.Hash), context.SelectedHash) {
				m.selection.LiveFeedIndex = index
				m.selection.LiveFeedScrollOffset = context.ScrollOffset
				m.status = Status{
					Level:   StatusInfo,
					Message: "Restored live feed monitoring context.",
				}
				return true
			}
		}
	}

	m.selection.LiveFeedIndex = clampLiveFeedIndex(context.SelectedIndex, len(m.liveFeed.RecentTransactions))
	m.selection.LiveFeedScrollOffset = context.ScrollOffset
	m.status = Status{
		Level:   StatusInfo,
		Message: "Restored live feed monitoring context.",
	}
	return true
}

// SetLookupReturnContext records a return target for the active lookup view.
func (m *Model) SetLookupReturnContext(screen Screen, label string) {
	m.lookup.ReturnContext = &LookupReturnContext{
		Screen: screen,
		Label:  strings.TrimSpace(label),
	}
}

// ClearLookupReturnContext removes any lookup return target.
func (m *Model) ClearLookupReturnContext() {
	m.lookup.ReturnContext = nil
}

// ReturnToLiveMonitoring switches back to the live feed and restores monitoring context.
func (m *Model) ReturnToLiveMonitoring() bool {
	if m.lookup.ReturnContext == nil || m.lookup.ReturnContext.Screen != ScreenLiveFeed {
		if !m.liveFeedMonitoring.Active {
			return false
		}
	}

	m.lookupHistory = nil
	m.lookupForward = nil
	m.lookup = LookupSnapshot{State: ViewStateIdle}
	m.current = ScreenLiveFeed
	m.helpVisible = false
	m.RestoreLiveFeedMonitoringContext()
	m.status = Status{
		Level:   StatusInfo,
		Message: "Returned to live feed monitoring.",
	}
	return true
}

// ApplyWatchSettings applies a saved watch preset to the live feed.
func (m *Model) ApplyWatchSettings(filterSpec LiveFeedFilterSpec, paused bool) {
	m.liveFeedFilterSpec = filterSpec
	m.liveFeed.Paused = paused
	m.liveFeed.ReplayOffset = 0
	m.applyLiveFeedView(0)
	m.status = Status{
		Level:   StatusInfo,
		Message: fmt.Sprintf("Applied watch settings: %s.", formatLiveFeedFilter(filterSpec)),
	}
}
