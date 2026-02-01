package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Patterns             []string  `yaml:"patterns"`
	Exclude              []string  `yaml:"exclude"`
	OnMissingFrontmatter string    `yaml:"on_missing_frontmatter"`
	Workers              int       `yaml:"workers"` // 0 means use CPU count
	Defaults             *Defaults `yaml:"defaults"`
}

type Defaults struct {
	Strategy string   `yaml:"strategy"`
	Interval string   `yaml:"interval"`
	Watch    []string `yaml:"watch"`
	Ignore   []string `yaml:"ignore"`
}

func DefaultConfig() *Config {
	return &Config{
		Patterns: []string{
			"**/doc/**/*.md",
			"**/docs/**/*.md",
		},
		Exclude: []string{
			"**/node_modules/**",
			"**/vendor/**",
		},
		OnMissingFrontmatter: "warn",
		Defaults: &Defaults{
			Strategy: "interval",
			Interval: "180d",
			// Watch and Ignore are nil by default; smart defaults are computed
			// based on document location (watch parent dir, ignore docs dir)
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return cfg, nil
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, err
	}

	// Merge with defaults
	if len(fileCfg.Patterns) > 0 {
		cfg.Patterns = fileCfg.Patterns
	}
	if len(fileCfg.Exclude) > 0 {
		cfg.Exclude = fileCfg.Exclude
	}
	if fileCfg.OnMissingFrontmatter != "" {
		cfg.OnMissingFrontmatter = fileCfg.OnMissingFrontmatter
	}
	if fileCfg.Workers > 0 {
		cfg.Workers = fileCfg.Workers
	}
	if fileCfg.Defaults != nil {
		if fileCfg.Defaults.Strategy != "" {
			cfg.Defaults.Strategy = fileCfg.Defaults.Strategy
		}
		if fileCfg.Defaults.Interval != "" {
			cfg.Defaults.Interval = fileCfg.Defaults.Interval
		}
		if len(fileCfg.Defaults.Watch) > 0 {
			cfg.Defaults.Watch = fileCfg.Defaults.Watch
		}
		if len(fileCfg.Defaults.Ignore) > 0 {
			cfg.Defaults.Ignore = fileCfg.Defaults.Ignore
		}
	}

	return cfg, nil
}
