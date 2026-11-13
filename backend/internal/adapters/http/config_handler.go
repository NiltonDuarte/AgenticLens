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

		var sanitizedRoleOrder []string
		if len(cfg.RoleOrder) == 0 {
			sanitizedRoleOrder = make([]string, 0)
		} else {
			sanitizedRoleOrder = make([]string, len(cfg.RoleOrder))
			copy(sanitizedRoleOrder, cfg.RoleOrder)
		}

		version := domain.GetBuildInfo().Version

		roots := make([]domain.WorkspaceRootConfig, len(cfg.Workspaces.Roots))
		for i, root := range cfg.Workspaces.Roots {
			roots[i] = domain.WorkspaceRootConfig{Path: root.Path, Discovery: root.Discovery}
		}

		dto := ProjectConfigDTO{
			ProjectName:     projectName,
			Prefixes:        sanitizedPrefixes,
			CaseInsensitive: cfg.Ticket.CaseInsensitive,
			RoleMappings:    cfg.RoleMappings,
			WorkspacesRoots: roots,
			RoleOrder:       sanitizedRoleOrder,
			Version:         version,
		}

		// Unused live policy may be absent for persisted-ledger views. Preserve
		// an explicitly supplied zero/negative value for validation at use.
		if cfg.Settings == nil || cfg.Settings["live.poll_interval_ms"] {
			interval := cfg.Live.PollIntervalMs
			dto.LivePollIntervalMS = &interval
		}

		WriteProjectConfigResponse(w, dto)
	}
}
