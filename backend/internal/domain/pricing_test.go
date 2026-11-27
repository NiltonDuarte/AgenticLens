package domain_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func assertCostPtrEqual(t *testing.T, label string, got *float64, want *float64) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatalf("%s: got %v, want nil", label, *got)
		}
		return
	}
	if got == nil {
		t.Fatalf("%s: got nil, want %v", label, *want)
	}
	if math.Abs(*got-*want) > 1e-9 {
		t.Fatalf("%s: got %.12f, want %.12f", label, *got, *want)
	}
}

func TestValidateModelPricingConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     domain.ModelPricingConfig
		wantErr string
	}{
		{
			name: "empty config is valid",
			cfg:  domain.ModelPricingConfig{},
		},
		{
			name: "explicit empty slice is valid",
			cfg:  domain.ModelPricingConfig{Rates: []domain.ModelPricingRate{}},
		},
		{
			name: "valid exact and glob rates including zero",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
					{Model: "*gemini-3.8*", InputPerMillion: 0.0, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
					{Model: "claude-?-sonnet", InputPerMillion: 3.0, CacheReadPerMillion: 0.3, OutputPerMillion: 15.0},
				},
			},
		},
		{
			name: "blank model rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "   ", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "must be nonblank",
		},
		{
			name: "negative input rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: -0.01, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "input_per_million",
		},
		{
			name: "negative cache_read rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: 1.0, CacheReadPerMillion: -0.5, OutputPerMillion: 2.0},
				},
			},
			wantErr: "cache_read_per_million",
		},
		{
			name: "negative output rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: -2.0},
				},
			},
			wantErr: "output_per_million",
		},
		{
			name: "NaN rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: math.NaN(), CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "finite non-negative",
		},
		{
			name: "positive Inf rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: 1.0, CacheReadPerMillion: math.Inf(1), OutputPerMillion: 2.0},
				},
			},
			wantErr: "finite non-negative",
		},
		{
			name: "negative Inf rate rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5.4", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: math.Inf(-1)},
				},
			},
			wantErr: "finite non-negative",
		},
		{
			name: "bracket glob syntax rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-[5]", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "invalid glob syntax",
		},
		{
			name: "backslash glob syntax rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: `gpt-\5`, InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "invalid glob syntax",
		},
		{
			name: "control char in model rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "gpt-5\x00", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},
				},
			},
			wantErr: "control characters",
		},
		{
			name: "duplicate normalized model rejected",
			cfg: domain.ModelPricingConfig{
				Rates: []domain.ModelPricingRate{
					{Model: "GPT-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
					{Model: "  gpt-5.4  ", InputPerMillion: 3.0, CacheReadPerMillion: 0.5, OutputPerMillion: 10.0},
				},
			},
			wantErr: "duplicates normalized rule",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateModelPricingConfig(tc.cfg)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateModelPricingConfig unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateModelPricingConfig error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestEffectiveCatalogIDDeterministicAndSensitive(t *testing.T) {
	if got := (domain.ModelPricingConfig{}).EffectiveCatalogID(); got != "" {
		t.Fatalf("empty catalog EffectiveCatalogID = %q, want empty string", got)
	}

	negZero := math.Copysign(0, -1)
	cfgA := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "GPT-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			{Model: "free-model", InputPerMillion: 0.0, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
		},
	}
	cfgReordered := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "  free-model ", InputPerMillion: negZero, CacheReadPerMillion: negZero, OutputPerMillion: negZero},
			{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
		},
	}

	idA := cfgA.EffectiveCatalogID()
	idReordered := cfgReordered.EffectiveCatalogID()
	if !strings.HasPrefix(idA, "sha256:") || len(idA) != len("sha256:")+64 {
		t.Fatalf("EffectiveCatalogID format invalid: %q", idA)
	}
	if idA != idReordered {
		t.Fatalf("EffectiveCatalogID differs across reorder/-0.0 normalization: %q vs %q", idA, idReordered)
	}

	cfgRateEdit := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "gpt-5.4", InputPerMillion: 2.01, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			{Model: "free-model", InputPerMillion: 0.0, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
		},
	}
	if idEdit := cfgRateEdit.EffectiveCatalogID(); idEdit == idA {
		t.Fatalf("EffectiveCatalogID did not change on rate edit: %q", idEdit)
	}

	cfgModelEdit := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "gpt-5.5", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			{Model: "free-model", InputPerMillion: 0.0, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
		},
	}
	if idModelEdit := cfgModelEdit.EffectiveCatalogID(); idModelEdit == idA {
		t.Fatalf("EffectiveCatalogID did not change on model edit: %q", idModelEdit)
	}
}

