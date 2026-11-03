package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReadProjectNameForMaintenance interprets the shared label using the same raw
// DTO, format rules and merge semantics as operational loading. It intentionally
// does not validate the remaining configuration or supply an operational state
// path: explicit maintenance must be able to repair a missing required key.
func ReadProjectNameForMaintenance(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading project name for maintenance: %w", err)
	}
	var dto rawProjectConfigDTO
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &dto)
	case ".json":
		err = json.Unmarshal(data, &dto)
	default:
		err = yaml.Unmarshal(data, &dto)
		if err != nil {
			err = json.Unmarshal(data, &dto)
		}
	}
	if err != nil {
		return "", fmt.Errorf("decoding project name from %s: %w", filePath, err)
	}
	if dto.ProjectName == nil || strings.TrimSpace(*dto.ProjectName) == "" {
		return "", errors.New("configured project_name is required before updating capture state")
	}
	return mergeConfig(dto).ProjectName, nil
}
