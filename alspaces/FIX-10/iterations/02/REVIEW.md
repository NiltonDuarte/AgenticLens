# Iteration 02 Review: Schema v4 Hierarchy Normalization and Backfill Integrity

## Verdict: STATUS: APPROVED

Iteration 02 successfully normalized the container-execution relationship between `SessionFact` and `InvocationFact` across Go domain models, boundary DTOs, and TypeScript definitions. All 5 defects identified during the initial audit have been resolved and verified with targeted regression tests. Full automated verification (`make`, `go test -race ./...`, `npm test`, `npm run typecheck`, `npm run lint`) passes with zero failures, zero warnings, zero `any`/`interface{}`, and zero emojis.

---

## Axis 1: Spec & Contract Alignment

### 1.1 Schema v4 Hierarchy Normalization
- **Domain Structs (`backend/internal/domain/models.go`)**:
  - `InvocationFact` correctly removed execution-irrelevant container metadata (`ParentID` and `User`).
  - `SessionFact` correctly hoisted `User string` alongside existing `ParentID *string` and `Source SourceHarness`.
  - Pure domain aggregations (`ComputeWorkflowEfficiency`, `ComputeThreadSummariesFromSessions`, `DeriveMetricsLedger`) derive context window peaks, role effort, and contributor attribution directly from `session.ParentID` and `session.User`.
- **Boundary DTOs (`backend/internal/adapters/schemas.go`)**:
  - `InvocationFactDTO` retains `ParentID *string` and `User string` with `omitempty` to deserialize legacy ledgers without unknown field rejections under `DisallowUnknownFields()`.
  - `SessionFactDTO` adds `User string` (`omitempty`).
  - `AgenticFactLedgerDTO.ToDomain()` gracefully hoists `User` and `ParentID` from child invocations to the session level if omitted at the session container level.
  - `ValidateRuntimeFactLedger` updated its schema validation rules: `parent_id` and `user` are no longer required on `invocations`.

### 1.2 Multi-Root Sessions & Explicit Fallback CLI Flag
- `backend/internal/adapters/backfill/engine.go` correctly extracts all candidate root session IDs across `openCodeSessions`.
- Multi-root session discovery resolves candidate root IDs against SQLite using `resolveSessionRootInDB`.
- Extraction queries facts for all resolved root IDs simultaneously via `opencode.ExtractSessions`.
- Partial multi-machine resolution preserves unresolved session facts without data loss when `--allow-fallbacks` is omitted.
- Heuristic session resolution (title matching and commit date range) is strictly guarded behind `opts.AllowFallbacks`.

### 1.3 Distinct Invocation Identifiers
- OpenCode extractor (`backend/internal/adapters/opencode/extractor.go`) assigns distinct IDs formatted as `fmt.Sprintf("%s-inv-%d", s.id, idx+1)`.
- Codex rollout extractor (`backend/internal/adapters/codex/rollout.go`) assigns distinct IDs formatted as `fmt.Sprintf("%s-inv-%d", transcript.invocationID(), len(transcript.Invocations)+1)`.
- Model aggregations (`ComputeTotals`, `ComputeModelMetrics`) accumulate distinct invocation counts via `sids[inv.ID]` without collapsing multiple model calls within a session.

### 1.4 Verification of Previous Defect Resolutions

#### Defect 1.4.1 (Resolved): Codex Rollout Backfill Drops `ParentID` on Child Sessions
- **Resolution**: `backend/internal/adapters/backfill/engine.go:533` now invokes `combined.SessionFacts()` instead of `domain.SessionsFromFactsWithSource(...)`. In addition, lines 535-556 preserve existing `User` and `ParentID` mappings when refreshed facts omit them.
- **Verification**: Verified by `TestBackfillEngine_CodexPreservesChildParentIDAndUser` in `backend/internal/adapters/backfill/engine_test.go`. The test confirms `rootSess.ParentID == nil`, `rootSess.User == "dev-lead"`, `childSess.ParentID == "codex-root"`, and `childSess.User == "dev-subagent"`.

