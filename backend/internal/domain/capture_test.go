package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestCaptureBindingPreservesExplicitSessionOwnership(t *testing.T) {
	for _, tc := range []struct {
		name, existing, requested, inherited, want string
		fail                                       bool
	}{
		{"unknown stays unassigned", "", "", "", "", false},
		{"explicit response", "", "TEST-101", "", "TEST-101", false},
		{"proven parent", "", "", "TEST-101", "TEST-101", false},
		{"duplicate response", "TEST-101", "TEST-101", "", "TEST-101", false},
		{"no response preserves binding", "TEST-101", "", "", "TEST-101", false},
		{"conflicting response", "TEST-101", "TEST-202", "", "", true},
		{"conflicting proven parent", "TEST-101", "", "TEST-202", "", true},
		{"unsafe ticket", "", "TEST-1;touch bad", "", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := domain.CaptureBinding(tc.existing, tc.requested, tc.inherited)
			if (err != nil) != tc.fail || got != tc.want {
				t.Fatalf("binding=%q err=%v; want %q failure=%v", got, err, tc.want, tc.fail)
			}
		})
	}
}
