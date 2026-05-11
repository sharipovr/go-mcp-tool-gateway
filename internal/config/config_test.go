package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
server:
  addr: ":9090"
  log_level: "debug"
tools:
  echo:
    enabled: true
    category: "utility"
    tags: ["test"]
  hash:
    enabled: false
    category: "crypto"
`), 0644)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.Server.Addr)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected debug, got %s", cfg.Server.LogLevel)
	}
	if !cfg.Tools["echo"].IsEnabled() {
		t.Error("echo should be enabled")
	}
	if cfg.Tools["hash"].IsEnabled() {
		t.Error("hash should be disabled")
	}
	if cfg.Tools["echo"].Category != "utility" {
		t.Errorf("expected category utility, got %s", cfg.Tools["echo"].Category)
	}
	if len(cfg.Tools["echo"].Tags) != 1 || cfg.Tools["echo"].Tags[0] != "test" {
		t.Errorf("unexpected tags: %v", cfg.Tools["echo"].Tags)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte(`{{{invalid`), 0644)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Server.Addr != ":8080" {
		t.Errorf("expected :8080, got %s", cfg.Server.Addr)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("expected info, got %s", cfg.Server.LogLevel)
	}
}

func TestEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
server:
  addr: ":8080"
`), 0644)

	t.Setenv("ADDR", ":7070")
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != ":7070" {
		t.Errorf("expected :7070, got %s", cfg.Server.Addr)
	}
	if cfg.Server.LogLevel != "warn" {
		t.Errorf("expected warn, got %s", cfg.Server.LogLevel)
	}
}

func TestToolEnabledDefault(t *testing.T) {
	tc := config.ToolConfig{}
	if !tc.IsEnabled() {
		t.Error("tool should be enabled by default when Enabled is nil")
	}
}
