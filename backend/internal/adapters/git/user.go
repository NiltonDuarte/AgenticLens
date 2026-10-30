package git

import "strings"

// IsKnownUser reports whether value contains an identity rather than the
// placeholder used when no identity is available.
func IsKnownUser(value string) bool {
	clean := strings.TrimSpace(value)
	return clean != "" && !strings.EqualFold(clean, "unknown")
}

// ResolveSessionUser selects the most specific available identity for a
// session, falling back to the configured Git user when necessary.
func ResolveSessionUser(extracted, existing, gitUser string) string {
	if IsKnownUser(extracted) {
		return strings.TrimSpace(extracted)
	}
	if IsKnownUser(existing) {
		return strings.TrimSpace(existing)
	}
	if IsKnownUser(gitUser) {
		return strings.TrimSpace(gitUser)
	}
	return "unknown"
}
