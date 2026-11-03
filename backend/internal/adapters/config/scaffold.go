package config

import (
	"bytes"
	"errors"
)

// BootstrapScaffold is deliberately incomplete; guarded native init supplies
// explicitly known project/ticket/code-root settings before publication readiness.
func BootstrapScaffold(version string) ([]byte, error) {
	if err := ValidateReleaseVersion(version); err != nil {
		return nil, err
	}
	return []byte("# AgenticLens bootstrap scaffold v1: incomplete project configuration\nagenticlens:\n  version: " + version + "\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\n    commit_warning: true\n"), nil
}

// ValidateBootstrapScaffold permits completion of only the exact owned scaffold.
// A changed comment or existing unrelated setting requires explicit reconciliation.
func ValidateBootstrapScaffold(data []byte, version string) error {
	expected, err := BootstrapScaffold(version)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, expected) {
		return errors.New("configuration is not the exact owned bootstrap scaffold for this release; existing files preserved")
	}
	return nil
}
