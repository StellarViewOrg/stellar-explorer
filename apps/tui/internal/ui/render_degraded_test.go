package ui

import (
	"strings"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/ui/fixtures"
)

func TestRenderDegradedCacheFallbackLookup(t *testing.T) {
	view := RenderWithSize(fixtures.CacheFallbackTransactionModel().Snapshot(), 140, 60)
	for _, needle := range []string{"degraded", "local cache", "tx-cache-degraded", "State: ready"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected cache fallback render to include %q, got %q", needle, view)
		}
	}
}

func TestRenderDegradedHybridFallbackLedger(t *testing.T) {
	view := RenderWithSize(fixtures.HybridFallbackLedgerModel().Snapshot(), 140, 60)
	for _, needle := range []string{"degraded", "rpc fallback", "12345", "ledger-fallback"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected hybrid fallback render to include %q, got %q", needle, view)
		}
	}
}

func TestRenderDegradedLiveFeedUnavailable(t *testing.T) {
	view := RenderWithSize(fixtures.LiveFeedUnavailableModel().Snapshot(), 140, 60)
	for _, needle := range []string{"backend unavailable", "connection refused", "degraded"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected live feed degraded render to include %q, got %q", needle, view)
		}
	}
}

func TestRenderRPCTransactionOmitsIndexedEffects(t *testing.T) {
	view := RenderWithSize(fixtures.RPCEffectsUnavailableModel().Snapshot(), 140, 60)
	if strings.Contains(view, "account_credited") {
		t.Fatalf("expected rpc-only transaction render to omit indexed effects, got %q", view)
	}
	if !strings.Contains(view, "Effects") || !strings.Contains(view, "Unavailable") || !strings.Contains(view, "indexed backend") {
		t.Fatalf("expected effects unavailable notice, got %q", view)
	}
}

func TestRenderSorobanRawDecodeMode(t *testing.T) {
	model := fixtures.SorobanRawDecodeModel()
	view := RenderWithSize(model.Snapshot(), 140, 60)
	if !strings.Contains(view, "decode=raw") {
		t.Fatalf("expected raw soroban render to include decode=raw, got %q", view)
	}

	lookup := model.Snapshot().Lookup
	lookup.ContractTab = app.ContractWorkspaceTabStorage
	for _, section := range lookupContractTabSections(lookup) {
		if strings.Contains(section.Body, "raw-key-xdr") && strings.Contains(section.Body, "raw-value-xdr") {
			return
		}
	}
	t.Fatalf("expected storage tab to render raw xdr values, got %#v", lookupContractTabSections(lookup))
}

func TestRenderHeaderShowsCacheDegradedWhenUnavailable(t *testing.T) {
	model := fixtures.CacheFallbackTransactionModel()
	model.UpdateCacheSnapshot(app.CacheSnapshot{
		Enabled:   true,
		Available: false,
		Path:      "/tmp/cache.db",
	})
	view := RenderWithSize(model.Snapshot(), 140, 60)
	if !strings.Contains(view, "cache:degraded") {
		t.Fatalf("expected cache degraded header label, got %q", view)
	}
}
