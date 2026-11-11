package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// ReadProjectNameForMaintenance reads the shared label without requiring the
// operational policy that explicit maintenance may be repairing.
func ReadProjectNameForMaintenance(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return "", fmt.Errorf("configuration %s must have a .yaml or .yml suffix", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading project name for maintenance: %w", err)
	}
	nodes, err := configurationNodes(data)
	if err != nil {
		return "", err
	}
	n := nodes["project_name"]
	if err := requireNodeType(n, "project_name", yaml.ScalarNode, "!!str"); err != nil {
		return "", err
	}
	name := strings.TrimSpace(n.Value)
	if name == "" {
		return "", fmt.Errorf("configured project_name is required before updating capture state")
	}
	return name, nil
}
