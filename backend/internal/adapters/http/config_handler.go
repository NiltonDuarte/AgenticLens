package http

import (
	"net/http"
	"strings"

	"agenticlens/internal/domain"
)

// ConfigHandler returns an http.HandlerFunc that responds with sanitized project configuration metadata.
func ConfigHandler(cfg domain.ProjectConfig, workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectName := strings.TrimSpace(cfg.ProjectName)
		if projectName == "" {
			projectName = "AgenticLens"
		}

		prefixes := cfg.Ticket.Prefixes
		if len(prefixes) == 0 {
			prefixes = domain.DefaultProjectConfig().Ticket.Prefixes
		}
		// Ensure non-nil slice
		sanitizedPrefixes := make([]domain.TicketPrefix, len(prefixes))
		copy(sanitizedPrefixes, prefixes)

		dir := strings.TrimSpace(workspacesDir)
		if dir == "" {
			dir = strings.TrimSpace(cfg.Workspaces.Dir)
		}
		if dir == "" {
			dir = "workspaces"
		}

		roleOrder := cfg.RoleOrder
		if len(roleOrder) == 0 {
			roleOrder = domain.DefaultProjectConfig().RoleOrder
		}
		sanitizedRoleOrder := make([]string, len(roleOrder))
		copy(sanitizedRoleOrder, roleOrder)

		version := domain.GetBuildInfo().Version

		dto := ProjectConfigDTO{
			ProjectName:   projectName,
			Prefixes:      sanitizedPrefixes,
			WorkspacesDir: dir,
			RoleOrder:     sanitizedRoleOrder,
			Version:       version,
		}

		WriteProjectConfigResponse(w, dto)
	}
}
