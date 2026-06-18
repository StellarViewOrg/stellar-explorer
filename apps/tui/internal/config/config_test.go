package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
	if cfg.Cache.Driver != defaultCacheDriver {
		t.Fatalf("expected default cache driver %q, got %q", defaultCacheDriver, cfg.Cache.Driver)
	}
}

func TestProfileLookup(t *testing.T) {
	cfg := Default()
	profile, ok := cfg.Profile(defaultProfileName)
	if !ok {
		t.Fatal("expected default profile to exist")
	}

	if profile.Network != defaultNetworkName {
		t.Fatalf("expected network %q, got %q", defaultNetworkName, profile.Network)
	}
	if profile.IndexerURL != "" {
		t.Fatalf("expected standalone default profile to omit indexer_url")
	}
}

func TestResolveCachePathUsesDefaultLocation(t *testing.T) {
	path, err := ResolveCachePath("")
	if err != nil {
		t.Fatalf("resolve cache path: %v", err)
	}

	if !strings.HasSuffix(path, "/stellar-tui/"+defaultCacheFileName) {
		t.Fatalf("expected default cache suffix, got %q", path)
	}
}

func TestResolvePathPrefersWorkingDirectoryConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, defaultConfigFileName)
	if err := os.WriteFile(configPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(original)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	path, err := ResolvePath("")
	if err != nil {
		t.Fatalf("ResolvePath() error = %v", err)
	}
	samePath := func(left, right string) bool {
		var evalErr error
		left, evalErr = filepath.EvalSymlinks(left)
		if evalErr != nil {
			t.Fatalf("eval symlinks %q: %v", left, evalErr)
		}
		right, evalErr = filepath.EvalSymlinks(right)
		if evalErr != nil {
			t.Fatalf("eval symlinks %q: %v", right, evalErr)
		}
		return left == right
	}
	if !samePath(path, configPath) {
		t.Fatalf("ResolvePath() = %q, want %q", path, configPath)
	}
}

func TestApplyEnvOverridesUpdatesDefaultProfile(t *testing.T) {
	t.Setenv(envRPCURL, "https://rpc.example.test")
	t.Setenv(envNetwork, "public")
	t.Setenv(envBackendMode, BackendModeHybrid)
	t.Setenv(envIndexerURL, "http://127.0.0.1:9090")
	t.Setenv(envHorizonURL, "https://horizon.example.test")
	t.Setenv(envProfile, defaultProfileName)

	cfg := ApplyEnvOverrides(Default())
	profile, ok := cfg.Profile(defaultProfileName)
	if !ok {
		t.Fatal("expected default profile")
	}

	if profile.RPCEndpoint != "https://rpc.example.test" {
		t.Fatalf("rpc endpoint = %q", profile.RPCEndpoint)
	}
	if profile.Network != "public" {
		t.Fatalf("network = %q", profile.Network)
	}
	if profile.BackendMode != BackendModeHybrid {
		t.Fatalf("backend mode = %q", profile.BackendMode)
	}
	if profile.IndexerURL != "http://127.0.0.1:9090" {
		t.Fatalf("indexer url = %q", profile.IndexerURL)
	}
	if profile.HorizonURL != "https://horizon.example.test" {
		t.Fatalf("horizon url = %q", profile.HorizonURL)
	}
}

func TestLoadCreatesMissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, defaultConfigFileName)

	cfg, resolved, created, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !created {
		t.Fatal("expected first load to create config file")
	}
	if resolved != configPath {
		t.Fatalf("resolved path = %q", resolved)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("created config invalid: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read created config: %v", err)
	}
	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode created config: %v", err)
	}
}

func TestValidateRejectsInvalidIndexerURL(t *testing.T) {
	cfg := Default()
	cfg.Profiles[0].IndexerURL = "localhost:8081"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid indexer_url to fail validation")
	}
}

func TestLoadLabelsFileSupportsStringAndStructuredEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, defaultLabelsFileName)
	content := `
[accounts]
"GABC" = "Treasury"
"GDEF" = { name = "Ops Wallet", tags = ["ops", "treasury"] }

[transactions]
"deadbeef" = "Deploy"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write labels: %v", err)
	}

	attachments, err := LoadLabelsFile(path)
	if err != nil {
		t.Fatalf("LoadLabelsFile() error = %v", err)
	}
	if len(attachments) != 3 {
		t.Fatalf("expected 3 attachments, got %d", len(attachments))
	}
}
