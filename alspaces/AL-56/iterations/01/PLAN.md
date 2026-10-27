# AL-56 Iteration 01 Plan: OpenCode Non-Breakage Audit and Compatibility Hardening

## Overview

Ticket AL-56 introduced OpenAI Codex rollout parsing, session discovery, and ref-based Git code facts analysis to AgenticLens. While Codex support is validated via synthetic fixtures and local probe runs, the changes made to cross-cutting packages affect core operational pathways shared with OpenCode.

This iteration conducts a systematic audit of four critical integration surfaces:
1. CLI entrypoint and flag dispatch in backend/cmd/agentic-lens/main.go.
2. File classification rules and empty pattern fallbacks in backend/internal/adapters/git/classifier.go.
3. Batch Git code facts recomputation in backend/internal/adapters/backfill/engine.go.
4. Test coverage gaps across existing OpenCode test suites and historical workspace fixtures.

The objective is to establish hard guarantees that OpenCode users experience zero regressions, zero silent data corruptions during backfills, and no degradation in code metric classifications.

## Cross-Cutting Audit & Risk Analysis

### 1. CLI Entrypoint and Flag Dispatch (backend/cmd/agentic-lens/main.go)
- Flag Defaults:
  The flag --source defaults to "opencode" and --code-ref defaults to "HEAD". The source argument is normalized and strictly validated against "opencode" and "codex". If omitted, execution flows through the legacy OpenCode database client.
- Database Extraction Pathway:
  The OpenCode pathway opens the SQLite database at --db-path (or configured database.path), calls opencode.ResolveActiveRootSession, opencode.ExtractFacts, and opencode.ExtractSteeringTurns.
- Risk Analysis:
  In the new implementation, codeFacts is computed unconditionally via git.ComputeCodeFactsAtRef using --code-ref. In legacy versions, when git diff discovery was unable to locate a ticket-matching branch, it fell back to existing code facts stored in cached fact ledgers. In the current implementation, if --code-ref is executed in an environment where git history has been squashed, rebased, or already merged into the base branch, lines added and deleted drop to zero. Furthermore, if git.ComputeCodeFactsAtRef encounters an error (such as an unresolvable base branch in a minimal container environment), the command exits with code 1 instead of allowing existing code metrics to be retained.

### 2. File Classification and Empty Pattern Fallback (backend/internal/adapters/git/classifier.go)
- Rule Changes:
  AL-56 stripped all hardcoded classification heuristics from ClassifyFile, requiring all categorization to be driven by ClassificationConfig (TestPatterns, ProdPatterns, Include, Exclude).
- Impact on OpenCode Repositories:
  Inspection of backend/internal/adapters/config/writer.go reveals that GenerateDefaultYAML (called by agentic-lens init) writes empty lists for test_patterns: [] and prod_patterns: []. Existing OpenCode repositories that initialized configuration prior to AL-56 or used agentic-lens init have empty pattern slices.
- Risk Analysis:
  When TestPatterns and ProdPatterns are empty, ClassifyFile categorizes every modified Go and TypeScript file as CategoryOther. Under git.ComputeCodeFactsAtRef, CategoryOther files increment LinesAdded and FilesChanged, but TestLinesAdded and ProdLinesAdded remain zero. As a result, domain.DeriveMetricsLedger computes a TestToProductionRatio of "0:0". This silently breaks the test-to-production metric across all existing OpenCode projects unless users manually craft comprehensive regexes. Furthermore, test files in non-source directories (such as testdata/ or integration test folders) are rejected by IsMonitoredSourcePath if they are not inside src_folders and do not match TestPatterns.

### 3. Backfill Engine Guardrails (backend/internal/adapters/backfill/engine.go)
- Flag Enforcement:
  AL-56 updated the backfill CLI and engine so that --recompute-git requires --code-ref. If --code-ref is absent, backfill aborts immediately.
- Multi-Ticket Contamination Risk:
  The backfill command is frequently invoked across a directory of multiple tickets (for example, agentic-lens backfill --workspaces workspaces --migrate-schema --update-summaries). If a user specifies --recompute-git with --code-ref HEAD across an entire workspace directory, the engine iterates over every ticket in that directory and assigns the exact same Git diff facts from HEAD to every historical ticket.
- Risk Analysis:
  This constitutes silent multi-ticket telemetry corruption. Batch backfill should either disallow --recompute-git when --ticket-id is omitted, or resolve Git refs per ticket rather than applying a single commit diff to the entire repository history.

### 4. OpenCode Tests and Parity Assertions
- Current Coverage:
  Existing tests in backend/cmd/agentic-lens/main_test.go and backend/internal/adapters/parity_test.go focus heavily on SQLite querying, domain metric derivations, and schema migrations.
