// TODO: Fix that the config brackets are being stringified

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
)

type LoggingConfig struct {
	LogLevel   string `toml:"log_level"`
	LogFile    string `toml:"log_file"`
	LiveReload bool   `toml:"live_reload"`
}

type TCPConfig struct {
	Port int `toml:"port"`
}

type Config struct {
	Logging LoggingConfig `toml:"hyperatomic.logging"`
	TCP     TCPConfig     `toml:"hyperatomic.tcp"`
}

var (
	cfg      *Config
	cfgMutex sync.RWMutex
)

func DefaultConfig() *Config {
	return &Config{
		Logging: LoggingConfig{
			LogLevel:   "info",
			LogFile:    filepath.Join(os.Getenv("HOME"), ".config", "hyperatomic", "hyperatomic.log"),
			LiveReload: false,
		},
		TCP: TCPConfig{
			Port: 9001,
		},
	}
}

func EnsureConfigExists(configPath string) error {
	configDir := filepath.Dir(configPath)

	// Create config directory if it does not exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// If config file does not exist, write the default config
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
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	newCfg := &Config{}
	if err := toml.Unmarshal(data, newCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfgMutex.Lock()
	cfg = newCfg
	cfgMutex.Unlock()

	if cfg.Logging.LiveReload {
		go watchConfig(configPath)
	}

	return cfg, nil
}

func watchConfig(configPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("❌ Failed to create config watcher:", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(configPath)
	if err != nil {
		fmt.Println("❌ Failed to watch config file:", err)
		return
	}

	fmt.Println("🔄 Live config reloading enabled...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Reload config if the file is modified
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("⚡ Config file changed, reloading...")

				data, err := os.ReadFile(configPath)
				if err != nil {
					fmt.Println("❌ Failed to read updated config:", err)
					continue
				}

				newCfg := &Config{}
				if err := toml.Unmarshal(data, newCfg); err != nil {
					fmt.Println("❌ Failed to parse updated config:", err)
					continue
				}

				cfgMutex.Lock()
				cfg = newCfg
				cfgMutex.Unlock()

				fmt.Println("✅ Config reloaded successfully!")
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("❌ Watcher error:", err)
		}
	}
}

func GetConfig() *Config {
	cfgMutex.RLock()
	defer cfgMutex.RUnlock()

	if cfg == nil {
		fmt.Println("⚠️ Warning: Config is not initialized, returning default config")
		return DefaultConfig()
	}
	return cfg
}
