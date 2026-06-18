package ui

import (
	"fmt"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

var sidebarScreens = []app.Screen{
	app.ScreenHome,
	app.ScreenLiveFeed,
	app.ScreenLookup,
	app.ScreenSettings,
}

func renderSidebarPanel(snapshot app.Snapshot, width, height int, selected app.Screen, focused bool) string {
	contentWidth := panelContentWidth(width)
	lines := []string{
		fmt.Sprintf("Screen: %s", snapshot.Current),
		keyValue("Profile", snapshot.Profile.Name, contentWidth),
		keyValue("Network", snapshot.Profile.Network, contentWidth),
		keyValue("Mode", snapshot.Profile.BackendMode, contentWidth),
		"",
		tableHeaderStyle.Render("Workflow"),
		navLine(snapshot, selected, app.ScreenHome, "Home"),
		navLine(snapshot, selected, app.ScreenLiveFeed, "Live feed"),
		navLine(snapshot, selected, app.ScreenLookup, "Lookup explorer"),
		navLine(snapshot, selected, app.ScreenSettings, "Settings"),
		"",
		tableHeaderStyle.Render("Source"),
		truncate(activeDataSource(snapshot), contentWidth),
		"",
		tableHeaderStyle.Render("Keys"),
		"h/l/u/s views",
		"j/k move",
		"enter open",
		"/ search",
		"b/f history",
		"r refresh",
		"? help",
		"q quit",
	}
	return styledPanel("Navigation", width, height, lines, focused)
}

func navLine(snapshot app.Snapshot, selected app.Screen, screen app.Screen, label string) string {
	prefix := "  "
	if selected == screen {
		prefix = "> "
	}
	line := prefix + label
	if snapshot.Current == screen && selected == screen {
		return selectedRowStyle.Render(line + "  (current)")
	}
	if selected == screen {
		return selectedRowStyle.Render(line)
	}
	if snapshot.Current == screen {
		return line + "  (current)"
	}
	return line
}
