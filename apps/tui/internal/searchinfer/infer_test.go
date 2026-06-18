package searchinfer

import "testing"

func TestFromQueryInfersAssetAndPartialHash(t *testing.T) {
	asset := FromQuery("USDC:GBBD47IF6LWK7P7MDEVSCWR7DPUWV3NY3DTQEVFL4NAT4AQH3ZLLFLA5")
	if len(asset) != 1 || asset[0].Kind != "asset" {
		t.Fatalf("expected asset inference, got %#v", asset)
	}

	partial := FromQuery("abc12345")
	if len(partial) != 1 || partial[0].Kind != "transaction" {
		t.Fatalf("expected partial tx inference, got %#v", partial)
	}
}

func TestFromQueryIgnoresInvalidInput(t *testing.T) {
	if results := FromQuery("not-a-target"); len(results) != 0 {
		t.Fatalf("expected no inference, got %#v", results)
	}
}
