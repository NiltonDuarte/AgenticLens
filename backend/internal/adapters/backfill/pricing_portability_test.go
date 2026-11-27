package backfill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func setProjectModelPricing(t *testing.T, projectRoot string, rates []domain.ModelPricingRate) domain.ProjectConfig {
	t.Helper()
	cfgPath := filepath.Join(projectRoot, ".agenticlens-config.yaml")
	cfg, err := adaptersconfig.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ModelPricing = domain.ModelPricingConfig{Rates: rates}
	data, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	reloaded, err := adaptersconfig.LoadFromFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	return reloaded
}

func writePricedCodexRollout(t *testing.T, dir, id, role, model string, totalInput, cachedInput, output, reasoning int64) string {
	t.Helper()
	meta := struct {
		ID   string `json:"id"`
		Role string `json:"agent_role,omitempty"`
	}{id, role}
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("{\"timestamp\":\"2026-10-20T09:00:00Z\",\"type\":\"session_meta\",\"payload\":%s}\n", encoded)
	body += fmt.Sprintf(
		"{\"timestamp\":\"2026-10-20T09:00:05Z\",\"type\":\"token_usage_record\",\"payload\":{\"turn_id\":\"turn-1\",\"model\":%q,\"usage\":{\"input_tokens\":%d,\"cached_input_tokens\":%d,\"output_tokens\":%d,\"reasoning_output_tokens\":%d}}}\n",
		model, totalInput, cachedInput, output, reasoning,
	)
	path := filepath.Join(dir, id+".jsonl")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPartialRefreshPreservesUnavailableOwnerEstimatesAndIsIdempotent(t *testing.T) {
	project := newBackfillProject(t)
	ticket := "AL2-6"

	// Machine A originally saved session-a with Rate Card A (catalog A, $4.25).
	rateCardA := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     1.5,
		CacheReadPerMillion: 0.15,
		OutputPerMillion:    6.0,
		EffectiveCatalogID:  "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	estimateA := &domain.CostEstimate{
		AmountUSD: 4.25,
		Rate:      rateCardA,
	}

	// Machine B configures a DIFFERENT active Rate Card B.
	cfgB := setProjectModelPricing(t, project, []domain.ModelPricingRate{
		{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
	})
	catalogB := cfgB.ModelPricing.EffectiveCatalogID()
	if catalogB == "" || catalogB == rateCardA.EffectiveCatalogID {
		t.Fatalf("unexpected catalogB: %q", catalogB)
	}

	schema := domain.CurrentSchemaVersion
	storedSessions := []domain.SessionFact{
		{
			ID:     "session-a-unavailable",
			Source: domain.SourceCodex,
			User:   "developer-a",
			Invocations: []domain.InvocationFact{
				{
					Role:         "planner",
					Model:        "gpt-5.4",
					Tokens:       domain.InvocationTokens{Input: 300_000, CacheRead: 600_000, Output: 50_000},
					CostEvidence: domain.UnavailableCost(),
					CostEstimate: domain.CloneCostEstimate(estimateA),
				},
			},
			Tools:     []domain.SessionToolFact{},
			Anomalies: []domain.Anomaly{},
		},
		{
			ID:     "session-b-local",
			Source: domain.SourceCodex,
			User:   "developer-b",
			Invocations: []domain.InvocationFact{
				{
					Role:         "coder",
					Model:        "gpt-5.4",
					Tokens:       domain.InvocationTokens{Input: 10, CacheRead: 0, Output: 5},
					CostEvidence: domain.UnavailableCost(),
				},
			},
			Tools:     []domain.SessionToolFact{},
			Anomalies: []domain.Anomaly{},
		},
		{
			ID:     "session-oc-other-harness",
			Source: domain.SourceOpenCode,
			User:   "developer-a",
			Invocations: []domain.InvocationFact{
				{
					Role:         "orchestrator",
					Model:        "claude-sonnet-4",
					Tokens:       domain.InvocationTokens{Input: 1000, Output: 200},
					CostUSD:      3.50,
					CostEvidence: domain.ObservedCost(3.50),
				},
			},
			Tools:     []domain.SessionToolFact{},
			Anomalies: []domain.Anomaly{},
		},
	}

	createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &schema,
		TicketID:      &ticket,
		UpdatedAt:     "2026-10-20T08:00:00Z",
		Sessions:      storedSessions,
		Anomalies:     []domain.Anomaly{},
	})
	ledgerPath := filepath.Join(project, ticket, "agentic_metrics.json")

	// Machine B only has session-b-local in its Codex store (session-a-unavailable is NOT present).
	// Input 1,000,000 total with 800,000 cached => 200,000 uncached input, 800,000 cache_read, 50,000 output.
	// Under Rate Card B (2.0 / 0.25 / 8.0): (200k*2 + 800k*0.25 + 50k*8)/1e6 = $1.00.
	codexHome := t.TempDir()
	sessionsDir := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(sessionsDir, 0700); err != nil {
		t.Fatal(err)
	}
	writePricedCodexRollout(t, sessionsDir, "session-b-local", "coder", "gpt-5.4", 1_000_000, 800_000, 50_000, 10_000)

	var firstRefreshSessions []domain.SessionFact
	for pass := 1; pass <= 2; pass++ {
		report, err := NewEngine().Execute(BackfillOptions{
			Project:     project,
			TicketID:    ticket,
			Source:      "codex",
			CodexHome:   codexHome,
			RecomputeDB: true,
			Force:       true,
		})
		if err != nil {
			t.Fatalf("pass %d: backfill error: %v", pass, err)
		}
		if len(report.Results) != 1 || report.Results[0].Error != nil {
			t.Fatalf("pass %d: unexpected ticket results: %+v", pass, report.Results)
		}

		after, err := adapters.LoadExistingFactLedger(ledgerPath)
		if err != nil {
			t.Fatalf("pass %d: reload ledger: %v", pass, err)
		}
		if len(after.Sessions) != 3 {
			t.Fatalf("pass %d: expected 3 sessions, got %d", pass, len(after.Sessions))
		}

		byID := map[string]domain.SessionFact{}
		for _, s := range after.Sessions {
			byID[s.ID] = s
		}

		// 1. Unavailable Developer A session retains its EXACT persisted CostEstimate (Rate Card A, $4.25).
		gotA := byID["session-a-unavailable"]
		if len(gotA.Invocations) != 1 || gotA.Invocations[0].CostEstimate == nil {
			t.Fatalf("pass %d: session-a lost CostEstimate: %+v", pass, gotA)
		}
		if !reflect.DeepEqual(*gotA.Invocations[0].CostEstimate, *estimateA) {
			t.Fatalf("pass %d: session-a CostEstimate altered: got %+v want %+v", pass, *gotA.Invocations[0].CostEstimate, *estimateA)
		}

		// 2. Locally refreshed Developer B session gets Rate Card B ($1.00, catalogB).
		gotB := byID["session-b-local"]
		if len(gotB.Invocations) != 1 || gotB.Invocations[0].CostEstimate == nil {
			t.Fatalf("pass %d: session-b missing CostEstimate: %+v", pass, gotB)
		}
		estB := gotB.Invocations[0].CostEstimate
		if math.Abs(estB.AmountUSD-1.0) > 1e-9 || estB.Rate.EffectiveCatalogID != catalogB || estB.Rate.MatchedPattern != "gpt-5.4" {
			t.Fatalf("pass %d: session-b estimate mismatch: %+v (want $1.00, catalog %s)", pass, estB, catalogB)
		}

		// 3. Other-harness OpenCode session stays observed $3.50 with nil CostEstimate.
		gotOC := byID["session-oc-other-harness"]
		if len(gotOC.Invocations) != 1 || gotOC.Invocations[0].CostEstimate != nil || gotOC.Invocations[0].CostUSD != 3.50 {
			t.Fatalf("pass %d: OpenCode session corrupted: %+v", pass, gotOC)
		}

		// 4. Whole-ledger derived aggregates combine measured ($3.50) + preserved A ($4.25) + refreshed B ($1.00) => hybrid $8.75.
		rawLedgerBytes, err := os.ReadFile(ledgerPath)
		if err != nil {
			t.Fatalf("pass %d: read ledger bytes: %v", pass, err)
		}
		var savedDTO adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(rawLedgerBytes, &savedDTO); err != nil {
			t.Fatalf("pass %d: decode saved ledger DTO: %v", pass, err)
		}
		if err := savedDTO.Validate(); err != nil {
			t.Fatalf("pass %d: validate saved ledger DTO: %v", pass, err)
		}
		derived := domain.DeriveMetricsLedger(savedDTO.ToDomain(), cfgB)
		equiv := derived.Totals.EquivalentCostSummary
		if equiv.Provenance != domain.CostProvenanceHybrid || equiv.Measured != 1 || equiv.Estimated != 2 || equiv.Total != 3 {
			t.Fatalf("pass %d: whole-ledger equiv = %+v", pass, equiv)
		}
		if equiv.MeasuredCostUSD == nil || math.Abs(*equiv.MeasuredCostUSD-3.50) > 1e-9 {
			t.Fatalf("pass %d: whole-ledger measured = %v, want 3.50", pass, equiv.MeasuredCostUSD)
		}
		if equiv.EstimatedCostUSD == nil || math.Abs(*equiv.EstimatedCostUSD-5.25) > 1e-9 {
			t.Fatalf("pass %d: whole-ledger estimated = %v, want 5.25", pass, equiv.EstimatedCostUSD)
		}
		if equiv.TotalEquivalentCostUSD == nil || math.Abs(*equiv.TotalEquivalentCostUSD-8.75) > 1e-9 {
			t.Fatalf("pass %d: whole-ledger total_equivalent = %v, want 8.75", pass, equiv.TotalEquivalentCostUSD)
		}

		if pass == 1 {
			firstRefreshSessions = after.Sessions
		} else if !reflect.DeepEqual(after.Sessions, firstRefreshSessions) {
			t.Fatalf("pass 2 not idempotent:\nfirst:  %+v\nsecond: %+v", firstRefreshSessions, after.Sessions)
		}
	}
}

