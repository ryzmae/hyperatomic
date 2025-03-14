package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
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

func LoadConfig() (*Config, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "hyperatomic", "config.toml")

	if err := EnsureConfigExists(configPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if cfg.Logging.LiveReload {
		go watchConfig(configPath)
	}

	return cfg, nil
}

func watchConfig(configPath string) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		fmt.Println("Failed to create config watcher: ", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				data, err := os.ReadFile(configPath)

				if err != nil {
					fmt.Println("Failed to read updated config:", err)
					continue
				}

				newCfg := &Config{}

				if err := toml.Unmarshal(data, newCfg); err != nil {
					fmt.Println("Failed to parse updated config:", err)
					continue
				}

				cfgMutex.Lock()
				cfg = newCfg
				cfgMutex.Unlock()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}
}

func GetConfig() *Config {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()
	return cfg
}
