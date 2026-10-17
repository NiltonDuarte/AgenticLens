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
		toks     int64
		ceiling  int64
		expected domain.ContextHealthStatus
	}{
		{toks: 0, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{toks: 50000, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{toks: 99999, ceiling: ceiling, expected: domain.ContextHealthHealthy},
		{toks: 100000, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{toks: 199999, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{toks: 200000, ceiling: ceiling, expected: domain.ContextHealthElevated},
		{toks: 200001, ceiling: ceiling, expected: domain.ContextHealthBloated},
		{toks: 400000, ceiling: ceiling, expected: domain.ContextHealthBloated},
		{toks: 400001, ceiling: ceiling, expected: domain.ContextHealthCritical},
		{toks: 600000, ceiling: ceiling, expected: domain.ContextHealthCritical},
	}

	for _, tt := range tests {
		got := domain.GetBaseHealthStatus(tt.toks, tt.ceiling)
		if got != tt.expected {
			t.Errorf("GetBaseHealthStatus(%d, %d) = %q, want %q", tt.toks, tt.ceiling, got, tt.expected)
		}
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
