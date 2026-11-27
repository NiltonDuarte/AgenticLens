package adapters_test

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestSchemaV6CostEstimateDTORoundTripAndValidation(t *testing.T) {
	baseLedger := func(invJSON string) string {
		return `{
			"schema_version": 6,
			"ticket_id": "AL-99",
			"ticket_title": "Cost Estimate DTO",
			"ticket_type": "feat",
			"started_at": "2026-10-17T10:00:00Z",
			"completed_at": "2026-10-17T11:00:00Z",
			"updated_at": "2026-10-17T11:00:00Z",
			"code": {
				"lines_added": 10, "lines_deleted": 2,
				"test_lines_added": 5, "test_lines_deleted": 0,
				"prod_lines_added": 5, "prod_lines_deleted": 2,
				"files_changed": 1, "sdlc_iterations": 1
			},
			"sessions": [{
				"id": "ses-1",
				"source": "codex",
				"user": "tester",
				"invocations": [` + invJSON + `],
				"tools": [],
				"anomalies": []
			}]
		}`
	}

	t.Run("existing v6 ledger without cost_estimate decodes and serializes without cost_estimate key", func(t *testing.T) {
		raw := baseLedger(`{
			"role": "coder",
			"model": "gpt-5.4",
			"started_at": null,
			"completed_at": null,
			"active_duration_seconds": 30,
			"turns": 2,
			"tokens": {"input": 1000, "output": 200, "cache_read": 500, "reasoning": 0, "peak_context": 4000},
			"cost_usd": null,
			"cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
			"compactions": 0,
			"user_iterations": 1
		}`)
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("unexpected validate error: %v", err)
		}
		dom := dto.ToDomain()
		if dom.Sessions[0].Invocations[0].CostEstimate != nil {
			t.Fatalf("expected nil CostEstimate, got %+v", dom.Sessions[0].Invocations[0].CostEstimate)
		}
		reencoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(dom))
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		if strings.Contains(string(reencoded), `"cost_estimate"`) {
			t.Fatalf("expected cost_estimate key to be omitted when nil, got %s", string(reencoded))
		}
	})

	t.Run("valid unavailable invocation with cost_estimate including zero estimate round-trips cleanly", func(t *testing.T) {
		for _, amountStr := range []string{"1.25", "0"} {
			raw := baseLedger(`{
				"role": "coder",
				"model": "gpt-5.4",
				"started_at": null,
				"completed_at": null,
				"active_duration_seconds": 30,
				"turns": 2,
				"tokens": {"input": 200000, "output": 50000, "cache_read": 800000, "reasoning": 10000, "peak_context": 90000},
				"cost_usd": null,
				"cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": ` + amountStr + `,
					"rate": {
						"matched_pattern": "gpt-5.4",
						"input_per_million": 2.0,
						"cache_read_per_million": 0.25,
						"output_per_million": 8.0,
						"effective_catalog_id": "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
					}
				},
				"compactions": 0,
				"user_iterations": 1
			}`)
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal([]byte(raw), &dto); err != nil {
				t.Fatalf("amount %s: unexpected decode error: %v", amountStr, err)
			}
			if err := dto.Validate(); err != nil {
				t.Fatalf("amount %s: unexpected validate error: %v", amountStr, err)
			}
			dom := dto.ToDomain()
			est := dom.Sessions[0].Invocations[0].CostEstimate
			if est == nil {
				t.Fatalf("amount %s: expected non-nil CostEstimate", amountStr)
			}
			if est.Rate.MatchedPattern != "gpt-5.4" || est.Rate.InputPerMillion != 2.0 || est.Rate.CacheReadPerMillion != 0.25 || est.Rate.OutputPerMillion != 8.0 {
				t.Fatalf("amount %s: rate snapshot corrupted: %+v", amountStr, est.Rate)
			}
			reencoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(dom))
			if err != nil {
				t.Fatalf("amount %s: marshal failed: %v", amountStr, err)
			}
			var decodedAgain adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(reencoded, &decodedAgain); err != nil {
				t.Fatalf("amount %s: round-trip decode failed: %v", amountStr, err)
			}
			est2 := decodedAgain.ToDomain().Sessions[0].Invocations[0].CostEstimate
			if est2 == nil || math.Abs(est2.AmountUSD-est.AmountUSD) > 1e-9 || est2.Rate != est.Rate {
				t.Fatalf("amount %s: round-trip mismatch: got %+v want %+v", amountStr, est2, est)
			}
		}
	})

	invalidInvocations := []struct {
		name    string
		invJSON string
	}{
		{
			name: "cost_estimate on observed invocation rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": 1.0, "cost_evidence": {"state": "observed", "known_subtotal_usd": 1.0},
				"cost_estimate": {
					"amount_usd": 1.0,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "cost_estimate on partial invocation rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "partial", "known_subtotal_usd": 0.5},
				"cost_estimate": {
					"amount_usd": 1.0,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "missing amount_usd rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "null amount_usd rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": null,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "negative amount_usd rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": -0.01,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "missing rate rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {"amount_usd": 1.0},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "blank matched_pattern rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": 1.0,
					"rate": {"matched_pattern": "  ", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "negative per-million rate rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": 1.0,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": -0.1, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "unknown field inside cost_estimate rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": 1.0,
					"extra_field": true,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc"}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
		{
			name: "unknown field inside cost_estimate.rate rejected",
			invJSON: `{
				"role": "coder", "model": "gpt-5.4", "started_at": null, "completed_at": null,
				"active_duration_seconds": 10, "turns": 1,
				"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
				"cost_usd": null, "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
				"cost_estimate": {
					"amount_usd": 1.0,
					"rate": {"matched_pattern": "gpt-5.4", "input_per_million": 2.0, "cache_read_per_million": 0.25, "output_per_million": 8.0, "effective_catalog_id": "sha256:abc", "cache_write_per_million": 1.0}
				},
				"compactions": 0, "user_iterations": 1
			}`,
		},
	}

	for _, tc := range invalidInvocations {
		t.Run(tc.name, func(t *testing.T) {
			var dto adapters.AgenticFactLedgerDTO
			err := json.Unmarshal([]byte(baseLedger(tc.invJSON)), &dto)
			if err == nil {
				err = dto.Validate()
			}
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}

	t.Run("explicit null cost_estimate rejected on unmarshal", func(t *testing.T) {
		raw := baseLedger(`{
			"role": "coder",
			"model": "gpt-5.4",
			"started_at": null,
			"completed_at": null,
			"active_duration_seconds": 10,
			"turns": 1,
			"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 200},
			"cost_usd": null,
			"cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
			"cost_estimate": null,
			"compactions": 0,
			"user_iterations": 1
		}`)
		var dto adapters.AgenticFactLedgerDTO
		err := json.Unmarshal([]byte(raw), &dto)
		if err == nil || !strings.Contains(err.Error(), "cost_estimate must be omitted instead of null") {
			t.Fatalf("expected explicit null cost_estimate rejection, got: %v", err)
		}
	})
}

func TestFactLedgerDTODirectValidateCostEstimate(t *testing.T) {
	schema := domain.CurrentSchemaVersion
	tid := "AL-99"
	title := "Direct DTO Validation"
	validRate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	makeDTO := func(inv adapters.InvocationFactDTO) adapters.AgenticFactLedgerDTO {
		return adapters.AgenticFactLedgerDTO{
			SchemaVersion: &schema,
			TicketID:      &tid,
			TicketTitle:   &title,
			UpdatedAt:     "2026-10-17T11:00:00Z",
			Sessions: []adapters.SessionFactDTO{
				{
					ID:          "ses-direct-1",
					Source:      string(domain.SourceCodex),
					User:        "tester",
					Invocations: []adapters.InvocationFactDTO{inv},
					Tools:       []adapters.SessionToolFactDTO{},
				},
			},
		}
	}

	baseInv := func(costUSD *float64, evidence domain.CostEvidence, estimate *domain.CostEstimate) adapters.InvocationFactDTO {
		ev := evidence
		return adapters.InvocationFactDTO{
			Role:                  "coder",
			Model:                 "gpt-5.4",
			ActiveDurationSeconds: 10,
			Turns:                 1,
			Tokens:                adapters.InvocationTokensDTO{Input: 100, Output: 50},
			CostUSD:               costUSD,
			CostEvidence:          &ev,
			CostEstimate:          estimate,
			UserIterations:        1,
		}
	}

	// Positive sanity check: valid direct DTO passes Validate().
	validDTO := makeDTO(baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{AmountUSD: 1.25, Rate: validRate}))
	if err := validDTO.Validate(); err != nil {
		t.Fatalf("valid direct DTO failed Validate(): %v", err)
	}

	cases := []struct {
		name    string
		inv     adapters.InvocationFactDTO
		wantErr string
	}{
		{
			name:    "observed invocation with cost_estimate rejected",
			inv:     baseInv(domain.CostNumber(2.0), domain.ObservedCost(2.0), &domain.CostEstimate{AmountUSD: 1.0, Rate: validRate}),
			wantErr: "cost_estimate is only allowed when cost_evidence.state is unavailable",
		},
		{
			name:    "partial invocation with cost_estimate rejected",
			inv:     baseInv(nil, domain.PartialCost(0.75), &domain.CostEstimate{AmountUSD: 1.0, Rate: validRate}),
			wantErr: "cost_estimate is only allowed when cost_evidence.state is unavailable",
		},
		{
			name:    "negative amount_usd rejected",
			inv:     baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{AmountUSD: -0.5, Rate: validRate}),
			wantErr: "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:    "NaN amount_usd rejected",
			inv:     baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{AmountUSD: math.NaN(), Rate: validRate}),
			wantErr: "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name:    "Inf amount_usd rejected",
			inv:     baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{AmountUSD: math.Inf(1), Rate: validRate}),
			wantErr: "cost_estimate.amount_usd must be a finite non-negative number",
		},
		{
			name: "negative input_per_million rejected",
			inv: baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     -2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			}),
			wantErr: "cost_estimate.rate.input_per_million must be a finite non-negative number",
		},
		{
			name: "negative cache_read_per_million rejected",
			inv: baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: -0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			}),
			wantErr: "cost_estimate.rate.cache_read_per_million must be a finite non-negative number",
		},
		{
			name: "negative output_per_million rejected",
			inv: baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    -8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			}),
			wantErr: "cost_estimate.rate.output_per_million must be a finite non-negative number",
		},
		{
			name: "blank matched_pattern rejected",
			inv: baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "   ",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  validRate.EffectiveCatalogID,
				},
			}),
			wantErr: "cost_estimate.rate.matched_pattern must be nonblank",
		},
		{
			name: "blank effective_catalog_id rejected",
			inv: baseInv(nil, domain.UnavailableCost(), &domain.CostEstimate{
				AmountUSD: 1.0,
				Rate: domain.AppliedRateSnapshot{
					MatchedPattern:      "gpt-5.4",
					InputPerMillion:     2.0,
					CacheReadPerMillion: 0.25,
					OutputPerMillion:    8.0,
					EffectiveCatalogID:  "   ",
				},
			}),
			wantErr: "cost_estimate.rate.effective_catalog_id must be nonblank",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := makeDTO(tc.inv).Validate()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("direct Validate() error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestFormatMarkdownMetricsTableEquivalentCosts(t *testing.T) {
	cfg := domain.ProjectConfig{
		AnomalyThresholds: domain.AnomalyThresholdsConfig{
			ContextCeilingWarningTokens:  160000,
			ContextCeilingCriticalTokens: 200000,
		},
	}
	schema := domain.CurrentSchemaVersion
	tid := "AL2-6"
	rate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:1111111111111111111111111111111111111111111111111111111111111111",
	}

	buildLedger := func(invs []domain.InvocationFact) domain.AgenticMetricsLedger {
		facts := domain.AgenticFactLedger{
			SchemaVersion: &schema,
			TicketID:      &tid,
			UpdatedAt:     "2026-10-17T11:00:00Z",
			Anomalies:     []domain.Anomaly{},
			Sessions: []domain.SessionFact{
				{ID: "ses-1", Source: domain.SourceCodex, User: "tester", Invocations: invs, Tools: []domain.SessionToolFact{}, Anomalies: []domain.Anomaly{}},
			},
		}
		return domain.DeriveMetricsLedger(facts, cfg)
	}

	t.Run("pure measured keeps baseline format with no EST or rate-card footnote", func(t *testing.T) {
		ledger := buildLedger([]domain.InvocationFact{
			{
				Role:         "orchestrator",
				Model:        "claude-sonnet-4",
				Tokens:       domain.InvocationTokens{Input: 5000, Output: 1000},
				CostUSD:      4.50,
				CostEvidence: domain.ObservedCost(4.50),
			},
		})
		md := adapters.FormatMarkdownMetricsTable(ledger)
		if !strings.Contains(md, "| $4.50 |") || !strings.Contains(md, "| **$4.50** |") {
			t.Fatalf("expected measured $4.50 cells in markdown:\n%s", md)
		}
		if strings.Contains(md, "EST") || strings.Contains(md, "Applied API-equivalent rate card") || strings.Contains(md, "CacheWrite") {
			t.Fatalf("pure measured markdown must not contain EST/rate-card/CacheWrite text:\n%s", md)
		}
	})

	t.Run("pure estimated renders EST and applied rate card footnote with CacheWrite note", func(t *testing.T) {
		ledger := buildLedger([]domain.InvocationFact{
			{
				Role:         "coder",
				Model:        "gpt-5.4",
				Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: &domain.CostEstimate{AmountUSD: 6.74, Rate: rate},
			},
		})
		md := adapters.FormatMarkdownMetricsTable(ledger)
		for _, want := range []string{
			"| $6.74 EST |",
			"| **$6.74 EST** |",
			"**Applied API-equivalent rate card:**",
			"`gpt-5.4`: $2.00 / $0.25 / $8.00 per 1M tokens (input / cache_read / output; catalog `sha256:1111111111111111111111111111111111111111111111111111111111111111`)",
			"CacheWrite",
		} {
			if !strings.Contains(md, want) {
				t.Fatalf("pure estimated markdown missing %q:\n%s", want, md)
			}
		}
	})

	t.Run("hybrid renders measured, EST, and equiv total", func(t *testing.T) {
		ledger := buildLedger([]domain.InvocationFact{
			{
				Role:         "orchestrator",
				Model:        "claude-sonnet-4",
				Tokens:       domain.InvocationTokens{Input: 5000, Output: 1000},
				CostUSD:      23.20,
				CostEvidence: domain.ObservedCost(23.20),
			},
			{
				Role:         "coder",
				Model:        "gpt-5.4",
				Tokens:       domain.InvocationTokens{Input: 200_000, CacheRead: 800_000, Output: 50_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: &domain.CostEstimate{AmountUSD: 7.12, Rate: rate},
			},
		})
		md := adapters.FormatMarkdownMetricsTable(ledger)
		for _, want := range []string{
			"| **claude-sonnet-4** | 1 | 5,000 | 1,000 | 0 | 0.0% | $23.20 |",
			"| **gpt-5.4** | 1 | 200,000 | 50,000 | 800,000 | 80.0% | $7.12 EST |",
			"**$23.20 measured + $7.12 EST ($30.32 equiv)**",
			"**Applied API-equivalent rate card:**",
		} {
			if !strings.Contains(md, want) {
				t.Fatalf("hybrid markdown missing %q:\n%s", want, md)
			}
		}
	})

	t.Run("partial mixed renders explicit coverage without complete equiv total", func(t *testing.T) {
		ledger := buildLedger([]domain.InvocationFact{
			{
				Role:         "orchestrator",
				Model:        "claude-sonnet-4",
				Tokens:       domain.InvocationTokens{Input: 5000, Output: 1000},
				CostUSD:      2.00,
				CostEvidence: domain.ObservedCost(2.00),
			},
			{
				Role:         "coder",
				Model:        "gpt-5.4",
				Tokens:       domain.InvocationTokens{Input: 100_000, Output: 10_000},
				CostEvidence: domain.UnavailableCost(),
				CostEstimate: &domain.CostEstimate{AmountUSD: 3.00, Rate: rate},
			},
			{
				Role:         "tester",
				Model:        "unmatched-model",
				Tokens:       domain.InvocationTokens{Input: 1000, Output: 500},
				CostEvidence: domain.UnavailableCost(),
			},
		})
		md := adapters.FormatMarkdownMetricsTable(ledger)
		wantPartial := "$2.00 measured + $3.00 EST (partial; 1 measured, 1 estimated, 0 partial, 1 unavailable / 3 invocations)"
		if !strings.Contains(md, wantPartial) {
			t.Fatalf("partial markdown missing %q:\n%s", wantPartial, md)
		}
		if strings.Contains(md, "equiv)") {
			t.Fatalf("partial markdown must not claim complete equiv total:\n%s", md)
		}
	})
}
