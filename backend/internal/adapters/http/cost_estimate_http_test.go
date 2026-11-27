package http_test

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
)

func assertHTTPFloatPtr(t *testing.T, label string, got *float64, want *float64) {
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
		t.Fatalf("%s: got %.9f, want %.9f", label, *got, *want)
	}
}

func TestHTTPEquivalentCostEndpointsAndCrossMachineImmutability(t *testing.T) {
	const catalogA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	// Machine A saved ledger: AL-10 has measured OpenCode ($23.20) + estimated Codex ($7.12) => hybrid ($30.32).
	// AL-11 is pure estimated ($6.74) with rate snapshot from Catalog A.
	al10JSON := `{
  "schema_version": 6,
  "ticket_id": "AL-10",
  "ticket_title": "Hybrid Ticket",
  "ticket_type": "feat",
  "started_at": "2026-10-15T10:00:00Z",
  "completed_at": "2026-10-15T11:00:00Z",
  "updated_at": "2026-10-15T11:00:00Z",
  "code": {"lines_added": 50, "lines_deleted": 10, "test_lines_added": 25, "test_lines_deleted": 0, "prod_lines_added": 25, "prod_lines_deleted": 10, "files_changed": 2, "sdlc_iterations": 1},
  "sessions": [
    {
      "id": "ses-oc",
      "source": "opencode",
      "user": "dev-a",
      "invocations": [{
        "role": "orchestrator",
        "model": "claude-sonnet-4",
        "started_at": "2026-10-15T10:05:00Z",
        "completed_at": "2026-10-15T10:20:00Z",
        "active_duration_seconds": 900,
        "turns": 2,
        "tokens": {"input": 10000, "output": 2000, "cache_read": 5000, "reasoning": 0, "peak_context": 40000},
        "cost_usd": 23.20,
        "cost_evidence": {"state": "observed", "known_subtotal_usd": 23.20},
        "compactions": 0,
        "user_iterations": 1
      }],
      "tools": [],
      "anomalies": []
    },
    {
      "id": "ses-codex",
      "source": "codex",
      "user": "dev-a",
      "invocations": [{
        "role": "coder",
        "model": "gpt-5.4",
        "started_at": "2026-10-15T10:25:00Z",
        "completed_at": "2026-10-15T10:45:00Z",
        "active_duration_seconds": 1200,
        "turns": 4,
        "tokens": {"input": 200000, "output": 50000, "cache_read": 800000, "reasoning": 10000, "peak_context": 90000},
        "cost_usd": null,
        "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
        "cost_estimate": {
          "amount_usd": 7.12,
          "rate": {
            "matched_pattern": "gpt-5.4",
            "input_per_million": 2.0,
            "cache_read_per_million": 0.25,
            "output_per_million": 8.0,
            "effective_catalog_id": "` + catalogA + `"
          }
        },
        "compactions": 0,
        "user_iterations": 1
      }],
      "tools": [],
      "anomalies": []
    }
  ]
}`

	al11JSON := `{
  "schema_version": 6,
  "ticket_id": "AL-11",
  "ticket_title": "Estimated Ticket",
  "ticket_type": "feat",
  "started_at": "2026-10-16T10:00:00Z",
  "completed_at": "2026-10-16T11:00:00Z",
  "updated_at": "2026-10-16T11:00:00Z",
  "code": {"lines_added": 20, "lines_deleted": 0, "test_lines_added": 10, "test_lines_deleted": 0, "prod_lines_added": 10, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
  "sessions": [
    {
      "id": "ses-codex-2",
      "source": "codex",
      "user": "dev-a",
      "invocations": [{
        "role": "coder",
        "model": "gpt-5.4",
        "started_at": "2026-10-16T10:10:00Z",
        "completed_at": "2026-10-16T10:40:00Z",
        "active_duration_seconds": 1800,
        "turns": 3,
        "tokens": {"input": 100000, "output": 20000, "cache_read": 400000, "reasoning": 5000, "peak_context": 60000},
        "cost_usd": null,
        "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
        "cost_estimate": {
          "amount_usd": 6.74,
          "rate": {
            "matched_pattern": "gpt-5.4",
            "input_per_million": 2.0,
            "cache_read_per_million": 0.25,
            "output_per_million": 8.0,
            "effective_catalog_id": "` + catalogA + `"
          }
        },
        "compactions": 0,
        "user_iterations": 1
      }],
      "tools": [],
      "anomalies": []
    }
  ]
}`

	// Test on Machine B with (a) absent local rate card and (b) wildly different local Rate Card B.
	modes := []struct {
		name  string
		rates []domain.ModelPricingRate
	}{
		{name: "machine B with absent rate card", rates: nil},
		{
			name: "machine B with different rate card",
			rates: []domain.ModelPricingRate{
				{Model: "gpt-5.4", InputPerMillion: 999.0, CacheReadPerMillion: 500.0, OutputPerMillion: 1999.0},
			},
		},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			project := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			cfg := httpFixturePolicy()
			cfg.ConfigPath = filepath.Join(project, ".agenticlens-config.yaml")
			cfg.ModelPricing = domain.ModelPricingConfig{Rates: mode.rates}
			writeHTTPFixturePolicy(t, project, cfg)

			ws10 := filepath.Join(project, "workspaces", "AL-10")
			ws11 := filepath.Join(project, "workspaces", "AL-11")
			if err := os.MkdirAll(ws10, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(ws11, 0700); err != nil {
				t.Fatal(err)
			}
			al10Path := filepath.Join(ws10, "agentic_metrics.json")
			al11Path := filepath.Join(ws11, "agentic_metrics.json")
			if err := os.WriteFile(al10Path, []byte(al10JSON), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(al11Path, []byte(al11JSON), 0600); err != nil {
				t.Fatal(err)
			}
			al10Before, err := os.ReadFile(al10Path)
			if err != nil {
				t.Fatal(err)
			}

			server := adaptershttp.NewServerWithConfig(project, fstest.MapFS{}, cfg)
			get := func(target string) *httptest.ResponseRecorder {
				t.Helper()
				rec := httptest.NewRecorder()
				server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
				if rec.Code != http.StatusOK {
					t.Fatalf("GET %s returned %d: %s", target, rec.Code, rec.Body.String())
				}
				return rec
			}

			// 1. /api/workspaces/AL-10
			var detail adaptershttp.WorkspaceDetailResponse
			if err := json.Unmarshal(get("/api/workspaces/AL-10").Body.Bytes(), &detail); err != nil {
				t.Fatalf("unmarshal workspace detail: %v", err)
			}
			tot := detail.Ledger.Totals
			// Unchanged measured cost_usd / cost_summary contract
			assertHTTPFloatPtr(t, "AL-10 cost_usd", tot.CostUSD, nil)
			if tot.CostSummary.State != domain.CostSummaryPartial {
				t.Fatalf("AL-10 cost_summary.state = %q, want partial", tot.CostSummary.State)
			}
			// New equivalent fields match Machine A's exact saved values (NOT Machine B's 999 rate!)
			if tot.CostProvenance != domain.CostProvenanceHybrid || tot.EquivalentCostSummary.Provenance != domain.CostProvenanceHybrid {
				t.Fatalf("AL-10 provenance = %q / %+v, want hybrid", tot.CostProvenance, tot.EquivalentCostSummary)
			}
			assertHTTPFloatPtr(t, "AL-10 measured_cost_usd", tot.MeasuredCostUSD, domain.CostNumber(23.20))
			assertHTTPFloatPtr(t, "AL-10 estimated_cost_usd", tot.EstimatedCostUSD, domain.CostNumber(7.12))
			assertHTTPFloatPtr(t, "AL-10 total_equivalent_cost_usd", tot.TotalEquivalentCostUSD, domain.CostNumber(30.32))

			var foundGPT bool
			for _, m := range detail.Ledger.Models {
				if m.Model == "gpt-5.4" {
					foundGPT = true
					if m.CostProvenance != domain.CostProvenanceEstimated {
						t.Fatalf("gpt-5.4 provenance = %q, want estimated", m.CostProvenance)
					}
					assertHTTPFloatPtr(t, "gpt-5.4 estimated", m.EstimatedCostUSD, domain.CostNumber(7.12))
					if len(m.AppliedRates) != 1 || m.AppliedRates[0].EffectiveCatalogID != catalogA || m.AppliedRates[0].InputPerMillion != 2.0 {
						t.Fatalf("gpt-5.4 applied_rates altered on Machine B: %+v", m.AppliedRates)
					}
				}
			}
			if !foundGPT {
				t.Fatalf("gpt-5.4 model row missing from /api/workspaces/AL-10")
			}

			// 2. /api/fleet
			var fleet adaptershttp.FleetSummaryDTO
			if err := json.Unmarshal(get("/api/fleet").Body.Bytes(), &fleet); err != nil {
				t.Fatalf("unmarshal fleet: %v", err)
			}
			assertHTTPFloatPtr(t, "fleet total_cost_usd", fleet.TotalCostUSD, nil)
			if fleet.CostSummary.State != domain.CostSummaryPartial {
				t.Fatalf("fleet cost_summary.state = %q, want partial", fleet.CostSummary.State)
			}
			if fleet.CostProvenance != domain.CostProvenanceHybrid {
				t.Fatalf("fleet cost_provenance = %q, want hybrid", fleet.CostProvenance)
			}
			assertHTTPFloatPtr(t, "fleet measured_cost_usd", fleet.MeasuredCostUSD, domain.CostNumber(23.20))
			assertHTTPFloatPtr(t, "fleet estimated_cost_usd", fleet.EstimatedCostUSD, domain.CostNumber(13.86))
			assertHTTPFloatPtr(t, "fleet total_equivalent_cost_usd", fleet.TotalEquivalentCostUSD, domain.CostNumber(37.06))
			if fleet.ModelEquivalentCostSummaries["gpt-5.4"].Provenance != domain.CostProvenanceEstimated {
				t.Fatalf("fleet model gpt-5.4 equiv = %+v", fleet.ModelEquivalentCostSummaries["gpt-5.4"])
			}

			// 3. /api/progression
			var buckets []adaptershttp.TimeBucketMetricDTO
			if err := json.Unmarshal(get("/api/progression?granularity=daily").Body.Bytes(), &buckets); err != nil {
				t.Fatalf("unmarshal progression: %v", err)
			}
			if len(buckets) != 2 {
				t.Fatalf("expected 2 daily buckets, got %d", len(buckets))
			}
			if buckets[0].CostProvenance != domain.CostProvenanceHybrid || buckets[1].CostProvenance != domain.CostProvenanceEstimated {
				t.Fatalf("unexpected bucket provenances: [0]=%q [1]=%q", buckets[0].CostProvenance, buckets[1].CostProvenance)
			}
			if buckets[1].CumulativeCostProvenance != domain.CostProvenanceHybrid {
				t.Fatalf("bucket[1] cumulative provenance = %q, want hybrid", buckets[1].CumulativeCostProvenance)
			}
			assertHTTPFloatPtr(t, "bucket[1] cumulative total_equivalent", buckets[1].CumulativeTotalEquivalentCostUSD, domain.CostNumber(37.06))

			// 4. /api/comparison
			var rows []adaptershttp.TicketComparisonRowDTO
			if err := json.Unmarshal(get("/api/comparison").Body.Bytes(), &rows); err != nil {
				t.Fatalf("unmarshal comparison: %v", err)
			}
			if len(rows) != 2 {
				t.Fatalf("expected 2 comparison rows, got %d", len(rows))
			}
			rowByID := map[string]adaptershttp.TicketComparisonRowDTO{}
			for _, r := range rows {
				rowByID[r.TicketID] = r
			}
			if rowByID["AL-10"].CostProvenance != domain.CostProvenanceHybrid {
				t.Fatalf("comparison AL-10 provenance = %q, want hybrid", rowByID["AL-10"].CostProvenance)
			}
			assertHTTPFloatPtr(t, "comparison AL-10 total_equivalent", rowByID["AL-10"].TotalEquivalentCostUSD, domain.CostNumber(30.32))
			if rowByID["AL-11"].CostProvenance != domain.CostProvenanceEstimated {
				t.Fatalf("comparison AL-11 provenance = %q, want estimated", rowByID["AL-11"].CostProvenance)
			}
			assertHTTPFloatPtr(t, "comparison AL-11 total_equivalent", rowByID["AL-11"].TotalEquivalentCostUSD, domain.CostNumber(6.74))

			// 5. Ledger bytes on disk MUST be completely untouched by HTTP reads.
			al10After, err := os.ReadFile(al10Path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(al10Before, al10After) {
				t.Fatalf("HTTP reads modified persisted ledger file bytes on disk")
			}
		})
	}
}

func assertRawJSONNullKey(t *testing.T, context string, obj map[string]json.RawMessage, key string) {
	t.Helper()
	raw, exists := obj[key]
	if !exists {
		t.Fatalf("%s: expected explicit key %q in raw JSON, but key was omitted", context, key)
	}
	if string(bytes.TrimSpace(raw)) != "null" {
		t.Fatalf("%s: key %q = %s, want explicit null", context, key, string(raw))
	}
}

func TestHTTPEquivalentCostAllStatesAndExplicitNullKeys(t *testing.T) {
	const catalogA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	// Five tickets covering all 5 non-empty provenance states:
	// AL-10: hybrid (orchestrator/claude-sonnet-4 measured $23.20 + coder/gpt-5.4 estimated $7.12 => $30.32 equiv)
	// AL-11: estimated (coder/gpt-5.4 estimated $6.74)
	// AL-12: measured (planner/claude-sonnet-4 measured $4.50)
	// AL-13: partial (reviewer/claude-sonnet-4 measured $2.00 + coder/gpt-5.4 estimated $3.00 + tester/unpriced-model unavailable)
	// AL-14: unavailable (tester/unpriced-model unavailable)
	ledgersJSON := map[string]string{
		"AL-10": `{
  "schema_version": 6,
  "ticket_id": "AL-10",
  "ticket_title": "Hybrid Ticket",
  "ticket_type": "feat",
  "started_at": "2026-10-15T10:00:00Z",
  "completed_at": "2026-10-15T11:00:00Z",
  "updated_at": "2026-10-15T11:00:00Z",
  "code": {"lines_added": 50, "lines_deleted": 10, "test_lines_added": 25, "test_lines_deleted": 0, "prod_lines_added": 25, "prod_lines_deleted": 10, "files_changed": 2, "sdlc_iterations": 1},
  "sessions": [
    {
      "id": "ses-10-oc",
      "source": "opencode",
      "user": "dev-a",
      "invocations": [{
        "role": "orchestrator",
        "model": "claude-sonnet-4",
        "started_at": "2026-10-15T10:05:00Z",
        "completed_at": "2026-10-15T10:20:00Z",
        "active_duration_seconds": 900,
        "turns": 2,
        "tokens": {"input": 10000, "output": 2000, "cache_read": 5000, "reasoning": 0, "peak_context": 40000},
        "cost_usd": 23.20,
        "cost_evidence": {"state": "observed", "known_subtotal_usd": 23.20},
        "compactions": 0,
        "user_iterations": 1
      }],
      "tools": [],
      "anomalies": []
    },
    {
      "id": "ses-10-codex",
      "source": "codex",
      "user": "dev-a",
      "invocations": [{
        "role": "coder",
        "model": "gpt-5.4",
        "started_at": "2026-10-15T10:25:00Z",
        "completed_at": "2026-10-15T10:45:00Z",
        "active_duration_seconds": 1200,
        "turns": 4,
        "tokens": {"input": 200000, "output": 50000, "cache_read": 800000, "reasoning": 10000, "peak_context": 90000},
        "cost_usd": null,
        "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
        "cost_estimate": {
          "amount_usd": 7.12,
          "rate": {
            "matched_pattern": "gpt-5.4",
            "input_per_million": 2.0,
            "cache_read_per_million": 0.25,
            "output_per_million": 8.0,
            "effective_catalog_id": "` + catalogA + `"
          }
        },
        "compactions": 0,
        "user_iterations": 1
      }],
      "tools": [],
      "anomalies": []
    }
  ]
}`,
		"AL-11": `{
  "schema_version": 6,
  "ticket_id": "AL-11",
  "ticket_title": "Estimated Ticket",
  "ticket_type": "feat",
  "started_at": "2026-10-16T10:00:00Z",
  "completed_at": "2026-10-16T11:00:00Z",
  "updated_at": "2026-10-16T11:00:00Z",
  "code": {"lines_added": 20, "lines_deleted": 0, "test_lines_added": 10, "test_lines_deleted": 0, "prod_lines_added": 10, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
  "sessions": [{
    "id": "ses-11-codex",
    "source": "codex",
    "user": "dev-a",
    "invocations": [{
      "role": "coder",
      "model": "gpt-5.4",
      "started_at": "2026-10-16T10:10:00Z",
      "completed_at": "2026-10-16T10:40:00Z",
      "active_duration_seconds": 1800,
      "turns": 3,
      "tokens": {"input": 100000, "output": 20000, "cache_read": 400000, "reasoning": 5000, "peak_context": 60000},
      "cost_usd": null,
      "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
      "cost_estimate": {
        "amount_usd": 6.74,
        "rate": {
          "matched_pattern": "gpt-5.4",
          "input_per_million": 2.0,
          "cache_read_per_million": 0.25,
          "output_per_million": 8.0,
          "effective_catalog_id": "` + catalogA + `"
        }
      },
      "compactions": 0,
      "user_iterations": 1
    }],
    "tools": [],
    "anomalies": []
  }]
}`,
		"AL-12": `{
  "schema_version": 6,
  "ticket_id": "AL-12",
  "ticket_title": "Measured Ticket",
  "ticket_type": "fix",
  "started_at": "2026-10-17T10:00:00Z",
  "completed_at": "2026-10-17T11:00:00Z",
  "updated_at": "2026-10-17T11:00:00Z",
  "code": {"lines_added": 15, "lines_deleted": 5, "test_lines_added": 10, "test_lines_deleted": 0, "prod_lines_added": 5, "prod_lines_deleted": 5, "files_changed": 1, "sdlc_iterations": 1},
  "sessions": [{
    "id": "ses-12-oc",
    "source": "opencode",
    "user": "dev-a",
    "invocations": [{
      "role": "planner",
      "model": "claude-sonnet-4",
      "started_at": "2026-10-17T10:10:00Z",
      "completed_at": "2026-10-17T10:40:00Z",
      "active_duration_seconds": 600,
      "turns": 2,
      "tokens": {"input": 5000, "output": 1000, "cache_read": 2000, "reasoning": 0, "peak_context": 20000},
      "cost_usd": 4.50,
      "cost_evidence": {"state": "observed", "known_subtotal_usd": 4.50},
      "compactions": 0,
      "user_iterations": 1
    }],
    "tools": [],
    "anomalies": []
  }]
}`,
		"AL-13": `{
  "schema_version": 6,
  "ticket_id": "AL-13",
  "ticket_title": "Partial Ticket",
  "ticket_type": "feat",
  "started_at": "2026-10-18T10:00:00Z",
  "completed_at": "2026-10-18T11:00:00Z",
  "updated_at": "2026-10-18T11:00:00Z",
  "code": {"lines_added": 30, "lines_deleted": 2, "test_lines_added": 15, "test_lines_deleted": 0, "prod_lines_added": 15, "prod_lines_deleted": 2, "files_changed": 2, "sdlc_iterations": 1},
  "sessions": [{
    "id": "ses-13-mixed",
    "source": "codex",
    "user": "dev-a",
    "invocations": [
      {
        "role": "reviewer",
        "model": "claude-sonnet-4",
        "started_at": "2026-10-18T10:05:00Z",
        "completed_at": "2026-10-18T10:15:00Z",
        "active_duration_seconds": 300,
        "turns": 1,
        "tokens": {"input": 3000, "output": 500, "cache_read": 1000, "reasoning": 0, "peak_context": 15000},
        "cost_usd": 2.00,
        "cost_evidence": {"state": "observed", "known_subtotal_usd": 2.00},
        "compactions": 0,
        "user_iterations": 1
      },
      {
        "role": "coder",
        "model": "gpt-5.4",
        "started_at": "2026-10-18T10:20:00Z",
        "completed_at": "2026-10-18T10:35:00Z",
        "active_duration_seconds": 600,
        "turns": 2,
        "tokens": {"input": 100000, "output": 10000, "cache_read": 200000, "reasoning": 0, "peak_context": 40000},
        "cost_usd": null,
        "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
        "cost_estimate": {
          "amount_usd": 3.00,
          "rate": {
            "matched_pattern": "gpt-5.4",
            "input_per_million": 2.0,
            "cache_read_per_million": 0.25,
            "output_per_million": 8.0,
            "effective_catalog_id": "` + catalogA + `"
          }
        },
        "compactions": 0,
        "user_iterations": 1
      },
      {
        "role": "tester",
        "model": "unpriced-model",
        "started_at": "2026-10-18T10:40:00Z",
        "completed_at": "2026-10-18T10:50:00Z",
        "active_duration_seconds": 300,
        "turns": 1,
        "tokens": {"input": 4000, "output": 1000, "cache_read": 0, "reasoning": 0, "peak_context": 10000},
        "cost_usd": null,
        "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
        "compactions": 0,
        "user_iterations": 1
      }
    ],
    "tools": [],
    "anomalies": []
  }]
}`,
		"AL-14": `{
  "schema_version": 6,
  "ticket_id": "AL-14",
  "ticket_title": "Unavailable Ticket",
  "ticket_type": "chore",
  "started_at": "2026-10-19T10:00:00Z",
  "completed_at": "2026-10-19T11:00:00Z",
  "updated_at": "2026-10-19T11:00:00Z",
  "code": {"lines_added": 5, "lines_deleted": 0, "test_lines_added": 0, "test_lines_deleted": 0, "prod_lines_added": 5, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
  "sessions": [{
    "id": "ses-14-unavail",
    "source": "codex",
    "user": "dev-a",
    "invocations": [{
      "role": "tester",
      "model": "unpriced-model",
      "started_at": "2026-10-19T10:10:00Z",
      "completed_at": "2026-10-19T10:20:00Z",
      "active_duration_seconds": 300,
      "turns": 1,
      "tokens": {"input": 2000, "output": 400, "cache_read": 0, "reasoning": 0, "peak_context": 8000},
      "cost_usd": null,
      "cost_evidence": {"state": "unavailable", "known_subtotal_usd": null},
      "compactions": 0,
      "user_iterations": 1
    }],
    "tools": [],
    "anomalies": []
  }]
}`,
	}

	project := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	cfg := httpFixturePolicy()
	cfg.ConfigPath = filepath.Join(project, ".agenticlens-config.yaml")
	writeHTTPFixturePolicy(t, project, cfg)

	for id, body := range ledgersJSON {
		dir := filepath.Join(project, "workspaces", id)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}

	server := adaptershttp.NewServerWithConfig(project, fstest.MapFS{}, cfg)
	get := func(target string) []byte {
		t.Helper()
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s returned %d: %s", target, rec.Code, rec.Body.String())
		}
		return rec.Body.Bytes()
	}

	// 1. /api/workspaces (list endpoint) - typed checks + explicit null key checks
	wsListRawBytes := get("/api/workspaces")
	var wsList []adaptershttp.MetricsLedgerResponseDTO
	if err := json.Unmarshal(wsListRawBytes, &wsList); err != nil {
		t.Fatalf("unmarshal /api/workspaces: %v", err)
	}
	if len(wsList) != 5 {
		t.Fatalf("/api/workspaces returned %d ledgers, want 5", len(wsList))
	}
	var wsListRaw []map[string]json.RawMessage
	if err := json.Unmarshal(wsListRawBytes, &wsListRaw); err != nil {
		t.Fatalf("unmarshal raw /api/workspaces: %v", err)
	}
	rawLedgerByID := map[string]map[string]json.RawMessage{}
	for _, entry := range wsListRaw {
		var tid string
		if err := json.Unmarshal(entry["ticket_id"], &tid); err != nil {
			t.Fatalf("decode ticket_id in list: %v", err)
		}
		rawLedgerByID[tid] = entry
	}
	ledgerByID := map[string]adaptershttp.MetricsLedgerResponseDTO{}
	for _, l := range wsList {
		if l.TicketID != nil {
			ledgerByID[*l.TicketID] = l
		}
	}

	wantStates := map[string]struct {
		prov      domain.CostProvenance
		measured  *float64
		estimated *float64
		total     *float64
		nullKeys  []string
	}{
		"AL-10": {domain.CostProvenanceHybrid, domain.CostNumber(23.20), domain.CostNumber(7.12), domain.CostNumber(30.32), nil},
		"AL-11": {domain.CostProvenanceEstimated, nil, domain.CostNumber(6.74), domain.CostNumber(6.74), []string{"measured_cost_usd"}},
		"AL-12": {domain.CostProvenanceMeasured, domain.CostNumber(4.50), nil, domain.CostNumber(4.50), []string{"estimated_cost_usd"}},
		"AL-13": {domain.CostProvenancePartial, domain.CostNumber(2.00), domain.CostNumber(3.00), nil, []string{"total_equivalent_cost_usd"}},
		"AL-14": {domain.CostProvenanceUnavailable, nil, nil, nil, []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"}},
	}

	for id, want := range wantStates {
		got, ok := ledgerByID[id]
		if !ok {
			t.Fatalf("/api/workspaces missing %s", id)
		}
		if got.Totals.CostProvenance != want.prov || got.Totals.EquivalentCostSummary.Provenance != want.prov {
			t.Fatalf("/api/workspaces %s provenance = %q / %q, want %q", id, got.Totals.CostProvenance, got.Totals.EquivalentCostSummary.Provenance, want.prov)
		}
		assertHTTPFloatPtr(t, id+" list measured", got.Totals.MeasuredCostUSD, want.measured)
		assertHTTPFloatPtr(t, id+" list estimated", got.Totals.EstimatedCostUSD, want.estimated)
		assertHTTPFloatPtr(t, id+" list total_equivalent", got.Totals.TotalEquivalentCostUSD, want.total)

		var rawTotals map[string]json.RawMessage
		if err := json.Unmarshal(rawLedgerByID[id]["totals"], &rawTotals); err != nil {
			t.Fatalf("unmarshal raw totals for %s: %v", id, err)
		}
		var rawEquiv map[string]json.RawMessage
		if err := json.Unmarshal(rawTotals["equivalent_cost_summary"], &rawEquiv); err != nil {
			t.Fatalf("unmarshal raw equivalent_cost_summary for %s: %v", id, err)
		}
		for _, k := range want.nullKeys {
			assertRawJSONNullKey(t, "/api/workspaces "+id+" totals", rawTotals, k)
			assertRawJSONNullKey(t, "/api/workspaces "+id+" equivalent_cost_summary", rawEquiv, k)
		}
	}

	// 2. /api/workspaces/{id} - inspect Roles, Models, and Threads[].Roles[].Models[] + raw null keys
	for id, want := range wantStates {
		detailBytes := get("/api/workspaces/" + id)
		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(detailBytes, &detail); err != nil {
			t.Fatalf("unmarshal detail %s: %v", id, err)
		}
		if detail.Ledger.Totals.CostProvenance != want.prov {
			t.Fatalf("detail %s provenance = %q, want %q", id, detail.Ledger.Totals.CostProvenance, want.prov)
		}
		assertHTTPFloatPtr(t, "detail "+id+" total_equivalent", detail.Ledger.Totals.TotalEquivalentCostUSD, want.total)
		if len(detail.Ledger.Roles) == 0 || len(detail.Ledger.Threads) == 0 {
			t.Fatalf("detail %s missing roles or threads", id)
		}

		// Inspect raw detail totals for explicit null keys
		var rawDetail map[string]json.RawMessage
		if err := json.Unmarshal(detailBytes, &rawDetail); err != nil {
			t.Fatal(err)
		}
		var rawLedger map[string]json.RawMessage
		if err := json.Unmarshal(rawDetail["ledger"], &rawLedger); err != nil {
			t.Fatal(err)
		}
		var rawTotals map[string]json.RawMessage
		if err := json.Unmarshal(rawLedger["totals"], &rawTotals); err != nil {
			t.Fatal(err)
		}
		for _, k := range want.nullKeys {
			assertRawJSONNullKey(t, "/api/workspaces/"+id+" totals", rawTotals, k)
		}

		if id == "AL-10" {
			roleByName := map[domain.AgentRole]string{}
			for _, r := range detail.Ledger.Roles {
				roleByName[r.Role] = string(r.CostProvenance)
			}
			if roleByName["orchestrator"] != "measured" || roleByName["coder"] != "estimated" {
				t.Fatalf("AL-10 detail roles provenances = %+v", roleByName)
			}
			// Verify Threads[].Roles[].Models[] carries AppliedRates and equivalent summary for gpt-5.4
			var foundThreadEstimated bool
			for _, th := range detail.Ledger.Threads {
				for _, r := range th.Roles {
					for _, m := range r.Models {
						if m.Model == "gpt-5.4" {
							foundThreadEstimated = true
							if m.CostProvenance != domain.CostProvenanceEstimated {
								t.Fatalf("AL-10 thread gpt-5.4 provenance = %q, want estimated", m.CostProvenance)
							}
							assertHTTPFloatPtr(t, "AL-10 thread gpt-5.4 estimated", m.EstimatedCostUSD, domain.CostNumber(7.12))
							if len(m.AppliedRates) != 1 || m.AppliedRates[0].EffectiveCatalogID != catalogA {
								t.Fatalf("AL-10 thread gpt-5.4 applied_rates = %+v", m.AppliedRates)
							}
						}
					}
				}
			}
			if !foundThreadEstimated {
				t.Fatalf("AL-10 thread missing estimated gpt-5.4 model metric")
			}
		}

		if id == "AL-14" {
			// Check explicit null keys on Roles[0] and Threads[0].Roles[0].Models[0]
			var rawRoles []map[string]json.RawMessage
			if err := json.Unmarshal(rawLedger["roles"], &rawRoles); err != nil || len(rawRoles) == 0 {
				t.Fatalf("AL-14 raw roles decode failed: %v", err)
			}
			for _, k := range []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"} {
				assertRawJSONNullKey(t, "AL-14 roles[0]", rawRoles[0], k)
			}
			var rawThreads []struct {
				Roles []struct {
					Models []map[string]json.RawMessage `json:"models"`
				} `json:"roles"`
			}
			if err := json.Unmarshal(rawLedger["threads"], &rawThreads); err != nil || len(rawThreads) == 0 || len(rawThreads[0].Roles) == 0 || len(rawThreads[0].Roles[0].Models) == 0 {
				t.Fatalf("AL-14 raw threads decode failed: %v", err)
			}
			for _, k := range []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"} {
				assertRawJSONNullKey(t, "AL-14 threads[0].roles[0].models[0]", rawThreads[0].Roles[0].Models[0], k)
			}
		}
	}

	// 3. /api/fleet - partial overall (2 unpriced unavailable tester invocations remain), plus Role/Model maps
	fleetBytes := get("/api/fleet")
	var fleet adaptershttp.FleetSummaryDTO
	if err := json.Unmarshal(fleetBytes, &fleet); err != nil {
		t.Fatalf("unmarshal /api/fleet: %v", err)
	}
	if fleet.CostProvenance != domain.CostProvenancePartial {
		t.Fatalf("fleet provenance = %q, want partial", fleet.CostProvenance)
	}
	// Measured: 23.20 + 4.50 + 2.00 = 29.70; Estimated: 7.12 + 6.74 + 3.00 = 16.86
	assertHTTPFloatPtr(t, "fleet measured_cost_usd", fleet.MeasuredCostUSD, domain.CostNumber(29.70))
	assertHTTPFloatPtr(t, "fleet estimated_cost_usd", fleet.EstimatedCostUSD, domain.CostNumber(16.86))
	assertHTTPFloatPtr(t, "fleet total_equivalent_cost_usd", fleet.TotalEquivalentCostUSD, nil)
	assertHTTPFloatPtr(t, "fleet known_subtotal_usd", fleet.EquivalentCostSummary.KnownSubtotalUSD, domain.CostNumber(46.56))

	if fleet.RoleEquivalentCostSummaries["orchestrator"].Provenance != domain.CostProvenanceMeasured {
		t.Fatalf("fleet role orchestrator = %+v", fleet.RoleEquivalentCostSummaries["orchestrator"])
	}
	if fleet.RoleEquivalentCostSummaries["planner"].Provenance != domain.CostProvenanceMeasured {
		t.Fatalf("fleet role planner = %+v", fleet.RoleEquivalentCostSummaries["planner"])
	}
	if fleet.RoleEquivalentCostSummaries["coder"].Provenance != domain.CostProvenanceEstimated {
		t.Fatalf("fleet role coder = %+v", fleet.RoleEquivalentCostSummaries["coder"])
	}
	if fleet.RoleEquivalentCostSummaries["tester"].Provenance != domain.CostProvenanceUnavailable {
		t.Fatalf("fleet role tester = %+v", fleet.RoleEquivalentCostSummaries["tester"])
	}
	if fleet.ModelEquivalentCostSummaries["claude-sonnet-4"].Provenance != domain.CostProvenanceMeasured {
		t.Fatalf("fleet model claude-sonnet-4 = %+v", fleet.ModelEquivalentCostSummaries["claude-sonnet-4"])
	}
	if fleet.ModelEquivalentCostSummaries["gpt-5.4"].Provenance != domain.CostProvenanceEstimated {
		t.Fatalf("fleet model gpt-5.4 = %+v", fleet.ModelEquivalentCostSummaries["gpt-5.4"])
	}
	if fleet.ModelEquivalentCostSummaries["unpriced-model"].Provenance != domain.CostProvenanceUnavailable {
		t.Fatalf("fleet model unpriced-model = %+v", fleet.ModelEquivalentCostSummaries["unpriced-model"])
	}

	// Verify explicit null wire keys on /api/fleet
	var rawFleet map[string]json.RawMessage
	if err := json.Unmarshal(fleetBytes, &rawFleet); err != nil {
		t.Fatal(err)
	}
	assertRawJSONNullKey(t, "/api/fleet", rawFleet, "total_equivalent_cost_usd")
	var rawRoleEquiv map[string]map[string]json.RawMessage
	if err := json.Unmarshal(rawFleet["role_equivalent_cost_summaries"], &rawRoleEquiv); err != nil {
		t.Fatal(err)
	}
	assertRawJSONNullKey(t, "/api/fleet role_equivalent_cost_summaries[coder]", rawRoleEquiv["coder"], "measured_cost_usd")
	assertRawJSONNullKey(t, "/api/fleet role_equivalent_cost_summaries[planner]", rawRoleEquiv["planner"], "estimated_cost_usd")
	for _, k := range []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"} {
		assertRawJSONNullKey(t, "/api/fleet role_equivalent_cost_summaries[tester]", rawRoleEquiv["tester"], k)
	}
	var rawModelEquiv map[string]map[string]json.RawMessage
	if err := json.Unmarshal(rawFleet["model_equivalent_cost_summaries"], &rawModelEquiv); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"} {
		assertRawJSONNullKey(t, "/api/fleet model_equivalent_cost_summaries[unpriced-model]", rawModelEquiv["unpriced-model"], k)
	}

	// 4. /api/progression?granularity=daily - 5 daily buckets matching hybrid, estimated, measured, partial, unavailable
	progBytes := get("/api/progression?granularity=daily")
	var buckets []adaptershttp.TimeBucketMetricDTO
	if err := json.Unmarshal(progBytes, &buckets); err != nil {
		t.Fatalf("unmarshal /api/progression: %v", err)
	}
	if len(buckets) != 5 {
		t.Fatalf("/api/progression returned %d buckets, want 5", len(buckets))
	}
	wantBucketProvs := []domain.CostProvenance{
		domain.CostProvenanceHybrid,
		domain.CostProvenanceEstimated,
		domain.CostProvenanceMeasured,
		domain.CostProvenancePartial,
		domain.CostProvenanceUnavailable,
	}
	wantCumProvs := []domain.CostProvenance{
		domain.CostProvenanceHybrid,
		domain.CostProvenanceHybrid,
		domain.CostProvenanceHybrid,
		domain.CostProvenancePartial,
		domain.CostProvenancePartial,
	}
	var rawBuckets []map[string]json.RawMessage
	if err := json.Unmarshal(progBytes, &rawBuckets); err != nil {
		t.Fatal(err)
	}
	for i := range wantBucketProvs {
		if buckets[i].CostProvenance != wantBucketProvs[i] || buckets[i].CumulativeCostProvenance != wantCumProvs[i] {
			t.Fatalf("bucket[%d] prov=%q cum=%q, want %q / %q", i, buckets[i].CostProvenance, buckets[i].CumulativeCostProvenance, wantBucketProvs[i], wantCumProvs[i])
		}
	}
	assertRawJSONNullKey(t, "progression bucket[1] (estimated)", rawBuckets[1], "measured_cost_usd")
	assertRawJSONNullKey(t, "progression bucket[2] (measured)", rawBuckets[2], "estimated_cost_usd")
	assertRawJSONNullKey(t, "progression bucket[3] (partial)", rawBuckets[3], "total_equivalent_cost_usd")
	assertRawJSONNullKey(t, "progression bucket[3] cumulative (partial)", rawBuckets[3], "cumulative_total_equivalent_cost_usd")
	for _, k := range []string{"measured_cost_usd", "estimated_cost_usd", "total_equivalent_cost_usd"} {
		assertRawJSONNullKey(t, "progression bucket[4] (unavailable)", rawBuckets[4], k)
	}

	// 5. /api/comparison - all 5 rows + explicit null wire keys
	compBytes := get("/api/comparison")
	var compRows []adaptershttp.TicketComparisonRowDTO
	if err := json.Unmarshal(compBytes, &compRows); err != nil {
		t.Fatalf("unmarshal /api/comparison: %v", err)
	}
	if len(compRows) != 5 {
		t.Fatalf("/api/comparison returned %d rows, want 5", len(compRows))
	}
	var rawCompRows []map[string]json.RawMessage
	if err := json.Unmarshal(compBytes, &rawCompRows); err != nil {
		t.Fatal(err)
	}
	rawCompByID := map[string]map[string]json.RawMessage{}
	for _, r := range rawCompRows {
		var tid string
		if err := json.Unmarshal(r["ticket_id"], &tid); err != nil {
			t.Fatal(err)
		}
		rawCompByID[tid] = r
	}
	compByID := map[string]adaptershttp.TicketComparisonRowDTO{}
	for _, r := range compRows {
		compByID[r.TicketID] = r
	}
	for id, want := range wantStates {
		row := compByID[id]
		if row.CostProvenance != want.prov {
			t.Fatalf("comparison %s provenance = %q, want %q", id, row.CostProvenance, want.prov)
		}
		assertHTTPFloatPtr(t, "comparison "+id+" measured", row.MeasuredCostUSD, want.measured)
		assertHTTPFloatPtr(t, "comparison "+id+" estimated", row.EstimatedCostUSD, want.estimated)
		assertHTTPFloatPtr(t, "comparison "+id+" total_equivalent", row.TotalEquivalentCostUSD, want.total)
		for _, k := range want.nullKeys {
			assertRawJSONNullKey(t, "/api/comparison "+id, rawCompByID[id], k)
		}
	}
}