func TestMatchModelPricingRatePrecedenceAndAmbiguity(t *testing.T) {
	t.Run("exact beats glob with case and whitespace normalization", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5*", InputPerMillion: 10.0, CacheReadPerMillion: 1.0, OutputPerMillion: 20.0},
				{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			},
		}
		rate, matched, err := cfg.MatchModelPricingRate("  GPT-5.4  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !matched || rate.Model != "gpt-5.4" || rate.InputPerMillion != 2.0 {
			t.Fatalf("got (%+v, %v), want exact gpt-5.4 rule", rate, matched)
		}
	})

	t.Run("provider prefix matching via glob across slash", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "*gemini-3.8*", InputPerMillion: 0.5, CacheReadPerMillion: 0.05, OutputPerMillion: 1.5},
			},
		}
		rate, matched, err := cfg.MatchModelPricingRate("litellm/gemini-3.8-flash")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !matched || rate.Model != "*gemini-3.8*" {
			t.Fatalf("got (%+v, %v), want *gemini-3.8*", rate, matched)
		}
	})

	t.Run("higher literal count beats lower literal count", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "*-mini", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},  // 5 literals
				{Model: "gpt-5-*", InputPerMillion: 3.0, CacheReadPerMillion: 0.3, OutputPerMillion: 6.0}, // 6 literals
			},
		}
		rate, matched, err := cfg.MatchModelPricingRate("gpt-5-mini")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !matched || rate.Model != "gpt-5-*" {
			t.Fatalf("got (%+v, %v), want gpt-5-* (6 literals beats 5)", rate, matched)
		}
	})

	t.Run("longer normalized pattern beats shorter on literal tie", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5*", InputPerMillion: 1.0, CacheReadPerMillion: 0.1, OutputPerMillion: 2.0},   // 5 literals, len 6
				{Model: "*gpt-5*", InputPerMillion: 4.0, CacheReadPerMillion: 0.4, OutputPerMillion: 8.0},  // 5 literals, len 7
			},
		}
		rate, matched, err := cfg.MatchModelPricingRate("gpt-5-pro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !matched || rate.Model != "*gpt-5*" {
			t.Fatalf("got (%+v, %v), want *gpt-5* (len 7 beats len 6 on 5-literal tie)", rate, matched)
		}
	})

	t.Run("equal specificity glob tie fails closed with AmbiguousModelPricingError", func(t *testing.T) {
		// Both "gpt-5*" (5 literals, len 6) and "*-mini" (5 literals, len 6) match "gpt-5-mini".
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5*", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
				{Model: "*-mini", InputPerMillion: 0.5, CacheReadPerMillion: 0.05, OutputPerMillion: 1.5},
			},
		}
		_, matched, err := cfg.MatchModelPricingRate("gpt-5-mini")
		if matched {
			t.Fatalf("expected no match on ambiguity")
		}
		var ambigErr *domain.AmbiguousModelPricingError
		if !errors.As(err, &ambigErr) {
			t.Fatalf("expected AmbiguousModelPricingError, got %T: %v", err, err)
		}
		if ambigErr.Model != "gpt-5-mini" || len(ambigErr.Patterns) != 2 || ambigErr.Patterns[0] != "*-mini" || ambigErr.Patterns[1] != "gpt-5*" {
			t.Fatalf("unexpected ambiguity details: %+v", ambigErr)
		}
	})

	t.Run("unmatched model returns false without error", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			},
		}
		_, matched, err := cfg.MatchModelPricingRate("claude-sonnet-4")
		if err != nil || matched {
			t.Fatalf("got (matched=%v, err=%v), want (false, nil)", matched, err)
		}
	})

	t.Run("question mark glob matches exactly one rune", func(t *testing.T) {
		cfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5.?", InputPerMillion: 2.5, CacheReadPerMillion: 0.25, OutputPerMillion: 10.0},
			},
		}
		rate, matched, err := cfg.MatchModelPricingRate("gpt-5.4")
		if err != nil {
			t.Fatalf("unexpected error for gpt-5.4: %v", err)
		}
		if !matched || rate.Model != "gpt-5.?" || rate.InputPerMillion != 2.5 {
			t.Fatalf("gpt-5.4: got (%+v, %v), want gpt-5.?", rate, matched)
		}
		for _, nonMatch := range []string{"gpt-5.44", "gpt-5.", "gpt-5.4-mini"} {
			_, m, err := cfg.MatchModelPricingRate(nonMatch)
			if err != nil || m {
				t.Fatalf("%q: got (matched=%v, err=%v), want (false, nil)", nonMatch, m, err)
			}
		}
	})
}

