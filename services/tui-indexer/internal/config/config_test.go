package config

import (
	"os"
	"testing"
)

func TestLoad_RPCEndpointOptional(t *testing.T) {
	os.Unsetenv("RPC_ENDPOINT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RPCEndpoint != "" {
		t.Errorf("expected empty RPCEndpoint, got '%s'", cfg.RPCEndpoint)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("RPC_ENDPOINT", "https://rpc.example.com")
	defer os.Unsetenv("RPC_ENDPOINT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Network != "public" {
		t.Errorf("expected network 'public', got '%s'", cfg.Network)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("expected batch size 100, got %d", cfg.BatchSize)
	}
	if cfg.WorkerCount != 8 {
		t.Errorf("expected worker count 8, got %d", cfg.WorkerCount)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("RPC_ENDPOINT", "https://rpc.example.com")
	os.Setenv("NETWORK", "testnet")
	os.Setenv("BATCH_SIZE", "200")
	os.Setenv("SEARCH_BACKEND", "typesense")
	defer func() {
		os.Unsetenv("RPC_ENDPOINT")
		os.Unsetenv("NETWORK")
		os.Unsetenv("BATCH_SIZE")
		os.Unsetenv("SEARCH_BACKEND")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Network != "testnet" {
		t.Errorf("expected network 'testnet', got '%s'", cfg.Network)
	}
	if cfg.BatchSize != 200 {
		t.Errorf("expected batch size 200, got %d", cfg.BatchSize)
	}
	if cfg.SearchBackendMode() != SearchBackendTypesense {
		t.Errorf("expected search backend %q, got %q", SearchBackendTypesense, cfg.SearchBackendMode())
	}
	if !cfg.UsesDedicatedSearchIndex() {
		t.Fatal("expected typesense mode to use dedicated search index")
	}
}

func TestSearchBackendModeDefaultsToPostgres(t *testing.T) {
	cfg := Config{SearchBackend: "unknown"}
	if cfg.SearchBackendMode() != SearchBackendPostgres {
		t.Fatalf("expected unknown search backend to default to postgres, got %q", cfg.SearchBackendMode())
	}
}

func TestShouldUseDedicatedSearchIndexPolicy(t *testing.T) {
	if ShouldUseDedicatedSearchIndex(SearchIndexMetrics{IndexedEntities: 1000, P95LatencyMS: 40, QueriesPerMinute: 20}) {
		t.Fatal("expected small exact-search workload to stay on Postgres")
	}

	cases := []SearchIndexMetrics{
		{RequiresFuzzyRanking: true},
		{RequiresSavedInvestigationSearch: true},
		{IndexedEntities: DedicatedSearchEntityThreshold},
		{P95LatencyMS: DedicatedSearchP95LatencyMS},
		{QueriesPerMinute: DedicatedSearchQueriesPerMin},
	}
	for _, metrics := range cases {
		if !ShouldUseDedicatedSearchIndex(metrics) {
			t.Fatalf("expected dedicated index for metrics %#v", metrics)
		}
	}
}
