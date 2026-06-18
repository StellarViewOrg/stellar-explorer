package config

import "testing"

func TestResolveHorizonURLUsesProfileOverride(t *testing.T) {
	got := ResolveHorizonURL(Profile{
		Network:    "testnet",
		HorizonURL: "https://horizon.example.test",
	})
	if got != "https://horizon.example.test" {
		t.Fatalf("ResolveHorizonURL() = %q", got)
	}
}

func TestResolveHorizonURLDefaultsByNetwork(t *testing.T) {
	cases := map[string]string{
		"public":    "https://horizon.stellar.org",
		"mainnet":   "https://horizon.stellar.org",
		"testnet":   "https://horizon-testnet.stellar.org",
		"futurenet": "https://horizon-futurenet.stellar.org",
	}
	for network, want := range cases {
		got := ResolveHorizonURL(Profile{Network: network})
		if got != want {
			t.Fatalf("network %q = %q, want %q", network, got, want)
		}
	}
}

func TestResolveHorizonURLEmptyForUnknownNetwork(t *testing.T) {
	if got := ResolveHorizonURL(Profile{Network: "custom"}); got != "" {
		t.Fatalf("ResolveHorizonURL() = %q, want empty", got)
	}
}
