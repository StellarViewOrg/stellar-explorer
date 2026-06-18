package ui

import (
	"fmt"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

func homeView(snapshot app.Snapshot, width int) []string {
	lines := []string{
		titleStyle.Render("Terminal-native Stellar explorer"),
		"",
		statusInfoStyle.Render("State: ready"),
		"Open the live feed for recent ledger activity.",
		"Use lookup or / search to jump to ledgers, transactions, accounts, assets, and contracts.",
		"Press ? for the key map.",
		"",
		tableHeaderStyle.Render("Current target"),
		truncate(fmt.Sprintf("%s on %s", snapshot.Profile.BackendMode, activeDataSource(snapshot)), width),
		"",
		tableHeaderStyle.Render("Track 1 Runtime"),
		"Per-view keyboard state now lives in the Bubble Tea models.",
		"Main pane views support local scrolling and focus-specific routing.",
	}
	return lines
}

func settingsView(snapshot app.Snapshot, width int) []string {
	return []string{
		tableHeaderStyle.Render("Local configuration"),
		"",
		statusInfoStyle.Render("State: ready"),
		keyValue("Config", snapshot.ConfigPath, width),
		keyValue("Profile", snapshot.Profile.Name, width),
		keyValue("Cache", valueOrFallback(snapshot.Cache.Path, "not configured"), width),
		keyValue("Profiles", fmt.Sprintf("%d", snapshot.Cache.Profiles), width),
		keyValue("Last view", valueOrFallback(snapshot.Cache.LastScreen, "none"), width),
		keyValue("Focus", string(snapshot.Focus), width),
		keyValue("Data", activeDataSource(snapshot), width),
	}
}
