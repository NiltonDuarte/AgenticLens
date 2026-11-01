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

func TestCaptureSnapshotPrecedenceRejectsTruncationAndOlderSource(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		previous, next         int64
		previousTime, nextTime string
		accept                 bool
	}{
		{"first", 0, 100, "", "2026-01-01T00:00:01Z", true},
		{"unchanged", 100, 100, "2026-01-01T00:00:01Z", "2026-01-01T00:00:01Z", true},
		{"growing resumed source", 100, 200, "2026-01-01T00:00:01Z", "2026-01-01T00:00:02Z", true},
		{"truncated later source", 200, 100, "2026-01-01T00:00:01Z", "2026-01-01T00:00:02Z", false},
		{"larger but older source", 100, 200, "2026-01-01T00:00:02Z", "2026-01-01T00:00:01Z", false},
		{"fractional later source", 100, 200, "2026-01-01T00:00:01Z", "2026-01-01T00:00:01.1Z", true},
		{"offset equivalent source", 100, 200, "2026-01-01T00:00:01Z", "2025-12-31T19:00:01-05:00", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := domain.CaptureSnapshotAccept(tc.previous, tc.next, tc.previousTime, tc.nextTime); got != tc.accept {
				t.Fatalf("accept=%v, want %v", got, tc.accept)
			}
		})
	}
}
