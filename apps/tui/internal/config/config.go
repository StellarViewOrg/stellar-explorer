package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultProductName    = "stellar-tui"
	defaultProfileName    = "default"
	defaultNetworkName    = "testnet"
	defaultRPCEndpoint    = "https://soroban-testnet.stellar.org"
	defaultIndexerURL     = "http://127.0.0.1:8081"
	defaultRedisURL       = "redis://127.0.0.1:63890"
	defaultBackendMode    = "rpc"
	defaultCacheDriver    = "sqlite"
	defaultConfigFileName = "config.json"
	defaultCacheFileName  = "cache.db"
)

const (
	BackendModeRPC     = "rpc"
	BackendModeHybrid  = "hybrid"
	BackendModeIndexer = "indexer"
)

// Config captures the minimum local application configuration required to boot
// the terminal client before persistence and remote backends are integrated.
type Config struct {
	ProductName    string    `json:"product_name"`
	DefaultProfile string    `json:"default_profile"`
	Profiles       []Profile `json:"profiles"`
	Cache          Cache     `json:"cache"`
}

// Cache controls the local persistence backend used by the terminal client.
type Cache struct {
	Driver string `json:"driver"`
	Path   string `json:"path"`
}

// Profile defines a runnable network profile for the terminal client.
type Profile struct {
	Name         string `json:"name"`
	Network      string `json:"network"`
	RPCEndpoint  string `json:"rpc_endpoint"`
	HorizonURL   string `json:"horizon_url,omitempty"`
	IndexerURL   string `json:"indexer_url,omitempty"`
	RedisURL     string `json:"redis_url,omitempty"`
	BackendMode  string `json:"backend_mode"`
	DefaultWatch string `json:"default_watch,omitempty"`
}

func (p Profile) NormalizedBackendMode() string {
	switch value := normalizeBackendMode(p.BackendMode); value {
	case BackendModeRPC, BackendModeHybrid, BackendModeIndexer:
		return value
	default:
		if p.IndexerURL != "" {
			return BackendModeIndexer
		}
		return BackendModeRPC
	}
}

func (p Profile) PreferredSource() string {
	switch p.NormalizedBackendMode() {
	case BackendModeHybrid, BackendModeIndexer:
		return "indexer"
	default:
		return "rpc"
	}
}

func (p Profile) FallbackSource() string {
	if p.NormalizedBackendMode() == BackendModeHybrid {
		return "rpc"
	}
	return ""
}

func normalizeBackendMode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// Default returns a minimal but valid boot configuration for standalone installs.
func Default() Config {
	return Config{
		ProductName:    defaultProductName,
		DefaultProfile: defaultProfileName,
		Profiles: []Profile{
			{
				Name:        defaultProfileName,
				Network:     defaultNetworkName,
				RPCEndpoint: defaultRPCEndpoint,
				BackendMode: defaultBackendMode,
			},
		},
		Cache: Cache{
			Driver: defaultCacheDriver,
		},
	}
}

// ResolvePath returns an explicit config path or searches cwd then ~/.config/stellar-tui/.
func ResolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	if value := strings.TrimSpace(os.Getenv(envConfigPath)); value != "" {
		return value, nil
	}
	return searchPath(defaultConfigFileName)
}

// ResolveCachePath returns the requested cache path or a deterministic default path.
func ResolveCachePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	configDir, err := UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, defaultCacheFileName), nil
}

// EnsureFile writes cfg to path when the file does not exist yet.
func EnsureFile(path string, cfg Config) (bool, error) {
	if path == "" {
		return false, errors.New("config path is required")
	}

	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("stat config file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("create config directory: %w", err)
	}

	encoded, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Errorf("encode config file: %w", err)
	}
	encoded = append(encoded, '\n')

	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return false, fmt.Errorf("write config file: %w", err)
	}

	return true, nil
}

// Load reads a config file when present and falls back to sane defaults when it
// does not exist yet.
func Load(path string) (Config, string, bool, error) {
	resolvedPath, err := ResolvePath(path)
	if err != nil {
		return Config{}, "", false, err
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := ApplyEnvOverrides(Default())
			created, ensureErr := EnsureFile(resolvedPath, cfg)
			if ensureErr != nil {
				return Config{}, "", false, ensureErr
			}
			return cfg, resolvedPath, created, cfg.Validate()
		}
		return Config{}, "", false, fmt.Errorf("read config file: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, "", false, fmt.Errorf("decode config file: %w", err)
	}

	cfg = ApplyEnvOverrides(cfg)
	return cfg, resolvedPath, false, cfg.Validate()
}

// Validate ensures the app can boot with the available config.
func (c Config) Validate() error {
	if c.ProductName == "" {
		return errors.New("product_name is required")
	}

	if c.DefaultProfile == "" {
		return errors.New("default_profile is required")
	}

	if len(c.Profiles) == 0 {
		return errors.New("at least one profile is required")
	}

	_, ok := c.Profile(c.DefaultProfile)
	if !ok {
		return fmt.Errorf("default_profile %q not found", c.DefaultProfile)
	}

	for _, profile := range c.Profiles {
		if profile.Name == "" {
			return errors.New("profile name is required")
		}
		if profile.Network == "" {
			return fmt.Errorf("profile %q network is required", profile.Name)
		}
		if profile.RPCEndpoint == "" {
			return fmt.Errorf("profile %q rpc_endpoint is required", profile.Name)
		}
		switch profile.NormalizedBackendMode() {
		case BackendModeRPC, BackendModeHybrid, BackendModeIndexer:
		default:
			return fmt.Errorf("profile %q backend_mode %q is invalid", profile.Name, profile.BackendMode)
		}
		if profile.HorizonURL != "" {
			parsed, err := url.Parse(profile.HorizonURL)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return fmt.Errorf("profile %q horizon_url is invalid", profile.Name)
			}
		}
		if profile.IndexerURL != "" {
			parsed, err := url.Parse(profile.IndexerURL)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return fmt.Errorf("profile %q indexer_url is invalid", profile.Name)
			}
		}
		if profile.RedisURL != "" {
			parsed, err := url.Parse(profile.RedisURL)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return fmt.Errorf("profile %q redis_url is invalid", profile.Name)
			}
		}
	}

	if c.Cache.Driver != "" {
		if _, err := ResolveCachePath(c.Cache.Path); err != nil {
			return err
		}
	}

	return nil
}

// Profile returns the named profile if it exists.
func (c Config) Profile(name string) (Profile, bool) {
	for _, profile := range c.Profiles {
		if profile.Name == name {
			return profile, true
		}
	}

	return Profile{}, false
}
