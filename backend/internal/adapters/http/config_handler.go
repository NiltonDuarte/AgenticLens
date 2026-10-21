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

		var sanitizedPrefixes []domain.TicketPrefix
		if len(cfg.Ticket.Prefixes) == 0 {
			sanitizedPrefixes = make([]domain.TicketPrefix, 0)
		} else {
			sanitizedPrefixes = make([]domain.TicketPrefix, len(cfg.Ticket.Prefixes))
			copy(sanitizedPrefixes, cfg.Ticket.Prefixes)
		}

		dir := strings.TrimSpace(workspacesDir)
		if dir == "" {
			dir = strings.TrimSpace(cfg.Workspaces.Dir)
		}

		var sanitizedRoleOrder []string
		if len(cfg.RoleOrder) == 0 {
			sanitizedRoleOrder = make([]string, 0)
		} else {
			sanitizedRoleOrder = make([]string, len(cfg.RoleOrder))
			copy(sanitizedRoleOrder, cfg.RoleOrder)
		}

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
