package main

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestResolveCodexSessionUsersForGeneration(t *testing.T) {
	tests := []struct {
		name       string
		freshUser  string
		cachedUser string
		gitUser    string
		want       string
	}{
		{name: "unknown cached identity falls through to Git", cachedUser: " UnKnOwN ", gitUser: "Git User", want: "Git User"},
		{name: "known same-session cache takes precedence over Git", cachedUser: " Cached User ", gitUser: "Git User", want: "Cached User"},
		{name: "fresh known identity takes precedence over cache and Git", freshUser: " Fresh User ", cachedUser: "Cached User", gitUser: "Git User", want: "Fresh User"},
		{name: "unknown final fallback", cachedUser: "unknown", gitUser: " UNKNOWN ", want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fresh := []domain.SessionFact{{ID: "session-1", User: tt.freshUser}}
			cachedUsers := map[string]string{"session-1": tt.cachedUser}

			resolveCodexSessionUsers(fresh, cachedUsers, tt.gitUser)

			if fresh[0].User != tt.want {
				t.Fatalf("resolved generation user = %q, want %q", fresh[0].User, tt.want)
			}
		})
	}
}
