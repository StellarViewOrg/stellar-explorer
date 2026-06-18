package cache

import (
	"testing"
	"time"
)

func TestEntityFreshImmutableKindsNeverExpire(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	entity := EntityCache{
		Kind:      "ledger",
		UpdatedAt: now.Add(-48 * time.Hour),
	}

	if !EntityFresh(entity, now) {
		t.Fatal("expected ledger cache to remain fresh")
	}
	if EntityStale(entity, now) {
		t.Fatal("expected ledger cache to never be stale")
	}
}

func TestEntityFreshMutableKindsExpire(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	fresh := EntityCache{Kind: "account", UpdatedAt: now.Add(-2 * time.Minute)}
	stale := EntityCache{Kind: "account", UpdatedAt: now.Add(-10 * time.Minute)}

	if !EntityFresh(fresh, now) {
		t.Fatal("expected recent account cache to be fresh")
	}
	if EntityFresh(stale, now) {
		t.Fatal("expected old account cache to be stale")
	}
	if !EntityStale(stale, now) {
		t.Fatal("expected old account cache to report stale")
	}
}
