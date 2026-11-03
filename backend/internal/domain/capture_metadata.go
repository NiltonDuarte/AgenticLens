package domain

import "fmt"

// CaptureTicketMetadata rejects conflicting explicit context independently of row
// order. Rebinding the conflicting registrations explicitly resolves a conflict.
func CaptureTicketMetadata(rows []CaptureRegistration, ticket string) (string, string, error) {
	titles := map[string]bool{}
	kinds := map[string]bool{}
	var title, kind string
	for _, r := range rows {
		if r.TicketID != ticket {
			continue
		}
		if r.TicketTitle != "" {
			titles[r.TicketTitle] = true
			title = r.TicketTitle
		}
		if r.TicketType != "" {
			kinds[r.TicketType] = true
			kind = r.TicketType
		}
	}
	if len(titles) > 1 {
		return "", "", fmt.Errorf("ticket %s has conflicting explicit titles; rebind conflicting registrations with one explicit title", ticket)
	}
	if len(kinds) > 1 {
		return "", "", fmt.Errorf("ticket %s has conflicting explicit types; rebind conflicting registrations with one explicit type", ticket)
	}
	return title, kind, nil
}
