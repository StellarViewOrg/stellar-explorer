package readapi

import (
	"context"
	"strings"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/transform"
)

type specEnrichStore interface {
	transform.ContractSpecRegistryLoader
	LoadTransactionEntriesByHash(ctx context.Context, hashes []string) (map[string]store.ReadTransactionDetail, error)
	BackfillOperationDetails(ctx context.Context, transactionHash string, applicationOrder int32, createdAt time.Time, details string) error
	BackfillContractEventSpecDecode(ctx context.Context, event store.ReadContractEventSummary) error
}

func (s *Server) enrichReadOperations(ctx context.Context, ops []store.ReadOperationSummary) error {
	enrichStore, ok := s.store.(specEnrichStore)
	if !ok || len(ops) == 0 {
		return nil
	}

	originalDetails := make([]string, len(ops))
	for index, op := range ops {
		if transform.OperationNeedsSpecReEnrich(op) {
			originalDetails[index] = op.Details
		}
	}

	hashes := uniqueOperationTransactionHashes(ops)
	if len(hashes) == 0 {
		return nil
	}
	detailsByHash, err := enrichStore.LoadTransactionEntriesByHash(ctx, hashes)
	if err != nil {
		return err
	}
	txByHash := make(map[string]source.TransactionEntry, len(detailsByHash))
	for hash, detail := range detailsByHash {
		txByHash[hash] = transform.TransactionEntryFromReadDetail(&detail)
	}
	loader := transform.NewCachingSpecRegistryLoader(enrichStore)
	if err := transform.ReEnrichReadOperationSummaries(ctx, loader, ops, txByHash); err != nil {
		return err
	}
	s.backfillImprovedOperations(ctx, enrichStore, ops, originalDetails)
	return nil
}

func (s *Server) enrichReadContractEvents(ctx context.Context, events []store.ReadContractEventSummary) error {
	enrichStore, ok := s.store.(specEnrichStore)
	if !ok || len(events) == 0 {
		return nil
	}

	originalValues := make([]*string, len(events))
	for index, event := range events {
		if transform.EventNeedsSpecReEnrich(event) {
			originalValues[index] = cloneOptionalString(event.ValueDecoded)
		}
	}

	loader := transform.NewCachingSpecRegistryLoader(enrichStore)
	if err := transform.ReEnrichReadContractEventSummaries(ctx, loader, events); err != nil {
		return err
	}
	for index := range events {
		store.RepopulateContractEventSummary(&events[index])
	}
	s.backfillImprovedEvents(ctx, enrichStore, events, originalValues)
	return nil
}

func (s *Server) backfillImprovedOperations(
	ctx context.Context,
	enrichStore specEnrichStore,
	ops []store.ReadOperationSummary,
	originalDetails []string,
) {
	for index, op := range ops {
		before := originalDetails[index]
		if before == "" {
			continue
		}
		if !transform.OperationDetailsSpecImproved(before, op.Details) {
			continue
		}
		if err := enrichStore.BackfillOperationDetails(
			ctx,
			op.TransactionHash,
			op.ApplicationOrder,
			op.CreatedAt,
			op.Details,
		); err != nil && s.log != nil {
			s.log.Printf("spec backfill operation %s:%d: %v", op.TransactionHash, op.ApplicationOrder, err)
		}
	}
}

func (s *Server) backfillImprovedEvents(
	ctx context.Context,
	enrichStore specEnrichStore,
	events []store.ReadContractEventSummary,
	originalValues []*string,
) {
	for index, event := range events {
		before := originalValues[index]
		if before == nil {
			continue
		}
		if !transform.EventValueSpecImproved(before, event.ValueDecoded) {
			continue
		}
		if err := enrichStore.BackfillContractEventSpecDecode(ctx, event); err != nil && s.log != nil {
			s.log.Printf("spec backfill event %s %s: %v", event.ContractID, event.TransactionHash, err)
		}
	}
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	clone := strings.TrimSpace(*value)
	return &clone
}

func uniqueOperationTransactionHashes(ops []store.ReadOperationSummary) []string {
	seen := make(map[string]struct{}, len(ops))
	hashes := make([]string, 0, len(ops))
	for _, op := range ops {
		hash := strings.TrimSpace(op.TransactionHash)
		if hash == "" {
			continue
		}
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}
		hashes = append(hashes, hash)
	}
	return hashes
}
