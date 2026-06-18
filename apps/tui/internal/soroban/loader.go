package soroban

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/sordecode"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// SpecLoader caches contract spec registries loaded from Stellar RPC.
type SpecLoader struct {
	client *rpcclient.Client
	mu     sync.RWMutex
	cache  map[string]*sordecode.SpecRegistry
	raw    map[string]string
}

// NewSpecLoader constructs a spec loader backed by RPC ledger entry reads.
func NewSpecLoader(client *rpcclient.Client) *SpecLoader {
	return &SpecLoader{
		client: client,
		cache:  make(map[string]*sordecode.SpecRegistry),
		raw:    make(map[string]string),
	}
}

// Registry returns a cached spec registry for one contract.
func (l *SpecLoader) Registry(ctx context.Context, contractID string) (*sordecode.SpecRegistry, error) {
	if l == nil || l.client == nil {
		return nil, nil
	}
	l.mu.RLock()
	if registry, ok := l.cache[contractID]; ok {
		l.mu.RUnlock()
		return registry, nil
	}
	l.mu.RUnlock()

	registry, raw, err := l.loadRegistry(ctx, contractID)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.cache[contractID] = registry
	if raw != "" {
		l.raw[contractID] = raw
	}
	l.mu.Unlock()
	return registry, nil
}

// RawSpecJSON returns the parsed spec JSON for one contract when available.
func (l *SpecLoader) RawSpecJSON(contractID string) string {
	if l == nil {
		return ""
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.raw[contractID]
}

func (l *SpecLoader) loadRegistry(ctx context.Context, contractID string) (*sordecode.SpecRegistry, string, error) {
	instanceKey, err := contractInstanceLedgerKey(contractID)
	if err != nil {
		return nil, "", err
	}
	instanceResult, err := l.client.GetLedgerEntries(ctx, []string{instanceKey})
	if err != nil || len(instanceResult.Entries) == 0 {
		return nil, "", err
	}

	wasmHash, _, err := extractWasmHashFromInstance(instanceResult.Entries[0].DataXDR)
	if err != nil {
		return nil, "", nil
	}

	codeKey, err := contractCodeLedgerKey(wasmHash)
	if err != nil {
		return nil, "", err
	}
	codeResult, err := l.client.GetLedgerEntries(ctx, []string{codeKey})
	if err != nil || len(codeResult.Entries) == 0 {
		return nil, "", err
	}

	wasmBytes, err := extractWasmBytecode(codeResult.Entries[0].DataXDR)
	if err != nil {
		return nil, "", err
	}

	specBytes, err := sordecode.ExtractSpecFromWASM(wasmBytes)
	if err != nil || len(specBytes) == 0 {
		return nil, "", nil
	}
	entries, err := sordecode.DecodeEntries(specBytes)
	if err != nil || len(entries) == 0 {
		return nil, "", err
	}

	parsed, err := json.Marshal(sordecode.EntriesToJSON(entries))
	if err != nil {
		return sordecode.NewRegistry(entries), "", nil
	}
	return sordecode.NewRegistry(entries), string(parsed), nil
}

// Instance loads the contract instance ScMap when present.
func (l *SpecLoader) Instance(ctx context.Context, contractID string) (*xdr.ScContractInstance, error) {
	if l == nil || l.client == nil {
		return nil, nil
	}
	instanceKey, err := contractInstanceLedgerKey(contractID)
	if err != nil {
		return nil, err
	}
	result, err := l.client.GetLedgerEntries(ctx, []string{instanceKey})
	if err != nil || len(result.Entries) == 0 {
		return nil, err
	}
	_, instance, err := extractWasmHashFromInstance(result.Entries[0].DataXDR)
	if err != nil {
		return instance, nil
	}
	return instance, nil
}
