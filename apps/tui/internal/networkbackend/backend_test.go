package networkbackend

import (
	"testing"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/backendclient"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func TestMergeTransactionLookupPrefersHorizonAndFillsRPCMeta(t *testing.T) {
	meta := "AAAA"
	merged := mergeTransactionLookup(
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				Hash:    "abc",
				Account: "GABC",
			},
		},
		backendclient.TransactionLookupResponse{
			Transaction: &backendclient.TransactionDetail{
				EnvelopeXDR:   "envelope",
				ResultXDR:     "result",
				ResultMetaXDR: &meta,
				IsSoroban:     true,
			},
		},
	)

	if merged.Transaction == nil {
		t.Fatal("expected merged transaction")
	}
	if merged.Transaction.Account != "GABC" {
		t.Fatalf("account = %q", merged.Transaction.Account)
	}
	if merged.Transaction.EnvelopeXDR != "envelope" {
		t.Fatalf("envelope = %q", merged.Transaction.EnvelopeXDR)
	}
	if merged.Transaction.ResultMetaXDR == nil || *merged.Transaction.ResultMetaXDR != meta {
		t.Fatalf("result meta = %#v", merged.Transaction.ResultMetaXDR)
	}
	if !merged.Transaction.IsSoroban {
		t.Fatal("expected soroban flag")
	}
}

func TestCombinedLabelIncludesHorizonAndRPC(t *testing.T) {
	profile := config.Profile{
		Network:     "testnet",
		HorizonURL:  "https://horizon-testnet.stellar.org",
		RPCEndpoint: "https://rpc.test",
	}
	label := combinedLabel(profile, true)
	want := "https://horizon-testnet.stellar.org + https://rpc.test"
	if label != want {
		t.Fatalf("label = %q, want %q", label, want)
	}
}
