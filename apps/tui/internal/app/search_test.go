package app

import "testing"

func TestParseMetadataSearchQuery(t *testing.T) {
	filter, query := ParseMetadataSearchQuery("label:suspicious")
	if filter != "label" || query != "suspicious" {
		t.Fatalf("expected label filter, got filter=%q query=%q", filter, query)
	}

	filter, query = ParseMetadataSearchQuery("GABC")
	if filter != "" || query != "gabc" {
		t.Fatalf("expected plain query, got filter=%q query=%q", filter, query)
	}
}

func TestGroupSearchResultsBucketsBySourceAndKind(t *testing.T) {
	groups := GroupSearchResults([]SearchResult{
		{Kind: "bookmark", Title: "A", Source: "local", Enabled: true},
		{Kind: "account", Title: "B", Source: "indexer", Enabled: true},
		{Kind: "note", Title: "C", Source: "local", Enabled: true},
	})
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Source != "indexer" || groups[0].Kind != "account" {
		t.Fatalf("expected indexer account group first, got %#v", groups[0])
	}
}

func TestRankSearchResultsPrefersExactBookmarkTitle(t *testing.T) {
	results := RankSearchResults("whale", []SearchResult{
		{Kind: "note", Title: "mentions whale activity", Command: "lookup account GOTHER", Enabled: true, Source: "local"},
		{Kind: "bookmark", Title: "whale", Command: "lookup account GABC", Enabled: true, Source: "local"},
	})
	if len(results) != 2 || results[0].Title != "whale" {
		t.Fatalf("expected exact bookmark title first, got %#v", results)
	}
}

func TestRankSearchResultsDeprioritizesDisabledRows(t *testing.T) {
	results := RankSearchResults("suspicious", []SearchResult{
		{Kind: "label", Title: "suspicious", Enabled: false, Source: "local"},
		{Kind: "label", Title: "suspicious", Command: "lookup account GABC", Enabled: true, Source: "local"},
	})
	if len(results) != 2 || !results[0].Enabled {
		t.Fatalf("expected enabled label target first, got %#v", results)
	}
}

func TestInferSearchResultsSupportsAssetQuery(t *testing.T) {
	results := InferSearchResults("USDC:GBBD47IF6LWK7P7MDEVSCWR7DPUWV3NY3DTQEVFL4NAT4AQH3ZLLFLA5")
	if len(results) != 1 || results[0].Kind != "asset" {
		t.Fatalf("expected asset inference, got %#v", results)
	}
}
