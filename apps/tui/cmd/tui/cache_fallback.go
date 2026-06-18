package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func tryRestoreLookupFromCache(ctx context.Context, model *app.Model, store *cache.Store, kind app.LookupKind, target string, mode lookupCacheMode) bool {
	if model == nil || store == nil {
		return false
	}

	entity, err := store.GetEntityCache(ctx, model.Snapshot().Profile.Name, string(kind), strings.TrimSpace(target))
	if err != nil {
		if errors.Is(err, cache.ErrNoRows()) {
			return false
		}
		model.SetWarningStatus(fmt.Sprintf("Cache lookup failed: %v", err))
		return false
	}

	if restoreLookupFromCache(model, kind, target, *entity, mode) {
		afterCacheRestore(ctx, model, store, string(kind), target)
		return true
	}
	return false
}

func restoreLookupFromCache(model *app.Model, kind app.LookupKind, target string, entity cache.EntityCache, mode lookupCacheMode) bool {
	source := cacheSourceMetadata(model.Snapshot().Profile, entity, mode)
	target = strings.TrimSpace(target)

	switch kind {
	case app.LookupLedger:
		var response backendclient.LedgerLookupResponse
		if err := json.Unmarshal([]byte(entity.Payload), &response); err != nil {
			return false
		}
		model.UpdateLookupLedger(target, response, source)
	case app.LookupTransaction:
		var response backendclient.TransactionLookupResponse
		if err := json.Unmarshal([]byte(entity.Payload), &response); err != nil {
			return false
		}
		model.UpdateLookupTransaction(target, response, source)
	case app.LookupAccount:
		var response backendclient.AccountLookupResponse
		if err := json.Unmarshal([]byte(entity.Payload), &response); err != nil {
			return false
		}
		model.UpdateLookupAccount(target, response, source)
	case app.LookupAsset:
		var response backendclient.AssetLookupResponse
		if err := json.Unmarshal([]byte(entity.Payload), &response); err != nil {
			return false
		}
		model.UpdateLookupAsset(target, response, source)
	case app.LookupContract:
		var response backendclient.ContractLookupResponse
		if err := json.Unmarshal([]byte(entity.Payload), &response); err != nil {
			return false
		}
		model.UpdateLookupContract(target, response, source)
	default:
		return false
	}

	switch mode {
	case lookupCacheHit:
		model.SetInfoStatus(fmt.Sprintf("Loaded cached %s %s.", kind, target))
	case lookupCacheStale:
		model.SetWarningStatus(fmt.Sprintf("Backend unavailable; showing stale cached %s %s from %s.", kind, target, entity.UpdatedAt.UTC().Format(time.RFC3339)))
	default:
		model.SetInfoStatus(fmt.Sprintf("Loaded cached %s %s from local workspace.", kind, target))
	}
	return true
}

func executeOpenCacheCommand(ctx context.Context, model *app.Model, store *cache.Store, args []string) (bool, error) {
	if store == nil {
		model.SetWarningStatus("Local cache is unavailable.")
		return true, nil
	}

	kind := ""
	target := ""
	if len(args) >= 2 {
		kind = strings.TrimSpace(args[0])
		target = strings.TrimSpace(strings.Join(args[1:], " "))
	} else {
		snapshot := model.Snapshot()
		if snapshot.Current != app.ScreenLookup || snapshot.Lookup.State != app.ViewStateReady {
			model.SetWarningStatus("open cache requires an active lookup or explicit kind and target.")
			return true, nil
		}
		kind = string(snapshot.Lookup.Kind)
		target = strings.TrimSpace(snapshot.Lookup.Query)
	}

	normalized := normalizeLookupKind(kind)
	if normalized == "" || target == "" {
		model.SetWarningStatus("Usage: open cache  |  open cache <kind> <target>")
		return true, nil
	}
	if !tryRestoreLookupFromCache(ctx, model, store, normalized, target, lookupCacheHit) {
		model.SetWarningStatus(fmt.Sprintf("No cached payload found for %s %s.", normalized, target))
	}
	return true, nil
}

func cacheSourceMetadata(profile config.Profile, entity cache.EntityCache, mode lookupCacheMode) app.SourceMetadata {
	label := strings.TrimSpace(entity.SourceLabel)
	if label == "" {
		label = "local cache"
	}

	switch mode {
	case lookupCacheHit:
		return app.SourceMetadata{
			Mode:       profile.NormalizedBackendMode(),
			Operation:  "lookup",
			Policy:     "cache-first",
			Preferred:  profile.PreferredSource(),
			Actual:     "cache",
			Label:      label,
			CacheState: "hit",
		}
	case lookupCacheStale:
		return app.SourceMetadata{
			Mode:           profile.NormalizedBackendMode(),
			Operation:      "lookup",
			Policy:         "cache-first",
			Preferred:      profile.PreferredSource(),
			Actual:         "cache",
			Label:          label,
			CacheState:     "stale",
			FallbackUsed:   true,
			Degraded:       true,
			DegradedReason: fmt.Sprintf("backend lookup unavailable; showing cached payload from %s", entity.UpdatedAt.UTC().Format(time.RFC3339)),
		}
	default:
		return app.SourceMetadata{
			Mode:           profile.NormalizedBackendMode(),
			Operation:      "lookup-cache",
			Policy:         "cache-fallback",
			Preferred:      profile.PreferredSource(),
			Actual:         "cache",
			Label:          label,
			CacheState:     "stale",
			FallbackUsed:   true,
			Degraded:       true,
			DegradedReason: "backend lookup unavailable; showing cached payload",
		}
	}
}
