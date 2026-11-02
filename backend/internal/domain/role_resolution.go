package domain

import "strings"

// PersonaEvidence accumulates explicit assistant persona names without choosing
// between conflicting source records. Its zero value represents no evidence.
type PersonaEvidence struct {
	Role        AgentRole
	Conflicting bool
}

// WithCandidate returns new evidence, retaining disagreement even if later
// records repeat the first persona. Blank candidates provide no evidence.
func (e PersonaEvidence) WithCandidate(raw string) PersonaEvidence {
	role := AgentRole(strings.TrimSpace(raw))
	if role == "" || e.Conflicting {
		return e
	}
	if e.Role != "" && e.Role != role {
		return PersonaEvidence{Conflicting: true}
	}
	return PersonaEvidence{Role: role}
}

// Resolve preserves explicit session metadata before considering unambiguous
// assistant evidence. Missing or conflicting evidence remains unknown.
func (e PersonaEvidence) Resolve(sessionPersona string) AgentRole {
	if role := AgentRole(strings.TrimSpace(sessionPersona)); role != "" {
		return role
	}
	if e.Conflicting {
		return ""
	}
	return e.Role
}
