package main

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

type lookupCacheMode int

const (
	lookupCacheFallback lookupCacheMode = iota
	lookupCacheHit
	lookupCacheStale
)

type lookupCacheOptions struct {
	bypassCache bool
}

func performLedgerLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	query string,
	sequence uint32,
	opts lookupCacheOptions,
) (bool, error) {
	return performEntityLookup(ctx, model, store, backend, app.LookupLedger, query, query, opts, func(ctx context.Context) (any, error) {
		return backend.Ledger(ctx, sequence)
	}, func(model *app.Model, target string, payload any, source app.SourceMetadata) {
		model.UpdateLookupLedger(target, payload.(backendclient.LedgerLookupResponse), source)
	})
}

func performTransactionLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	query string,
	opts lookupCacheOptions,
) (bool, error) {
	return performEntityLookup(ctx, model, store, backend, app.LookupTransaction, query, query, opts, func(ctx context.Context) (any, error) {
		return backend.Transaction(ctx, query)
	}, func(model *app.Model, target string, payload any, source app.SourceMetadata) {
		model.UpdateLookupTransaction(target, payload.(backendclient.TransactionLookupResponse), source)
	})
}

func performAccountLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	query string,
	opts lookupCacheOptions,
) (bool, error) {
	return performEntityLookup(ctx, model, store, backend, app.LookupAccount, query, query, opts, func(ctx context.Context) (any, error) {
		return backend.Account(ctx, query)
	}, func(model *app.Model, target string, payload any, source app.SourceMetadata) {
		model.UpdateLookupAccount(target, payload.(backendclient.AccountLookupResponse), source)
	})
}

func performAssetLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	displayQuery string,
	target string,
	code string,
	issuer string,
	opts lookupCacheOptions,
) (bool, error) {
	return performEntityLookup(ctx, model, store, backend, app.LookupAsset, displayQuery, target, opts, func(ctx context.Context) (any, error) {
		return backend.Asset(ctx, code, issuer)
	}, func(model *app.Model, target string, payload any, source app.SourceMetadata) {
		model.UpdateLookupAsset(target, payload.(backendclient.AssetLookupResponse), source)
	})
}

func performContractLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	query string,
	opts lookupCacheOptions,
) (bool, error) {
	return performEntityLookup(ctx, model, store, backend, app.LookupContract, query, query, opts, func(ctx context.Context) (any, error) {
		return backend.Contract(ctx, query)
	}, func(model *app.Model, target string, payload any, source app.SourceMetadata) {
		model.UpdateLookupContract(target, payload.(backendclient.ContractLookupResponse), source)
	})
}

type lookupFetcher func(ctx context.Context) (any, error)

type lookupApplier func(model *app.Model, target string, payload any, source app.SourceMetadata)

func performEntityLookup(
	ctx context.Context,
	model *app.Model,
	store *cache.Store,
	backend lookupBackend,
	kind app.LookupKind,
	displayQuery string,
	cacheTarget string,
	opts lookupCacheOptions,
	fetch lookupFetcher,
	apply lookupApplier,
) (bool, error) {
	profile := model.Snapshot().Profile
	source := sourceMetadataFor(profile, string(kind), backend)
	cacheTarget = strings.TrimSpace(cacheTarget)
	displayQuery = strings.TrimSpace(displayQuery)

	if store != nil && !opts.bypassCache {
		entity, err := store.GetEntityCache(ctx, profile.Name, string(kind), cacheTarget)
		if err == nil && cache.EntityFresh(*entity, time.Now().UTC()) {
			if restoreLookupFromCache(model, kind, displayQuery, *entity, lookupCacheHit) {
				afterCacheRestore(ctx, model, store, string(kind), cacheTarget)
				return true, nil
			}
		}
	}

	model.SetLookupLoading(kind, displayQuery, source)
	payload, err := fetch(ctx)
	if err == nil {
		apply(model, displayQuery, payload, source)
		afterLookupLoaded(ctx, model, store, string(kind), cacheTarget, payload)
		return true, nil
	}

	if store != nil && tryRestoreLookupFromCache(ctx, model, store, kind, cacheTarget, lookupCacheStale) {
		return true, nil
	}

	model.SetLookupError(kind, displayQuery, err, source)
	return true, nil
}

func afterCacheRestore(ctx context.Context, model *app.Model, store *cache.Store, kind string, target string) {
	if model == nil {
		return
	}
	metadata, err := loadLookupMetadata(ctx, store, model.Snapshot().Profile.Name, kind, target)
	if err != nil {
		model.SetWarningStatus(err.Error())
		return
	}
	model.SetLookupMetadata(metadata)
}

func refreshActiveLookup(ctx context.Context, cfg config.Config, model *app.Model, store *cache.Store) {
	if model == nil {
		return
	}
	snapshot := model.Snapshot()
	if snapshot.Current != app.ScreenLookup || snapshot.Lookup.State != app.ViewStateReady {
		return
	}
	if strings.TrimSpace(snapshot.Lookup.Query) == "" || snapshot.Lookup.Kind == "" {
		return
	}

	backend, err := initializeLookupBackend(cfg)
	if err != nil || backend == nil {
		return
	}

	opts := lookupCacheOptions{bypassCache: true}
	kind := snapshot.Lookup.Kind
	query := snapshot.Lookup.Query

	switch kind {
	case app.LookupLedger:
		sequence, err := strconvParseUint32(query)
		if err != nil {
			return
		}
		_, _ = performLedgerLookup(ctx, model, store, backend, query, sequence, opts)
	case app.LookupTransaction:
		_, _ = performTransactionLookup(ctx, model, store, backend, query, opts)
	case app.LookupAccount:
		_, _ = performAccountLookup(ctx, model, store, backend, query, opts)
	case app.LookupAsset:
		code, issuer, ok := strings.Cut(query, ":")
		if !ok {
			return
		}
		_, _ = performAssetLookup(ctx, model, store, backend, query, query, strings.TrimSpace(code), strings.TrimSpace(issuer), opts)
	case app.LookupContract:
		_, _ = performContractLookup(ctx, model, store, backend, query, opts)
	}
}

func strconvParseUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, err
	}
	if parsed == 0 {
		return 0, errors.New("sequence must be positive")
	}
	return uint32(parsed), nil
}
