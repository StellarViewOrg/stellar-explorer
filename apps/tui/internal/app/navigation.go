package app

import "strings"

// LookupRouteStep describes one stop in the lookup navigation path.
type LookupRouteStep struct {
	Kind         LookupKind
	Query        string
	ExplorerKind LookupExplorerKind
	Title        string
	ListLimit    int
	ListOffset   int
}

// SetLookupSelection stores renderer-facing lookup cursor state in the app model.
func (m *Model) SetLookupSelection(section, offset int) {
	if section < 0 {
		section = 0
	}
	if offset < 0 {
		offset = 0
	}
	m.lookup.SelectedSection = section
	m.lookup.ScrollOffset = offset
}

func (m *Model) buildLookupRoute() []LookupRouteStep {
	route := make([]LookupRouteStep, 0, len(m.lookupHistory)+1)
	for _, snapshot := range m.lookupHistory {
		route = append(route, lookupSnapshotToRouteStep(snapshot))
	}
	if m.lookup.State == ViewStateReady && strings.TrimSpace(m.lookup.Query) != "" {
		route = append(route, lookupSnapshotToRouteStep(m.lookup))
	}
	return route
}

func lookupSnapshotToRouteStep(snapshot LookupSnapshot) LookupRouteStep {
	step := LookupRouteStep{
		Kind:  snapshot.Kind,
		Query: strings.TrimSpace(snapshot.Query),
	}
	if snapshot.Explorer == nil {
		return step
	}
	explorer := snapshot.Explorer
	step.ExplorerKind = explorer.Kind
	step.Title = strings.TrimSpace(explorer.Title)
	step.ListLimit = explorer.ListLimit
	step.ListOffset = explorer.ListOffset
	return step
}

func resetLookupSelection(snapshot *LookupSnapshot) {
	snapshot.SelectedSection = 0
	snapshot.ScrollOffset = 0
}
