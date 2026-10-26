# AL-56 Iteration 02 Plan: Classification Policy Alignment, Guardrail Retention, and Resilience Clarification

## Overview

Iteration 02 addresses user steering regarding file classification, configuration generation, and operational guardrails.

Key adjustments:
1. Revert runtime fallback heuristics in backend/internal/adapters/git/classifier.go. When classification patterns are empty or unconfigured, the system must not guess or fall back to internal lists. Instead, it must emit an explicit warning to the user and classify unconfigured files as other.
2. Maintain template-only defaults in backend/internal/adapters/config/writer.go. Writing sensible defaults into the generated configuration file during project initialization (agentic-lens init) is confirmed as desired, giving the user explicit ownership of the patterns on disk.
3. Retain the backfill safety guardrails in backend/internal/adapters/backfill/engine.go and backend/cmd/agentic-lens/main.go requiring an explicit ticket ID and code ref when recomputing Git facts.
4. Retain the non-destructive ledger loader in backend/internal/adapters/repository.go.
5. Clarify CLI entrypoint resilience in backend/cmd/agentic-lens/main.go for shallow or detached Git environments.

## Architectural Boundaries & Principles

- Functional Core, Imperative Shell (FCIS):
  The domain layer in backend/internal/domain/ remains completely untouched. Classification rules and warnings remain confined to adapters/git/ and adapters/config/.
- Value Semantics & Pure Classification:
  ClassifyFile in backend/internal/adapters/git/classifier.go remains a pure deterministic function that maps path and configuration to a category without hidden global defaults.
- Zero Code Blocks:
  This plan specifies architecture, behavioral invariants, and verification steps in pure descriptive prose.

## Phased Implementation Tasks

### Phase 1: Revert Runtime Classification Fallbacks (classifier.go)
- Target: backend/internal/adapters/git/classifier.go
- Tasks:
  1. Remove runtime fallback logic that uses DefaultTestPatterns or DefaultProdPatterns when cfg.TestPatterns or cfg.ProdPatterns are empty or nil.
  2. Remove EffectiveTestPatterns and EffectiveProdPatterns helpers if they inject defaults into classification.
  3. Ensure ClassifyFile evaluates only explicit configuration:
     - Check workspace directory exclusion.
     - Check explicit Exclude patterns.
     - Check explicit TestPatterns; if matched, return CategoryTest.
     - Check explicit ProdPatterns; if matched, return CategoryProd.
     - Check explicit Include patterns; if matched, return CategoryProd.
     - Check standard non-code lockfiles and dotfiles; return CategoryIgnored.
     - When no patterns match, return CategoryOther.
  4. Implement an explicit diagnostic warning function (for example, WarnUnconfiguredPatterns) that checks if TestPatterns or ProdPatterns are empty and outputs an actionable warning to stderr during code facts computation, notifying the user to configure patterns in .agenticlens-config.yaml.
  5. Ground against backend/internal/adapters/git/analyzer.go to ensure IsMonitoredSourcePath respects the strict classification without expecting runtime default arrays.

### Phase 2: Template-Only Initialization Defaults (writer.go)
- Target: backend/internal/adapters/config/writer.go
- Tasks:
  1. Retain DefaultTestPatterns and DefaultProdPatterns in backend/internal/adapters/config/writer.go strictly as template scaffolding values for GenerateDefaultYAML and GenerateDefaultConfigYAML.
  2. Verify that running agentic-lens init generates a .agenticlens-config.yaml file containing explicit, pre-populated test_patterns and prod_patterns arrays.
  3. Verify that the runtime metric extraction pipeline does not import or execute these defaults if a user removes or empties the arrays in their configuration file.

### Phase 3: Verify CLI Entrypoint Resilience (main.go)
- Target: backend/cmd/agentic-lens/main.go
- Tasks:
  1. Validate the fallback handling in runGenerate: when git.ComputeCodeFactsAtRef returns an error and a cached ledger exists (fallbackCode != nil), log an explicit warning to stderr and reuse cached facts.
  2. Ensure that if git.ComputeCodeFactsAtRef fails on an uninitialized run without a cached ledger, the command halts with exit code 1 and prints the error message.
  3. Connect pattern configuration warning checks into runGenerate so users running with empty patterns receive a diagnostic warning before fact generation completes.
  4. Ensure --source defaults to opencode and dispatches cleanly.

### Phase 4: Reaffirm Backfill Safety Guardrails & Ledger Loader
- Targets:
  - backend/internal/adapters/backfill/engine.go
  - backend/cmd/agentic-lens/main.go (runBackfill)
  - backend/internal/adapters/repository.go (LoadExistingFactLedger)
- Tasks:
  1. Confirm that Execute in BackfillEngine checks that if RecomputeGit is true, TicketID must be non-empty and CodeRef must be non-empty, returning an explicit validation error if either is omitted.
  2. Confirm that runBackfill in main.go checks for these flags before initiating the engine and outputs explicit error messages to stderr.
  3. Confirm that LoadExistingFactLedger in repository.go safely parses existing ledgers (both schema-v3 and legacy formats) without mutating them, enabling safe schema migration and DB fact updates without zeroing Git code metrics.

### Phase 5: Adversarial Testing & Parity Verification
- Targets:
  - backend/internal/adapters/git/classifier_test.go
  - backend/internal/adapters/config/writer_test.go
  - backend/internal/adapters/backfill/engine_test.go
  - backend/cmd/agentic-lens/main_test.go
- Tasks:
  1. In classifier_test.go, add tests asserting that when ClassificationConfig has empty or nil TestPatterns and ProdPatterns, Go test files (such as main_test.go) and TypeScript files are classified as CategoryOther, not CategoryTest or CategoryProd.
  2. In classifier_test.go, verify warning generation when patterns are empty.
  3. In writer_test.go, verify that generated YAML templates contain the standard patterns, but direct configuration parsing with empty arrays remains empty.
  4. In backfill/engine_test.go, verify that multi-ticket backfill without --ticket-id fails fast when --recompute-git is supplied, while succeeding for schema-only migrations.
  5. In main_test.go, verify CLI runs with empty patterns emit the unconfigured warning to stderr.

### Phase 6: Workspace Documentation Synchronization
- Update alspaces/AL-56/CONTRACT.md with the zero-fallback file classification policy and user warning requirements.
- Retain this delta plan in alspaces/AL-56/iterations/02/PLAN.md.
- Ensure all artifacts strictly obey the Zero Emojis Mandate.
