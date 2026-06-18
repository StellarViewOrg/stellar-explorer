package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
)

func TestLookupOperationDetailsRendersSummaryAndRelations(t *testing.T) {
	dest := "GDEST"
	sections := lookupOperationDetails(app.LookupSnapshot{
		Kind:  app.LookupOperation,
		Query: "tx-op-1:1",
		Operation: &backendclient.OperationLookupSnapshot{
			ParentTransactionHash: "tx-op-1",
			Operation: backendclient.OperationSummary{
				TransactionHash:  "tx-op-1",
				ApplicationOrder: 1,
				TypeName:         "payment",
				Destination:      &dest,
				Details:          `{"amount":"10"}`,
				CreatedAt:        time.Unix(1700000100, 0).UTC(),
			},
		},
	})

	foundSummary := false
	foundParent := false
	for _, section := range sections {
		if section.Title == "Operation:" && strings.Contains(section.Body, "payment #1") {
			foundSummary = true
		}
		if section.Title == "Parent Transaction" && section.Command == "lookup tx tx-op-1" {
			foundParent = true
		}
	}
	if !foundSummary || !foundParent {
		t.Fatalf("expected operation summary and parent relation, got %#v", sections)
	}
}

func TestAppendTransactionEffectsSectionsShowsIndexedEffects(t *testing.T) {
	lines := appendTransactionEffectsSections(nil, app.LookupSnapshot{
		Source: app.SourceMetadata{Label: "tui-indexer"},
		Transaction: &backendclient.TransactionLookupResponse{
			Effects: []backendclient.EffectSummary{
				{TypeName: "account_credited", Account: "GDEST", Details: `{"asset":"USDC"}`},
			},
		},
	})

	foundEffect := false
	for _, line := range lines {
		if line.Title == "Effect 1" && strings.Contains(line.Body, "account_credited") && line.Command == "lookup account GDEST" {
			foundEffect = true
		}
	}
	if !foundEffect {
		t.Fatalf("expected indexed effect row, got %#v", lines)
	}
}

func TestAppendTransactionEffectsSectionsShowsUnavailableForRPC(t *testing.T) {
	lines := appendTransactionEffectsSections(nil, app.LookupSnapshot{
		Source: app.SourceMetadata{Label: "stellar-rpc"},
		Transaction: &backendclient.TransactionLookupResponse{
			Effects: []backendclient.EffectSummary{
				{TypeName: "account_credited", Account: "GDEST"},
			},
		},
	})

	foundUnavailable := false
	for _, line := range lines {
		if line.Title == "Unavailable" && strings.Contains(line.Body, "indexed backend") {
			foundUnavailable = true
		}
	}
	if !foundUnavailable {
		t.Fatalf("expected RPC degradation message, got %#v", lines)
	}
}

func TestAppendTransactionEffectsSectionsShowsEmptyIndexedState(t *testing.T) {
	lines := appendTransactionEffectsSections(nil, app.LookupSnapshot{
		Source:      app.SourceMetadata{Actual: "hybrid-indexer"},
		Transaction: &backendclient.TransactionLookupResponse{},
	})

	foundEmpty := false
	for _, line := range lines {
		if line.Title == "Indexed" && strings.Contains(line.Body, "No effects indexed") {
			foundEmpty = true
		}
	}
	if !foundEmpty {
		t.Fatalf("expected empty indexed effects message, got %#v", lines)
	}
}
