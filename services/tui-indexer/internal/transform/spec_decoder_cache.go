package transform

import (
	"context"
	"sync"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
)

// CachingSpecRegistryLoader memoizes contract spec registries for one ingest batch.
type CachingSpecRegistryLoader struct {
	inner ContractSpecRegistryLoader
	mu    sync.Mutex
	cache map[string]*sordecode.SpecRegistry
}

func NewCachingSpecRegistryLoader(inner ContractSpecRegistryLoader) *CachingSpecRegistryLoader {
	return &CachingSpecRegistryLoader{
		inner: inner,
		cache: make(map[string]*sordecode.SpecRegistry),
	}
}

func (l *CachingSpecRegistryLoader) GetSpecRegistryForContract(ctx context.Context, contractID string) (*sordecode.SpecRegistry, error) {
	if l == nil || l.inner == nil {
		return nil, nil
	}
	l.mu.Lock()
	if cached, ok := l.cache[contractID]; ok {
		l.mu.Unlock()
		return cached, nil
	}
	l.mu.Unlock()

	registry, err := l.inner.GetSpecRegistryForContract(ctx, contractID)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.cache[contractID] = registry
	l.mu.Unlock()
	return registry, nil
}