- Risk Analysis:
  There are minimal assertions verifying that agentic-lens generate without flags operates cleanly against an OpenCode database in repositories with minimal or default .agenticlens-config.yaml files. Specifically, test coverage must assert that projects with empty test_patterns and prod_patterns still compute non-zero test and production code lines when standard extensions are present.

## Phased Implementation Plan

### Phase 1: Architectural Baseline & Verification Setup
- Ground all adapter modifications against TECHNICAL_GUIDELINES.md.
- Ensure that the pure functional domain in backend/internal/domain/ remains untouched, maintaining zero I/O, zero external dependencies, and value semantics.
- Audit all OpenCode test suites in backend/internal/adapters/opencode/ and backend/cmd/agentic-lens/ to identify test runs that depend on Git state.

### Phase 2: Configuration & Classifier Fallback Hardening
- Audit backend/internal/adapters/git/classifier.go:
  Evaluate the introduction of sensible convention fallbacks when cfg.TestPatterns and cfg.ProdPatterns are empty slices. If a project does not declare custom patterns, the classifier should apply standard language conventions (such as *_test.go, *.test.ts, *.test.tsx, *.spec.ts for test code; and standard source extensions under monitored folders for production code).
- Audit backend/internal/adapters/config/writer.go:
  Ensure GenerateDefaultYAML populates sensible default patterns for test_patterns and prod_patterns rather than empty arrays, aligning newly initialized projects with repository standards.
- Audit backend/internal/adapters/git/analyzer.go:
  Verify IsMonitoredSourcePath behaves predictably when test patterns are populated via defaults versus explicit overrides.

### Phase 3: CLI Entrypoint Resilience (backend/cmd/agentic-lens/main.go)
- Preserve OpenCode Default Invariant:
  Verify that when --source is omitted, activeSource evaluates strictly to "opencode" without any side-effects from Codex logic.
- Safe Code Facts Handling:
  In runGenerate, ensure that if git.ComputeCodeFactsAtRef returns an error (e.g. unresolvable base branch or shallow environment), the CLI provides a graceful fallback to cached code facts from existing ledgers when available, warning the user rather than hard-crashing.
- Timestamp Derivation Parity:
  Confirm that started_at and completed_at derivation from invocations behaves identically between OpenCode SQLite and Codex rollouts.

### Phase 4: Backfill Engine Guardrails (backend/internal/adapters/backfill/)
- Enforce Scoping on --recompute-git:
  In backend/cmd/agentic-lens/main.go (runBackfill) and backend/internal/adapters/backfill/engine.go, enforce that if --recompute-git is supplied, --ticket-id must also be provided, preventing accidental overwrite of all workspace tickets with a single commit ref.
- Clear Diagnostics:
  If a user attempts to run batch backfill with --recompute-git across all tickets, return an explicit error stating that Git recomputation must be targeted to a single ticket via --ticket-id.
- Non-Destructive Preservation:
  Verify that backfill operations with --migrate-schema or --recompute-db preserve existing code facts without modification when --recompute-git is not set.

### Phase 5: Adversarial Testing & Parity Verification
- Unit Testing for Empty Pattern Fallbacks:
  Add tests in backend/internal/adapters/git/classifier_test.go verifying classification behavior when ClassificationConfig has empty slices versus nil versus populated patterns.
- OpenCode End-to-End CLI Regressions:
  In backend/cmd/agentic-lens/main_test.go, add tests executing agentic-lens generate without --source, verifying full SQLite extraction, metric table formatting, and summary updates.
- Backfill Scoping Tests:
  Add tests in backend/internal/adapters/backfill/engine_test.go asserting that multi-ticket backfill rejects global --recompute-git without --ticket-id, while allowing schema migrations across all tickets safely.
- Parity Workspace Suite:
  Run backend/internal/adapters/parity_test.go to verify that all historical schema fixtures parse, validate, and compute identical metrics.

### Phase 6: Documentation & Workspace Lifecycle Sync
- Synchronize CONTRACT.md to reflect OpenCode non-breakage guarantees alongside Codex support.
- Maintain alspaces/AL-56/SUMMARY.md and PULL_REQUEST.md.
- Ensure all artifacts strictly obey the Zero Emojis Mandate.

## Go Backend & TypeScript Frontend Boundaries

- Go Backend Pure Domain (backend/internal/domain/):
  Zero modifications required. Domain models, aggregations, and metrics derivations remain pure functional core.
- Go Backend Imperative Shell (backend/internal/adapters/, backend/cmd/):
  Hardening confined to adapters/git/ (classifier fallbacks), adapters/config/ (default template patterns), adapters/backfill/ (batch protection), and cmd/agentic-lens/ (CLI fallbacks).
- TypeScript Frontend (web/src/):
  Zero modifications required. Frontend consumes standard schema-v3 agentic_metrics.json ledgers identically for both OpenCode and Codex sources.
