package app

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestDefaultSourceMetadataRPCMode(t *testing.T) {
	profile := config.Profile{
		Name:        "rpc",
		Network:     "testnet",
		RPCEndpoint: "https://rpc.test",
		BackendMode: config.BackendModeRPC,
	}
	meta := DefaultSourceMetadata(profile, "lookup")
	if meta.Mode != config.BackendModeRPC || meta.Preferred != "rpc" || meta.Actual != "rpc" {
		t.Fatalf("unexpected rpc metadata: %#v", meta)
	}
	if meta.FallbackUsed || meta.Degraded {
		t.Fatalf("rpc mode should not be degraded by default: %#v", meta)
	}
}

func TestDefaultSourceMetadataHybridMode(t *testing.T) {
	profile := config.Profile{
		Name:        "hybrid",
		Network:     "testnet",
		RPCEndpoint: "https://rpc.test",
		IndexerURL:  "http://indexer.test",
		BackendMode: config.BackendModeHybrid,
	}
	meta := DefaultSourceMetadata(profile, "lookup")
	if meta.Preferred != "indexer" || meta.Actual != "indexer" {
		t.Fatalf("expected indexer preferred metadata, got %#v", meta)
	}
	if meta.Policy != "single-source: indexer -> rpc; no field merge" {
		t.Fatalf("unexpected hybrid policy: %q", meta.Policy)
	}
}

func TestDefaultSourceMetadataIndexerMode(t *testing.T) {
	profile := config.Profile{
		Name:        "indexer",
		Network:     "testnet",
		RPCEndpoint: "https://rpc.test",
		IndexerURL:  "http://indexer.test",
		BackendMode: config.BackendModeIndexer,
	}
	meta := DefaultSourceMetadata(profile, "search")
	if meta.Mode != config.BackendModeIndexer || meta.Label != "http://indexer.test" {
		t.Fatalf("unexpected indexer metadata: %#v", meta)
	}
}
