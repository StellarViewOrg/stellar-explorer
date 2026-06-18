package cache

import (
	"strings"
	"time"
)

const (
	// DefaultEntityCacheLimit bounds visited entity payloads kept per profile.
	DefaultEntityCacheLimit = 500
)

// EntityTTL returns how long a cached entity remains fresh before revalidation.
// Zero duration means the entry never expires while present.
func EntityTTL(kind string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "ledger", "transaction", "operation":
		return 0
	case "account":
		return 5 * time.Minute
	case "asset", "contract":
		return 10 * time.Minute
	default:
		return 15 * time.Minute
	}
}

// EntityFresh reports whether an entity cache row can be served without revalidation.
func EntityFresh(entity EntityCache, now time.Time) bool {
	if entity.UpdatedAt.IsZero() {
		return false
	}
	ttl := EntityTTL(entity.Kind)
	if ttl == 0 {
		return true
	}
	return now.Sub(entity.UpdatedAt) <= ttl
}

// EntityStale reports whether a cache row exists but is outside its fresh window.
func EntityStale(entity EntityCache, now time.Time) bool {
	if entity.UpdatedAt.IsZero() {
		return false
	}
	ttl := EntityTTL(entity.Kind)
	if ttl == 0 {
		return false
	}
	return now.Sub(entity.UpdatedAt) > ttl
}
