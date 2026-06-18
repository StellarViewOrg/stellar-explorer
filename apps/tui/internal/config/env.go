package config

import (
	"os"
	"strings"
)

const (
	envConfigPath  = "STELLAR_TUI_CONFIG"
	envProfile     = "STELLAR_TUI_PROFILE"
	envLabelsPath  = "STELLAR_TUI_LABELS"
	envRPCURL      = "STELLAR_RPC_URL"
	envHorizonURL  = "STELLAR_HORIZON_URL"
	envNetwork     = "STELLAR_NETWORK"
	envBackendMode = "STELLAR_BACKEND_MODE"
	envIndexerURL  = "STELLAR_INDEXER_URL"
	envRedisURL    = "STELLAR_REDIS_URL"
)

// ApplyEnvOverrides mutates cfg using supported STELLAR_* environment variables.
func ApplyEnvOverrides(cfg Config) Config {
	if value := strings.TrimSpace(os.Getenv(envProfile)); value != "" {
		cfg.DefaultProfile = value
	}

	profileName := cfg.DefaultProfile
	profile, ok := cfg.Profile(profileName)
	if !ok && len(cfg.Profiles) > 0 {
		profile = cfg.Profiles[0]
		profileName = profile.Name
		ok = true
	}
	if !ok {
		return cfg
	}

	if value := strings.TrimSpace(os.Getenv(envRPCURL)); value != "" {
		profile.RPCEndpoint = value
	}
	if value := strings.TrimSpace(os.Getenv(envHorizonURL)); value != "" {
		profile.HorizonURL = value
	}
	if value := strings.TrimSpace(os.Getenv(envNetwork)); value != "" {
		profile.Network = value
	}
	if value := strings.TrimSpace(os.Getenv(envBackendMode)); value != "" {
		profile.BackendMode = value
	}
	if value := strings.TrimSpace(os.Getenv(envIndexerURL)); value != "" {
		profile.IndexerURL = value
	}
	if value := strings.TrimSpace(os.Getenv(envRedisURL)); value != "" {
		profile.RedisURL = value
	}

	cfg = cfg.withProfile(profileName, profile)
	return cfg
}

func (c Config) withProfile(name string, profile Profile) Config {
	for i, existing := range c.Profiles {
		if existing.Name == name {
			c.Profiles[i] = profile
			return c
		}
	}
	return c
}

// ResolveLabelsPath returns the labels file path from STELLAR_TUI_LABELS, cwd, or user config.
func ResolveLabelsPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	if value := strings.TrimSpace(os.Getenv(envLabelsPath)); value != "" {
		return value, nil
	}
	return searchPath(defaultLabelsFileName)
}
