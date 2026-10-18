package domain

import (
	"fmt"
)

// Default build metadata constants applied when compilation linker flags are omitted.
const (
	DefaultVersion   = "v0.1.0-dev"
	DefaultGitCommit = "unknown"
	DefaultBuildDate = "unknown"
)

// Package-level variables populated at compile time via linker flags (-ldflags -X).
var (
	Version   = DefaultVersion
	GitCommit = DefaultGitCommit
	BuildDate = DefaultBuildDate
)

// BuildInfo represents the immutable compilation and release metadata for AgenticLens.
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"commit"`
	BuildDate string `json:"build_date"`
}

// GetBuildInfo returns an immutable BuildInfo value object constructed from package variables,
// falling back to defaults if any variable is empty.
func GetBuildInfo() BuildInfo {
	v := Version
	if v == "" {
		v = DefaultVersion
	}
	commit := GitCommit
	if commit == "" {
		commit = DefaultGitCommit
	}
	date := BuildDate
	if date == "" {
		date = DefaultBuildDate
	}
	return BuildInfo{
		Version:   v,
		GitCommit: commit,
		BuildDate: date,
	}
}

// FormatBuildInfo returns a standard formatted version string:
// agentic-lens version <version> (commit: <commit>, built: <build_date>)
func FormatBuildInfo(info BuildInfo) string {
	v := info.Version
	if v == "" {
		v = DefaultVersion
	}
	commit := info.GitCommit
	if commit == "" {
		commit = DefaultGitCommit
	}
	date := info.BuildDate
	if date == "" {
		date = DefaultBuildDate
	}
	return fmt.Sprintf("agentic-lens version %s (commit: %s, built: %s)", v, commit, date)
}

// FormatShortVersion returns the short semantic version string.
func FormatShortVersion(info BuildInfo) string {
	if info.Version == "" {
		return DefaultVersion
	}
	return info.Version
}
