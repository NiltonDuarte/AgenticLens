package capture_test

import (
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
	"testing"
)

func fixtureSharedYAML(t *testing.T, name, workspace string) ([]byte, error) {
	t.Helper()
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ProjectName = name
	cfg.Generation.Metrics = domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
	cfg.Workspaces.Dir = workspace
	return adaptersconfig.GenerateConfigYAML(cfg)
}

// Explicit fixture input for bounded diagnostic/reconciliation operations.
const fixtureByteBudget int64 = 64 * 1024 * 1024

func fixtureGenerationPolicy() domain.GenerationPolicy {
	return domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
}
