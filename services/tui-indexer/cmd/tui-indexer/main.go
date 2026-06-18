package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/pipeline"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/publisher"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/readapi"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/source"
	"github.com/miguelnietoa/stellar-explorer/tui-indexer/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Stellar Explorer TUI Indexer\n")
	fmt.Printf("  Network:    %s\n", cfg.Network)
	fmt.Printf("  RPC:        %s\n", cfg.RPCEndpoint)
	fmt.Printf("  Database:   %s\n", cfg.DatabaseURL)
	fmt.Printf("  Workers:    %d\n", cfg.WorkerCount)
	fmt.Printf("  Read API:   %s\n", cfg.ReadAPIAddr)

	if len(os.Args) < 2 {
		fmt.Println("Usage: tui-indexer <live|backfill|s3backfill|migrate|serve>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "live":
		if cfg.RPCEndpoint == "" {
			log.Fatal("RPC_ENDPOINT is required for live command")
		}
		runLive(cfg)
	case "backfill":
		if cfg.RPCEndpoint == "" {
			log.Fatal("RPC_ENDPOINT is required for backfill command")
		}
		runBackfill(cfg)
	case "s3backfill":
		runS3Backfill(cfg)
	case "migrate":
		runMigrate(cfg.DatabaseURL)
	case "serve":
		runServe(cfg)
	default:
		log.Fatalf("Unknown command: %s. Use: live, backfill, s3backfill, migrate, serve", os.Args[1])
	}
}

func setupContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()
	return ctx, cancel
}

func initDeps(cfg *config.Config) (*store.PostgresStore, *source.RPCClient) {
	db, err := store.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	var rpc *source.RPCClient
	if cfg.RPCEndpoint != "" {
		rpc = source.NewRPCClient(cfg.RPCEndpoint)
	}
	return db, rpc
}

func runLive(cfg *config.Config) {
	ctx, cancel := setupContext()
	defer cancel()

	db, rpc := initDeps(cfg)
	defer db.Close()

	passphrase, err := cfg.NetworkPassphrase()
	if err != nil {
		log.Fatalf("Failed to resolve network passphrase: %v", err)
	}
	p := pipeline.NewLivePipeline(rpc, db, passphrase, cfg.BatchSize)

	// Attach Redis publisher if configured
	if cfg.RedisURL != "" {
		pub, err := publisher.NewRedisPublisher(cfg.RedisURL)
		if err != nil {
			log.Printf("Warning: Redis publisher unavailable: %v", err)
		} else {
			defer pub.Close()
			p.SetPublisher(pub)
			log.Println("Redis publisher attached")
		}
	}

	log.Println("Starting live ingestion...")
	if err := p.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Live pipeline failed: %v", err)
	}
	log.Println("Shutdown complete.")
}

func runBackfill(cfg *config.Config) {
	startLedger, endLedger := parseBackfillFlags()

	ctx, cancel := setupContext()
	defer cancel()

	db, rpc := initDeps(cfg)
	defer db.Close()

	passphrase, err := cfg.NetworkPassphrase()
	if err != nil {
		log.Fatalf("Failed to resolve network passphrase: %v", err)
	}
	p := pipeline.NewBackfillPipeline(rpc, db, passphrase, cfg.BatchSize, cfg.WorkerCount)

	log.Printf("Starting backfill from ledger %d to %d...", startLedger, endLedger)
	if err := p.Run(ctx, startLedger, endLedger); err != nil && err != context.Canceled {
		log.Fatalf("Backfill failed: %v", err)
	}
	log.Println("Backfill complete.")
}

func runS3Backfill(cfg *config.Config) {
	startLedger, endLedger := parseBackfillFlags()

	ctx, cancel := setupContext()
	defer cancel()

	db, err := store.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	p := pipeline.NewS3BackfillPipeline(db, cfg.WorkerCount)

	log.Printf("Starting S3 data lake backfill from ledger %d to %d...", startLedger, endLedger)
	if err := p.Run(ctx, startLedger, endLedger); err != nil && err != context.Canceled {
		log.Fatalf("S3 backfill failed: %v", err)
	}
	log.Println("S3 backfill complete.")
}

func parseBackfillFlags() (uint32, uint32) {
	var startLedger, endLedger uint32

	for i := 2; i < len(os.Args)-1; i++ {
		switch os.Args[i] {
		case "--start":
			n, err := strconv.ParseUint(os.Args[i+1], 10, 32)
			if err != nil {
				log.Fatalf("Invalid --start value: %v", err)
			}
			startLedger = uint32(n)
		case "--end":
			n, err := strconv.ParseUint(os.Args[i+1], 10, 32)
			if err != nil {
				log.Fatalf("Invalid --end value: %v", err)
			}
			endLedger = uint32(n)
		}
	}

	if startLedger == 0 || endLedger == 0 {
		log.Fatal("Usage: tui-indexer backfill --start <ledger> --end <ledger>")
	}

	return startLedger, endLedger
}

func runServe(cfg *config.Config) {
	ctx, cancel := setupContext()
	defer cancel()

	db, rpc := initDeps(cfg)
	defer db.Close()

	server := &http.Server{
		Addr:              cfg.ReadAPIAddr,
		Handler:           readapi.NewServer(db, rpc, log.Default()).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Read API shutdown failed: %v", err)
		}
	}()

	log.Printf("Starting read API on %s", cfg.ReadAPIAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Read API failed: %v", err)
	}
	log.Println("Read API shutdown complete.")
}
