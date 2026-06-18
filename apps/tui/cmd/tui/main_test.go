package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/networkbackend"
)

func TestNormalizeArgs(t *testing.T) {
	args := normalizeArgs([]string{" home ", "lookup"})
	if args[0] != "home" {
		t.Fatalf("expected trimmed arg, got %q", args[0])
	}
}

func TestRunNonInteractiveWithConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","backend_mode":"rpc"}]}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--screen", "lookup"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}
}

func TestRunNonInteractiveLiveScreenFetchesBackendSummary(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","indexer_url":"http://127.0.0.1:8081","backend_mode":"hybrid"}]}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	previousFactory := openLookupBackend
	t.Cleanup(func() {
		openLookupBackend = previousFactory
	})
	openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
		return fakeLookupBackend{
			liveFeedSummary: backendclient.LiveFeedSummaryResponse{
				LastIngestedLedger: 321,
				RecentTransactions: []backendclient.TransactionSummary{
					{Hash: "tx-live-1", LedgerSequence: 321, Account: "GABC"},
				},
			},
		}, nil
	}

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--screen", "live-feed"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}

	data, err := os.ReadFile(stdout.Name())
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "Last ingested ledger: 321") {
		t.Fatalf("expected live feed output, got %q", output)
	}
}

func TestInitializeStartupStateRestoresSessionFromCache(t *testing.T) {
	driverName := registerFakeSQLiteDriver(t)
	cfg := testConfigWithCache(driverName)

	store, err := openCacheStore(context.Background(), driverName, "shared")
	if err != nil {
		t.Fatalf("open cache store: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := persistSessionState(store, "lookup"); err != nil {
		t.Fatalf("persist session state: %v", err)
	}

	startup, err := initializeStartupState(cfg)
	if err != nil {
		t.Fatalf("initialize startup state: %v", err)
	}
	defer func() { _ = startup.store.Close() }()

	if startup.restoredScreen != "lookup" {
		t.Fatalf("expected lookup screen to be restored, got %q", startup.restoredScreen)
	}

	if !startup.snapshot.Available {
		t.Fatal("expected cache snapshot to report available cache")
	}

	if startup.snapshot.Profiles != 1 {
		t.Fatalf("expected one cached profile, got %d", startup.snapshot.Profiles)
	}
}

func TestRunPersistsScreenSelectionIntoCache(t *testing.T) {
	dir := t.TempDir()
	driverName := registerFakeSQLiteDriver(t)
	configPath := filepath.Join(dir, "config.json")
	content := fmt.Sprintf(`{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","backend_mode":"rpc"}],"cache":{"driver":"%s","path":"shared"}}`, driverName)

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--screen", "settings"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}

	startup, err := initializeStartupState(testConfigWithCache(driverName))
	if err != nil {
		t.Fatalf("initialize startup state: %v", err)
	}
	defer func() { _ = startup.store.Close() }()

	if startup.restoredScreen != "settings" {
		t.Fatalf("expected settings screen to be restored, got %q", startup.restoredScreen)
	}
}

func TestRunLookupCommandRendersTransactionResult(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","indexer_url":"http://127.0.0.1:8081","backend_mode":"rpc"}]}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	original := openLookupBackend
	openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
		return stubLookupBackend{
			transaction: backendclient.TransactionLookupResponse{
				Transaction: &backendclient.TransactionDetail{
					Hash:           "tx-123",
					Account:        "GTEST",
					LedgerSequence: 77,
					OperationCount: 2,
					Status:         1,
				},
			},
		}, nil
	}
	defer func() { openLookupBackend = original }()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--command", "lookup tx tx-123"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}

	if _, err := stdout.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek stdout: %v", err)
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if !strings.Contains(string(output), "Route: transaction tx-123") {
		t.Fatalf("expected transaction lookup route, got %q", string(output))
	}
	if !strings.Contains(string(output), "Loaded transaction tx-123") {
		t.Fatalf("expected transaction lookup status, got %q", string(output))
	}
}

func TestRunLookupCommandRendersLedgerResult(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","backend_mode":"rpc"}]}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	original := openLookupBackend
	openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
		return stubLookupBackend{
			ledger: backendclient.LedgerLookupResponse{
				Ledger: &backendclient.LedgerSummary{
					Sequence:         12345,
					Hash:             "ledger-hash",
					TransactionCount: 2,
					OperationCount:   5,
				},
			},
		}, nil
	}
	defer func() { openLookupBackend = original }()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--command", "lookup ledger 12345"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}

	if _, err := stdout.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek stdout: %v", err)
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if !strings.Contains(string(output), "Route: ledger 12345") {
		t.Fatalf("expected ledger lookup route, got %q", string(output))
	}
	if !strings.Contains(string(output), "Loaded ledger 12345") {
		t.Fatalf("expected ledger lookup status, got %q", string(output))
	}
}

