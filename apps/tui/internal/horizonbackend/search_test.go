package horizonbackend

import (
	"context"
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestSearchInfersKnownIdentifiers(t *testing.T) {
	backend, err := New(config.Profile{
		Network:     "testnet",
		HorizonURL:  "https://horizon-testnet.stellar.org",
		RPCEndpoint: "https://soroban-testnet.stellar.org",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	response, err := backend.Search(context.Background(), "12345", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(response.Results) != 1 || response.Results[0].Kind != "ledger" {
		t.Fatalf("expected ledger inference, got %#v", response.Results)
	}
	if response.Results[0].Source != "horizon" {
		t.Fatalf("expected horizon source, got %q", response.Results[0].Source)
	}
}
