package domain

import (
	"fmt"
	"strings"
)

// CaptureTicketMetadata returns the agreed display hint, if any, and rejects
// conflicting ticket types independently of row order.
func CaptureTicketMetadata(rows []CaptureRegistration, ticket string) (string, string, error) {
	titles := map[string]bool{}
	kinds := map[string]bool{}
	var title, kind string
	for _, r := range rows {
		if r.TicketID != ticket {
			continue
		}
		if hint := strings.TrimSpace(r.TicketTitle); hint != "" {
			titles[hint] = true
			title = hint
		}
		if r.TicketType != "" {
			kinds[r.TicketType] = true
			kind = r.TicketType
		}
	}
	if len(kinds) > 1 {
		return "", "", fmt.Errorf("ticket %s has conflicting explicit types; rebind conflicting registrations with one explicit type", ticket)
	}
	if len(titles) > 1 {
		title = ""
	}
	return title, kind, nil
}