func TestEstimateInvocationCostFormulaAndGuardrails(t *testing.T) {
	cfg := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			{Model: "zero-rate-model", InputPerMillion: 0.0, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
		},
	}
	wantCatalog := cfg.EffectiveCatalogID()

	t.Run("exact formula without double-charging cache_read or reasoning", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "GPT-5.4",
			Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000, Reasoning: 10_000},
			CostEvidence: domain.UnavailableCost(),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if est == nil {
			t.Fatalf("expected estimate, got nil")
		}
		// (200000*2.0 + 800000*0.25 + 50000*8.0)/1e6 = (400000 + 200000 + 400000)/1e6 = 1.00
		if math.Abs(est.AmountUSD-1.0) > 1e-9 {
			t.Fatalf("AmountUSD = %.9f, want 1.0", est.AmountUSD)
		}
		if est.Rate.MatchedPattern != "gpt-5.4" || est.Rate.EffectiveCatalogID != wantCatalog {
			t.Fatalf("unexpected rate snapshot: %+v", est.Rate)
		}
		if est.Rate.InputPerMillion != 2.0 || est.Rate.CacheReadPerMillion != 0.25 || est.Rate.OutputPerMillion != 8.0 {
			t.Fatalf("unexpected rate numbers: %+v", est.Rate)
		}
	})

	t.Run("observed zero spend is never overwritten or estimated", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "gpt-5.4",
			Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
			CostUSD:      0,
			CostEvidence: domain.ObservedCost(0),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if est != nil {
			t.Fatalf("observed $0.00 must not be estimated, got %+v", est)
		}
	})

	t.Run("positive observed spend is never overwritten or estimated", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "gpt-5.4",
			Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
			CostUSD:      3.25,
			CostEvidence: domain.ObservedCost(3.25),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if est != nil {
			t.Fatalf("positive observed spend must not be estimated, got %+v", est)
		}
	})

	t.Run("partial measured invocation gets no token estimate", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "gpt-5.4",
			Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
			CostEvidence: domain.PartialCost(2.0),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if est != nil {
			t.Fatalf("partial invocation must not be estimated, got %+v", est)
		}
	})

	t.Run("unmatched model stays unavailable", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "unknown-model",
			Tokens:       domain.InvocationTokens{Input: 1000, CacheRead: 500, Output: 200},
			CostEvidence: domain.UnavailableCost(),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil || est != nil {
			t.Fatalf("got (%+v, %v), want (nil, nil)", est, err)
		}
	})

	t.Run("negative token counts stay unavailable", func(t *testing.T) {
		for _, tokens := range []domain.InvocationTokens{
			{Input: -1, CacheRead: 100, Output: 100},
			{Input: 100, CacheRead: -1, Output: 100},
			{Input: 100, CacheRead: 100, Output: -1},
		} {
			inv := domain.InvocationFact{
				Role:         "coder",
				Model:        "gpt-5.4",
				Tokens:       tokens,
				CostEvidence: domain.UnavailableCost(),
			}
			est, err := domain.EstimateInvocationCost(inv, cfg)
			if err != nil || est != nil {
				t.Fatalf("negative tokens %+v: got (%+v, %v), want (nil, nil)", tokens, est, err)
			}
		}
	})

	t.Run("configured zero rate produces genuine $0.00 estimate distinct from nil", func(t *testing.T) {
		inv := domain.InvocationFact{
			Role:         "coder",
			Model:        "zero-rate-model",
			Tokens:       domain.InvocationTokens{Input: 100_000, CacheRead: 50_000, Output: 10_000},
			CostEvidence: domain.UnavailableCost(),
		}
		est, err := domain.EstimateInvocationCost(inv, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if est == nil {
			t.Fatalf("expected non-nil $0.00 estimate")
		}
		if est.AmountUSD != 0 {
			t.Fatalf("AmountUSD = %v, want 0", est.AmountUSD)
		}
		if est.Rate.MatchedPattern != "zero-rate-model" {
			t.Fatalf("MatchedPattern = %q, want zero-rate-model", est.Rate.MatchedPattern)
		}
	})

	t.Run("6-decimal rounding on sub-microdollar fractions", func(t *testing.T) {
		microCfg := domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "micro-down", InputPerMillion: 0.4, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
				{Model: "micro-up", InputPerMillion: 1.5, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
				{Model: "micro-mid", InputPerMillion: 2.4, CacheReadPerMillion: 0.0, OutputPerMillion: 0.0},
			},
		}
		cases := []struct {
			model string
			input int64
			want  float64
		}{
			// 1 * 0.4 / 1e6 = 0.0000004 -> rounds to 0.000000
			{model: "micro-down", input: 1, want: 0.0},
			// 1 * 1.5 / 1e6 = 0.0000015 -> rounds to 0.000002
			{model: "micro-up", input: 1, want: 0.000002},
			// 1 * 2.4 / 1e6 = 0.0000024 -> rounds to 0.000002
			{model: "micro-mid", input: 1, want: 0.000002},
			// 3 * 1.5 / 1e6 = 0.0000045 -> rounds to 0.000005
			{model: "micro-up", input: 3, want: 0.000005},
		}
		for _, tc := range cases {
			inv := domain.InvocationFact{
				Role:         "coder",
				Model:        tc.model,
				Tokens:       domain.InvocationTokens{Input: tc.input},
				CostEvidence: domain.UnavailableCost(),
			}
			est, err := domain.EstimateInvocationCost(inv, microCfg)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tc.model, err)
			}
			if est == nil {
				t.Fatalf("%s: expected non-nil estimate", tc.model)
			}
			if math.Abs(est.AmountUSD-tc.want) > 1e-12 {
				t.Fatalf("%s (input=%d): AmountUSD = %.12f, want %.12f", tc.model, tc.input, est.AmountUSD, tc.want)
			}
		}
	})
}