#### Defect 1.4.2 (Resolved): `hydrateCodexEvents` Fails Closed Due to Checking `inv.ID` Against `parsedThreadIDs`
- **Resolution**: `backend/internal/adapters/live/service.go:1005-1011` iterates over `codexSessions` and adds only `session.ID` to `expectedThreadIDs`. `inv.ID` (`<session_id>-inv-<index>`) is no longer checked against rollout thread IDs.
- **Verification**: Verified by `TestHydrateWorkspaceEvents_CodexSucceedsWithDistinctInvocationIDs` in `backend/internal/adapters/live/sqlite_test.go`. The test verifies successful hydration (`resp.IsHydrated == true`, `resp.DataSource == DataSourceLocalCodex`) when invocations possess distinct IDs `thread-codex-inv-1` and `thread-codex-inv-2`.

#### Defect 1.4.3 (Resolved): `hydrateAnomalyAssociations` Drops `ParentID` and `User`
- **Resolution**: `backend/internal/adapters/live/service.go:1058-1088` clones `facts.Sessions` directly and populates each session's `Events` from `eventsBySession[s.ID]`, preserving container-level `ParentID` and `User` metadata without calling `SessionsFromFacts`.
- **Verification**: Verified by `TestHydrateWorkspaceEvents_AnomalyHydrationRetainsSessionHierarchy` in `backend/internal/adapters/live/sqlite_test.go`. The test confirms parent-child session hierarchy is preserved during anomaly detection.

#### Defect 1.4.4 (Resolved): `detectContextSaturationSpikes` Fallback Sets `SessionID` to `inv.ID`
- **Resolution**: `backend/internal/domain/anomalies.go:353-361` iterates over `ledger.Sessions` and prefers `session.ID` for `sid` over `inv.ID`.
- **Verification**: Verified by `TestDetectAnomalies_ContextSaturationSpike/v1 fallback without events associates anomaly with session ID and persists correctly` in `backend/internal/domain/anomalies_test.go`. `AnomalyContext.SessionID` matches `sessionID`, and `PersistAnomalySummaries` preserves the anomaly in `session.Anomalies`.

#### Defect 1.4.5 (Resolved): `sessionRoleMap` Key Mismatch in `LiveService`
- **Resolution**: `backend/internal/adapters/live/service.go:758-773` populates `sessionRoleMap` from `ledger.Facts.Sessions` using `session.ID` (with fallback stripping `-inv-` suffix), ensuring keys match `verifiedSessions` lookups.
- **Verification**: Verified by `TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID` in `backend/internal/adapters/live/sqlite_test.go`. The test confirms that role `planner` from the session invocation is resolved without falling back to database queries.

---

## Axis 2: Engineering Standards (TECHNICAL_GUIDELINES.md)

### 2.1 Go FCIS Layer Separation & Determinism
- **Domain Layer (`backend/internal/domain/`)**:
  - Zero imports of outer layers (`adapters`, `cmd`).
  - Zero imports of I/O packages (`net/http`, `os`, `database/sql`).
  - Deterministic algorithms with pure value semantics.
- **Adapters Layer (`backend/internal/adapters/`)**:
  - Imperative shell properly handles boundary JSON marshaling, SQLite queries, and filesystem access.

### 2.2 TypeScript Strictness & Boundary Typing
- **Frontend Layer (`web/src/domain/types.ts`)**:
  - `InvocationFact` properly defines `readonly parent_id?: string | null` and `readonly user?: string`.
  - `SessionFact` properly defines `readonly user?: string`.
  - `tsc --noEmit` and `eslint` pass with 0 errors and 0 warnings.
  - Component views correctly render hoisted thread user attributions.

### 2.3 Zero `any` and Zero `interface{}` Invariant
- Backend: 0 occurrences of `interface{}` and 0 untyped `any` primitives.
- Frontend: 0 occurrences of `any`, `as any`, `<any>`, or `any[]`.

### 2.4 Zero Emojis Compliance
- Complete repository diff scanned for unicode emojis: 0 emojis found in code, docstrings, tests, summaries, or commit messages.

### 2.5 Automated Verification Results
- `go vet ./...`: Passed.
- `go test -race ./...`: All backend packages passed without race conditions or memory leaks.
- `npm run typecheck`: Passed (zero TypeScript compiler diagnostics).
- `npm run lint`: Passed (zero ESLint warnings or errors).
- `npm test`: Passed (32 test files, 734 tests passing).
- Production multi-platform build: Single binary compiled successfully across `linux-amd64`, `linux-arm64`, `darwin-amd64`, and `darwin-arm64`.

---

## Final Verdict
**STATUS: APPROVED**
All acceptance criteria satisfied. Defects 1.4.1 through 1.4.5 resolved. Quality gates passed. Ready for telemetry generation and PR handoff.
