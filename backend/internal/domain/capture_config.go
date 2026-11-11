package domain

import (
	"fmt"
	"strings"
)

// CaptureConfig is selected from one whole machine-local file. Durations are milliseconds.
type CaptureConfig struct {
	Settings       map[string]bool `yaml:"-" json:"-"`
	Enabled        bool            `yaml:"enabled" json:"enabled"`
	StateDir       string          `yaml:"state_dir" json:"state_dir"`
	RefreshMS      int             `yaml:"refresh_ms" json:"refresh_ms"`
	DebounceMS     int             `yaml:"debounce_ms" json:"debounce_ms"`
	IdleMS         int             `yaml:"idle_ms" json:"idle_ms"`
	MaxLifetimeMS  int             `yaml:"max_lifetime_ms" json:"max_lifetime_ms"`
	PassTimeoutMS  int             `yaml:"pass_timeout_ms" json:"pass_timeout_ms"`
	MaxSessions    int             `yaml:"max_sessions" json:"max_sessions"`
	MaxSourceBytes int64           `yaml:"max_source_bytes" json:"max_source_bytes"`
	CommitWarning  bool            `yaml:"commit_warning" json:"commit_warning"`
}

func (cfg CaptureConfig) RequireSettings(keys ...string) error {
	for _, key := range keys {
		if cfg.Settings != nil && !cfg.Settings[key] {
			return fmt.Errorf("telemetry.capture.%s must be explicitly configured in the selected local YAML", key)
		}
	}
	return nil
}
func (cfg CaptureConfig) ValidateState() error {
	if err := cfg.RequireSettings("enabled", "state_dir"); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.StateDir) == "" || strings.IndexByte(cfg.StateDir, 0) >= 0 {
		return fmt.Errorf("telemetry.capture.state_dir must be an explicitly configured nonblank path")
	}
	return nil
}

// Validate checks state and any provided settings without requiring unused worker policy.
func (cfg CaptureConfig) Validate() error {
	if err := cfg.ValidateState(); err != nil {
		return err
	}
	present := func(key string) bool { return cfg.Settings == nil || cfg.Settings[key] }
	for _, v := range []struct {
		key   string
		value int
	}{{"refresh_ms", cfg.RefreshMS}, {"idle_ms", cfg.IdleMS}, {"max_lifetime_ms", cfg.MaxLifetimeMS}, {"pass_timeout_ms", cfg.PassTimeoutMS}, {"max_sessions", cfg.MaxSessions}} {
		if present(v.key) && v.value <= 0 {
			return fmt.Errorf("telemetry.capture.%s must be positive", v.key)
		}
	}
	if present("max_source_bytes") && cfg.MaxSourceBytes <= 0 {
		return fmt.Errorf("telemetry.capture.max_source_bytes must be positive")
	}
	if present("refresh_ms") && cfg.RefreshMS < 100 {
		return fmt.Errorf("telemetry.capture.refresh_ms must be at least 100")
	}
	if present("debounce_ms") && (cfg.DebounceMS < 0 || (present("refresh_ms") && cfg.DebounceMS > cfg.RefreshMS)) {
		return fmt.Errorf("telemetry.capture.debounce_ms must be from 0 through refresh_ms")
	}
	if present("idle_ms") && present("refresh_ms") && cfg.IdleMS < cfg.RefreshMS {
		return fmt.Errorf("telemetry.capture.idle_ms must be at least refresh_ms")
	}
	if present("max_lifetime_ms") && (cfg.MaxLifetimeMS > 86400000 || (present("idle_ms") && cfg.MaxLifetimeMS < cfg.IdleMS)) {
		return fmt.Errorf("telemetry.capture.max_lifetime_ms must be from idle_ms through 86400000")
	}
	if present("pass_timeout_ms") && (cfg.PassTimeoutMS < 100 || cfg.PassTimeoutMS > 300000) {
		return fmt.Errorf("telemetry.capture.pass_timeout_ms must be from 100 through 300000")
	}
	return nil
}
func (cfg CaptureConfig) ValidateWorker() error {
	if err := cfg.RequireSettings("refresh_ms", "debounce_ms", "idle_ms", "max_lifetime_ms", "pass_timeout_ms", "max_sessions", "max_source_bytes"); err != nil {
		return err
	}
	return cfg.Validate()
}
func (cfg CaptureConfig) ValidateBudgets() error {
	if err := cfg.RequireSettings("max_sessions", "max_source_bytes"); err != nil {
		return err
	}
	if cfg.MaxSessions < 1 || cfg.MaxSourceBytes < 1 {
		return fmt.Errorf("telemetry.capture.max_sessions and max_source_bytes must be positive integers")
	}
	return nil
}
func (cfg CaptureConfig) ValidatePass() error {
	if err := cfg.RequireSettings("pass_timeout_ms"); err != nil {
		return err
	}
	if cfg.PassTimeoutMS < 100 || cfg.PassTimeoutMS > 300000 {
		return fmt.Errorf("telemetry.capture.pass_timeout_ms must be from 100 through 300000")
	}
	return nil
}
