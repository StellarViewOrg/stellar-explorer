package soroban

import (
	"context"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/rpcclient"
)

// Enricher applies in-client Soroban decoding to network-backed lookup payloads.
type Enricher struct {
	client *rpcclient.Client
	loader *SpecLoader
}

// NewEnricher constructs an enricher for one RPC client.
func NewEnricher(client *rpcclient.Client) *Enricher {
	if client == nil {
		return nil
	}
	return &Enricher{
		client: client,
		loader: NewSpecLoader(client),
	}
}

// EnrichTransaction adds spec-decoded Soroban operation details when possible.
func (e *Enricher) EnrichTransaction(ctx context.Context, response *backendclient.TransactionLookupResponse) error {
	if e == nil {
		return nil
	}
	return EnrichTransactionOperations(ctx, e.loader, response)
}

// EnrichContract adds spec, storage, and recent events to a contract lookup response.
func (e *Enricher) EnrichContract(ctx context.Context, response *backendclient.ContractLookupResponse) error {
	if e == nil || response == nil || response.Contract == nil {
		return nil
	}

	contractID := response.Contract.ContractID
	rawSpec := e.loader.RawSpecJSON(contractID)
	if rawSpec == "" {
		if _, err := e.loader.Registry(ctx, contractID); err == nil {
			rawSpec = e.loader.RawSpecJSON(contractID)
		}
	}
	if response.Spec == nil && rawSpec != "" {
		response.Spec = BuildContractSpec(contractID, rawSpec)
	}

	if len(response.Storage) == 0 {
		instance, err := e.loader.Instance(ctx, contractID)
		if err == nil && instance != nil {
			response.Storage = StorageFromInstance(ctx, e.loader, contractID, instance, response.Contract.LastModifiedLedger)
			response.Contract.StorageEntryCount = int32(len(response.Storage))
		}
	}

	if len(response.RecentEvents) == 0 {
		events, err := EventsFromRPC(ctx, e.client, e.loader, contractID, 10)
		if err == nil {
			response.RecentEvents = events
			response.Contract.EventCount = int64(len(events))
		}
	}

	return nil
}

// ContractEvents fetches decoded contract events directly from RPC.
func (e *Enricher) ContractEvents(ctx context.Context, contractID string, limit int, offset int) ([]backendclient.ContractEventSummary, error) {
	if e == nil {
		return nil, nil
	}
	events, err := EventsFromRPC(ctx, e.client, e.loader, contractID, limit+offset)
	if err != nil {
		return nil, err
	}
	if offset >= len(events) {
		return nil, nil
	}
	end := offset + limit
	if end > len(events) {
		end = len(events)
	}
	return append([]backendclient.ContractEventSummary(nil), events[offset:end]...), nil
}

// ContractStorage returns decoded instance storage entries for one contract.
func (e *Enricher) ContractStorage(ctx context.Context, contractID string, limit int, offset int) ([]backendclient.ContractStorageSummary, error) {
	if e == nil {
		return nil, nil
	}
	instance, err := e.loader.Instance(ctx, contractID)
	if err != nil || instance == nil {
		return nil, err
	}
	entries := StorageFromInstance(ctx, e.loader, contractID, instance, 0)
	if offset >= len(entries) {
		return nil, nil
	}
	end := offset + limit
	if end > len(entries) {
		end = len(entries)
	}
	return append([]backendclient.ContractStorageSummary(nil), entries[offset:end]...), nil
}

// ContractSpec builds a decoded contract spec for one contract.
func (e *Enricher) ContractSpec(ctx context.Context, contractID string) (backendclient.ContractSpec, error) {
	if e == nil {
		return backendclient.ContractSpec{}, nil
	}
	if _, err := e.loader.Registry(ctx, contractID); err != nil {
		return backendclient.ContractSpec{}, err
	}
	spec := BuildContractSpec(contractID, e.loader.RawSpecJSON(contractID))
	if spec == nil {
		return backendclient.ContractSpec{ContractID: contractID, DecodeStatus: "missing"}, nil
	}
	return *spec, nil
}
