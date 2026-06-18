package ui

import (
	"fmt"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
)

func relatedEntityRow(title, body, copyValue, command, hint string) lookupSection {
	return lookupSection{
		Title:   title,
		Body:    body,
		Copy:    copyValue,
		Command: command,
		Hint:    hint,
	}
}

func explorerContextHeader(lookup app.LookupSnapshot) lookupSection {
	explorer := lookup.Explorer
	if explorer == nil {
		return lookupSection{}
	}

	parent := strings.TrimSpace(explorer.ParentLabel)
	if parent == "" {
		parent = string(lookup.Kind) + " " + strings.TrimSpace(lookup.Query)
	}
	title := strings.TrimSpace(explorer.Title)
	if title == "" {
		title = string(explorer.Kind)
	}

	body := parent + " › " + title
	if explorer.ListLimit > 0 || explorer.ListOffset > 0 {
		body += fmt.Sprintf("  (limit %d offset %d)", explorer.ListLimit, explorer.ListOffset)
	}
	if lookup.Source.Degraded {
		body += "  degraded"
	}

	return lookupSection{
		Title: "Context",
		Body:  body,
		Emph:  true,
	}
}

func routeStepLabel(step app.LookupRouteStep, active bool) string {
	label := strings.ToLower(string(step.Kind))
	if query := formatBreadcrumbQuery(step.Query); query != "" {
		label += " " + query
	}
	if step.ExplorerKind != "" {
		title := strings.TrimSpace(step.Title)
		if title == "" {
			title = strings.ToLower(string(step.ExplorerKind))
		}
		label += " › " + strings.ToLower(title)
	}
	return label
}
