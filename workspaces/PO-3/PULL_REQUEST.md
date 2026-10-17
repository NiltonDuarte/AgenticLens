# PR: PO-3: Native Go Extraction, Backfill Command, and Multi-Language Git Metrics

## Summary
Defines the product specification, architectural contracts, and 3 vertical tracer-bullet Jira tickets for decoupling active session telemetry extraction and ledger backfilling into separate Go CLI subcommands (`generate` and `backfill`), alongside fixing defective git code metric calculations across Python and Go.

## Root Cause Analysis: Git Metrics
Investigation identified why code metrics (`lines_added`, `lines_deleted`, test/prod breakdown) were evaluating to zero:
1. `scripts/extract_agentic_metrics.py` lines 363 & 398 hardcoded prefix checks `startswith(("src/", "tests/", "infra/", "cli/"))`, rejecting all Go and TypeScript files (`cmd/`, `internal/`, `web/`, `scripts/`).
2. Test vs prod classification (lines 594-597) hardcoded `startswith("tests/")` and `startswith("src/")`, failing to identify Go `*_test.go` and TypeScript `*.test.ts`/`*.spec.ts` files.
3. Revision resolution queried `origin/main` directly, missing local branch commits and unpushed commits on `main`.

## Ticket Breakdown
New tickets created under `workspaces/plan/to-do/`:
- `AL-10`: Fix Git Metrics Calculation in Python Extractor and Backfill Existing Workspace Ledgers (Ready / Unblocked)
- `AL-11`: Implement Native Go Telemetry Extractor and Active Session Metric Generation (`agentic-lens generate`)
- `AL-12`: Implement Go Telemetry Backfill Engine and Schema Migration Command (`agentic-lens backfill`)

## Architectural Invariants
- FCIS compliance: pure file categorization heuristics in `internal/domain`, imperative git and SQLite execution in `internal/adapters`.
- Hybrid path classification: robust multi-language defaults (`*_test.go`, `*.test.ts`, etc.) coupled with optional config overrides in `.agenticlens/config.json`.
- Pure Go SQLite driver (`modernc.org/sqlite`) preserves zero-CGO single-binary compilation.
- Zero emojis across all code, logs, tables, and documentation.
- Zero `any` or empty `interface{}` across Go.

## Pointers
- Specification: `workspaces/PO-3/SPEC.md`
- Shared Roadmap Index: `workspaces/plan/TICKETS.md`
- Session Summary: `workspaces/PO-3/SUMMARY.md`
