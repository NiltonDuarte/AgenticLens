package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

type rawPricingLedger struct {
	Sessions []rawPricingSession `json:"sessions"`
}

type rawPricingSession struct {
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`
	Invocations []rawPricingInvocation `json:"invocations"`
}

type rawPricingInvocation struct {
	Model        string               `json:"model"`
	CostUSD      json.RawMessage      `json:"cost_usd"`
	CostEvidence *domain.CostEvidence `json:"cost_evidence"`
	CostEstimate json.RawMessage      `json:"cost_estimate"`
}

type pricingCLIFixture struct {
	project  string
	home     string
	output   string
	delivery string
}

func newPricingCLIFixture(t *testing.T, modelPricingYAML string) pricingCLIFixture {
	t.Helper()
	f := pricingCLIFixture{
		project: t.TempDir(),
		home:    t.TempDir(),
	}
	f.output = filepath.Join(f.project, "records", "agentic_metrics.json")
	f.delivery = filepath.Join(f.project, "records", "delivery.md")

	sharedCfg := claudeCLISharedConfig()
	if strings.TrimSpace(modelPricingYAML) != "" {
		sharedCfg += "\n" + strings.TrimSpace(modelPricingYAML) + "\n"
	}
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), []byte(sharedCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	localCfg := fmt.Sprintf(
		"telemetry:\n  default_source: claude\n  sources:\n    claude:\n      home: %q\n  capture:\n    enabled: false\n    state_dir: .agenticlens-capture\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\nagenticlens:\n  state_dir: %q\n",
		f.home,
		filepath.Join(f.project, ".publication-state"),
	)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(localCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f pricingCLIFixture) runGenerate(t *testing.T, ticketID, title string, extra ...string) (string, error) {
	t.Helper()
	args := append([]string{
		"generate",
		"--ticket-id", ticketID,
		"--ticket-title", title,
		"--ticket-type", "feat",
		"--json-output", f.output,
		"--append", f.delivery,
	}, extra...)
	cmd := exec.Command(cliBinaryPath, args...)
	cmd.Dir = f.project
	prepareScopedGenerateFixture(t, cmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (f pricingCLIFixture) readRawLedger(t *testing.T) ([]byte, rawPricingLedger, domain.AgenticFactLedger) {
	t.Helper()
	data, err := os.ReadFile(f.output)
	if err != nil {
		t.Fatalf("read ledger %s: %v", f.output, err)
	}
	var raw rawPricingLedger
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw ledger: %v", err)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatalf("unmarshal dto ledger: %v", err)
	}
	return data, raw, dto.ToDomain()
}

func writeObservedOpenCodeDB(t *testing.T, projectDir, sessionID, model string, input, output, cacheRead int64, costUSD float64) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "observed_opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	defer db.Close()

	schema := `
		CREATE TABLE session (
			id TEXT PRIMARY KEY,
			parent_id TEXT,
			agent TEXT,
			model TEXT,
			title TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			directory TEXT
		);
		CREATE TABLE message (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
		CREATE TABLE part (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			message_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create sqlite schema: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated, directory) VALUES (?, NULL, 'orchestrator', ?, 'Observed pricing fixture', 1760000000000, 1760000030000, ?)`,
		sessionID, model, projectDir,
	); err != nil {
		t.Fatalf("insert session: %v", err)
	}
	msgData := fmt.Sprintf(
		`{"role":"assistant","modelID":%q,"time":{"created":1760000005000,"completed":1760000025000},"tokens":{"input":%d,"output":%d,"cache":{"read":%d},"reasoning":0},"cost":%.6f}`,
		model, input, output, cacheRead, costUSD,
	)
	if _, err := db.Exec(
		`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('msg_obs_1', ?, 1760000005000, 1760000025000, ?)`,
		sessionID, msgData,
	); err != nil {
		t.Fatalf("insert message: %v", err)
	}
	return dbPath
}

func TestGeneratePricingUnavailableSessionPersistsEstimateAndRendersDelivery(t *testing.T) {
	pricingYAML := `
model_pricing:
  rates:
    - model: claude-sonnet-4-5
      input_per_million: 3.0
      cache_read_per_million: 0.30
      output_per_million: 15.0
    - model: "*gpt-5.4*"
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`
	expectedCatalogID := (domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "claude-sonnet-4-5", InputPerMillion: 3.0, CacheReadPerMillion: 0.30, OutputPerMillion: 15.0},
			{Model: "*gpt-5.4*", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
		},
	}).EffectiveCatalogID()

	f := newPricingCLIFixture(t, pricingYAML)
	if err := os.MkdirAll(filepath.Dir(f.delivery), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.delivery, []byte("# feat: pricing delivery\n\nInitial delivery notes.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Expected amount: (400_000 * 3.0 + 1_200_000 * 0.30 + 150_000 * 15.0) / 1_000_000 = 1.20 + 0.36 + 2.25 = 3.81
	const rootID = "claude-pricing-root"
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-pricing", rootID), rootID, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:10Z", "r-1", "claude-sonnet-4-5", "end_turn", 400_000, 150_000, 1_200_000),
	})

	out, err := f.runGenerate(t, "TEST-206", "Pricing E2E", "--source", "claude", "--claude-home", f.home, "--session-id", rootID)
	if err != nil {
		t.Fatalf("generate failed: %v\n%s", err, out)
	}

	_, raw, ledger := f.readRawLedger(t)
	if len(raw.Sessions) != 1 || len(raw.Sessions[0].Invocations) != 1 {
		t.Fatalf("expected 1 session with 1 invocation, got %+v", raw.Sessions)
	}
	rawInv := raw.Sessions[0].Invocations[0]
	if string(bytes.TrimSpace(rawInv.CostUSD)) != "null" {
		t.Fatalf("wire cost_usd = %s, want null", string(rawInv.CostUSD))
	}
	if rawInv.CostEvidence == nil || rawInv.CostEvidence.State != domain.CostUnavailable {
		t.Fatalf("wire cost_evidence = %+v, want state=%q", rawInv.CostEvidence, domain.CostUnavailable)
	}

	inv := ledger.Sessions[0].Invocations[0]
	if inv.CostEstimate == nil {
		t.Fatalf("expected persisted cost_estimate, got nil")
	}
	wantEstimate := domain.CostEstimate{
		AmountUSD: 3.81,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "claude-sonnet-4-5",
			InputPerMillion:     3.0,
			CacheReadPerMillion: 0.30,
			OutputPerMillion:    15.0,
			EffectiveCatalogID:  expectedCatalogID,
		},
	}
	if *inv.CostEstimate != wantEstimate {
		t.Fatalf("cost_estimate = %+v, want %+v", *inv.CostEstimate, wantEstimate)
	}

	deliveryBytes, err := os.ReadFile(f.delivery)
	if err != nil {
		t.Fatalf("read delivery.md: %v", err)
	}
	md := string(deliveryBytes)
	wantRateLine := fmt.Sprintf(
		"- `claude-sonnet-4-5`: $3.00 / $0.30 / $15.00 per 1M tokens (input / cache_read / output; catalog `%s`)",
		expectedCatalogID,
	)
	for _, want := range []string{
		"| $3.81 EST |",
		"| **$3.81 EST** |",
		"**Applied API-equivalent rate card:**",
		wantRateLine,
		"currently exclude `CacheWrite` tokens",
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("delivery.md missing %q:\n%s", want, md)
		}
	}
}

func TestGeneratePricingObservedAndHybridSessions(t *testing.T) {
	pricingYAML := `
model_pricing:
  rates:
    - model: claude-sonnet-4-5
      input_per_million: 3.0
      cache_read_per_million: 0.30
      output_per_million: 15.0
    - model: "*gpt-5.4*"
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`
	expectedCatalogID := (domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "claude-sonnet-4-5", InputPerMillion: 3.0, CacheReadPerMillion: 0.30, OutputPerMillion: 15.0},
			{Model: "*gpt-5.4*", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
		},
	}).EffectiveCatalogID()

	f := newPricingCLIFixture(t, pricingYAML)
	if err := os.MkdirAll(filepath.Dir(f.delivery), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.delivery, []byte("# feat: hybrid pricing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Step 1: Observed-cost OpenCode session using a model that also appears in model_pricing.rates.
	// Measured spend must remain authoritative and must never receive a cost_estimate.
	const obsSessionID = "ses_observed_pricing"
	dbPath := writeObservedOpenCodeDB(t, f.project, obsSessionID, "claude-sonnet-4-5", 10_000, 2_000, 5_000, 4.25)

	out, err := f.runGenerate(t, "TEST-207", "Hybrid pricing", "--source", "opencode", "--db-path", dbPath, "--session-id", obsSessionID)
	if err != nil {
		t.Fatalf("observed generate failed: %v\n%s", err, out)
	}

	_, rawObs, ledgerObs := f.readRawLedger(t)
	if len(rawObs.Sessions) != 1 || len(rawObs.Sessions[0].Invocations) != 1 {
		t.Fatalf("expected 1 observed session, got %+v", rawObs.Sessions)
	}
	if len(bytes.TrimSpace(rawObs.Sessions[0].Invocations[0].CostEstimate)) != 0 {
		t.Fatalf("observed invocation must omit cost_estimate, got %s", string(rawObs.Sessions[0].Invocations[0].CostEstimate))
	}
	if rawObs.Sessions[0].Invocations[0].CostEvidence == nil || rawObs.Sessions[0].Invocations[0].CostEvidence.State != domain.CostObserved {
		t.Fatalf("observed invocation cost_evidence = %+v, want observed", rawObs.Sessions[0].Invocations[0].CostEvidence)
	}
	if ledgerObs.Sessions[0].Invocations[0].CostUSD != 4.25 {
		t.Fatalf("observed cost_usd = %v, want 4.25", ledgerObs.Sessions[0].Invocations[0].CostUSD)
	}

	obsMDBytes, err := os.ReadFile(f.delivery)
	if err != nil {
		t.Fatalf("read observed delivery.md: %v", err)
	}
	obsMD := string(obsMDBytes)
	if !strings.Contains(obsMD, "| $4.25 |") || !strings.Contains(obsMD, "| **$4.25** |") {
		t.Fatalf("observed delivery.md missing $4.25 measured cost:\n%s", obsMD)
	}
	if strings.Contains(obsMD, "EST") || strings.Contains(obsMD, "Applied API-equivalent rate card") || strings.Contains(obsMD, "CacheWrite") {
		t.Fatalf("pure observed delivery.md must not emit EST or rate-card footnote:\n%s", obsMD)
	}

	// Step 2: Add an unavailable-cost Claude Code session matched via glob rule "*gpt-5.4*".
	// Expected estimate: (200_000 * 2.0 + 800_000 * 0.25 + 50_000 * 8.0) / 1_000_000 = 0.40 + 0.20 + 0.40 = 1.00.
	const claudeSessionID = "claude-hybrid-root"
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-hybrid", claudeSessionID), claudeSessionID, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T11:00:00Z"),
		claudeCLIResponse("2026-01-01T11:00:15Z", "r-1", "openai/gpt-5.4-codex", "end_turn", 200_000, 50_000, 800_000),
	})

	out, err = f.runGenerate(t, "TEST-207", "Hybrid pricing", "--source", "claude", "--claude-home", f.home, "--session-id", claudeSessionID)
	if err != nil {
		t.Fatalf("hybrid generate failed: %v\n%s", err, out)
	}

	_, rawHybrid, ledgerHybrid := f.readRawLedger(t)
	if len(rawHybrid.Sessions) != 2 {
		t.Fatalf("expected 2 sessions in hybrid ledger, got %d", len(rawHybrid.Sessions))
	}

	rawByID := make(map[string]rawPricingSession, len(rawHybrid.Sessions))
	for _, s := range rawHybrid.Sessions {
		rawByID[s.ID] = s
	}
	domByID := claudeCLISessions(ledgerHybrid)

	obsRawInv := rawByID[obsSessionID].Invocations[0]
	if obsRawInv.CostEvidence == nil || obsRawInv.CostEvidence.State != domain.CostObserved || len(bytes.TrimSpace(obsRawInv.CostEstimate)) != 0 {
		t.Fatalf("hybrid observed session corrupted: %+v", obsRawInv)
	}
	if domByID[obsSessionID].Invocations[0].CostUSD != 4.25 || domByID[obsSessionID].Invocations[0].CostEstimate != nil {
		t.Fatalf("hybrid observed domain invocation corrupted: %+v", domByID[obsSessionID].Invocations[0])
	}

	estRawInv := rawByID[claudeSessionID].Invocations[0]
	if string(bytes.TrimSpace(estRawInv.CostUSD)) != "null" {
		t.Fatalf("hybrid estimated wire cost_usd = %s, want null", string(estRawInv.CostUSD))
	}
	if estRawInv.CostEvidence == nil || estRawInv.CostEvidence.State != domain.CostUnavailable {
		t.Fatalf("hybrid estimated wire cost_evidence = %+v, want unavailable", estRawInv.CostEvidence)
	}
	estDomInv := domByID[claudeSessionID].Invocations[0]
	if estDomInv.CostEstimate == nil {
		t.Fatalf("hybrid estimated invocation missing cost_estimate")
	}
	wantHybridEst := domain.CostEstimate{
		AmountUSD: 1.0,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "*gpt-5.4*",
			InputPerMillion:     2.0,
			CacheReadPerMillion: 0.25,
			OutputPerMillion:    8.0,
			EffectiveCatalogID:  expectedCatalogID,
		},
	}
	if *estDomInv.CostEstimate != wantHybridEst {
		t.Fatalf("hybrid cost_estimate = %+v, want %+v", *estDomInv.CostEstimate, wantHybridEst)
	}

	hybridMDBytes, err := os.ReadFile(f.delivery)
	if err != nil {
		t.Fatalf("read hybrid delivery.md: %v", err)
	}
	hybridMD := string(hybridMDBytes)
	wantGlobRateLine := fmt.Sprintf(
		"- `*gpt-5.4*`: $2.00 / $0.25 / $8.00 per 1M tokens (input / cache_read / output; catalog `%s`)",
		expectedCatalogID,
	)
	for _, want := range []string{
		"| $4.25 |",
		"| $1.00 EST |",
		"**$4.25 measured + $1.00 EST ($5.25 equiv)**",
		"**Applied API-equivalent rate card:**",
		wantGlobRateLine,
		"currently exclude `CacheWrite` tokens",
	} {
		if !strings.Contains(hybridMD, want) {
			t.Fatalf("hybrid delivery.md missing %q:\n%s", want, hybridMD)
		}
	}
	if strings.Contains(hybridMD, "- `claude-sonnet-4-5`:") {
		t.Fatalf("hybrid delivery.md must not list observed-only model in applied rate card:\n%s", hybridMD)
	}
}

func TestGeneratePricingAmbiguousGlobFailsClosed(t *testing.T) {
	// Both "gpt-5*" and "*4-pro" have 5 literal runes and normalized length 6,
	// so they tie deterministically on "gpt-5.4-pro" and must fail closed.
	ambiguousYAML := `
model_pricing:
  rates:
    - model: "gpt-5*"
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
    - model: "*4-pro"
      input_per_million: 4.0
      cache_read_per_million: 0.50
      output_per_million: 12.0
`

	t.Run("fresh generation fails without writing ledger or delivery", func(t *testing.T) {
		f := newPricingCLIFixture(t, ambiguousYAML)
		const rootID = "claude-ambiguous-fresh"
		writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-ambiguous", rootID), rootID, "", []claudeCLIRecord{
			claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
			claudeCLIResponse("2026-01-01T10:00:05Z", "r-1", "gpt-5.4-pro", "end_turn", 100_000, 10_000, 50_000),
		})

		out, err := f.runGenerate(t, "TEST-208", "Ambiguous pricing", "--source", "claude", "--claude-home", f.home, "--session-id", rootID)
		if err == nil {
			t.Fatalf("expected generate to fail on ambiguous glob match, succeeded with output:\n%s", out)
		}
		for _, want := range []string{"ambiguous model pricing match", "gpt-5.4-pro", "*4-pro", "gpt-5*"} {
			if !strings.Contains(out, want) {
				t.Fatalf("diagnostic output missing %q:\n%s", want, out)
			}
		}
		if _, statErr := os.Stat(f.output); !os.IsNotExist(statErr) {
			t.Fatalf("ambiguous pricing must not publish a ledger, stat err = %v", statErr)
		}
		if _, statErr := os.Stat(f.delivery); !os.IsNotExist(statErr) {
			t.Fatalf("ambiguous pricing must not create delivery.md, stat err = %v", statErr)
		}
	})

	t.Run("refresh preserves existing ledger and delivery unchanged", func(t *testing.T) {
		f := newPricingCLIFixture(t, ambiguousYAML)
		if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
			t.Fatal(err)
		}
		version := domain.CurrentSchemaVersion
		ticket, title := "TEST-208", "Ambiguous pricing"
		if err := adapters.SaveFactLedger(f.output, domain.AgenticFactLedger{
			SchemaVersion: &version,
			TicketID:      &ticket,
			TicketTitle:   &title,
			Sessions: []domain.SessionFact{
				{
					ID:     "existing-owner",
					Source: domain.SourceCodex,
					User:   "remote",
					Invocations: []domain.InvocationFact{
						{Role: "coder", Model: "codex-model", Tokens: domain.InvocationTokens{Input: 100}},
					},
				},
			},
		}); err != nil {
			t.Fatal(err)
		}
		initialDelivery := []byte("# Existing delivery\n\nUntouched content.\n")
		if err := os.WriteFile(f.delivery, initialDelivery, 0o600); err != nil {
			t.Fatal(err)
		}
		beforeLedger := readCaptureFile(t, f.output)

		const rootID = "claude-ambiguous-refresh"
		writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-ambiguous-refresh", rootID), rootID, "", []claudeCLIRecord{
			claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
			claudeCLIResponse("2026-01-01T10:00:05Z", "r-1", "gpt-5.4-pro", "end_turn", 100_000, 10_000, 50_000),
		})

		out, err := f.runGenerate(t, "TEST-208", "Ambiguous pricing", "--source", "claude", "--claude-home", f.home, "--session-id", rootID)
		if err == nil {
			t.Fatalf("expected refresh generate to fail on ambiguous glob match, succeeded with output:\n%s", out)
		}
		if !strings.Contains(out, "ambiguous model pricing match") {
			t.Fatalf("diagnostic output missing ambiguous pricing error:\n%s", out)
		}
		if !bytes.Equal(beforeLedger, readCaptureFile(t, f.output)) {
			t.Fatalf("ambiguous pricing failure modified existing ledger")
		}
		if !bytes.Equal(initialDelivery, readCaptureFile(t, f.delivery)) {
			t.Fatalf("ambiguous pricing failure modified existing delivery.md")
		}
	})
}

func TestGeneratePricingCaptureAutoPersistsEstimatesAndRecordsPublication(t *testing.T) {
	pricingYAML := `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`
	expectedCatalogID := (domain.ModelPricingConfig{
		Rates: []domain.ModelPricingRate{
			{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
		},
	}).EffectiveCatalogID()

	project := t.TempDir()
	store := t.TempDir()
	state := filepath.Join(project, ".agenticlens-capture")
	output := filepath.Join(project, "records", "TEST-209", "agentic_metrics.json")
	delivery := filepath.Join(project, "records", "TEST-209", "delivery.md")
	if err := os.MkdirAll(filepath.Dir(output), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(delivery, []byte("# feat: capture-auto pricing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	sharedCfg := strings.Replace(claudeCLISharedConfig(), "agenticlens:\n", "agenticlens:\n  pending_dir: .agenticlens-pending\n", 1)
	sharedCfg = strings.Replace(sharedCfg, "max_depth: 0", "max_depth: 1", 1) + "\n" + strings.TrimSpace(pricingYAML) + "\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(sharedCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	localCfg := fmt.Sprintf(
		"telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n  capture:\n    enabled: true\n    state_dir: %q\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\nagenticlens:\n  state_dir: %q\n",
		store,
		state,
		filepath.Join(project, ".publication-state"),
	)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(localCfg), 0o600); err != nil {
		t.Fatal(err)
	}

	// Write a Codex rollout JSONL with 1,000,000 total input (800,000 cached => 200,000 uncached) and 50,000 output.
	// Expected estimate: (200_000 * 2.0 + 800_000 * 0.25 + 50_000 * 8.0) / 1_000_000 = 1.00.
	rolloutPath := filepath.Join(store, "sessions", "capture-auto-root.jsonl")
	if err := os.MkdirAll(filepath.Dir(rolloutPath), 0o700); err != nil {
		t.Fatal(err)
	}
	rollout := fmt.Sprintf(
		`{"timestamp":"2026-01-01T10:00:00Z","type":"session_meta","payload":{"id":"capture-auto-root","cwd":%q,"timestamp":"2026-01-01T10:00:00Z"}}`+"\n"+
			`{"timestamp":"2026-01-01T10:00:01Z","type":"turn_context","payload":{"model":"gpt-5.4","turn_id":"turn-1"}}`+"\n"+
			`{"timestamp":"2026-01-01T10:00:02Z","type":"token_usage_record","payload":{"thread_id":"capture-auto-root","model":"gpt-5.4","turn_id":"turn-1","usage":{"input_tokens":1000000,"cached_input_tokens":800000,"output_tokens":50000}}}`+"\n",
		project,
	)
	if err := os.WriteFile(rolloutPath, []byte(rollout), 0o600); err != nil {
		t.Fatal(err)
	}

	registry, err := capture.OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: "capture-auto-root"}
	row, err := registry.Register(context.Background(), identity, rolloutPath, "SessionEnd")
	if err != nil {
		registry.Close()
		t.Fatal(err)
	}
	if _, err := registry.Bind(context.Background(), identity, "TEST-209", row.Request); err != nil {
		registry.Close()
		t.Fatal(err)
	}
	if err := registry.Close(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--capture-auto",
		"--source", "codex",
		"--capture-dir", state,
		"--ticket-id", "TEST-209",
		"--ticket-title", "Capture auto pricing",
		"--ticket-type", "feat",
		"--json-output", output,
		"--append", delivery,
	)
	cmd.Dir = project
	prepareScopedGenerateFixture(t, cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generate --capture-auto failed: %v\n%s", err, string(out))
	}

	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(readCaptureFile(t, output), &dto); err != nil {
		t.Fatalf("decode published ledger: %v", err)
	}
	ledger := dto.ToDomain()
	if len(ledger.Sessions) != 1 || len(ledger.Sessions[0].Invocations) != 1 {
		t.Fatalf("expected 1 session with 1 invocation, got %+v", ledger.Sessions)
	}
	inv := ledger.Sessions[0].Invocations[0]
	if inv.CostEstimate == nil {
		t.Fatalf("expected persisted cost_estimate from --capture-auto, got nil")
	}
	wantEstimate := domain.CostEstimate{
		AmountUSD: 1.0,
		Rate: domain.AppliedRateSnapshot{
			MatchedPattern:      "gpt-5.4",
			InputPerMillion:     2.0,
			CacheReadPerMillion: 0.25,
			OutputPerMillion:    8.0,
			EffectiveCatalogID:  expectedCatalogID,
		},
	}
	if *inv.CostEstimate != wantEstimate {
		t.Fatalf("capture-auto cost_estimate = %+v, want %+v", *inv.CostEstimate, wantEstimate)
	}

	md := string(readCaptureFile(t, delivery))
	for _, want := range []string{
		"| $1.00 EST |",
		"| **$1.00 EST** |",
		"**Applied API-equivalent rate card:**",
		expectedCatalogID,
	} {
		if !strings.Contains(md, want) {
			t.Fatalf("capture-auto delivery.md missing %q:\n%s", want, md)
		}
	}

	// Verify RecordPublication accepted the priced owner via CaptureOwnerFactsEqual parity.
	assertCapturePublishedOwners(t, captureDiagnostics(t, project, state), output, delivery, ledger.Sessions)
}

func TestListPricingEquivalentCosts(t *testing.T) {
	project := t.TempDir()
	sharedCfg := strings.Replace(claudeCLISharedConfig(), "max_depth: 0", "max_depth: 1", 1)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(sharedCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	localCfg := fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(project, ".publication-state"))
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(localCfg), 0o600); err != nil {
		t.Fatal(err)
	}

	rate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:1111111111111111111111111111111111111111111111111111111111111111",
	}
	schema := domain.CurrentSchemaVersion
	save := func(id, title string, invs []domain.InvocationFact) {
		t.Helper()
		dir := filepath.Join(project, "records", id)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		tid, tt := id, title
		if err := adapters.SaveFactLedger(filepath.Join(dir, "agentic_metrics.json"), domain.AgenticFactLedger{
			SchemaVersion: &schema,
			TicketID:      &tid,
			TicketTitle:   &tt,
			UpdatedAt:     "2026-10-17T12:00:00Z",
			Sessions: []domain.SessionFact{
				{
					ID:          "ses-" + id,
					Source:      domain.SourceCodex,
					User:        "tester",
					Invocations: invs,
					Tools:       []domain.SessionToolFact{},
					Anomalies:   []domain.Anomaly{},
				},
			},
			Anomalies: []domain.Anomaly{},
		}); err != nil {
			t.Fatal(err)
		}
	}

	save("AL-301", "Measured Only", []domain.InvocationFact{
		{Role: "orchestrator", Model: "claude-sonnet-4", Tokens: domain.InvocationTokens{Input: 1000, Output: 500}, CostUSD: 4.50, CostEvidence: domain.ObservedCost(4.50)},
	})
	save("AL-302", "Estimated Only", []domain.InvocationFact{
		{Role: "coder", Model: "gpt-5.4", Tokens: domain.InvocationTokens{Input: 100_000, Output: 20_000}, CostEvidence: domain.UnavailableCost(), CostEstimate: &domain.CostEstimate{AmountUSD: 6.74, Rate: rate}},
	})
	save("AL-303", "Hybrid Ticket", []domain.InvocationFact{
		{Role: "orchestrator", Model: "claude-sonnet-4", Tokens: domain.InvocationTokens{Input: 1000, Output: 500}, CostUSD: 23.20, CostEvidence: domain.ObservedCost(23.20)},
		{Role: "coder", Model: "gpt-5.4", Tokens: domain.InvocationTokens{Input: 200_000, Output: 50_000}, CostEvidence: domain.UnavailableCost(), CostEstimate: &domain.CostEstimate{AmountUSD: 7.12, Rate: rate}},
	})
	save("AL-304", "Partial Ticket", []domain.InvocationFact{
		{Role: "orchestrator", Model: "claude-sonnet-4", Tokens: domain.InvocationTokens{Input: 1000, Output: 500}, CostUSD: 2.00, CostEvidence: domain.ObservedCost(2.00)},
		{Role: "coder", Model: "gpt-5.4", Tokens: domain.InvocationTokens{Input: 100_000, Output: 10_000}, CostEvidence: domain.UnavailableCost(), CostEstimate: &domain.CostEstimate{AmountUSD: 3.00, Rate: rate}},
		{Role: "tester", Model: "unpriced-model", Tokens: domain.InvocationTokens{Input: 500, Output: 100}, CostEvidence: domain.UnavailableCost()},
	})

	cmd := exec.Command(cliBinaryPath, "--list")
	cmd.Dir = project
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--list failed: %v\n%s", err, string(out))
	}
	text := string(out)
	for _, want := range []string{
		"AL-301", "$4.50",
		"AL-302", "$6.74 EST",
		"AL-303", "$23.20 measured + $7.12 EST ($30.32 equiv)",
		"AL-304", "$2.00 measured + $3.00 EST (partial; 1 measured, 1 estimated, 0 partial, 1 unavailable / 3 invocations)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("--list output missing %q:\n%s", want, text)
		}
	}
}

func TestSerializeLedgerAndComputedExportWithEstimates(t *testing.T) {
	rate := domain.AppliedRateSnapshot{
		MatchedPattern:      "gpt-5.4",
		InputPerMillion:     2.0,
		CacheReadPerMillion: 0.25,
		OutputPerMillion:    8.0,
		EffectiveCatalogID:  "sha256:3333333333333333333333333333333333333333333333333333333333333333",
	}
	schema := domain.CurrentSchemaVersion
	tid := "AL-305"
	title := "Computed Export Hybrid"
	facts := domain.AgenticFactLedger{
		SchemaVersion: &schema,
		TicketID:      &tid,
		TicketTitle:   &title,
		UpdatedAt:     "2026-10-17T12:00:00Z",
		Sessions: []domain.SessionFact{
			{
				ID:     "ses-export-1",
				Source: domain.SourceCodex,
				User:   "tester",
				Invocations: []domain.InvocationFact{
					{
						Role:         "orchestrator",
						Model:        "claude-sonnet-4",
						Tokens:       domain.InvocationTokens{Input: 10_000, Output: 2_000},
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
				},
				Tools:     []domain.SessionToolFact{},
				Anomalies: []domain.Anomaly{},
			},
		},
		Anomalies: []domain.Anomaly{},
	}

	// 1. Direct adapters.SerializeLedger verification
	derived := domain.DeriveMetricsLedger(facts, domain.ProjectConfig{})
	serialized := adapters.SerializeLedger(derived)
	rawBytes, err := json.Marshal(serialized)
	if err != nil {
		t.Fatalf("marshal SerializeLedger: %v", err)
	}
	var decoded adapters.AgenticMetricsLedgerDTO
	if err := json.Unmarshal(rawBytes, &decoded); err != nil {
		t.Fatalf("unmarshal SerializeLedger: %v", err)
	}
	if decoded.Totals.CostProvenance != domain.CostProvenanceHybrid {
		t.Fatalf("SerializeLedger totals provenance = %q, want hybrid", decoded.Totals.CostProvenance)
	}
	if decoded.Totals.MeasuredCostUSD == nil || math.Abs(*decoded.Totals.MeasuredCostUSD-23.20) > 1e-9 ||
		decoded.Totals.EstimatedCostUSD == nil || math.Abs(*decoded.Totals.EstimatedCostUSD-7.12) > 1e-9 ||
		decoded.Totals.TotalEquivalentCostUSD == nil || math.Abs(*decoded.Totals.TotalEquivalentCostUSD-30.32) > 1e-9 {
		t.Fatalf("SerializeLedger totals amounts mismatch: %+v", decoded.Totals)
	}
	var foundModelRate bool
	for _, m := range decoded.Models {
		if m.Model == "gpt-5.4" {
			foundModelRate = true
			if m.CostProvenance != domain.CostProvenanceEstimated || len(m.AppliedRates) != 1 || m.AppliedRates[0] != rate {
				t.Fatalf("SerializeLedger gpt-5.4 model mismatch: %+v", m)
			}
		}
	}
	if !foundModelRate {
		t.Fatalf("SerializeLedger missing gpt-5.4 model row")
	}

	// Verify explicit null wire keys on individual model rows (measured model has estimated_cost_usd: null, estimated model has measured_cost_usd: null).
	var rawRoot struct {
		Models []map[string]json.RawMessage `json:"models"`
	}
	if err := json.Unmarshal(rawBytes, &rawRoot); err != nil {
		t.Fatal(err)
	}
	for _, rawModel := range rawRoot.Models {
		var name string
		if err := json.Unmarshal(rawModel["model"], &name); err != nil {
			t.Fatal(err)
		}
		if name == "claude-sonnet-4" {
			if rawVal, ok := rawModel["estimated_cost_usd"]; !ok || string(bytes.TrimSpace(rawVal)) != "null" {
				t.Fatalf("measured model missing explicit estimated_cost_usd: null, got exists=%v val=%s", ok, string(rawVal))
			}
		}
		if name == "gpt-5.4" {
			if rawVal, ok := rawModel["measured_cost_usd"]; !ok || string(bytes.TrimSpace(rawVal)) != "null" {
				t.Fatalf("estimated model missing explicit measured_cost_usd: null, got exists=%v val=%s", ok, string(rawVal))
			}
		}
	}

	// 2. CLI computed JSON export path (`agentic-lens --ticket-id AL-305 --json-output <export>`)
	project := t.TempDir()
	sharedCfg := strings.Replace(claudeCLISharedConfig(), "max_depth: 0", "max_depth: 1", 1)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(sharedCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	localCfg := fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(project, ".publication-state"))
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(localCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	wsDir := filepath.Join(project, "records", "AL-305")
	if err := os.MkdirAll(wsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.SaveFactLedger(filepath.Join(wsDir, "agentic_metrics.json"), facts); err != nil {
		t.Fatal(err)
	}
	exportPath := filepath.Join(wsDir, "computed_export.json")
	cmd := exec.Command(cliBinaryPath, "--ticket-id", "AL-305", "--json-output", exportPath)
	cmd.Dir = project
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI computed export failed: %v\n%s", err, string(out))
	}
	var exported adapters.AgenticMetricsLedgerDTO
	if err := json.Unmarshal(readCaptureFile(t, exportPath), &exported); err != nil {
		t.Fatalf("decode CLI computed export: %v", err)
	}
	if exported.Totals.CostProvenance != domain.CostProvenanceHybrid ||
		exported.Totals.TotalEquivalentCostUSD == nil ||
		math.Abs(*exported.Totals.TotalEquivalentCostUSD-30.32) > 1e-9 {
		t.Fatalf("CLI computed export totals mismatch: %+v", exported.Totals)
	}
}
