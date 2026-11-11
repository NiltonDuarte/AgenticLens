package config

import (
	"agenticlens/internal/domain"
	"os"
	"testing"
)

func fixtureProjectPolicy(t *testing.T) domain.ProjectConfig {
	t.Helper()
	cfg, err := LoadFromFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func fixtureSharedYAML(t *testing.T, name, workspace string) ([]byte, error) {
	t.Helper()
	cfg := fixtureProjectPolicy(t)
	cfg.ProjectName = name
	cfg.Workspaces.Dir = workspace
	return GenerateConfigYAML(cfg)
}

func fixtureSharedBody(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
