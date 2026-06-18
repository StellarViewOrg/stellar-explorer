package transform

import (
	"context"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/sordecode"
)

// ContractSpecRegistryLoader resolves contract spec registries by contract ID.
type ContractSpecRegistryLoader interface {
	GetSpecRegistryForContract(ctx context.Context, contractID string) (*sordecode.SpecRegistry, error)
}

// StaticSpecRegistryLoader is an in-memory registry loader for tests.
type StaticSpecRegistryLoader struct {
	Registries map[string]*sordecode.SpecRegistry
}

func (l StaticSpecRegistryLoader) GetSpecRegistryForContract(_ context.Context, contractID string) (*sordecode.SpecRegistry, error) {
	if l.Registries == nil {
		return nil, nil
	}
	return l.Registries[contractID], nil
}