func TestRunOpenTxCommandUsesLiveFeedSelection(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"product_name":"tui","default_profile":"default","profiles":[{"name":"default","network":"testnet","rpc_endpoint":"https://example.com","indexer_url":"http://127.0.0.1:8081","backend_mode":"rpc"}]}`

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	original := openLookupBackend
	openLookupBackend = func(profile config.Profile) (lookupBackend, error) {
		return stubLookupBackend{
			liveFeed: backendclient.LiveFeedSummaryResponse{
				LastIngestedLedger: 111,
				RecentTransactions: []backendclient.TransactionSummary{
					{Hash: "tx-live-1", LedgerSequence: 111, Account: "GAAA"},
				},
			},
			transaction: backendclient.TransactionLookupResponse{
				Transaction: &backendclient.TransactionDetail{
					Hash:           "tx-live-1",
					Account:        "GAAA",
					LedgerSequence: 111,
					OperationCount: 1,
					Status:         1,
				},
			},
		}, nil
	}
	defer func() { openLookupBackend = original }()

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer stdin.Close()

	stdout, err := os.CreateTemp(dir, "stdout")
	if err != nil {
		t.Fatalf("create stdout temp file: %v", err)
	}
	defer stdout.Close()

	stderr, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatalf("create stderr temp file: %v", err)
	}
	defer stderr.Close()

	exitCode := run([]string{"--config", configPath, "--no-interactive", "--screen", "live-feed", "--command", "open tx 1"}, stdin, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("expected successful run, got exit code %d", exitCode)
	}

	if _, err := stdout.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek stdout: %v", err)
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if !strings.Contains(string(output), "Route: transaction tx-live-1") {
		t.Fatalf("expected selected transaction route, got %q", string(output))
	}
	if !strings.Contains(string(output), "Loaded transaction tx-live-1") {
		t.Fatalf("expected selected transaction status, got %q", string(output))
	}
}

func TestRunOpenAccountCommandUsesTransactionContext(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-ctx-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-ctx-1",
				Account:        "GCTX",
				LedgerSequence: 15,
				OperationCount: 2,
				Status:         1,
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{
		account: backendclient.AccountLookupResponse{
			Account:    &backendclient.AccountDetail{ID: "GCTX", Balance: "100.0", Sequence: 99},
			Trustlines: []backendclient.TrustlineSummary{{AssetCode: "USDC"}},
			Signers:    []backendclient.AccountSignerSummary{{SignerKey: "GCTX"}},
		},
	}, "open account")
	if err != nil {
		t.Fatalf("executeOpenCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Lookup.Account; got == nil || got.Account == nil || got.Account.ID != "GCTX" {
		t.Fatalf("expected account lookup context to load, got %#v", got)
	}
}

func TestExecuteOpenCommandUsesLedgerLookupTransactionContext(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"55",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 55, Hash: "ledger-55"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-2", LedgerSequence: 55, Account: "GLEDGER", OperationCount: 2, Status: 1},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{
		transaction: backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:           "tx-ledger-2",
				Account:        "GLEDGER",
				LedgerSequence: 55,
				OperationCount: 2,
				Status:         1,
			},
		},
	}, "open tx 1")
	if err != nil {
		t.Fatalf("executeOpenCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Lookup.Transaction; got == nil || got.Transaction == nil || got.Transaction.Hash != "tx-ledger-2" {
		t.Fatalf("expected ledger transaction context to open tx detail, got %#v", got)
	}
}

func TestExecuteNavigationChainPreservesLookupSelection(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"55",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 55, Hash: "ledger-55"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-1", LedgerSequence: 55, Account: "GLEDGER", OperationCount: 2, Status: 1},
				{Hash: "tx-ledger-2", LedgerSequence: 55, Account: "GLEDGER", OperationCount: 1, Status: 1},
			},
		},
	)
	model.SetLookupSelection(5, 1)

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open txs"); err != nil {
		t.Fatalf("executeOpenCommand(open txs) error = %v", err)
	}
	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{
		transaction: backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-ledger-1", LedgerSequence: 55, Account: "GLEDGER"},
		},
	}, "lookup tx tx-ledger-1"); err != nil {
		t.Fatalf("executeOpenCommand(lookup tx) error = %v", err)
	}
	if !model.Back() {
		t.Fatal("expected first back to return to transaction explorer")
	}
	if !model.Back() {
		t.Fatal("expected second back to return to ledger detail")
	}

	lookup := model.Snapshot().Lookup
	if lookup.Explorer != nil {
		t.Fatalf("expected ledger detail after back chain, got explorer %#v", lookup.Explorer)
	}
	if lookup.SelectedSection != 5 || lookup.ScrollOffset != 1 {
		t.Fatalf("expected restored ledger selection 5/1, got %d/%d", lookup.SelectedSection, lookup.ScrollOffset)
	}
	route := model.Snapshot().LookupRoute
	if len(route) != 1 || route[0].Kind != app.LookupLedger || route[0].Query != "55" {
		t.Fatalf("expected single ledger route step, got %#v", route)
	}
}

func TestExecuteOpenCommandOpensAndClosesTransactionExplorer(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"55",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 55, Hash: "ledger-55"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-1", LedgerSequence: 55, Account: "GLEDGER", OperationCount: 2, Status: 1},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open txs")
	if err != nil {
		t.Fatalf("executeOpenCommand(open txs) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Lookup.Explorer; got == nil || got.Kind != app.LookupExplorerTransactions {
		t.Fatalf("expected transaction explorer to open, got %#v", got)
	}

	keepRunning, err = executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open detail")
	if err != nil {
		t.Fatalf("executeOpenCommand(open detail) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Lookup.Explorer; got != nil {
		t.Fatalf("expected transaction explorer to close, got %#v", got)
	}
}

func TestExecuteOpenCommandOpensTopLevelLedgerExplorer(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})

	keepRunning, err := executeOpenCommand(context.Background(), model, stubExplorerBackend{
		stubLookupBackend: stubLookupBackend{},
		ledgers: []backendclient.LedgerSummary{
			{Sequence: 56, Hash: "ledger-56", TransactionCount: 1, OperationCount: 2},
			{Sequence: 55, Hash: "ledger-55", TransactionCount: 2, OperationCount: 4},
		},
	}, "open ledgers limit 2 before 57")
	if err != nil {
		t.Fatalf("executeOpenCommand(open ledgers) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerResults {
		t.Fatalf("expected generic explorer to open, got %#v", explorer)
	}
	if len(explorer.Results) != 3 || explorer.Results[0].Command != "lookup ledger 56" {
		t.Fatalf("expected ledger result command, got %#v", explorer.Results)
	}
	if got := explorer.Results[2].Command; got != "open ledgers limit 2 before 55" {
		t.Fatalf("expected next-page command, got %q", got)
	}
}

func TestExecuteOpenCommandOpensOperationExplorer(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", TypeName: "payment", Details: "{}"},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open ops")
	if err != nil {
		t.Fatalf("executeOpenCommand(open ops) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerOperations {
		t.Fatalf("expected operation explorer to open, got %#v", explorer)
	}
	if len(explorer.Operations) != 1 || explorer.Operations[0].TransactionHash != "tx-op-1" {
		t.Fatalf("expected operation explorer payload, got %#v", explorer.Operations)
	}
}

func TestExecuteOpenCommandOpensOperationDetailFromTransaction(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-deep-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-deep-1", LedgerSequence: 77, Account: "GSOURCE", Status: 1},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-deep-1", ApplicationOrder: 1, TypeName: "payment", Details: "{}"},
				{TransactionHash: "tx-deep-1", ApplicationOrder: 2, TypeName: "manage_data", Details: "{}"},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open op 2")
	if err != nil {
		t.Fatalf("executeOpenCommand(open op 2) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	lookup := model.Snapshot().Lookup
	if lookup.Kind != app.LookupOperation || lookup.Operation == nil {
		t.Fatalf("expected operation detail lookup, got %#v", lookup)
	}
	if got := lookup.Operation.Operation.TypeName; got != "manage_data" {
		t.Fatalf("expected second operation, got %q", got)
	}
	if got := lookup.Query; got != "tx-deep-1:2" {
		t.Fatalf("expected operation query tx-deep-1:2, got %q", got)
	}
}

func TestExecuteNavigationChainReturnsFromOperationDetailToTransaction(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupTransaction(
		"tx-deep-1",
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-deep-1", LedgerSequence: 77, Account: "GSOURCE", Status: 1},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-deep-1", ApplicationOrder: 1, TypeName: "payment", Details: "{}"},
			},
		},
	)
	model.SetLookupSelection(4, 0)

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open op 1"); err != nil {
		t.Fatalf("executeOpenCommand(open op 1) error = %v", err)
	}
	if got := model.Snapshot().Lookup.Kind; got != app.LookupOperation {
		t.Fatalf("expected operation detail, got %q", got)
	}

	if !model.Back() {
		t.Fatal("expected back to return to transaction detail")
	}

	lookup := model.Snapshot().Lookup
	if lookup.Kind != app.LookupTransaction || lookup.Operation != nil {
		t.Fatalf("expected transaction detail after back, got %#v", lookup)
	}
	if lookup.SelectedSection != 4 || lookup.ScrollOffset != 0 {
		t.Fatalf("expected restored transaction selection 4/0, got %d/%d", lookup.SelectedSection, lookup.ScrollOffset)
	}
}

func TestExecuteLookupOperationFetchesTransactionWhenNeeded(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	backend := stubLookupBackend{
		transaction: backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{Hash: "tx-remote-1", LedgerSequence: 10, Account: "GACC", Status: 1},
			Operations: []backendclient.OperationSummary{
				{TransactionHash: "tx-remote-1", ApplicationOrder: 1, TypeName: "payment", Details: "{}"},
			},
		},
	}

	keepRunning, err := executeLookupCommand(context.Background(), model, backend, "lookup op tx-remote-1:1", nil)
	if err != nil {
		t.Fatalf("executeLookupCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	lookup := model.Snapshot().Lookup
	if lookup.Kind != app.LookupOperation || lookup.Operation == nil {
		t.Fatalf("expected operation detail lookup, got %#v", lookup)
	}
	if got := lookup.Operation.ParentTransactionHash; got != "tx-remote-1" {
		t.Fatalf("expected parent tx tx-remote-1, got %q", got)
	}
	if got := lookup.Operation.Operation.TypeName; got != "payment" {
		t.Fatalf("expected payment operation, got %q", got)
	}
}

func TestExecuteOpenCommandOpensAccountTimeline(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	now := time.Unix(1715000000, 0).UTC()
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-1", LedgerSequence: 10, CreatedAt: now.Add(-time.Minute)},
			},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", TypeName: "payment", Details: "{}", CreatedAt: now},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open timeline")
	if err != nil {
		t.Fatalf("executeOpenCommand(open timeline) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerTimeline {
		t.Fatalf("expected timeline explorer to open, got %#v", explorer)
	}
	if len(explorer.Results) != 2 || explorer.Results[0].Kind != "op" {
		t.Fatalf("expected newest operation first, got %#v", explorer.Results)
	}
}

func TestExecuteOpenCommandFiltersAccountTimelineByCategory(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	now := time.Unix(1715000000, 0).UTC()
	model.UpdateLookupAccount(
		"GACC",
		backendclient.AccountLookupResponse{
			Account: &backendclient.AccountDetail{ID: "GACC", Balance: "10"},
			RecentTransactions: []backendclient.TransactionSummary{
				{Hash: "tx-1", LedgerSequence: 10, CreatedAt: now.Add(-time.Minute)},
			},
			RecentOperations: []backendclient.OperationSummary{
				{TransactionHash: "tx-op-1", TypeName: "payment", Details: "{}", CreatedAt: now},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open timeline type tx")
	if err != nil {
		t.Fatalf("executeOpenCommand(open timeline type tx) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerTimeline {
		t.Fatalf("expected timeline explorer to open, got %#v", explorer)
	}
	if explorer.Title != "Account Timeline: Transactions" {
		t.Fatalf("expected filtered timeline title, got %q", explorer.Title)
	}
	if len(explorer.Results) != 1 || explorer.Results[0].Kind != "tx" {
		t.Fatalf("expected only transaction timeline rows, got %#v", explorer.Results)
	}
}

func TestExecuteOpenCommandUsesBackendAssetTimeline(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupAsset(
		"USDC:GISS",
		backendclient.AssetLookupResponse{
			Asset: &backendclient.AssetDetail{AssetCode: "USDC", AssetIssuer: "GISS"},
			TopHolders: []backendclient.AssetHolderSummary{
				{AccountID: "GHOLDER", Balance: "10"},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubExplorerBackend{
		assetTimeline: []backendclient.TimelineItem{
			{Kind: "tx", Title: "Backend Asset Timeline", Command: "lookup tx tx-backend"},
		},
	}, "open timeline")
	if err != nil {
		t.Fatalf("executeOpenCommand(open timeline) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerTimeline {
		t.Fatalf("expected timeline explorer to open, got %#v", explorer)
	}
	if len(explorer.Results) != 1 || explorer.Results[0].Title != "Backend Asset Timeline" {
		t.Fatalf("expected backend asset timeline, got %#v", explorer.Results)
	}
}

func TestExecuteOpenCommandUsesBackendContractTimeline(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT"},
			RecentEvents: []backendclient.ContractEventSummary{
				{TransactionHash: "tx-event", LedgerSequence: 10},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubExplorerBackend{
		contractTimeline: []backendclient.TimelineItem{
			{Kind: "event", Title: "Backend Contract Timeline", Command: "lookup tx tx-backend"},
		},
	}, "open timeline")
	if err != nil {
		t.Fatalf("executeOpenCommand(open timeline) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerTimeline {
		t.Fatalf("expected timeline explorer to open, got %#v", explorer)
	}
	if len(explorer.Results) != 1 || explorer.Results[0].Title != "Backend Contract Timeline" {
		t.Fatalf("expected backend contract timeline, got %#v", explorer.Results)
	}
}

func TestExecuteOpenCommandSetsContractDecodeMode(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT"},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open decode raw")
	if err != nil {
		t.Fatalf("executeOpenCommand(open decode raw) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	if got := model.Snapshot().Lookup.DecodeMode; got != app.ContractDecodeModeRaw {
		t.Fatalf("expected raw decode mode, got %q", got)
	}
}

func TestExecuteOpenCommandOpensContractStorageExplorer(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT", StorageEntryCount: 2},
			Storage: []backendclient.ContractStorageSummary{
				{ContractID: "CCONTRACT", DisplayKey: "balance", DisplayValue: "100", DurabilityLabel: "persistent"},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubExplorerBackend{
		stubLookupBackend: stubLookupBackend{},
		contractStorage: []backendclient.ContractStorageSummary{
			{ContractID: "CCONTRACT", DisplayKey: "balance", DisplayValue: "100", DurabilityLabel: "persistent"},
			{ContractID: "CCONTRACT", DisplayKey: "owner", DisplayValue: "GOWNER", DurabilityLabel: "persistent"},
		},
	}, "open storage")
	if err != nil {
		t.Fatalf("executeOpenCommand(open storage) error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}

	explorer := model.Snapshot().Lookup.Explorer
	if explorer == nil || explorer.Kind != app.LookupExplorerStorage {
		t.Fatalf("expected storage explorer to open, got %#v", explorer)
	}
	if len(explorer.Storage) != 2 || explorer.Storage[0].DisplayKey != "balance" {
		t.Fatalf("expected storage explorer payload, got %#v", explorer.Storage)
	}
}

func TestExecuteOpenCommandOpensEventAndStorageDetail(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	topic := "transfer"
	value := "100"
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT", EventCount: 1, StorageEntryCount: 1},
			RecentEvents: []backendclient.ContractEventSummary{
				{ContractID: "CCONTRACT", TransactionHash: "tx-event-1", LedgerSequence: 88, Type: 1, Topic1: &topic, ValueDecoded: &value},
			},
			Storage: []backendclient.ContractStorageSummary{
				{ContractID: "CCONTRACT", DisplayKey: "balance", DisplayValue: "100", DurabilityLabel: "persistent", KeyXDR: "key", ValueXDR: "value"},
			},
		},
	)

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open event 1"); err != nil {
		t.Fatalf("executeOpenCommand(open event 1) error = %v", err)
	}
	if got := model.Snapshot().Lookup.Kind; got != app.LookupEvent {
		t.Fatalf("expected event detail, got %q", got)
	}

	if !model.Back() {
		t.Fatal("expected back to contract detail")
	}

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open storage 1"); err != nil {
		t.Fatalf("executeOpenCommand(open storage 1) error = %v", err)
	}
	if got := model.Snapshot().Lookup.Kind; got != app.LookupStorage {
		t.Fatalf("expected storage detail, got %q", got)
	}
}

func TestExecuteOpenCommandOpensInvocationDetail(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupContract(
		"CCONTRACT",
		backendclient.ContractLookupResponse{
			Contract: &backendclient.ContractDetail{ContractID: "CCONTRACT", InvocationCount: 1},
		},
	)
	transfer := "transfer"
	model.OpenLookupInvocationExplorer("Contract Invocations", "open detail", []backendclient.OperationSummary{
		{
			TransactionHash:  "tx-invoke-1",
			ApplicationOrder: 1,
			TypeName:         "invoke_host_function",
			FunctionName:     &transfer,
			Details:          `{"function_name":"transfer","auth_count":1,"authorizations":["auth 1: source_account"]}`,
		},
	}, 10, 0)

	if _, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open invocation 1"); err != nil {
		t.Fatalf("executeOpenCommand(open invocation 1) error = %v", err)
	}
	lookup := model.Snapshot().Lookup
	if lookup.Kind != app.LookupOperation || lookup.Operation == nil {
		t.Fatalf("expected invocation operation detail, got %#v", lookup)
	}
	if got := lookup.Operation.Operation.TypeName; got != "invoke_host_function" {
		t.Fatalf("expected invoke_host_function, got %q", got)
	}
}

func TestExecuteOpenCommandRejectsOutOfRangeLedgerLookupTransaction(t *testing.T) {
	model := app.NewModel(config.Default(), "/tmp/config.json", app.CacheSnapshot{})
	model.UpdateLookupLedger(
		"55",
		backendclient.LedgerLookupResponse{
			Ledger: &backendclient.LedgerSummary{Sequence: 55, Hash: "ledger-55"},
			Transactions: []backendclient.TransactionSummary{
				{Hash: "tx-ledger-1", LedgerSequence: 55},
			},
		},
	)

	keepRunning, err := executeOpenCommand(context.Background(), model, stubLookupBackend{}, "open tx 2")
	if err != nil {
		t.Fatalf("executeOpenCommand() error = %v", err)
	}
	if !keepRunning {
		t.Fatal("expected app to keep running")
	}
	if got := model.Snapshot().Lookup.State; got != app.ViewStateError {
		t.Fatalf("expected lookup error state, got %q", got)
	}
	if got := model.Snapshot().Lookup.Error; !strings.Contains(got, "current ledger lookup only has 1 transaction") {
		t.Fatalf("expected ledger range error, got %q", got)
	}
}

func TestOpenLookupBackendUsesNetworkBackendInRPCMode(t *testing.T) {
	profile := config.Profile{
		Name:        "rpc-only",
		Network:     "testnet",
		RPCEndpoint: "https://soroban-testnet.stellar.org",
		BackendMode: "rpc",
	}

	backend, err := openLookupBackend(profile)
	if err != nil {
		t.Fatalf("openLookupBackend() error = %v", err)
	}

	if _, ok := backend.(*networkbackend.Backend); !ok {
		t.Fatalf("expected network backend, got %T", backend)
	}
}

func TestOpenLookupBackendUsesHybridBackendInHybridMode(t *testing.T) {
	profile := config.Profile{
		Name:        "hybrid",
		Network:     "testnet",
		RPCEndpoint: "https://soroban-testnet.stellar.org",
		IndexerURL:  "http://127.0.0.1:8081",
		BackendMode: "hybrid",
	}

	backend, err := openLookupBackend(profile)
	if err != nil {
		t.Fatalf("openLookupBackend() error = %v", err)
	}

	if _, ok := backend.(*hybridLookupBackend); !ok {
		t.Fatalf("expected hybrid backend in hybrid mode, got %T", backend)
	}
}

func testConfigWithCache(driverName string) config.Config {
	cfg := config.Default()
	cfg.Cache.Driver = driverName
	cfg.Cache.Path = "shared"
	return cfg
}

type fakeLookupBackend struct {
	liveFeedSummary backendclient.LiveFeedSummaryResponse
	ledger          backendclient.LedgerLookupResponse
	transaction     backendclient.TransactionLookupResponse
	account         backendclient.AccountLookupResponse
	asset           backendclient.AssetLookupResponse
	contract        backendclient.ContractLookupResponse
	err             error
}

func (b fakeLookupBackend) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return b.liveFeedSummary, b.err
}

func (b fakeLookupBackend) Label() string {
	return "fake"
}

func (b fakeLookupBackend) Search(context.Context, string, int) (backendclient.SearchResponse, error) {
	return backendclient.SearchResponse{}, b.err
}

func (b fakeLookupBackend) Ledger(context.Context, uint32) (backendclient.LedgerLookupResponse, error) {
	return b.ledger, b.err
}

func (b fakeLookupBackend) Transaction(context.Context, string) (backendclient.TransactionLookupResponse, error) {
	return b.transaction, b.err
}

func (b fakeLookupBackend) Account(context.Context, string) (backendclient.AccountLookupResponse, error) {
	return b.account, b.err
}

func (b fakeLookupBackend) Asset(context.Context, string, string) (backendclient.AssetLookupResponse, error) {
	return b.asset, b.err
}

func (b fakeLookupBackend) Contract(context.Context, string) (backendclient.ContractLookupResponse, error) {
	return b.contract, b.err
}

type stubLookupBackend struct {
	liveFeed    backendclient.LiveFeedSummaryResponse
	ledger      backendclient.LedgerLookupResponse
	transaction backendclient.TransactionLookupResponse
	account     backendclient.AccountLookupResponse
	asset       backendclient.AssetLookupResponse
	contract    backendclient.ContractLookupResponse
	err         error
}

func (s stubLookupBackend) LiveFeedSummary(context.Context) (backendclient.LiveFeedSummaryResponse, error) {
	return s.liveFeed, s.err
}

func (s stubLookupBackend) Label() string {
	return "stub"
}

func (s stubLookupBackend) Search(context.Context, string, int) (backendclient.SearchResponse, error) {
	return backendclient.SearchResponse{}, s.err
}

func (s stubLookupBackend) Ledger(context.Context, uint32) (backendclient.LedgerLookupResponse, error) {
	return s.ledger, s.err
}

func (s stubLookupBackend) Transaction(context.Context, string) (backendclient.TransactionLookupResponse, error) {
	return s.transaction, s.err
}

func (s stubLookupBackend) Account(context.Context, string) (backendclient.AccountLookupResponse, error) {
	return s.account, s.err
}

func (s stubLookupBackend) Asset(context.Context, string, string) (backendclient.AssetLookupResponse, error) {
	return s.asset, s.err
}

func (s stubLookupBackend) Contract(context.Context, string) (backendclient.ContractLookupResponse, error) {
	return s.contract, s.err
}

type stubExplorerBackend struct {
	stubLookupBackend
	ledgers             []backendclient.LedgerSummary
	ledgerTxs           []backendclient.TransactionSummary
	accounts            []backendclient.AccountDetail
	accountOps          []backendclient.OperationSummary
	accountTimeline     []backendclient.TimelineItem
	assets              []backendclient.AssetDetail
	assetHolders        []backendclient.AssetHolderSummary
	assetTimeline       []backendclient.TimelineItem
	contracts           []backendclient.ContractDetail
	contractEvents      []backendclient.ContractEventSummary
	contractStorage     []backendclient.ContractStorageSummary
	contractInvocations []backendclient.OperationSummary
	contractTimeline    []backendclient.TimelineItem
}

func (s stubExplorerBackend) Ledgers(context.Context, int, uint32) ([]backendclient.LedgerSummary, error) {
	return s.ledgers, s.err
}

func (s stubExplorerBackend) LedgerTransactions(context.Context, uint32, int, int) ([]backendclient.TransactionSummary, error) {
	return s.ledgerTxs, s.err
}

func (s stubExplorerBackend) Accounts(context.Context, int) ([]backendclient.AccountDetail, error) {
	return s.accounts, s.err
}

func (s stubExplorerBackend) AccountOperations(context.Context, string, int, int) ([]backendclient.OperationSummary, error) {
	return s.accountOps, s.err
}

func (s stubExplorerBackend) AccountTimeline(context.Context, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return s.accountTimeline, s.err
}

func (s stubExplorerBackend) Assets(context.Context, int) ([]backendclient.AssetDetail, error) {
	return s.assets, s.err
}

func (s stubExplorerBackend) AssetHolders(context.Context, string, string, int, int) ([]backendclient.AssetHolderSummary, error) {
	return s.assetHolders, s.err
}

func (s stubExplorerBackend) AssetTimeline(context.Context, string, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return s.assetTimeline, s.err
}

func (s stubExplorerBackend) Contracts(context.Context, int) ([]backendclient.ContractDetail, error) {
	return s.contracts, s.err
}

func (s stubExplorerBackend) ContractEvents(context.Context, string, int, int) ([]backendclient.ContractEventSummary, error) {
	return s.contractEvents, s.err
}

func (s stubExplorerBackend) ContractStorage(context.Context, string, int, int) ([]backendclient.ContractStorageSummary, error) {
	return s.contractStorage, s.err
}

func (s stubExplorerBackend) ContractInvocations(context.Context, string, int, int) ([]backendclient.OperationSummary, error) {
	return s.contractInvocations, s.err
}

func (s stubExplorerBackend) ContractTimeline(context.Context, string, int, int, string) ([]backendclient.TimelineItem, error) {
	return s.contractTimeline, s.err
}

func registerFakeSQLiteDriver(t *testing.T) string {
	t.Helper()

	name := fmt.Sprintf("fake-sqlite-main-%d", time.Now().UnixNano())
	sql.Register(name, fakeSQLiteDriver{
		state: &fakeSQLiteState{
			state:         make(map[string]string),
			profiles:      make(map[string]cache.Profile),
			bookmarks:     make(map[string]cache.Bookmark),
			labels:        make(map[string]cache.Label),
			labelTargets:  make(map[string]cache.LabelTarget),
			notes:         make(map[string]cache.Note),
			liveTxs:       make(map[string]cache.LiveTransaction),
			entities:      make(map[string]cache.EntityCache),
			savedViews:    make(map[string]cache.SavedView),
			watchSettings: make(map[string]cache.WatchSetting),
		},
	})
	return name
}

type fakeSQLiteDriver struct {
	state *fakeSQLiteState
}

type fakeSQLiteConn struct {
	state *fakeSQLiteState
}

type fakeSQLiteState struct {
	mu            sync.Mutex
	versions      []int
	state         map[string]string
	profiles      map[string]cache.Profile
	bookmarks     map[string]cache.Bookmark
	labels        map[string]cache.Label
	labelTargets  map[string]cache.LabelTarget
	notes         map[string]cache.Note
	liveTxs       map[string]cache.LiveTransaction
	entities      map[string]cache.EntityCache
	savedViews    map[string]cache.SavedView
	watchSettings map[string]cache.WatchSetting
}

type fakeSQLiteTx struct{}

type fakeSQLiteRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (d fakeSQLiteDriver) Open(string) (driver.Conn, error) {
	return &fakeSQLiteConn{state: d.state}, nil
}

func (c *fakeSQLiteConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *fakeSQLiteConn) Close() error                        { return nil }
func (c *fakeSQLiteConn) Begin() (driver.Tx, error)           { return fakeSQLiteTx{}, nil }
func (c *fakeSQLiteConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return fakeSQLiteTx{}, nil
}

func (c *fakeSQLiteConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()

	switch {
	case strings.Contains(query, "INSERT INTO schema_migrations"):
		if len(args) > 0 {
			c.state.versions = append(c.state.versions, int(asInt64(args[0].Value)))
		}
	case strings.Contains(query, "INSERT INTO app_state"):
		if len(args) >= 2 {
			c.state.state[asString(args[0].Value)] = asString(args[1].Value)
		}
	case strings.Contains(query, "INSERT INTO profiles"):
		if len(args) >= 7 {
			profile := cache.Profile{
				ID:         asString(args[0].Value),
				Name:       asString(args[1].Value),
				Network:    asString(args[2].Value),
				RPCURL:     asString(args[3].Value),
				IndexerURL: asString(args[4].Value),
				CreatedAt:  asTime(args[5].Value),
				UpdatedAt:  asTime(args[6].Value),
			}
			c.state.profiles[profile.ID] = profile
		}
	case strings.Contains(query, "INSERT INTO bookmarks"):
		if len(args) >= 8 {
			bookmark := cache.Bookmark{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Kind:      asString(args[2].Value),
				Target:    asString(args[3].Value),
				Title:     asString(args[4].Value),
				Notes:     asString(args[5].Value),
				CreatedAt: asTime(args[6].Value),
				UpdatedAt: asTime(args[7].Value),
			}
			c.state.bookmarks[bookmark.ID] = bookmark
		}
	case strings.Contains(query, "INSERT INTO labels"):
		if len(args) >= 6 {
			label := cache.Label{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Name:      asString(args[2].Value),
				Color:     asString(args[3].Value),
				CreatedAt: asTime(args[4].Value),
				UpdatedAt: asTime(args[5].Value),
			}
			c.state.labels[label.ID] = label
		}
	case strings.Contains(query, "INSERT INTO label_targets"):
		if len(args) >= 7 {
			target := cache.LabelTarget{
				ID:        asString(args[0].Value),
				LabelID:   asString(args[1].Value),
				ProfileID: asString(args[2].Value),
				Kind:      asString(args[3].Value),
				Target:    asString(args[4].Value),
				CreatedAt: asTime(args[5].Value),
				UpdatedAt: asTime(args[6].Value),
			}
			c.state.labelTargets[target.ID] = target
		}
	case strings.Contains(query, "INSERT INTO notes"):
		if len(args) >= 7 {
			note := cache.Note{
				ID:        asString(args[0].Value),
				ProfileID: asString(args[1].Value),
				Target:    asString(args[2].Value),
				Title:     asString(args[3].Value),
				Body:      asString(args[4].Value),
				CreatedAt: asTime(args[5].Value),
				UpdatedAt: asTime(args[6].Value),
			}
			c.state.notes[note.ID] = note
		}
	case strings.Contains(query, "INSERT INTO live_transactions"):
		if len(args) >= 10 {
			tx := cache.LiveTransaction{
				ProfileID:        asString(args[0].Value),
				Hash:             asString(args[1].Value),
				LedgerSequence:   uint32(asInt64(args[2].Value)),
				ApplicationOrder: int32(asInt64(args[3].Value)),
				Account:          asString(args[4].Value),
				OperationCount:   int32(asInt64(args[5].Value)),
				Status:           int16(asInt64(args[6].Value)),
				IsSoroban:        asInt64(args[7].Value) != 0,
				CreatedAt:        asTime(args[8].Value),
				CachedAt:         asTime(args[9].Value),
			}
			c.state.liveTxs[tx.ProfileID+":"+tx.Hash] = tx
		}
	case strings.Contains(query, "INSERT INTO watch_settings"):
		if len(args) >= 8 {
			setting := cache.WatchSetting{
				ID:          asString(args[0].Value),
				ProfileID:   asString(args[1].Value),
				Name:        asString(args[2].Value),
				FiltersJSON: asString(args[3].Value),
				Paused:      asInt64(args[4].Value) != 0,
				AutoApply:   asInt64(args[5].Value) != 0,
				CreatedAt:   asTime(args[6].Value),
				UpdatedAt:   asTime(args[7].Value),
			}
			c.state.watchSettings[setting.ProfileID+":"+setting.Name] = setting
		}
	case strings.Contains(query, "INSERT INTO saved_views"):
		if len(args) >= 10 {
			view := cache.SavedView{
				ID:           asString(args[0].Value),
				ProfileID:    asString(args[1].Value),
				Name:         asString(args[2].Value),
				Command:      asString(args[3].Value),
				Screen:       asString(args[4].Value),
				EntityKind:   asString(args[5].Value),
				EntityTarget: asString(args[6].Value),
				FiltersJSON:  asString(args[7].Value),
				CreatedAt:    asTime(args[8].Value),
				UpdatedAt:    asTime(args[9].Value),
			}
			c.state.savedViews[view.ProfileID+":"+view.Name] = view
		}
	case strings.Contains(query, "INSERT INTO entity_cache"):
		if len(args) >= 9 {
			entity := cache.EntityCache{
				ProfileID:   asString(args[0].Value),
				Kind:        asString(args[1].Value),
				Target:      asString(args[2].Value),
				Title:       asString(args[3].Value),
				Summary:     asString(args[4].Value),
				Payload:     asString(args[5].Value),
				SourceLabel: asString(args[6].Value),
				CreatedAt:   asTime(args[7].Value),
				UpdatedAt:   asTime(args[8].Value),
			}
			c.state.entities[entity.ProfileID+":"+entity.Kind+":"+entity.Target] = entity
		}
	case strings.Contains(query, "DELETE FROM bookmarks"):
		if len(args) >= 1 {
			delete(c.state.bookmarks, asString(args[0].Value))
		}
	case strings.Contains(query, "DELETE FROM notes"):
		if len(args) >= 1 {
			delete(c.state.notes, asString(args[0].Value))
		}
	case strings.Contains(query, "DELETE FROM label_targets"):
		if len(args) >= 1 {
			delete(c.state.labelTargets, asString(args[0].Value))
		}
	case strings.Contains(query, "DELETE FROM labels"):
		if len(args) >= 1 {
			delete(c.state.labels, asString(args[0].Value))
		}
	case strings.Contains(query, "DELETE FROM saved_views"):
		if len(args) >= 2 {
			delete(c.state.savedViews, asString(args[0].Value)+":"+asString(args[1].Value))
		}
	case strings.Contains(query, "DELETE FROM watch_settings"):
		if len(args) >= 2 {
			delete(c.state.watchSettings, asString(args[0].Value)+":"+asString(args[1].Value))
		}
	}

	return driver.RowsAffected(1), nil
}

func (c *fakeSQLiteConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()

	switch {
	case strings.Contains(query, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"):
		version := 0
		for _, current := range c.state.versions {
			if current > version {
				version = current
			}
		}
		return &fakeSQLiteRows{columns: []string{"version"}, rows: [][]driver.Value{{int64(version)}}}, nil
	case strings.Contains(query, "SELECT value FROM app_state WHERE key = ?"):
		key := asString(args[0].Value)
		value, ok := c.state.state[key]
		if !ok {
			return &fakeSQLiteRows{columns: []string{"value"}}, nil
		}
		return &fakeSQLiteRows{columns: []string{"value"}, rows: [][]driver.Value{{value}}}, nil
	case strings.Contains(query, "FROM profiles"):
		profiles := make([]cache.Profile, 0, len(c.state.profiles))
		for _, profile := range c.state.profiles {
			profiles = append(profiles, profile)
		}
		sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
		rows := make([][]driver.Value, 0, len(profiles))
		for _, profile := range profiles {
			rows = append(rows, []driver.Value{
				profile.ID,
				profile.Name,
				profile.Network,
				profile.RPCURL,
				profile.IndexerURL,
				profile.CreatedAt,
				profile.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "name", "network", "rpc_url", "indexer_url", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM bookmarks"):
		bookmarks := make([]cache.Bookmark, 0, len(c.state.bookmarks))
		for _, bookmark := range c.state.bookmarks {
			bookmarks = append(bookmarks, bookmark)
		}
		sort.Slice(bookmarks, func(i, j int) bool {
			if bookmarks[i].Title == bookmarks[j].Title {
				return bookmarks[i].CreatedAt.Before(bookmarks[j].CreatedAt)
			}
			return bookmarks[i].Title < bookmarks[j].Title
		})
		rows := make([][]driver.Value, 0, len(bookmarks))
		for _, bookmark := range bookmarks {
			rows = append(rows, []driver.Value{
				bookmark.ID,
				bookmark.ProfileID,
				bookmark.Kind,
				bookmark.Target,
				bookmark.Title,
				bookmark.Notes,
				bookmark.CreatedAt,
				bookmark.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "kind", "target", "title", "notes", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM labels"):
		labels := make([]cache.Label, 0, len(c.state.labels))
		for _, label := range c.state.labels {
			labels = append(labels, label)
		}
		sort.Slice(labels, func(i, j int) bool { return labels[i].Name < labels[j].Name })
		rows := make([][]driver.Value, 0, len(labels))
		for _, label := range labels {
			rows = append(rows, []driver.Value{
				label.ID,
				label.ProfileID,
				label.Name,
				label.Color,
				label.CreatedAt,
				label.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "name", "color", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM label_targets"):
		targets := make([]cache.LabelTarget, 0, len(c.state.labelTargets))
		for _, target := range c.state.labelTargets {
			targets = append(targets, target)
		}
		sort.Slice(targets, func(i, j int) bool {
			if targets[i].LabelID == targets[j].LabelID {
				if targets[i].Kind == targets[j].Kind {
					return targets[i].Target < targets[j].Target
				}
				return targets[i].Kind < targets[j].Kind
			}
			return targets[i].LabelID < targets[j].LabelID
		})
		rows := make([][]driver.Value, 0, len(targets))
		for _, target := range targets {
			rows = append(rows, []driver.Value{
				target.ID,
				target.LabelID,
				target.ProfileID,
				target.Kind,
				target.Target,
				target.CreatedAt,
				target.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "label_id", "profile_id", "kind", "target", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM notes"):
		notes := make([]cache.Note, 0, len(c.state.notes))
		for _, note := range c.state.notes {
			notes = append(notes, note)
		}
		sort.Slice(notes, func(i, j int) bool { return notes[i].UpdatedAt.After(notes[j].UpdatedAt) })
		rows := make([][]driver.Value, 0, len(notes))
		for _, note := range notes {
			rows = append(rows, []driver.Value{
				note.ID,
				note.ProfileID,
				note.Target,
				note.Title,
				note.Body,
				note.CreatedAt,
				note.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "target", "title", "body", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM live_transactions"):
		profileID := asString(args[0].Value)
		limit := int(asInt64(args[1].Value))
		transactions := make([]cache.LiveTransaction, 0, len(c.state.liveTxs))
		for _, tx := range c.state.liveTxs {
			if tx.ProfileID == profileID {
				transactions = append(transactions, tx)
			}
		}
		sort.Slice(transactions, func(i, j int) bool {
			if transactions[i].LedgerSequence == transactions[j].LedgerSequence {
				return transactions[i].ApplicationOrder > transactions[j].ApplicationOrder
			}
			return transactions[i].LedgerSequence > transactions[j].LedgerSequence
		})
		if limit > 0 && len(transactions) > limit {
			transactions = transactions[:limit]
		}
		rows := make([][]driver.Value, 0, len(transactions))
		for _, tx := range transactions {
			rows = append(rows, []driver.Value{
				tx.ProfileID,
				tx.Hash,
				int64(tx.LedgerSequence),
				int64(tx.ApplicationOrder),
				tx.Account,
				int64(tx.OperationCount),
				int64(tx.Status),
				boolToInt(tx.IsSoroban),
				tx.CreatedAt,
				tx.CachedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"profile_id", "hash", "ledger_sequence", "application_order", "account", "operation_count", "status", "is_soroban", "created_at", "cached_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM watch_settings"):
		if strings.Contains(query, "auto_apply = 1") {
			profileID := asString(args[0].Value)
			var latest *cache.WatchSetting
			for _, setting := range c.state.watchSettings {
				if setting.ProfileID == profileID && setting.AutoApply {
					if latest == nil || setting.UpdatedAt.After(latest.UpdatedAt) {
						copy := setting
						latest = &copy
					}
				}
			}
			if latest == nil {
				return &fakeSQLiteRows{
					columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
				}, nil
			}
			paused := 0
			if latest.Paused {
				paused = 1
			}
			autoApply := 0
			if latest.AutoApply {
				autoApply = 1
			}
			return &fakeSQLiteRows{
				columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
				rows: [][]driver.Value{{
					latest.ID,
					latest.ProfileID,
					latest.Name,
					latest.FiltersJSON,
					int64(paused),
					int64(autoApply),
					latest.CreatedAt,
					latest.UpdatedAt,
				}},
			}, nil
		}
		profileID := asString(args[0].Value)
		settings := make([]cache.WatchSetting, 0, len(c.state.watchSettings))
		for _, setting := range c.state.watchSettings {
			if setting.ProfileID == profileID {
				settings = append(settings, setting)
			}
		}
		sort.Slice(settings, func(i, j int) bool { return settings[i].UpdatedAt.After(settings[j].UpdatedAt) })
		rows := make([][]driver.Value, 0, len(settings))
		for _, setting := range settings {
			paused := 0
			if setting.Paused {
				paused = 1
			}
			autoApply := 0
			if setting.AutoApply {
				autoApply = 1
			}
			rows = append(rows, []driver.Value{
				setting.ID,
				setting.ProfileID,
				setting.Name,
				setting.FiltersJSON,
				int64(paused),
				int64(autoApply),
				setting.CreatedAt,
				setting.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "name", "filters_json", "paused", "auto_apply", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM saved_views"):
		profileID := asString(args[0].Value)
		views := make([]cache.SavedView, 0, len(c.state.savedViews))
		for _, view := range c.state.savedViews {
			if view.ProfileID == profileID {
				views = append(views, view)
			}
		}
		sort.Slice(views, func(i, j int) bool { return views[i].UpdatedAt.After(views[j].UpdatedAt) })
		rows := make([][]driver.Value, 0, len(views))
		for _, view := range views {
			rows = append(rows, []driver.Value{
				view.ID,
				view.ProfileID,
				view.Name,
				view.Command,
				view.Screen,
				view.EntityKind,
				view.EntityTarget,
				view.FiltersJSON,
				view.CreatedAt,
				view.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"id", "profile_id", "name", "command", "screen", "entity_kind", "entity_target", "filters_json", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	case strings.Contains(query, "FROM entity_cache"):
		if len(args) == 3 {
			profileID := asString(args[0].Value)
			kind := asString(args[1].Value)
			target := asString(args[2].Value)
			entity, ok := c.state.entities[profileID+":"+kind+":"+target]
			if !ok {
				return &fakeSQLiteRows{
					columns: []string{"profile_id", "kind", "target", "title", "summary", "payload", "source_label", "created_at", "updated_at"},
				}, nil
			}
			return &fakeSQLiteRows{
				columns: []string{"profile_id", "kind", "target", "title", "summary", "payload", "source_label", "created_at", "updated_at"},
				rows: [][]driver.Value{{
					entity.ProfileID,
					entity.Kind,
					entity.Target,
					entity.Title,
					entity.Summary,
					entity.Payload,
					entity.SourceLabel,
					entity.CreatedAt,
					entity.UpdatedAt,
				}},
			}, nil
		}
		profileID := asString(args[0].Value)
		limit := int(asInt64(args[1].Value))
		entities := make([]cache.EntityCache, 0, len(c.state.entities))
		for _, entity := range c.state.entities {
			if entity.ProfileID == profileID {
				entities = append(entities, entity)
			}
		}
		sort.Slice(entities, func(i, j int) bool {
			return entities[i].UpdatedAt.After(entities[j].UpdatedAt)
		})
		if limit > 0 && len(entities) > limit {
			entities = entities[:limit]
		}
		rows := make([][]driver.Value, 0, len(entities))
		for _, entity := range entities {
			rows = append(rows, []driver.Value{
				entity.ProfileID,
				entity.Kind,
				entity.Target,
				entity.Title,
				entity.Summary,
				entity.Payload,
				entity.SourceLabel,
				entity.CreatedAt,
				entity.UpdatedAt,
			})
		}
		return &fakeSQLiteRows{
			columns: []string{"profile_id", "kind", "target", "title", "summary", "payload", "source_label", "created_at", "updated_at"},
			rows:    rows,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported query: %s", query)
	}
}

func (fakeSQLiteTx) Commit() error   { return nil }
func (fakeSQLiteTx) Rollback() error { return nil }

func (r *fakeSQLiteRows) Columns() []string { return r.columns }
func (r *fakeSQLiteRows) Close() error      { return nil }

func (r *fakeSQLiteRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func asString(value any) string {
	if current, ok := value.(string); ok {
		return current
	}
	return ""
}

func asInt64(value any) int64 {
	if current, ok := value.(int64); ok {
		return current
	}
	return 0
}

func asTime(value any) time.Time {
	if current, ok := value.(time.Time); ok {
		return current
	}
	return time.Time{}
}

func boolToInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
