package app

import (
	"fmt"
	"sort"
	"strings"
)

const defaultSearchResultLimit = 8

// SearchGroup is a renderer-friendly bucket of command palette results.
type SearchGroup struct {
	Source string
	Kind   string
	Items  []SearchResult
}

// ParseMetadataSearchQuery splits a free-text query from optional local metadata filters.
func ParseMetadataSearchQuery(input string) (filter string, query string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ""
	}
	prefix, remainder, found := strings.Cut(input, ":")
	if !found {
		return "", strings.ToLower(input)
	}
	switch strings.ToLower(strings.TrimSpace(prefix)) {
	case "label", "bookmark", "note", "cache", "view", "watch":
		return strings.ToLower(strings.TrimSpace(prefix)), strings.ToLower(strings.TrimSpace(remainder))
	default:
		return "", strings.ToLower(input)
	}
}

// GroupSearchResults buckets palette rows by source and kind for rendering.
func GroupSearchResults(results []SearchResult) []SearchGroup {
	if len(results) == 0 {
		return nil
	}

	order := make([]string, 0, len(results))
	groups := make(map[string]*SearchGroup)
	for _, result := range results {
		source := strings.ToLower(strings.TrimSpace(result.Source))
		if source == "" {
			source = "local"
		}
		kind := strings.ToLower(strings.TrimSpace(result.Kind))
		if kind == "" {
			kind = "result"
		}
		key := source + "/" + kind
		group, ok := groups[key]
		if !ok {
			group = &SearchGroup{Source: source, Kind: kind}
			groups[key] = group
			order = append(order, key)
		}
		group.Items = append(group.Items, result)
	}

	out := make([]SearchGroup, 0, len(order))
	for _, key := range order {
		out = append(out, *groups[key])
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := searchGroupPriority(out[i])
		right := searchGroupPriority(out[j])
		if left == right {
			return i < j
		}
		return left < right
	})
	return out
}

func searchGroupPriority(group SearchGroup) int {
	source := strings.ToLower(strings.TrimSpace(group.Source))
	kind := strings.ToLower(strings.TrimSpace(group.Kind))
	sourceScore := map[string]int{
		"indexer": 0,
		"horizon": 10,
		"local":   20,
	}[source]
	kindScore := map[string]int{
		"account":     0,
		"transaction": 10,
		"contract":    20,
		"asset":       30,
		"ledger":      40,
		"bookmark":    50,
		"note":        60,
		"label":       70,
		"cache":       80,
		"view":        90,
		"watch":       100,
		"command":     110,
		"result":      120,
	}[kind]
	return sourceScore + kindScore
}

// SearchMoreResult builds a palette row that requests the next backend page.
func SearchMoreResult(query string, offset int) SearchResult {
	return SearchResult{
		Kind:        "command",
		Title:       "Load more results",
		Description: fmt.Sprintf("%q offset %d", strings.TrimSpace(query), offset),
		Command:     fmt.Sprintf("search more %q %d", strings.TrimSpace(query), offset),
		Enabled:     true,
		Source:      "local",
	}
}

// SearchResultLimit returns the default backend/local search page size.
func SearchResultLimit() int {
	return defaultSearchResultLimit
}
