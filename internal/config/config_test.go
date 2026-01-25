package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_WithValidConfig(t *testing.T) {
	content := `patterns:
  - "**/doc/**/*.md"
  - "**/docs/**/*.md"
exclude:
  - "**/node_modules/**"
on_missing_frontmatter: warn
defaults:
  strategy: interval
  interval: 180d
  watch:
    - "**/*.rb"
    - "**/*.go"
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".docrot.yml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Patterns) != 2 {
		t.Errorf("Patterns length = %v, want %v", len(cfg.Patterns), 2)
	}

	if cfg.Patterns[0] != "**/doc/**/*.md" {
		t.Errorf("Patterns[0] = %v, want %v", cfg.Patterns[0], "**/doc/**/*.md")
	}

	if len(cfg.Exclude) != 1 {
		t.Errorf("Exclude length = %v, want %v", len(cfg.Exclude), 1)
	}

	if cfg.OnMissingFrontmatter != "warn" {
		t.Errorf("OnMissingFrontmatter = %v, want %v", cfg.OnMissingFrontmatter, "warn")
	}

	if cfg.Defaults.Strategy != "interval" {
		t.Errorf("Defaults.Strategy = %v, want %v", cfg.Defaults.Strategy, "interval")
	}

	if cfg.Defaults.Interval != "180d" {
		t.Errorf("Defaults.Interval = %v, want %v", cfg.Defaults.Interval, "180d")
	}
}

func TestLoad_FileNotFound_ReturnsDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/.docrot.yml")
	if err != nil {
		t.Fatalf("Load() error = %v, expected default config", err)
	}

	if len(cfg.Patterns) == 0 {
		t.Error("Expected default patterns, got empty")
	}

	if cfg.OnMissingFrontmatter != "warn" {
		t.Errorf("OnMissingFrontmatter = %v, want default 'warn'", cfg.OnMissingFrontmatter)
	}
}

func TestLoad_EmptyConfig_UsesDefaults(t *testing.T) {
	content := ``
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".docrot.yml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Patterns) == 0 {
		t.Error("Expected default patterns for empty config")
	}
}

func TestLoad_WithWorkers(t *testing.T) {
	content := `workers: 4
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".docrot.yml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Workers != 4 {
		t.Errorf("Workers = %v, want 4", cfg.Workers)
	}
}

func TestLoad_DefaultWorkers_IsZero(t *testing.T) {
	// Zero means "use CPU count" at runtime
	cfg, err := Load("/nonexistent/.docrot.yml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Workers != 0 {
		t.Errorf("Workers = %v, want 0 (default)", cfg.Workers)
	}
}

func TestLoad_PartialConfig_MergesWithDefaults(t *testing.T) {
	content := `patterns:
  - "custom/**/*.md"
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".docrot.yml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Custom patterns should be used
	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "custom/**/*.md" {
		t.Errorf("Patterns = %v, want [custom/**/*.md]", cfg.Patterns)
	}

	// Default on_missing_frontmatter should still apply
	if cfg.OnMissingFrontmatter != "warn" {
		t.Errorf("OnMissingFrontmatter = %v, want default 'warn'", cfg.OnMissingFrontmatter)
	}
}
