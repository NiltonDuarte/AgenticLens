package backfill

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func readMaintenanceLedger(owner io.Closer, path, requested string, ignoreCase bool) ([]byte, string, error) {
	var body []byte
	var err error
	if directory, ok := owner.(*capture.ReadDirectory); ok {
		if _, e := directory.Lstat(filepath.Base(path) + ".publication-pending"); e == nil {
			return nil, "", adapters.ErrPublicationPending
		} else if !errors.Is(e, os.ErrNotExist) {
			return nil, "", e
		}
		body, err = directory.ReadFile(filepath.Base(path))
	} else {
		body, err = capture.ReadPublicationFile(path)
	}
	if err != nil {
		return nil, "", err
	}
	identity, err := adapters.RecordedLedgerTicket(body)
	if err != nil {
		return nil, "", err
	}
	if requested != "" && !ticketMatches(identity, strings.TrimSpace(requested), ignoreCase) {
		return nil, "", fmt.Errorf("ledger identity changed while acquiring ownership; resolve and retry")
	}
	return body, identity, nil
}

func ticketMatches(recorded, selected string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.EqualFold(recorded, selected)
	}
	return recorded == selected
}

// Resolve every destination before publication admission or locking. Public
// callers provide the selected project's policy; internal ledger-only helpers
// do not need source or summary policy they never consume.
func backfillPublicationPaths(metricPath string, summary bool, cfg domain.ProjectConfig) (string, string, error) {
	ledger, err := capture.ProjectDirectory(metricPath)
	if err != nil {
		return "", "", err
	}
	if cfg.ConfigPath != "" {
		if err := adapters.ValidateLedgerDestinationScope(cfg, ledger); err != nil {
			return "", "", err
		}
	}
	if !summary {
		return ledger, "", nil
	}
	if err := cfg.RequireSettings("generation.summary.filename"); err != nil {
		return "", "", err
	}
	if err := config.ValidateSummaryFilename(cfg.Generation.Summary.Filename); err != nil {
		return "", "", err
	}
	if err := config.ValidateContextPolicy(cfg); err != nil {
		return "", "", err
	}
	destination, err := capture.ProjectDirectory(filepath.Join(filepath.Dir(ledger), cfg.Generation.Summary.Filename))
	if err != nil {
		return "", "", err
	}
	if cfg.ConfigPath != "" {
		if err := adapters.ValidateWorkspaceArtifactDestination(cfg, ledger, destination); err != nil {
			return "", "", err
		}
	}
	if destination == ledger || destination == ledger+".publication-pending" || strings.HasPrefix(destination, ledger+string(filepath.Separator)) || strings.HasPrefix(ledger, destination+string(filepath.Separator)) {
		return "", "", fmt.Errorf("configured summary destination overlaps the ledger or its publication journal")
	}
	info, err := os.Stat(destination)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	}
	if info != nil {
		if !info.Mode().IsRegular() {
			return "", "", fmt.Errorf("configured summary destination %s must be a regular file", destination)
		}
		ledgerInfo, err := os.Stat(ledger)
		if err != nil {
			return "", "", err
		}
		if os.SameFile(info, ledgerInfo) {
			return "", "", fmt.Errorf("configured summary destination aliases the ledger")
		}
	}
	return ledger, destination, nil
}

// Reuse the existing publication protocol for a ledger plus configured summary.
// A failed second write restores the exact prior authorized members; recovery
// failures retain the journal rather than conceal a partially published set.
func publishBackfillArtifacts(metricPath string, data []byte, summaryPath, table string) (err error) {
	if summaryPath == "" {
		return publishBackfillLedger(metricPath, data)
	}
	paths := []string{metricPath, summaryPath}
	if err := adapters.BeginPublication(metricPath, paths); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_, recoveryErr := adapters.RecoverPublication(metricPath, paths)
			err = errors.Join(err, recoveryErr)
		}
	}()
	if err := publishBackfillLedger(metricPath, data); err != nil {
		return err
	}
	if err := adapters.UpdateSummaryFile(summaryPath, table); err != nil {
		return fmt.Errorf("updating configured summary %s: %w", summaryPath, err)
	}
	return adapters.CommitPublication(metricPath, paths)
}

func publicationProject(cfg domain.ProjectConfig, ledger string) string {
	if cfg.ConfigPath != "" {
		return filepath.Dir(cfg.ConfigPath)
	}
	return filepath.Dir(ledger)
}