func TestBackfillFailsClosedOnAmbiguousPricingPatternWithoutCorruptingLedger(t *testing.T) {
	project := newBackfillProject(t)
	ticket := "AL2-7"

	// Configure two equal-specificity patterns that collide on "gpt-5-mini":
	// "gpt-5*" (5 literals, len 6) and "*-mini" (5 literals, len 6).
	setProjectModelPricing(t, project, []domain.ModelPricingRate{
		{Model: "gpt-5*", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
		{Model: "*-mini", InputPerMillion: 0.5, CacheReadPerMillion: 0.05, OutputPerMillion: 1.5},
	})

	schema := domain.CurrentSchemaVersion
	createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &schema,
		TicketID:      &ticket,
		UpdatedAt:     "2026-10-20T08:00:00Z",
		Sessions: []domain.SessionFact{
			{
				ID:     "ses-ambig",
				Source: domain.SourceCodex,
				User:   "dev",
				Invocations: []domain.InvocationFact{
					{
						Role:         "coder",
						Model:        "gpt-5-mini",
						Tokens:       domain.InvocationTokens{Input: 100, Output: 50},
						CostEvidence: domain.UnavailableCost(),
					},
				},
				Tools:     []domain.SessionToolFact{},
				Anomalies: []domain.Anomaly{},
			},
		},
		Anomalies: []domain.Anomaly{},
	})

	ledgerPath := filepath.Join(project, ticket, "agentic_metrics.json")
	beforeBytes, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}

	codexHome := t.TempDir()
	sessionsDir := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(sessionsDir, 0700); err != nil {
		t.Fatal(err)
	}
	writePricedCodexRollout(t, sessionsDir, "ses-ambig", "coder", "gpt-5-mini", 500_000, 100_000, 20_000, 0)

	report, err := NewEngine().Execute(BackfillOptions{
		Project:     project,
		TicketID:    ticket,
		Source:      "codex",
		CodexHome:   codexHome,
		RecomputeDB: true,
		Force:       true,
	})
	if err != nil {
		t.Fatalf("unexpected top-level error (expected per-ticket isolation error): %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].Error == nil {
		t.Fatalf("expected ambiguous pricing error on ticket result, got: %+v", report.Results)
	}
	if !strings.Contains(report.Results[0].Error.Error(), "ambiguous model pricing match") {
		t.Fatalf("expected ambiguity diagnostic, got: %v", report.Results[0].Error)
	}

	afterBytes, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBytes, afterBytes) {
		t.Fatalf("ambiguous pricing failure modified existing ledger on disk")
	}
}
