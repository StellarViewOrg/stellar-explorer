package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

func renderHeader(snapshot app.Snapshot, width int) string {
	cacheMode := "cache:off"
	if snapshot.Cache.Enabled {
		cacheMode = "cache:on"
		if !snapshot.Cache.Available {
			cacheMode = "cache:degraded"
		}
	}

	left := titleStyle.Render(fmt.Sprintf("STELLAR TUI  %s/%s", snapshot.Profile.Network, snapshot.Profile.BackendMode))
	sourceState := activeDataSource(snapshot)
	if snapshot.LiveFeed.Source.Degraded || snapshot.Lookup.Source.Degraded {
		sourceState += "  degraded"
	}
	if badge := lookupCacheBadge(snapshot); badge != "" {
		cacheMode = badge
	}
	right := mutedStyle.Render(fmt.Sprintf("%s  %s  focus:%s", cacheMode, sourceState, snapshot.Focus))
	top := fitPair(left, right, width)

	tabs := []string{
		renderTab(snapshot, app.ScreenHome, "h", "Home"),
		renderTab(snapshot, app.ScreenLiveFeed, "l", "Live"),
		renderTab(snapshot, app.ScreenLookup, "u", "Lookup"),
		renderTab(snapshot, app.ScreenSettings, "s", "Settings"),
		mutedStyle.Render(fmt.Sprintf("history %d/%d", len(snapshot.History), len(snapshot.Forward))),
	}
	return lipgloss.JoinVertical(lipgloss.Left, top, lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
}

func lookupCacheBadge(snapshot app.Snapshot) string {
	if snapshot.Current != app.ScreenLookup {
		return ""
	}
	switch snapshot.Lookup.Source.CacheState {
	case "hit":
		return "cache:hit"
	case "stale":
		return "cache:stale"
	default:
		return ""
	}
}

func renderTab(snapshot app.Snapshot, screen app.Screen, key string, label string) string {
	text := fmt.Sprintf("%s:%s", key, label)
	if snapshot.Current == screen {
		return activeTabStyle.Render(text)
	}
	return inactiveTabStyle.Render(text)
}

func renderMain(snapshot app.Snapshot, width, height int) string {
	if width < 92 {
		return NewBodyModel(snapshot, width, height).View()
	}

	sidebarWidth := clamp(width/3, 27, 34)
	bodyWidth := width - sidebarWidth - 1
	sidebar := NewSidebarModel(snapshot, sidebarWidth, height).View()
	body := NewBodyModel(snapshot, bodyWidth, height).View()
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", body)
}

func renderBodyPanel(snapshot app.Snapshot, width, height int) string {
	return NewBodyModel(snapshot, width, height).View()
}

func renderStatus(snapshot app.Snapshot, width, height int) string {
	level := "INFO"
	levelStyle := statusInfoStyle
	switch snapshot.Status.Level {
	case app.StatusError:
		level = "ERROR"
		levelStyle = statusErrorStyle
	case app.StatusWarn:
		level = "WARN"
		levelStyle = statusWarnStyle
	}

	contentWidth := panelContentWidth(width)
	lines := []string{levelStyle.Render(level) + "  " + truncate(snapshot.Status.Message, contentWidth-8)}
	if snapshot.Cache.Status != "" {
		lines = append(lines, truncate("Cache status: "+snapshot.Cache.Status, contentWidth))
	}
	if snapshot.Cache.Available {
		lines = append(lines, fmt.Sprintf("Cache schema: v%d", snapshot.Cache.Schema))
	}
	return styledPanel("Status", width, height, lines, snapshot.Focus == app.FocusStatus)
}

func renderFooter(snapshot app.Snapshot, width int) string {
	text := fmt.Sprintf("Focus: %s | tab cycle panes | h/l/u/s views | / search | y copy | ctrl+v paste | %s | ? help | q quit", snapshot.Focus, focusShortcutHint(snapshot))
	return mutedStyle.Width(width).Render(truncate(text, width))
}

func focusShortcutHint(snapshot app.Snapshot) string {
	switch snapshot.Focus {
	case app.FocusSidebar:
		return "j/k move nav | enter switch view"
	case app.FocusStatus:
		return "tab return to interactive panes"
	default:
		switch snapshot.Current {
		case app.ScreenLookup:
			return "j/k move | pgup/pgdn scroll | enter follow entity"
		case app.ScreenLiveFeed:
			return "j/k move | pgup/pgdn scroll | left/right copy field | enter open tx"
		default:
			return "j/k move | pgup/pgdn scroll"
		}
	}
}

func styledPanel(title string, width, height int, lines []string, focused bool) string {
	style := panelStyle
	if focused {
		style = focusedPanelStyle
	}
	innerWidth := panelContentWidth(width)
	innerHeight := max(1, height-2)

	content := make([]string, 0, innerHeight)
	content = append(content, tableHeaderStyle.Render(truncate(title, innerWidth)))
	for _, line := range lines {
		if len(content) >= innerHeight {
			break
		}
		content = append(content, truncate(line, innerWidth))
	}
	for len(content) < innerHeight {
		content = append(content, "")
	}

	return style.Width(width - 2).Height(height - 2).Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func bodyTitle(snapshot app.Snapshot) string {
	switch snapshot.Current {
	case app.ScreenLiveFeed:
		return "Live Feed"
	case app.ScreenLookup:
		return "Lookup"
	case app.ScreenSettings:
		return "Settings"
	default:
		return "Home"
	}
}
