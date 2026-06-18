package pipeline

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func TestS3BackfillSmallRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping S3 integration test")
	}

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgresql://explorer:explorer_dev@localhost:54320/stellar_explorer?sslmode=disable"
	}

	db, err := store.NewPostgresStore(dbURL)
	if err != nil {
		t.Skipf("Skipping: cannot connect to test database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	startLedger := uint32(100)
	endLedger := uint32(104)

	for seq := startLedger; seq <= endLedger; seq++ {
		db.CleanupTestData(ctx, seq)
	}

	p := NewS3BackfillPipeline(db, 2)
	err = p.Run(ctx, startLedger, endLedger)
	if err != nil {
		t.Fatalf("S3 backfill failed: %v", err)
	}

	for seq := startLedger; seq <= endLedger; seq++ {
		var count int
		row := db.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM ledgers WHERE sequence = %d", seq))
		if err := row.Scan(&count); err != nil {
			t.Fatalf("query ledger %d: %v", seq, err)
		}
		if count != 1 {
			t.Errorf("expected 1 ledger row for sequence %d, got %d", seq, count)
		}
	}

	for seq := startLedger; seq <= endLedger; seq++ {
		db.CleanupTestData(ctx, seq)
	}
}
