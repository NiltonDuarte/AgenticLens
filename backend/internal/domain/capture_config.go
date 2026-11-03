package domain

import "errors"

// CaptureConfig configures bounded native publication. Durations use milliseconds.
type CaptureConfig struct {
	DeprecatedDetailedReport bool   `yaml:"-" json:"-"`
	Enabled                  bool   `yaml:"enabled" json:"enabled"`
	StateDir                 string `yaml:"state_dir" json:"state_dir"`
	RefreshMS                int    `yaml:"refresh_ms" json:"refresh_ms"`
	DebounceMS               int    `yaml:"debounce_ms" json:"debounce_ms"`
	IdleMS                   int    `yaml:"idle_ms" json:"idle_ms"`
	MaxLifetimeMS            int    `yaml:"max_lifetime_ms" json:"max_lifetime_ms"`
	PassTimeoutMS            int    `yaml:"pass_timeout_ms" json:"pass_timeout_ms"`
	MaxSessions              int    `yaml:"max_sessions" json:"max_sessions"`
	MaxSourceBytes           int64  `yaml:"max_source_bytes" json:"max_source_bytes"`
	CommitWarning            bool   `yaml:"commit_warning" json:"commit_warning"`
	DetailedReport           bool   `yaml:"detailed_report,omitempty" json:"detailed_report"`
	Summary                  bool   `yaml:"summary" json:"summary"`
}

func DefaultCaptureConfig() CaptureConfig {
	return CaptureConfig{Enabled: true, StateDir: ".agenticlens-capture", RefreshMS: 15000, DebounceMS: 500, IdleMS: 120000, MaxLifetimeMS: 1800000, PassTimeoutMS: 30000, MaxSessions: 256, MaxSourceBytes: 64 * 1024 * 1024, DetailedReport: false, Summary: true, CommitWarning: true}
}

// Validate rejects unbounded or contradictory native worker settings.
func (cfg CaptureConfig) Validate() error {
	if cfg.StateDir == "" || cfg.RefreshMS < 100 || cfg.DebounceMS < 0 || cfg.DebounceMS > cfg.RefreshMS || cfg.IdleMS < cfg.RefreshMS || cfg.MaxLifetimeMS < cfg.IdleMS || cfg.MaxLifetimeMS > 86400000 || cfg.PassTimeoutMS < 100 || cfg.PassTimeoutMS > 300000 || cfg.MaxSessions < 1 || cfg.MaxSourceBytes < 1 {
		return errors.New("invalid telemetry.capture settings: positive bounded cadence, work and state path required")
	}
	return nil
}
