package ui

import "github.com/miguelnietoa/stellar-explorer/tui/internal/app"

// Render returns the full terminal view for the current app snapshot.
func Render(snapshot app.Snapshot) string {
	return RenderWithSize(snapshot, defaultWidth, defaultHeight)
}

// RenderWithSize renders the terminal dashboard using Lip Gloss components.
func RenderWithSize(snapshot app.Snapshot, width, height int) string {
	return NewDashboardModel(snapshot, width, height).View()
}
