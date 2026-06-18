package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appConfigDirName      = "stellar-tui"
	defaultLabelsFileName = "labels.toml"
)

// UserConfigDir returns the application config directory (~/.config/stellar-tui).
func UserConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, appConfigDirName), nil
}

func searchPath(filename string) (string, error) {
	if cwd, err := os.Getwd(); err == nil {
		local := filepath.Join(cwd, filename)
		if _, err := os.Stat(local); err == nil {
			return local, nil
		}
	}

	configDir, err := UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, filename), nil
}
