package domain_test

import (
	"agenticlens/internal/domain"
	"testing"
)

func TestDurationCoverageAndRounding(t *testing.T) {
	total, e := domain.MeasuredDuration(1600, 1600, true, "2026-01-01T00:00:03.2Z", "fixture:interval", "")
	if total != 3 || e.CompletedSeconds != 2 || e.ProvisionalSeconds != 1 {
		t.Fatalf("components must reconcile after invocation rounding: %d %+v", total, e)
	}
	if err := domain.ValidateDuration(total, e); err != nil {
		t.Fatal(err)
	}
	e.ProvisionalSeconds = 2
	if err := domain.ValidateDuration(total, e); err == nil {
		t.Fatal("mismatched components accepted")
	}
	summary := domain.MergeDuration(domain.DurationForInvocation(domain.InvocationFact{ActiveDurationSeconds: 7}), domain.DurationForInvocation(domain.InvocationFact{DurationEvidence: &domain.DurationEvidence{State: "provisional", ProvisionalSeconds: 10, Cutoff: "2026-01-01T00:00:10Z"}}))
	if summary.UnknownInvocations != 1 || summary.ProvisionalInvocations != 1 || summary.ProvisionalSeconds != 10 || len(summary.Cutoffs) != 1 {
		t.Fatalf("lost unknown/provisional coverage: %+v", summary)
	}
}

func TestDurationIntervalRequiresSupportedBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		start                  int64
		end                    *int64
		active                 bool
		completed, provisional int64
		limited                bool
	}{
		{name: "unsupported_missing_completion", start: 1000, limited: true},
		{name: "running", start: 1000, active: true, provisional: 9000},
		{name: "future_start", start: 11000, active: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, p, _, limit := domain.DurationInterval(tc.start, tc.end, 10000, tc.active)
			if c != tc.completed || p != tc.provisional || (limit != "") != tc.limited {
				t.Fatalf("interval %d %d %s", c, p, limit)
			}
		})
	}
	end := int64(20000)
	c, p, active, lim := domain.DurationInterval(1000, &end, 10000, false)
	if c != 0 || p != 9000 || !active || lim != "" {
		t.Fatalf("future completion leaked: %d %d %t %s", c, p, active, lim)
	}
}
