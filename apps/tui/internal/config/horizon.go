package config

import "strings"

// ResolveHorizonURL returns the configured Horizon endpoint or a network default.
func ResolveHorizonURL(profile Profile) string {
	if value := strings.TrimSpace(profile.HorizonURL); value != "" {
		return value
	}
	switch strings.ToLower(strings.TrimSpace(profile.Network)) {
	case "public", "mainnet":
		return "https://horizon.stellar.org"
	case "testnet":
		return "https://horizon-testnet.stellar.org"
	case "futurenet":
		return "https://horizon-futurenet.stellar.org"
	default:
		return ""
	}
}
