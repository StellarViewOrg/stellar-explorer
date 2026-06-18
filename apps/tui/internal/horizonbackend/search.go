package horizonbackend

import (
	"context"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/searchinfer"
)

func (b *Backend) Search(ctx context.Context, query string, limit int) (backendclient.SearchResponse, error) {
	_ = ctx
	query = strings.TrimSpace(query)
	if query == "" || limit <= 0 {
		return backendclient.SearchResponse{Results: []backendclient.SearchResult{}}, nil
	}

	candidates := searchinfer.FromQuery(query)
	if len(candidates) == 0 {
		return backendclient.SearchResponse{Results: []backendclient.SearchResult{}}, nil
	}
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]backendclient.SearchResult, 0, len(candidates))
	for _, candidate := range candidates {
		results = append(results, backendclient.SearchResult{
			Kind:        candidate.Kind,
			Title:       candidate.Title,
			Description: candidate.Description,
			Command:     candidate.Command,
			Source:      "horizon",
		})
	}
	return backendclient.SearchResponse{Results: results}, nil
}
