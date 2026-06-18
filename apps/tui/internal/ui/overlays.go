package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

func renderHelpOverlay(snapshot app.Snapshot, width int) string {
	overlayWidth := clamp(width-8, 36, 96)
	lines := []string{
		"q quit",
		"h home",
		"l live feed",
		"u lookup screen",
		"s settings",
		"b / f back and forward",
		"r refresh current screen",
		"/ open search and command overlay",
		"y copy current entity to clipboard",
		"ctrl+v paste clipboard into search",
		"tab cycle focus",
		"? toggle help",
	}
	if snapshot.Current == app.ScreenLiveFeed {
		lines = append(lines, "j/k or arrows move through live transactions", "pgup/pgdn and home/end move faster through the feed", "p pauses/resumes live refresh", "t cycles class filters (all/soroban/classic)", "[ and ] replay retained scrollback while paused", "left/right changes the live copy field", "enter opens the selected transaction", "b returns from a drilled-in transaction with monitoring context restored", "live filter account|contract|asset|operation <value> for advanced filters", "watch save|open|delete|auto <name> stores profile watch presets")
	}
	if snapshot.Current == app.ScreenLookup {
		lines = append(lines,
			"j/k move between lookup sections",
			"pgup/pgdn and home/end navigate larger lookup payloads",
			"enter follows the selected related entity when available",
			"m quick-bookmark current entity",
			"/ then bookmark add [title]  add a bookmark",
			"/ then bookmark remove  remove bookmarks for this entity",
			"/ then bookmark note <text>  annotate a bookmark",
			"/ then note add [title] [| body]  add a note (| separates body)",
			"/ then note remove [filter]  remove notes (optional title keyword)",
			"/ then note body [filter |] <text>  update note body (filter targets by title)",
			"/ then label add <name>  apply a label",
			"/ then label remove <name>  detach a label from this entity",
			"/ then label delete <name>  delete label definition entirely",
			"/ then label color <name> <color>  set label color",
			"/ then open recent  browse recently visited entities",
			"/ then open bookmarks  browse saved bookmarks",
			"/ then open notes  browse saved notes",
			"/ then open labels  browse labels and attached entities",
			"/ then open views  browse saved investigation views",
			"/ then open cache  reload cached payload for current entity",
			"/ then view save <name>  save current screen and filters",
			"search label:note:bookmark:cache:view:<term> filters local metadata results",
			"tab/shift+tab cycle contract workspace tabs on contract detail views",
			"c raw decode  d decoded decode  e expand fields  v visual navigation",
		)
	}
	return overlayStyle.Width(overlayWidth).Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{titleStyle.Render("Keyboard Help"), ""}, lines...)...))
}

func renderCommandOverlay(snapshot app.Snapshot, width int) string {
	return renderCommandOverlayState(width, snapshot.Command.Prompt, snapshot.Command.Input, snapshot.Command.Results, snapshot.Command.SelectedIndex)
}

func renderCommandOverlayState(width int, prompt string, input string, results []app.SearchResult, selectedIndex int) string {
	overlayWidth := clamp(width-8, 40, 110)
	lines := []string{
		titleStyle.Render(valueOrFallback(prompt, "Search / Command")),
		"> " + input,
	}
	if len(results) == 0 {
		lines = append(lines, "", mutedStyle.Render("Type a tx hash, account, contract, ledger, asset, or command."))
	} else {
		lines = append(lines, "", tableHeaderStyle.Render("Results"))
		flatIndex := 0
		for _, group := range app.GroupSearchResults(results) {
			header := fmt.Sprintf("%s / %s", strings.ToUpper(group.Source), strings.ToUpper(group.Kind))
			lines = append(lines, mutedStyle.Render(header))
			for _, result := range group.Items {
				state := "open"
				if !result.Enabled {
					state = "pending"
				}
				row := fmt.Sprintf("%s: %s (%s)", result.Title, result.Description, state)
				row = truncate(row, overlayWidth-4)
				if selectedIndex == flatIndex {
					row = "> " + row
					row = selectedRowStyle.Width(overlayWidth - 4).Render(row)
				}
				lines = append(lines, row)
				flatIndex++
			}
		}
	}
	return overlayStyle.Width(overlayWidth).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}