func TestApplyModelPricingToInvocationsEmptyRates(t *testing.T) {
	staleEstimate := &domain.CostEstimate{
		AmountUSD: 1.5,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "gpt-5.4",
			InputPerMillion:     2.0,
			CacheReadPerMillion: 0.25,
			OutputPerMillion:    8.0,
			EffectiveCatalogID:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	input := []domain.InvocationFact{
		{
			Role:         "coder",
			Model:        "gpt-5.4",
			Tokens:       domain.InvocationTokens{Input: 100_000, Output: 20_000},
			CostEvidence: domain.UnavailableCost(),
			CostEstimate: staleEstimate,
		},
		{
			Role:         "orchestrator",
			Model:        "claude-sonnet-4",
			Tokens:       domain.InvocationTokens{Input: 5000, Output: 1000},
			CostUSD:      2.50,
			CostEvidence: domain.ObservedCost(2.50),
		},
	}

	got, err := domain.ApplyModelPricingToInvocations(input, domain.ModelPricingConfig{Rates: []domain.ModelPricingRate{}})
	if err != nil {
		t.Fatalf("ApplyModelPricingToInvocations with empty rates failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d invocations, want 2", len(got))
	}
	if got[0].CostEstimate != nil {
		t.Fatalf("expected stale CostEstimate to be cleared with empty rates, got %+v", got[0].CostEstimate)
	}
	if got[1].CostUSD != 2.50 || got[1].CostEvidence.State != domain.CostObserved || got[1].CostEstimate != nil {
		t.Fatalf("observed invocation altered by empty rates: %+v", got[1])
	}
	// Original input slice must not be mutated.
	if input[0].CostEstimate != staleEstimate {
		t.Fatalf("ApplyModelPricingToInvocations mutated caller input slice")
	}

	// Nil slice stays nil without error.
	nilResult, err := domain.ApplyModelPricingToInvocations(nil, domain.ModelPricingConfig{Rates: []domain.ModelPricingRate{}})
	if err != nil || nilResult != nil {
		t.Fatalf("nil input: got (%v, %v), want (nil, nil)", nilResult, err)
	}
}

func TestValidateCostEstimateDirectCases(t *testing.T) {
	validRate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:1111111111111111111111111111111111111111111111111111111111111111",
	}

	if err := domain.ValidateCostEstimate(domain.ObservedCost(1.0), nil); err != nil {
		t.Fatalf("nil estimate on observed cost should pass: %v", err)
	}
	if err := domain.ValidateCostEstimate(domain.UnavailableCost(), &domain.CostEstimate{AmountUSD: 0, Rate: validRate}); err != nil {
		t.Fatalf("valid zero estimate on unavailable cost should pass: %v", err)
	}

	cases := []struct {
		name     string
		evidence domain.CostEvidence
		estimate *domain.CostEstimate
		wantErr  string
	}{
		{
			name:     "paired with observed evidence",
			evidence: domain.ObservedCost(2.5),
			estimate: &domain.CostEstimate{AmountUSD: 1.0, Rate: validRate},
			wantErr:  "cost_estimate is only allowed when cost_evidence.state is unavailable",
		},
		{
			name:     "paired with partial evidence",
			evidence: domain.PartialCost(1.25),
			estimate: &domain.CostEstimate{AmountUSD: 1.0, Rate: validRate},
			wantErr:  "cost_estimate is only allowed when cost_evidence.state is unavailable",
		},
		{
			name:     "negative amount_usd",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{AmountUSD: -0.01, Rate: validRate},
			wantErr:  "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:     "NaN amount_usd",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{AmountUSD: math.NaN(), Rate: validRate},
			wantErr:  "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:     "positive Inf amount_usd",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{AmountUSD: math.Inf(1), Rate: validRate},
			wantErr:  "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:     "negative Inf amount_usd",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{AmountUSD: math.Inf(-1), Rate: validRate},
			wantErr:  "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:     "blank matched_pattern",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "   ",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.matched_pattern must be nonblank",
		},
		{
			name:     "blank effective_catalog_id",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  "   ",
				},
			},
			wantErr: "cost_estimate.rate.effective_catalog_id must be nonblank",
		},
		{
			name:     "negative input_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     -1.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.input_per_million must be a finite non-negative number",
		},
		{
			name:     "NaN input_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     math.NaN(),
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.input_per_million must be a finite non-negative number",
		},
		{
			name:     "negative cache_read_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: -0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.cache_read_per_million must be a finite non-negative number",
		},
		{
			name:     "Inf cache_read_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: math.Inf(1),
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.cache_read_per_million must be a finite non-negative number",
		},
		{
			name:     "negative output_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    -8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.output_per_million must be a finite non-negative number",
		},
		{
			name:     "NaN output_per_million",
			evidence: domain.UnavailableCost(),
			estimate: &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    math.NaN(),
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			},
			wantErr: "cost_estimate.rate.output_per_million must be a finite non-negative number",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateCostEstimate(tc.evidence, tc.estimate)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateCostEstimate error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestCaptureOwnerFactsEqualIgnoresCostEstimateWithoutMutation(t *testing.T) {
	estimate := &domain.CostEstimate{
		AmountUSD: 3.81,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "gpt-5.4",
			InputPerMillion:     2.0,
			CacheReadPerMillion: 0.25,
			OutputPerMillion:    8.0,
			EffectiveCatalogID:  "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		},
	}
	ts := "2026-10-17T10:00:00Z"
	rawSnapshot := domain.SessionFact{
		ID:     "ses-capture-1",
		Source: domain.SourceCodex,
		User:   "capture-hook",
		Invocations: []domain.InvocationFact{
			{
				Role:         "coder",
				Model:        "gpt-5.4",
				StartedAt:    &ts,
				CompletedAt:  &ts,
				Turns:        2,
				Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: nil,
			},
		},
		Tools:     []domain.SessionToolFact{{Name: "bash", Calls: 3, Errors: 0}},
		Anomalies: []domain.Anomaly{},
	}
	pricedOwner := domain.SessionFact{
		ID:     "ses-capture-1",
		Source: domain.SourceCodex,
		User:   "canonical-user",
		Invocations: []domain.InvocationFact{
			{
				Role:         "coder",
				Model:        "gpt-5.4",
				StartedAt:    &ts,
				CompletedAt:  &ts,
				Turns:        2,
				Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: domain.CloneCostEstimate(estimate),
			},
		},
		Tools:     []domain.SessionToolFact{{Name: "bash", Calls: 3, Errors: 0}},
		Anomalies: []domain.Anomaly{},
	}

	equal, err := domain.CaptureOwnerFactsEqual(rawSnapshot, pricedOwner)
	if err != nil {
		t.Fatalf("CaptureOwnerFactsEqual unexpected error: %v", err)
	}
	if !equal {
		t.Fatalf("CaptureOwnerFactsEqual(rawSnapshot, pricedOwner) = false, want true when only CostEstimate/User differs")
	}

	// Verify neither input slice was mutated.
	if rawSnapshot.Invocations[0].CostEstimate != nil {
		t.Fatalf("rawSnapshot.Invocations[0].CostEstimate was mutated to %+v", rawSnapshot.Invocations[0].CostEstimate)
	}
	if pricedOwner.Invocations[0].CostEstimate == nil || *pricedOwner.Invocations[0].CostEstimate != *estimate {
		t.Fatalf("pricedOwner.Invocations[0].CostEstimate was mutated: got %+v, want %+v", pricedOwner.Invocations[0].CostEstimate, *estimate)
	}

	// Different token counts must still report unequal.
	diverged := pricedOwner
	diverged.Invocations = append([]domain.InvocationFact{}, pricedOwner.Invocations...)
	diverged.Invocations[0].Tokens.Input = 200_001
	equalDiverged, err := domain.CaptureOwnerFactsEqual(rawSnapshot, diverged)
	if err != nil {
		t.Fatalf("CaptureOwnerFactsEqual(diverged) unexpected error: %v", err)
	}
	if equalDiverged {
		t.Fatalf("CaptureOwnerFactsEqual should return false when invocation tokens differ")
	}
}

