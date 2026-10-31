# AL-55: Formal Schema v4 Evolution, Migration Step, and Backfill Runner

- **Subsystem**: Domain / Migration / Adapters / Backfill / CLI / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Status**: Pending
- **Blocked by**: AL-52, AL-54
- **Blocks**: None

## Planning status (2026-10-31)

Partially superseded; remains blocked by AL-52 and AL-54. AL-53 delivered the Ticket Deep Dive skills and subagents panel. AL-51 merged in [PR #62](https://github.com/NiltonDuarte/AgenticLens/pull/62), satisfying its ingestion dependency.

AL-56 already shipped schema v4, the Go v3-to-v4 migration, and explicit migration/source-backfill boundaries in [PR #38](https://github.com/NiltonDuarte/AgenticLens/pull/38) and [PR #43](https://github.com/NiltonDuarte/AgenticLens/pull/43). AL-51 deliberately kept skills/task links optional at that version. The existing migration does not establish the proposed required-array contract, so this ticket is not complete.

The scope below records the original proposal and needs replanning before implementation: do not bump to v4 again, replace the existing v3-to-v4 step, or introduce client-side legacy ingestion. `web/src/domain/migration.ts` does not exist in the current application. Reconcile any remaining skill/task normalization and historical backfill with the current schema and AL-51 contract, preserving incomplete source-backfill diagnostics. Do not treat a shape-only upgrade as successful source recovery.

## Context
With the skill and task observation features implemented, exercised, and stabilized across Live Observability, Trace Waterfall, Ticket Deep Dive, and Fleet Analytics (AL-51 through AL-54), the schema contracts are mature. Now the domain schema must be formally locked and bumped to Schema Version 4. A formal migration step (`StepV3ToV4`) and backfill runner must be executed to ensure historical ledgers are permanently upgraded and normalized.

## Goal
Bump the domain schema version to 4, implement the `StepV3ToV4` migration step in Go and TypeScript, and enhance the `agentic-lens backfill` CLI to backfill skill and task facts into historical workspace ledgers.

## Scope Across Layers

### 1. Functional Core & Migration Runner (`internal/domain/models.go`, `internal/domain/migration/runner.go`)
- Update `CurrentSchemaVersion = 4`.
- Implement `StepV3ToV4(raw []byte) ([]byte, error)`:
  - Validates `schema_version == 3`.
  - Ensures `skills` is initialized as a non-null empty array `[]` if missing or null.
  - Ensures `task_links` is initialized as a non-null empty array `[]` if missing or null.
  - Updates `schema_version` to `4`.
- Register `StepV3ToV4` in migration chain: `v1 -> v2 -> v3 -> v4`.

### 2. TypeScript Migration Engine (`web/src/domain/migration.ts`)
- Update `CURRENT_SCHEMA_VERSION = 4`.
- Implement TypeScript migration step:
  ```typescript
  export function migrateV3ToV4(ledger: Record<string, unknown>): Record<string, unknown> {
    return {
      ...ledger,
      schema_version: 4,
      skills: Array.isArray(ledger.skills) ? ledger.skills : [],
      task_links: Array.isArray(ledger.task_links) ? ledger.task_links : [],
    };
  }
  ```
- Register step in client-side migration pipeline.

### 3. Imperative Shell & Backfill Engine (`internal/adapters/backfill/runner.go`, `cmd/agentic-lens/`)
- Update `agentic-lens backfill` command:
  - Scans workspace ledgers for `schema_version < 4`.
  - Upgrades schema via `StepV3ToV4`.
  - Optionally re-scans raw OpenCode SQLite session logs to populate retroactive `skills` and `task_links` where historical sessions exist.
  - Writes back updated ledgers atomically with backups.

### 4. Automated Tests
- Table-driven migration tests in Go (`internal/domain/migration/runner_test.go`) validating:
  - Step v3 to v4 migration.
  - Full pipeline migration from v1 -> v4.
  - Idempotency when running migration on already-v4 ledgers.
- Vitest tests in `web/src/domain/migration.test.ts` mirroring Go migration test suite.
- CLI tests in `internal/adapters/backfill/runner_test.go` verifying workspace ledger upgrades and backup generation.

## Acceptance Criteria
1. `CurrentSchemaVersion` is bumped to 4 in both Go backend and TypeScript frontend.
2. `StepV3ToV4` cleanly upgrades v3 ledgers to v4, guaranteeing `skills` and `task_links` are valid arrays.
3. Full migration chain from v1 to v4 is completely verified and deterministic.
4. CLI backfill command successfully upgrades historical workspaces to schema v4.
5. All 700+ existing tests pass with zero regressions.

## Definition of Done
- Go and TypeScript migration engines synchronized and verified.
- CLI backfill runner tested with synthetic workspace fixtures.
- 100% test pass rate across `make check`.
- Zero emojis across all code, tests, and documentation.
