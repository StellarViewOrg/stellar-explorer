package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stellar/go-stellar-sdk/network"
)

const (
	SearchBackendPostgres  = "postgres"
	SearchBackendTypesense = "typesense"
)

const (
	DedicatedSearchEntityThreshold = 10_000_000
	DedicatedSearchP95LatencyMS    = 150
	DedicatedSearchQueriesPerMin   = 300
)

type Config struct {
	DatabaseURL   string
	RedisURL      string
	SearchBackend string
	TypesenseURL  string
	TypesenseKey  string
	RPCEndpoint   string
	ReadAPIAddr   string
	DataLakePath  string
	Network       string // "public", "testnet", "futurenet"
	StartLedger   uint32
	BatchSize     int
	WorkerCount   int
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://explorer:explorer_dev@localhost:54330/stellar_explorer_tui?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:63890"),
		SearchBackend: getEnv("SEARCH_BACKEND", SearchBackendPostgres),
		TypesenseURL:  getEnv("TYPESENSE_URL", "http://localhost:18118"),
		TypesenseKey:  getEnv("TYPESENSE_KEY", "explorer_dev_key"),
		RPCEndpoint:   getEnv("RPC_ENDPOINT", ""),
		ReadAPIAddr:   getEnv("READ_API_ADDR", ":8081"),
		DataLakePath:  getEnv("DATA_LAKE_PATH", "s3://aws-public-blockchain/v1.1/stellar/ledgers/pubnet"),
		Network:       getEnv("NETWORK", "public"),
		BatchSize:     getEnvInt("BATCH_SIZE", 100),
		WorkerCount:   getEnvInt("WORKER_COUNT", 8),
	}

	return cfg, nil
}

// SearchBackendMode returns the configured search engine policy.
//
// Policy:
//   - postgres is the default and remains the right choice while search is
//     mostly exact/prefix lookup over indexed entity tables.
//   - typesense becomes justified when the TUI needs fuzzy/ranked discovery,
//     saved-investigation search, or when observed Postgres search exceeds the
//     thresholds captured by ShouldUseDedicatedSearchIndex.
func (c *Config) SearchBackendMode() string {
	switch strings.ToLower(strings.TrimSpace(c.SearchBackend)) {
	case SearchBackendTypesense:
		return SearchBackendTypesense
	default:
		return SearchBackendPostgres
	}
}

// UsesDedicatedSearchIndex reports whether read/search paths should use a
// dedicated search service rather than direct Postgres-backed search.
func (c *Config) UsesDedicatedSearchIndex() bool {
	return c.SearchBackendMode() == SearchBackendTypesense
}

type SearchIndexMetrics struct {
	IndexedEntities                  int64
	P95LatencyMS                     int
	QueriesPerMinute                 int
	RequiresFuzzyRanking             bool
	RequiresSavedInvestigationSearch bool
}

// ShouldUseDedicatedSearchIndex captures the promotion rule from Postgres search
// to a dedicated index. Any one of these conditions is enough to justify the
// extra operational surface of Typesense or a future equivalent search service.
func ShouldUseDedicatedSearchIndex(metrics SearchIndexMetrics) bool {
	return metrics.RequiresFuzzyRanking ||
		metrics.RequiresSavedInvestigationSearch ||
		metrics.IndexedEntities >= DedicatedSearchEntityThreshold ||
		metrics.P95LatencyMS >= DedicatedSearchP95LatencyMS ||
		metrics.QueriesPerMinute >= DedicatedSearchQueriesPerMin
}

// NetworkPassphrase returns the Stellar network passphrase for the configured network.
func (c *Config) NetworkPassphrase() (string, error) {
	switch c.Network {
	case "public":
		return network.PublicNetworkPassphrase, nil
	case "testnet":
		return network.TestNetworkPassphrase, nil
	case "futurenet":
		return network.FutureNetworkPassphrase, nil
	default:
		return "", fmt.Errorf("unknown network: %s", c.Network)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}
