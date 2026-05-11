package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig          `yaml:"server"`
	Tools  map[string]ToolConfig `yaml:"tools"`
}

type ServerConfig struct {
	Addr     string `yaml:"addr"`
	LogLevel string `yaml:"log_level"`
}

type ToolConfig struct {
	Enabled  *bool    `yaml:"enabled"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
}

func (tc ToolConfig) IsEnabled() bool {
	if tc.Enabled == nil {
		return true
	}
	return *tc.Enabled
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if addr := os.Getenv("ADDR"); addr != "" {
		cfg.Server.Addr = addr
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Server.LogLevel = level
	}

	return cfg, nil
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:     ":8080",
			LogLevel: "info",
		},
		Tools: map[string]ToolConfig{},
	}
}
