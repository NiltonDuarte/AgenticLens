package domain_test

import (
	"agenticlens/internal/domain"
	"strings"
	"testing"
)

func TestDurationCoverageAndRounding(t *testing.T) {
	for _, tc := range []struct{ completed, provisional, total, portion int64 }{
		{1600, 1600, 3, 1}, {400, 400, 1, 1}, {600, 600, 1, 0}, {0, 500, 1, 1}, {1500, 0, 2, 0}, {0, 0, 0, 0},
	} {
		total, provisional := domain.MeasuredDuration(tc.completed, tc.provisional)
		if total != tc.total || provisional != tc.portion {
			t.Fatalf("rounding (%d,%d): got (%d,%d), want (%d,%d)", tc.completed, tc.provisional, total, provisional, tc.total, tc.portion)
		}
		if err := domain.ValidateDuration(total, provisional); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		total, provisional int64
		valid              bool
	}{{0, 0, true}, {10, 0, true}, {10, 10, true}, {10, 11, false}, {-1, 0, false}, {10, -1, false}, {9007199254740991, 9007199254740991, true}, {9007199254740992, 1, false}} {
		if err := domain.ValidateDuration(tc.total, tc.provisional); (err == nil) != tc.valid {
			t.Fatalf("validation %+v: %v", tc, err)
		}
	}
	for _, provisional := range []int64{0, 7} {
		description := domain.DurationDescription(provisional)
		if description == "" || strings.Contains(description, "complete") || strings.Contains(description, "verified") {
			t.Fatal("missing neutral numeric description")
		}
	}
}

func TestDurationIntervalRequiresSupportedBoundaries(t *testing.T) {
	reversed, future := int64(500), int64(20000)
	for _, tc := range []struct {
		name                   string
		start                  int64
		end                    *int64
		active                 bool
		completed, provisional int64
		invalid                bool
	}{
		{name: "unsupported_missing_completion", start: 1000},
		{name: "running", start: 1000, active: true, provisional: 9000},
		{name: "future_start", start: 11000, active: true},
		{name: "invalid_start", active: true, invalid: true},
		{name: "reversed_completion", start: 1000, end: &reversed, invalid: true},
		{name: "future_completion", start: 1000, end: &future, provisional: 9000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, p, active, err := domain.DurationInterval(tc.start, tc.end, 10000, tc.active)
			if c != tc.completed || p != tc.provisional || active != (p > 0) || (err != nil) != tc.invalid {
				t.Fatalf("interval %d %d %t %v", c, p, active, err)
			}
		})
	}
}
