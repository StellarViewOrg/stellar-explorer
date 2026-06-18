package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

// BackfillPipeline processes a historical range of ledgers using parallel workers.
type BackfillPipeline struct {
	rpc               *source.RPCClient
	store             *store.PostgresStore
	networkPassphrase string
	batchSize         int
	workerCount       int
}

func NewBackfillPipeline(rpc *source.RPCClient, store *store.PostgresStore, networkPassphrase string, batchSize, workerCount int) *BackfillPipeline {
	return &BackfillPipeline{
		rpc:               rpc,
		store:             store,
		networkPassphrase: networkPassphrase,
		batchSize:         batchSize,
		workerCount:       workerCount,
	}
}

// Run processes ledgers from startLedger to endLedger (inclusive) using parallel workers.
func (p *BackfillPipeline) Run(ctx context.Context, startLedger, endLedger uint32) error {
	if endLedger < startLedger {
		return fmt.Errorf("endLedger (%d) must be >= startLedger (%d)", endLedger, startLedger)
	}

	totalLedgers := endLedger - startLedger + 1
	log.Printf("backfill: processing ledgers %d to %d (%d total) with %d workers",
		startLedger, endLedger, totalLedgers, p.workerCount)

	// Split range across workers
	workers := p.workerCount
	if int(totalLedgers) < workers {
		workers = int(totalLedgers)
	}

	chunkSize := totalLedgers / uint32(workers)
	var processedTotal atomic.Int64
	var wg sync.WaitGroup
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		workerStart := startLedger + uint32(i)*chunkSize
		workerEnd := workerStart + chunkSize - 1
		if i == workers-1 {
			workerEnd = endLedger // last worker takes the remainder
		}

		go func(id int, start, end uint32) {
			defer wg.Done()
			err := p.runWorker(ctx, id, start, end, &processedTotal)
			if err != nil {
				errCh <- fmt.Errorf("worker %d: %w", id, err)
			}
		}(i, workerStart, workerEnd)
	}

	// Wait for all workers
	wg.Wait()
	close(errCh)

	// Collect errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	log.Printf("backfill: completed. processed %d ledgers", processedTotal.Load())

	if len(errs) > 0 {
		return fmt.Errorf("backfill completed with %d errors: %v", len(errs), errs[0])
	}
	return nil
}

func (p *BackfillPipeline) runWorker(ctx context.Context, id int, start, end uint32, processed *atomic.Int64) error {
	log.Printf("backfill worker %d: processing ledgers %d to %d", id, start, end)

	cursor := start
	for cursor <= end {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		remaining := int(end - cursor + 1)
		limit := p.batchSize
		if remaining < limit {
			limit = remaining
		}

		count, err := p.processLedgerBatch(ctx, cursor, limit)
		if err != nil {
			return fmt.Errorf("batch at ledger %d: %w", cursor, err)
		}

		if count == 0 {
			break
		}

		cursor += uint32(count)
		total := processed.Add(int64(count))

		if total%100 == 0 {
			log.Printf("backfill worker %d: progress %d/%d (total across workers: %d)",
				id, cursor-start, end-start+1, total)
		}
	}

	log.Printf("backfill worker %d: done", id)
	return nil
}

// processLedgerBatch fetches a batch of ledgers via getLedgers and extracts
// transactions from each ledger's MetadataXDR (LedgerCloseMeta).
func (p *BackfillPipeline) processLedgerBatch(ctx context.Context, startLedger uint32, limit int) (int, error) {
	ledgerResult, err := p.rpc.GetLedgers(ctx, source.GetLedgersParams{
		StartLedger: startLedger,
		Pagination:  &source.Pagination{Limit: limit},
	})
	if err != nil {
		return 0, fmt.Errorf("getLedgers: %w", err)
	}

	if len(ledgerResult.Ledgers) == 0 {
		return 0, nil
	}

	processed := 0
	for _, rpcLedger := range ledgerResult.Ledgers {
		lcm, err := source.DecodeMetadataXDR(rpcLedger.MetadataXDR)
		if err != nil {
			return processed, fmt.Errorf("decode ledger %d: %w", rpcLedger.Sequence, err)
		}

		txEntries, err := source.TransactionEntriesFromCloseMeta(lcm, p.networkPassphrase)
		if err != nil {
			return processed, fmt.Errorf("extract txs ledger %d: %w", rpcLedger.Sequence, err)
		}

		if err := ProcessOneLedger(ctx, nil, p.store, nil, p.networkPassphrase, rpcLedger, txEntries); err != nil {
			return processed, fmt.Errorf("process ledger %d: %w", rpcLedger.Sequence, err)
		}
		processed++
	}

	return processed, nil
}
