package main

import (
	"errors"
	"math"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestPendingPromotionCoverageWithPricedCanonical(t *testing.T) {
	ts := "2026-10-17T10:00:00Z"
	rate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	rawPendingOwner := domain.SessionFact{
		ID:     "pending-unavail-1",
		Source: domain.SourceCodex,
		User:   "capture-worker",
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
		Tools:     []domain.SessionToolFact{{Name: "bash", Calls: 2, Errors: 0}},
		Anomalies: []domain.Anomaly{},
	}

	pricedCanonicalOwner := rawPendingOwner
	pricedCanonicalOwner.User = "canonical-dev"
	pricedCanonicalOwner.Invocations = append([]domain.InvocationFact{}, rawPendingOwner.Invocations...)
	pricedCanonicalOwner.Invocations[0].CostEstimate = &domain.CostEstimate{
		AmountUSD: 1.0,
		Rate:      rate,
	}

	t.Run("priced canonical owner matches unpriced pending unavailable owner", func(t *testing.T) {
		err := verifyPendingPromotionCoverage(
			[]domain.SessionFact{pricedCanonicalOwner},
			nil,
			[]domain.SessionFact{rawPendingOwner},
		)
		if err != nil {
			t.Fatalf("verifyPendingPromotionCoverage rejected priced canonical vs unpriced pending: %v", err)
		}
	})

	t.Run("non-estimate fact divergence still fails closed for unavailable owner", func(t *testing.T) {
		diverged := pricedCanonicalOwner
		diverged.Invocations = append([]domain.InvocationFact{}, pricedCanonicalOwner.Invocations...)
		diverged.Invocations[0].Tokens.Input = 999_999

		err := verifyPendingPromotionCoverage(
			[]domain.SessionFact{diverged},
			nil,
			[]domain.SessionFact{rawPendingOwner},
		)
		if err == nil || !strings.Contains(err.Error(), "canonical and pending facts disagree for unavailable owner") {
			t.Fatalf("expected disagreement error, got: %v", err)
		}
	})

	t.Run("missing pending owner in canonical publication fails closed", func(t *testing.T) {
		err := verifyPendingPromotionCoverage(
			nil,
			nil,
			[]domain.SessionFact{rawPendingOwner},
		)
		if err == nil || !strings.Contains(err.Error(), "canonical publication does not cover pending owner") {
			t.Fatalf("expected missing coverage error, got: %v", err)
		}
	})
}

func TestFillMissingPricedCaptureSessions(t *testing.T) {
	// Pricing config contains an unambiguous rule for claude-sonnet-4-5 AND
	// an equal-specificity ambiguous glob pair ("gpt-5*" vs "*4-pro") for "gpt-5.4-pro".
	pricing := domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "claude-sonnet-4-5", InputPerMillion: 3.0, CacheReadPerMillion: 0.30, OutputPerMillion: 15.0},
			{Model: "gpt-5*", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
			{Model: "*4-pro", InputPerMillion: 4.0, CacheReadPerMillion: 0.50, OutputPerMillion: 12.0},
		},
	}
	wantCatalog := pricing.EffectiveCatalogID()

	persistedEstimate := &domain.CostEstimate{
		AmountUSD: 9.99,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "historical-rule",
			InputPerMillion:     1.0,
			CacheReadPerMillion: 0.1,
			OutputPerMillion:    4.0,
			EffectiveCatalogID:  "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
	}

	// Unrefreshed cachedFacts owner uses the model that would trigger ambiguity if re-priced!
	cachedOwner := domain.SessionFact{
		ID:     "cached-unrefreshed-owner",
		Source: domain.SourceCodex,
		User:   "remote-dev",
		Invocations: []domain.InvocationFact{
			{
				Role:         "coder",
				Model:        "gpt-5.4-pro",
				Tokens:       domain.InvocationTokens{Input: 100_000, Output: 20_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: domain.CloneCostEstimate(persistedEstimate),
			},
		},
	}

	// Pending-only owner not present in cachedFacts or freshSessions must be priced.
	// Expected estimate: (200_000*3.0 + 1_000_000*0.30 + 100_000*15.0)/1e6 = 0.60 + 0.30 + 1.50 = 2.40
	pendingOnlyOwner := domain.SessionFact{
		ID:     "pending-only-owner",
		Source: domain.SourceClaude,
		User:   "local-dev",
		Invocations: []domain.InvocationFact{
			{
				Role:         "reviewer",
				Model:        "claude-sonnet-4-5",
				Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 1_000_000, Output: 100_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: nil,
			},
		},
	}

	// Fresh owner already priced in freshSessions should not be re-priced when seen in fallback.
	freshOwner := domain.SessionFact{
		ID:     "fresh-owner",
		Source: domain.SourceCodex,
		User:   "local-dev",
		Invocations: []domain.InvocationFact{
			{
				Role:         "coder",
				Model:        "gpt-5.4-pro",
				Tokens:       domain.InvocationTokens{Input: 50_000, Output: 10_000},
				CostEvidence: domain.UnavailableCost(),
			},
		},
	}

	t.Run("preserves cachedFacts unrefreshed owner without spurious ambiguity while pricing pending-only owner", func(t *testing.T) {
		result, err := fillMissingPricedCaptureSessions(
			[]domain.SessionFact{cachedOwner},
			[]domain.SessionFact{freshOwner},
			pricing,
			[]domain.SessionFact{
				// Duplicate of cachedOwner (unpriced copy) + new pendingOnlyOwner + freshOwner
				{
					ID:     cachedOwner.ID,
					Source: cachedOwner.Source,
					User:   "pending-copy",
					Invocations: []domain.InvocationFact{
						{
							Role:         "coder",
							Model:        "gpt-5.4-pro",
							Tokens:       domain.InvocationTokens{Input: 100_000, Output: 20_000},
							CostEvidence: domain.UnavailableCost(),
						},
					},
				},
				pendingOnlyOwner,
				freshOwner,
			},
		)
		if err != nil {
			t.Fatalf("fillMissingPricedCaptureSessions failed unexpectedly: %v", err)
		}

		byID := make(map[string]domain.SessionFact, len(result))
		for _, s := range result {
			byID[s.ID] = s
		}

		gotCached, ok := byID[cachedOwner.ID]
		if !ok {
			t.Fatalf("cached owner missing from result")
		}
		if gotCached.Invocations[0].CostEstimate == nil || *gotCached.Invocations[0].CostEstimate != *persistedEstimate {
			t.Fatalf("cached owner estimate altered: got %+v, want %+v", gotCached.Invocations[0].CostEstimate, *persistedEstimate)
		}

		gotPending, ok := byID[pendingOnlyOwner.ID]
		if !ok {
			t.Fatalf("newly promoted pending-only owner missing from result")
		}
		est := gotPending.Invocations[0].CostEstimate
		if est == nil {
			t.Fatalf("expected newly promoted pending-only owner to be priced, got nil")
		}
		if math.Abs(est.AmountUSD-2.40) > 1e-9 || est.Rate.MatchedPattern != "claude-sonnet-4-5" || est.Rate.EffectiveCatalogID != wantCatalog {
			t.Fatalf("pending-only owner estimate mismatch: %+v", *est)
		}
	})

	t.Run("newly promoted pending-only owner with ambiguous model fails closed", func(t *testing.T) {
		ambiguousPending := domain.SessionFact{
			ID:     "pending-ambiguous",
			Source: domain.SourceCodex,
			User:   "local-dev",
			Invocations: []domain.InvocationFact{
				{
					Role:         "coder",
					Model:        "gpt-5.4-pro",
					Tokens:       domain.InvocationTokens{Input: 100_000, Output: 20_000},
					CostEvidence: domain.UnavailableCost(),
				},
			},
		}

		_, err := fillMissingPricedCaptureSessions(
			[]domain.SessionFact{cachedOwner},
			nil,
			pricing,
			[]domain.SessionFact{ambiguousPending},
		)
		var ambigErr *domain.AmbiguousModelPricingError
		if !errors.As(err, &ambigErr) {
			t.Fatalf("expected AmbiguousModelPricingError for newly promoted ambiguous owner, got %T: %v", err, err)
		}
	})
}