func TestEquivalentCostProvenanceAcrossAllRollups(t *testing.T) {
	rateA := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	rateB := domain.AppliedRateSnapshot{
		MatchedPattern:      "*gemini-3.8*",
		InputPerMillion:     0.5,
		CacheReadPerMillion: 0.05,
		OutputPerMillion:    1.5,
		EffectiveCatalogID:  "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	ts1 := "2026-10-15T10:00:00Z"
	ts2 := "2026-10-16T10:00:00Z"

	invMeasured := domain.InvocationFact{
		Role:         "orchestrator",
		Model:        "claude-sonnet-4",
		StartedAt:    &ts1,
		CompletedAt:  &ts1,
		Tokens:       domain.InvocationTokens{Input: 1000, Output: 500},
		CostUSD:      23.20,
		CostEvidence: domain.ObservedCost(23.20),
	}
	invEstimated := domain.InvocationFact{
		Role:         "coder",
		Model:        "gpt-5.4",
		StartedAt:    &ts2,
		CompletedAt:  &ts2,
		Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
		CostEvidence: domain.UnavailableCost(),
		CostEstimate: &domain.CostEstimate{AmountUSD: 7.12, Rate: rateA},
	}
	invPartial := domain.InvocationFact{
		Role:         "reviewer",
		Model:        "claude-sonnet-4",
		StartedAt:    &ts2,
		CompletedAt:  &ts2,
		Tokens:       domain.InvocationTokens{Input: 1000, Output: 200},
		CostEvidence: domain.PartialCost(2.00),
	}
	invUnavailable := domain.InvocationFact{
		Role:         "tester",
		Model:        "unpriced-model",
		StartedAt:    &ts2,
		CompletedAt:  &ts2,
		Tokens:       domain.InvocationTokens{Input: 1000, Output: 200},
		CostEvidence: domain.UnavailableCost(),
	}

	t.Run("all 6 provenance states on ComputeTotals", func(t *testing.T) {
		// 1. empty
		emptyTotals := domain.ComputeTotals(nil)
		if emptyTotals.EquivalentCostSummary.Provenance != domain.CostProvenanceEmpty || emptyTotals.EquivalentCostSummary.Total != 0 {
			t.Fatalf("empty provenance = %+v", emptyTotals.EquivalentCostSummary)
		}
		assertCostPtrEqual(t, "empty measured", emptyTotals.EquivalentCostSummary.MeasuredCostUSD, nil)
		assertCostPtrEqual(t, "empty estimated", emptyTotals.EquivalentCostSummary.EstimatedCostUSD, nil)
		assertCostPtrEqual(t, "empty total", emptyTotals.EquivalentCostSummary.TotalEquivalentCostUSD, nil)
		assertCostPtrEqual(t, "empty subtotal", emptyTotals.EquivalentCostSummary.KnownSubtotalUSD, nil)

		// 2. measured
		measuredTotals := domain.ComputeTotals([]domain.InvocationFact{invMeasured})
		mEquiv := measuredTotals.EquivalentCostSummary
		if mEquiv.Provenance != domain.CostProvenanceMeasured || mEquiv.Measured != 1 || mEquiv.Total != 1 {
			t.Fatalf("measured provenance = %+v", mEquiv)
		}
		assertCostPtrEqual(t, "measured.measured", mEquiv.MeasuredCostUSD, domain.CostNumber(23.20))
		assertCostPtrEqual(t, "measured.estimated", mEquiv.EstimatedCostUSD, nil)
		assertCostPtrEqual(t, "measured.total", mEquiv.TotalEquivalentCostUSD, domain.CostNumber(23.20))

		// 3. estimated
		estimatedTotals := domain.ComputeTotals([]domain.InvocationFact{invEstimated})
		eEquiv := estimatedTotals.EquivalentCostSummary
		if eEquiv.Provenance != domain.CostProvenanceEstimated || eEquiv.Estimated != 1 || eEquiv.Total != 1 {
			t.Fatalf("estimated provenance = %+v", eEquiv)
		}
		assertCostPtrEqual(t, "estimated.measured", eEquiv.MeasuredCostUSD, nil)
		assertCostPtrEqual(t, "estimated.estimated", eEquiv.EstimatedCostUSD, domain.CostNumber(7.12))
		assertCostPtrEqual(t, "estimated.total", eEquiv.TotalEquivalentCostUSD, domain.CostNumber(7.12))

		// 4. hybrid
		hybridTotals := domain.ComputeTotals([]domain.InvocationFact{invMeasured, invEstimated})
		hEquiv := hybridTotals.EquivalentCostSummary
		if hEquiv.Provenance != domain.CostProvenanceHybrid || hEquiv.Measured != 1 || hEquiv.Estimated != 1 || hEquiv.Total != 2 {
			t.Fatalf("hybrid provenance = %+v", hEquiv)
		}
		assertCostPtrEqual(t, "hybrid.measured", hEquiv.MeasuredCostUSD, domain.CostNumber(23.20))
		assertCostPtrEqual(t, "hybrid.estimated", hEquiv.EstimatedCostUSD, domain.CostNumber(7.12))
		assertCostPtrEqual(t, "hybrid.total", hEquiv.TotalEquivalentCostUSD, domain.CostNumber(30.32))
		assertCostPtrEqual(t, "hybrid.subtotal", hEquiv.KnownSubtotalUSD, domain.CostNumber(30.32))

		// 5. partial (mixed measured + estimated + partial + unavailable)
		partialTotals := domain.ComputeTotals([]domain.InvocationFact{invMeasured, invEstimated, invPartial, invUnavailable})
		pEquiv := partialTotals.EquivalentCostSummary
		if pEquiv.Provenance != domain.CostProvenancePartial || pEquiv.Measured != 1 || pEquiv.Estimated != 1 || pEquiv.PartialMeasured != 1 || pEquiv.Unavailable != 1 || pEquiv.Total != 4 {
			t.Fatalf("partial provenance = %+v", pEquiv)
		}
		assertCostPtrEqual(t, "partial.measured", pEquiv.MeasuredCostUSD, domain.CostNumber(25.20))
		assertCostPtrEqual(t, "partial.estimated", pEquiv.EstimatedCostUSD, domain.CostNumber(7.12))
		assertCostPtrEqual(t, "partial.total", pEquiv.TotalEquivalentCostUSD, nil)
		assertCostPtrEqual(t, "partial.subtotal", pEquiv.KnownSubtotalUSD, domain.CostNumber(32.32))

		// 6. unavailable
		unavailTotals := domain.ComputeTotals([]domain.InvocationFact{invUnavailable})
		uEquiv := unavailTotals.EquivalentCostSummary
		if uEquiv.Provenance != domain.CostProvenanceUnavailable || uEquiv.Unavailable != 1 || uEquiv.Total != 1 {
			t.Fatalf("unavailable provenance = %+v", uEquiv)
		}
		assertCostPtrEqual(t, "unavail.measured", uEquiv.MeasuredCostUSD, nil)
		assertCostPtrEqual(t, "unavail.estimated", uEquiv.EstimatedCostUSD, nil)
		assertCostPtrEqual(t, "unavail.total", uEquiv.TotalEquivalentCostUSD, nil)
		assertCostPtrEqual(t, "unavail.subtotal", uEquiv.KnownSubtotalUSD, nil)
	})

	t.Run("tiny-amount unrounded summation accuracy and deduplicated sorted AppliedRates", func(t *testing.T) {
		var tinyInvs []domain.InvocationFact
		for i := 0; i < 1000; i++ {
			rate := rateA
			if i%2 == 0 {
				rate = rateB
			}
			tinyInvs = append(tinyInvs, domain.InvocationFact{
				Role:         "coder",
				Model:        "gpt-5.4",
				Tokens:       domain.InvocationTokens{Input: 100, Output: 10},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: &domain.CostEstimate{AmountUSD: 0.000123, Rate: rate},
			})
		}
		models := domain.ComputeModelMetrics(tinyInvs)
		if len(models) != 1 {
			t.Fatalf("expected 1 model metric, got %d", len(models))
		}
		// 1000 * 0.000123 = 0.123 (no premature 2-decimal rounding)
		assertCostPtrEqual(t, "tiny estimated sum", models[0].EquivalentCostSummary.EstimatedCostUSD, domain.CostNumber(0.123))
		assertCostPtrEqual(t, "tiny total equiv", models[0].EquivalentCostSummary.TotalEquivalentCostUSD, domain.CostNumber(0.123))
		if len(models[0].AppliedRates) != 2 {
			t.Fatalf("expected 2 deduplicated applied rates, got %+v", models[0].AppliedRates)
		}
		// Sorted by MatchedPattern: "*gemini-3.8*" < "gpt-5.4"
		if models[0].AppliedRates[0] != rateB || models[0].AppliedRates[1] != rateA {
			t.Fatalf("AppliedRates not sorted deterministically: %+v", models[0].AppliedRates)
		}
	})

	t.Run("Roles, Models, FleetSummary, TimeBucketMetric cumulative, and TicketComparisonRow", func(t *testing.T) {
		cfg := testProjectConfig()
		tid1 := "AL-1"
		tid2 := "AL-2"
		schema := domain.CurrentSchemaVersion

		ledger1 := domain.DeriveMetricsLedger(domain.AgenticFactLedger{
			SchemaVersion: &schema,
			TicketID:      &tid1,
			StartedAt:     &ts1,
			CompletedAt:   &ts1,
			UpdatedAt:     ts1,
			Sessions: []domain.SessionFact{
				{ID: "s1", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{invMeasured}},
			},
		}, cfg)
		ledger2 := domain.DeriveMetricsLedger(domain.AgenticFactLedger{
			SchemaVersion: &schema,
			TicketID:      &tid2,
			StartedAt:     &ts2,
			CompletedAt:   &ts2,
			UpdatedAt:     ts2,
			Sessions: []domain.SessionFact{
				{ID: "s2", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{invEstimated}},
			},
		}, cfg)

		// Roles and Models on ledger2
		if len(ledger2.Roles) != 1 || ledger2.Roles[0].EquivalentCostSummary.Provenance != domain.CostProvenanceEstimated {
			t.Fatalf("ledger2 role equiv = %+v", ledger2.Roles)
		}
		if len(ledger2.Models) != 1 || ledger2.Models[0].EquivalentCostSummary.Provenance != domain.CostProvenanceEstimated || len(ledger2.Models[0].AppliedRates) != 1 {
			t.Fatalf("ledger2 model equiv = %+v", ledger2.Models)
		}

		// FleetSummary combines measured ledger1 + estimated ledger2 => hybrid
		fleet := domain.CalculateFleetSummary([]domain.AgenticMetricsLedger{ledger1, ledger2})
		if fleet.EquivalentCostSummary.Provenance != domain.CostProvenanceHybrid {
			t.Fatalf("fleet provenance = %q, want hybrid", fleet.EquivalentCostSummary.Provenance)
		}
		assertCostPtrEqual(t, "fleet.measured", fleet.EquivalentCostSummary.MeasuredCostUSD, domain.CostNumber(23.20))
		assertCostPtrEqual(t, "fleet.estimated", fleet.EquivalentCostSummary.EstimatedCostUSD, domain.CostNumber(7.12))
		assertCostPtrEqual(t, "fleet.total", fleet.EquivalentCostSummary.TotalEquivalentCostUSD, domain.CostNumber(30.32))
		if fleet.RoleEquivalentCostSummaries["coder"].Provenance != domain.CostProvenanceEstimated {
			t.Fatalf("fleet role coder equiv = %+v", fleet.RoleEquivalentCostSummaries["coder"])
		}
		if fleet.ModelEquivalentCostSummaries["gpt-5.4"].Provenance != domain.CostProvenanceEstimated {
			t.Fatalf("fleet model gpt-5.4 equiv = %+v", fleet.ModelEquivalentCostSummaries["gpt-5.4"])
		}

		// TimeBucketMetric: day 1 measured, day 2 estimated => cumulative day 2 hybrid
		buckets := domain.AggregateMetricsOverTime([]domain.AgenticMetricsLedger{ledger1, ledger2}, string(domain.GranularityDaily), nil, nil)
		if len(buckets) != 2 {
			t.Fatalf("expected 2 daily buckets, got %d", len(buckets))
		}
		if buckets[0].EquivalentCostSummary.Provenance != domain.CostProvenanceMeasured ||
			buckets[0].CumulativeEquivalentCostSummary.Provenance != domain.CostProvenanceMeasured {
			t.Fatalf("bucket 0 equiv = %+v / cum = %+v", buckets[0].EquivalentCostSummary, buckets[0].CumulativeEquivalentCostSummary)
		}
		if buckets[1].EquivalentCostSummary.Provenance != domain.CostProvenanceEstimated ||
			buckets[1].CumulativeEquivalentCostSummary.Provenance != domain.CostProvenanceHybrid {
			t.Fatalf("bucket 1 equiv = %+v / cum = %+v", buckets[1].EquivalentCostSummary, buckets[1].CumulativeEquivalentCostSummary)
		}
		assertCostPtrEqual(t, "bucket 1 cum total", buckets[1].CumulativeEquivalentCostSummary.TotalEquivalentCostUSD, domain.CostNumber(30.32))

		// TicketComparisonRow
		rows := domain.BuildComparisonRows([]domain.AgenticMetricsLedger{ledger1, ledger2}, cfg.Ticket)
		if len(rows) != 2 {
			t.Fatalf("expected 2 comparison rows, got %d", len(rows))
		}
		byID := map[string]domain.TicketComparisonRow{}
		for _, r := range rows {
			byID[r.TicketID] = r
		}
		if byID["AL-1"].EquivalentCostSummary.Provenance != domain.CostProvenanceMeasured {
			t.Fatalf("AL-1 comparison equiv = %+v", byID["AL-1"].EquivalentCostSummary)
		}
		if byID["AL-2"].EquivalentCostSummary.Provenance != domain.CostProvenanceEstimated {
			t.Fatalf("AL-2 comparison equiv = %+v", byID["AL-2"].EquivalentCostSummary)
		}
		assertCostPtrEqual(t, "AL-2 comparison equiv total", byID["AL-2"].EquivalentCostSummary.TotalEquivalentCostUSD, domain.CostNumber(7.12))
	})
}
