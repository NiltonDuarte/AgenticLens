package backfill

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"errors"
	"fmt"
	"strings"
)

// discoverWorkspaceLedgers uses shared maintenance resolution for targeted work,
// while unfiltered batches retain bounded partial results and diagnostics.
func discoverWorkspaceLedgers(cfg domain.ProjectConfig, ticket, selection string) ([]adapters.WorkspaceLocation, []string, error) {
	locations := []adapters.WorkspaceLocation{}
	issues := []domain.WorkspaceDiagnostic{}
	if strings.TrimSpace(ticket) != "" || selection != "" {
		location, err := adapters.ResolveMaintenancePublication(cfg, strings.TrimSpace(ticket), selection)
		if err != nil {
			var resolution *adapters.WorkspaceResolutionError
			if selection == "" && errors.As(err, &resolution) && resolution.Outcome == "not_found" {
				return nil, nil, nil
			}
			return nil, nil, err
		}
		locations = append(locations, location)
		issues = location.Diagnostics
	} else {
		locations, issues = adapters.DiscoverWorkspacePublicationLocations(cfg)
	}
	diagnostics := []string{}
	for _, issue := range issues {
		diagnostics = append(diagnostics, fmt.Sprintf("%s: %s", issue.Path, issue.Message))
	}
	return locations, diagnostics, nil
}

func closeWorkspaceSelections(locations []adapters.WorkspaceLocation, diagnostics *[]string) {
	for i := range locations {
		if err := locations[i].Close(); err != nil {
			*diagnostics = append(*diagnostics, fmt.Sprintf("Workspace %s: closing directory selection: %v", locations[i].Path, err))
		}
	}
}
