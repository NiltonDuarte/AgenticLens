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

func TestAuxiliaryExtractionUpgradeOnlyExpandsUnrepresentedFields(t *testing.T) {
	stamp := "2026-01-01T00:00:01Z"
	base := domain.SessionFact{ID: "owner", Source: domain.SourceCodex, User: "recorded", Invocations: []domain.InvocationFact{{Model: "recorded-model", CompletedAt: &stamp, CostUSD: 0, CostEvidence: domain.ObservedCost(0)}}}
	next := base
	next.AuxiliaryExtractionVersion = domain.CurrentAuxiliaryExtractionVersion
	next.Skills = []domain.SkillFact{{Name: "loaded", Evidence: "skill_file_read", Outcome: "unavailable"}}
	for _, tc := range []struct {
		name   string
		change func(*domain.SessionFact, *domain.SessionFact)
		want   bool
	}{
		{"source backed representation expansion", func(a, b *domain.SessionFact) {}, true},
		{"metric conflict", func(a, b *domain.SessionFact) {
			b.Invocations = append([]domain.InvocationFact{}, b.Invocations...)
			b.Invocations[0].Model = "changed"
		}, false},
		{"unknown versus observed zero cost", func(a, b *domain.SessionFact) {
			b.Invocations = append([]domain.InvocationFact{}, b.Invocations...)
			b.Invocations[0].CostEvidence = domain.UnavailableCost()
		}, false},
		{"different attribution", func(a, b *domain.SessionFact) { b.User = "other" }, false},
		{"different harness", func(a, b *domain.SessionFact) { b.Source = domain.SourceOpenCode }, false},
		{"unavailable source identity", func(a, b *domain.SessionFact) { a.Source = "" }, false},
		{"existing known occurrence", func(a, b *domain.SessionFact) { a.Skills = []domain.SkillFact{{Name: "accepted"}} }, false},
		{"current extractor cannot replace facts equally", func(a, b *domain.SessionFact) {
			a.AuxiliaryExtractionVersion = domain.CurrentAuxiliaryExtractionVersion
		}, false},
		{"accepted persisted events must survive", func(a, b *domain.SessionFact) { a.Events = []domain.SessionEventFact{{EventID: "recorded-event"}} }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := base, next
			tc.change(&a, &b)
			got, err := domain.CaptureAuxiliaryUpgrade(a, b)
			if err != nil || got != tc.want {
				t.Fatalf("upgrade=%v err=%v want=%v", got, err, tc.want)
			}
		})
	}
}
