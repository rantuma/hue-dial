package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	domainconfig "github.com/rantuma/hue-dial/domain/config"
	"github.com/rantuma/hue-dial/domain/ports"
)

type (
	store struct {
		path string
	}
)

func New() (ports.ConfigStore, error) {
	p := os.Getenv("CONFIG_PATH")
	if p == "" {
		p = resolveDefaultPath()
	}

	return &store{path: p}, nil
}

func resolveDefaultPath() string {
	candidates := []string{
		"/data/config.json",
		filepath.Join(userConfigDir(), "hue-dial", "config.json"),
		filepath.Join(".", "config.json"),
	}

	for _, p := range candidates {
		dir := filepath.Dir(p)

		err := os.MkdirAll(dir, 0o750)
		if err == nil {
			return p
		}
	}

	return candidates[len(candidates)-1]
}

func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return os.TempDir()
	}

	return dir
}

func (st *store) Exists() bool {
	_, err := os.Stat(st.path)
	return err == nil
}

func (st *store) Load() (domainconfig.SetupConfig, error) {
	var cfg domainconfig.SetupConfig

	data, err := os.ReadFile(st.path)
	if err != nil {
		return cfg, fmt.Errorf(
			"failed to read config file %q: %w", st.path, err,
		)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, fmt.Errorf(
			"failed to parse config file %q: %w", st.path, err,
		)
	}

	return cfg, nil
}

func (st *store) Save(cfg domainconfig.SetupConfig) error {
	//nolint:gosec // API key must be persisted to disk
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(st.path), 0o750)
	if err != nil {
		return fmt.Errorf(
			"failed to create config directory %q: %w",
			filepath.Dir(st.path), err,
		)
	}

	err = os.WriteFile(st.path, data, 0o600)
	if err != nil {
		return fmt.Errorf(
			"failed to write config file %q: %w", st.path, err,
		)
	}

	return nil
}

func (st *store) Path() string {
	return st.path
}
