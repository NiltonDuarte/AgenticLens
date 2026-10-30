package backfill

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestResolveCodexSessionUsersForBackfill(t *testing.T) {
	tests := []struct {
		name       string
		freshUser  string
		storedUser string
		gitUser    string
		want       string
	}{
		{name: "unknown stored identity falls through to Git", storedUser: " UNKNOWN ", gitUser: "Git User", want: "Git User"},
		{name: "known same-session identity takes precedence over Git", storedUser: " Stored User ", gitUser: "Git User", want: "Stored User"},
		{name: "fresh known identity takes precedence over stored and Git", freshUser: " Fresh User ", storedUser: "Stored User", gitUser: "Git User", want: "Fresh User"},
		{name: "unknown final fallback", storedUser: "unknown", gitUser: " unknown ", want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fresh := []domain.SessionFact{{ID: "session-1", User: tt.freshUser}}
			stored := []domain.SessionFact{{ID: "session-1", User: tt.storedUser}}

			resolveCodexSessionUsers(fresh, stored, tt.gitUser)

			if fresh[0].User != tt.want {
				t.Fatalf("resolved backfill user = %q, want %q", fresh[0].User, tt.want)
			}
		})
	}
}
