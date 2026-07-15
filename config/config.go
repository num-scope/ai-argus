package config

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Address        string `mapstructure:"address"`
	DatabasePath   string `mapstructure:"database_path"`
	LogLevel       string `mapstructure:"log_level"`
	LogFormat      string `mapstructure:"log_format"`
	GormLogLevel   string `mapstructure:"gorm_log_level"`
	MaxConcurrency int    `mapstructure:"max_concurrency"`
}

type rawConfig struct {
	Address        string `mapstructure:"address"`
	DatabasePath   string `mapstructure:"database_path"`
	LogLevel       string `mapstructure:"log_level"`
	LogFormat      string `mapstructure:"log_format"`
	GormLogLevel   string `mapstructure:"gorm_log_level"`
	MaxConcurrency string `mapstructure:"max_concurrency"`
}

func newViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetDefault("address", "127.0.0.1:8080")
	v.SetDefault("database_path", "data/ai-argus.db")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "console")
	v.SetDefault("gorm_log_level", "warn")
	v.SetDefault("max_concurrency", "1000")

	bindings := map[string]string{
		"address":         "ARGUS_ADDRESS",
		"database_path":   "ARGUS_DATABASE_PATH",
		"log_level":       "ARGUS_LOG_LEVEL",
		"log_format":      "ARGUS_LOG_FORMAT",
		"gorm_log_level":  "ARGUS_GORM_LOG_LEVEL",
		"max_concurrency": "ARGUS_MAX_CONCURRENCY",
	}
	for key, environment := range bindings {
		if err := v.BindEnv(key, environment); err != nil {
			return nil, fmt.Errorf("bind %s: %w", environment, err)
		}
	}
	return v, nil
}

func Load() (Config, error) {
	v, err := newViper()
	if err != nil {
		return Config{}, err
	}
	var raw rawConfig
	if err := v.Unmarshal(&raw); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}

	maxConcurrency, err := strconv.Atoi(strings.TrimSpace(raw.MaxConcurrency))
	if err != nil {
		return Config{}, fmt.Errorf("ARGUS_MAX_CONCURRENCY must be an integer: %w", err)
	}
	cfg := Config{
		Address:        strings.TrimSpace(raw.Address),
		DatabasePath:   strings.TrimSpace(raw.DatabasePath),
		LogLevel:       strings.ToLower(strings.TrimSpace(raw.LogLevel)),
		LogFormat:      strings.ToLower(strings.TrimSpace(raw.LogFormat)),
		GormLogLevel:   strings.ToLower(strings.TrimSpace(raw.GormLogLevel)),
		MaxConcurrency: maxConcurrency,
	}

	if cfg.Address == "" {
		return Config{}, fmt.Errorf("ARGUS_ADDRESS cannot be empty")
	}
	if cfg.DatabasePath == "" {
		return Config{}, fmt.Errorf("ARGUS_DATABASE_PATH cannot be empty")
	}
	if cfg.MaxConcurrency < 1 {
		return Config{}, fmt.Errorf("ARGUS_MAX_CONCURRENCY must be greater than zero")
	}
	if cfg.LogLevel != "debug" && cfg.LogLevel != "info" && cfg.LogLevel != "warn" && cfg.LogLevel != "error" {
		return Config{}, fmt.Errorf("ARGUS_LOG_LEVEL must be debug, info, warn, or error")
	}
	if cfg.LogFormat != "console" && cfg.LogFormat != "json" {
		return Config{}, fmt.Errorf("ARGUS_LOG_FORMAT must be console or json")
	}
	if cfg.GormLogLevel != "silent" && cfg.GormLogLevel != "error" && cfg.GormLogLevel != "warn" && cfg.GormLogLevel != "info" {
		return Config{}, fmt.Errorf("ARGUS_GORM_LOG_LEVEL must be silent, error, warn, or info")
	}

	absPath, err := filepath.Abs(cfg.DatabasePath)
	if err != nil {
		return Config{}, fmt.Errorf("resolve database path: %w", err)
	}
	cfg.DatabasePath = absPath
	return cfg, nil
}
