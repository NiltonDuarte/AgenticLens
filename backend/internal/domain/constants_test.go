package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestFormatTokensK(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{input: 0, expected: "0"},
		{input: 500, expected: "500"},
		{input: 999, expected: "999"},
		{input: 1000, expected: "1k"},
		{input: 1500, expected: "1.5k"},
		{input: 89337, expected: "89.3k"},
		{input: 200000, expected: "200k"},
		{input: 1000000, expected: "1M"},
		{input: 1200000, expected: "1.2M"},
		{input: 2550000, expected: "2.5M"},
	}

	for _, tt := range tests {
		got := domain.FormatTokensK(tt.input)
		if got != tt.expected {
			t.Errorf("FormatTokensK(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetBaseHealthStatus(t *testing.T) {
	ceiling := domain.HealthyCeilingTokens // 200,000

	tests := []struct {
		name     string
		toks     int64
		ceiling  int64
		expected domain.ContextHealthStatus
	}{
		// Standard ceiling (200,000) boundaries
		{name: "zero tokens standard ceiling", toks: 0, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{name: "negative tokens standard ceiling", toks: -100, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{name: "sub-100k tokens standard ceiling", toks: 50000, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{name: "boundary just under elevated", toks: 99999, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{name: "boundary exact elevated threshold", toks: 100000, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{name: "boundary just under standard ceiling", toks: 199999, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{name: "boundary exact standard ceiling", toks: 200000, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{name: "boundary just above standard ceiling", toks: 200001, ceiling: ceiling, expected: domain.ContextHealthBloated},
		{name: "boundary exact double standard ceiling", toks: 400000, ceiling: ceiling, expected: domain.ContextHealthBloated},
		{name: "boundary just above double standard ceiling", toks: 400001, ceiling: ceiling, expected: domain.ContextHealthCritical},
		{name: "substantially above critical threshold", toks: 600000, ceiling: ceiling, expected: domain.ContextHealthCritical},

		// Unassigned or non-positive ceiling defaulting to HealthyCeilingTokens (200,000)
		{name: "zero ceiling defaults to 200k (healthy)", toks: 50000, ceiling: 0, expected: domain.ContextHealthHealthy},
		{name: "zero ceiling defaults to 200k (elevated)", toks: 150000, ceiling: 0, expected: domain.ContextHealthElevated},
		{name: "zero ceiling defaults to 200k (bloated)", toks: 250000, ceiling: 0, expected: domain.ContextHealthBloated},
		{name: "zero ceiling defaults to 200k (critical)", toks: 450000, ceiling: 0, expected: domain.ContextHealthCritical},
		{name: "negative ceiling defaults to 200k (healthy)", toks: 50000, ceiling: -1, expected: domain.ContextHealthHealthy},
		{name: "negative ceiling defaults to 200k (bloated)", toks: 200001, ceiling: -50000, expected: domain.ContextHealthBloated},

		// Custom lower ceiling (50,000)
		{name: "custom 50k ceiling under elevated", toks: 24999, ceiling: 50000, expected: domain.ContextHealthHealthy},
		{name: "custom 50k ceiling at elevated boundary", toks: 25000, ceiling: 50000, expected: domain.ContextHealthElevated},
		{name: "custom 50k ceiling at ceiling boundary", toks: 50000, ceiling: 50000, expected: domain.ContextHealthElevated},
		{name: "custom 50k ceiling bloated", toks: 50001, ceiling: 50000, expected: domain.ContextHealthBloated},
		{name: "custom 50k ceiling at 2x boundary", toks: 100000, ceiling: 50000, expected: domain.ContextHealthBloated},
		{name: "custom 50k ceiling critical", toks: 100001, ceiling: 50000, expected: domain.ContextHealthCritical},

		// Custom higher ceiling (500,000)
		{name: "custom 500k ceiling under elevated", toks: 249999, ceiling: 500000, expected: domain.ContextHealthHealthy},
		{name: "custom 500k ceiling at elevated boundary", toks: 250000, ceiling: 500000, expected: domain.ContextHealthElevated},
		{name: "custom 500k ceiling at 500k boundary", toks: 500000, ceiling: 500000, expected: domain.ContextHealthElevated},
		{name: "custom 500k ceiling bloated", toks: 500001, ceiling: 500000, expected: domain.ContextHealthBloated},
		{name: "custom 500k ceiling critical", toks: 1000001, ceiling: 500000, expected: domain.ContextHealthCritical},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := domain.GetBaseHealthStatus(tt.toks, tt.ceiling)
			if got != tt.expected {
				t.Errorf("GetBaseHealthStatus(%d, %d) = %q, want %q", tt.toks, tt.ceiling, got, tt.expected)
			}
		})
	}
}

func TestGetHealthLabel(t *testing.T) {
	ceiling := domain.HealthyCeilingTokens

	tests := []struct {
		toks     int64
		ceiling  int64
		expected string
	}{
		{toks: 89337, ceiling: ceiling, expected: "HEALTHY"},
		{toks: 150000, ceiling: ceiling, expected: "ELEVATED"},
		{toks: 250000, ceiling: ceiling, expected: "BLOATED (125%)"},
		{toks: 500000, ceiling: ceiling, expected: "CRITICAL (250%)"},
	}

	for _, tt := range tests {
		got := domain.GetHealthLabel(tt.toks, tt.ceiling)
		if got != tt.expected {
			t.Errorf("GetHealthLabel(%d, %d) = %q, want %q", tt.toks, tt.ceiling, got, tt.expected)
		}
	}
}
