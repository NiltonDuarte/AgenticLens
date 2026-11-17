package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
)

// resolveGenerationSummaryPath selects one destination before publication admission.
func resolveGenerationSummaryPath(project string, cfg domain.ProjectConfig, appendPath string, appendProvided bool, ledger, ticket string) (string, error) {
	if appendProvided {
		if strings.TrimSpace(appendPath) == "" {
			return "", fmt.Errorf("--append requires a nonblank summary path")
		}
		var err error
		appendPath, err = adaptersconfig.ExpandHomePath(appendPath)
		if err != nil {
			return "", err
		}
		appendPath, err = filesystem.PhysicalPath(appendPath)
		if err != nil {
			return "", err
		}
		if ledger != "" {
			a, err := filesystem.PhysicalPath(appendPath)
			if err != nil {
				return "", err
			}
			b, err := filesystem.PhysicalPath(ledger)
			if err != nil {
				return "", err
			}
			if a == b {
				return "", fmt.Errorf("summary destination collides with the ledger")
			}
		}
		return validateEffectiveSummaryPath(appendPath, ledger)
	}
	if err := adaptersconfig.ValidateSummaryPolicy(cfg); err != nil {
		return "", err
	}
	if !cfg.Generation.Summary.Enabled {
		return "", nil
	}
	name := cfg.Generation.Summary.Filename
	if ledger != "" {
		if strings.EqualFold(filepath.Base(ledger), name) {
			return "", fmt.Errorf("generation.summary.filename collides with ledger %s", ledger)
		}
		physical, err := filesystem.PhysicalPath(ledger)
		if err != nil {
			return "", err
		}
		return validateEffectiveSummaryPath(filepath.Join(filepath.Dir(physical), name), ledger)
	}
	if strings.TrimSpace(ticket) == "" {
		return "", fmt.Errorf("configured summary output requires --ticket-id when --json-output is absent")
	}
	if ticket == "." || ticket == ".." || strings.ContainsAny(ticket, "/\\\x00") {
		return "", fmt.Errorf("ticket ID is unsafe for configured summary workspace")
	}
	location, err := adapters.ResolveTicketLocation(cfg, ticket, "")
	if err != nil {
		return "", err
	}
	return validateEffectiveSummaryPath(filepath.Join(location.Path, name), ledger)
}

func validateEffectiveSummaryPath(path, ledger string) (string, error) {
	physical, err := filesystem.PhysicalPath(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(physical)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("selected summary destination %s: %w", physical, err)
	}
	if err == nil && !info.Mode().IsRegular() {
		return "", fmt.Errorf("selected summary destination %s must be a regular file", physical)
	}
	if ledger != "" {
		target, pathErr := filesystem.PhysicalPath(ledger)
		if pathErr != nil {
			return "", pathErr
		}
		if physical == target {
			return "", fmt.Errorf("summary destination collides with the ledger")
		}
		ledgerInfo, ledgerErr := os.Stat(target)
		if err == nil && ledgerErr == nil && os.SameFile(info, ledgerInfo) {
			return "", fmt.Errorf("summary destination aliases the ledger")
		}
	}
	return physical, nil
}
