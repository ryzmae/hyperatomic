package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Logging struct {
		LogLevel   string `toml:"log_level"`
		LogFile    string `toml:"log_file"`
		LiveReload bool   `toml:"live_reload"`
	} `toml:"hyperatomic.logging"`
}

var (
	cfg      *Config
	cfgMutex *sync.RWMutex
)

func DefaultConfig() *Config {
	return &Config{
		Logging: struct {
			LogLevel   string `toml:"log_level"`
			LogFile    string `toml:"log_file"`
			LiveReload bool   `toml:"live_reload"`
		}{
			LogLevel:   "info",
			LogFile:    filepath.Join(os.Getenv("HOME"), ".config/hyperatomic/hyperatomic.log"),
			LiveReload: false,
		},
	}
}

func EnsureConfigExists(configPath string) error {
	configDir := filepath.Dir(configPath)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultCfg := DefaultConfig()
		data, err := toml.Marshal(defaultCfg)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	return nil
}
