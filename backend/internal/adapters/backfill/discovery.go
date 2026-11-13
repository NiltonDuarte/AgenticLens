package backfill

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// discoverWorkspaceLedgers uses shared maintenance resolution for targeted work,
// while unfiltered batches retain bounded partial results and diagnostics.
func discoverWorkspaceLedgers(cfg domain.ProjectConfig, ticket, selection string) ([]string, []string, error) {
	locations := []adapters.WorkspaceLocation{}
	issues := []domain.WorkspaceDiagnostic{}
	if strings.TrimSpace(ticket) != "" || selection != "" {
		location, err := adapters.ResolveMaintenanceLocation(cfg, strings.TrimSpace(ticket), selection)
		if err != nil {
			var resolution *adapters.WorkspaceResolutionError
			if selection == "" && errors.As(err, &resolution) && resolution.Outcome == "not_found" {
				return []string{}, []string{}, nil
			}
			return nil, nil, err
		}
		locations = append(locations, location)
		issues = location.Diagnostics
	} else {
		locations, issues = adapters.DiscoverWorkspaceLocations(cfg)
	}
	paths := []string{}
	diagnostics := []string{}
	for _, location := range locations {
		paths = append(paths, location.LedgerPath)
	}
	for _, issue := range issues {
		diagnostics = append(diagnostics, fmt.Sprintf("%s: %s", issue.Path, issue.Message))
	}
	return paths, diagnostics, nil
}
func recordedTicket(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	body, readErr := io.ReadAll(io.LimitReader(f, 128<<20+1))
	closeErr := f.Close()
	if readErr != nil {
		return "", readErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if len(body) > 128<<20 {
		return "", fmt.Errorf("ledger exceeds read bound")
	}
	return adapters.RecordedLedgerTicket(body)
}
