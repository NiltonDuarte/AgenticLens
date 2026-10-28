# Iteration 02 Plan: Normalize Session-Invocation Hierarchy in Schema v4, Migration, and Backfill

## Overview
Normalize the relationship between SessionFact and InvocationFact across Go domain, schema v4, migration, adapters, and TypeScript frontend:
1. `sessions`: Container-level metadata: `id`, `parent_id`, `source`, `user`.
2. `invocations`: Execution-level facts: `id` as distinct invocation identifier (`<session_id>-inv-<index>`); remove `user` and `parent_id`.
3. Schema v4 migration: In `step_v3_v4.go` and `InvariantValidator`, migrate user to session level, assign distinct invocation IDs, strip `parent_id` and `user` from invocations, and ensure strict metric invariants.
4. Adapters & Backfill: Update OpenCode and Codex extractors to populate `session.User` and distinct invocation IDs; update Backfill engine to maintain normalized hierarchy.
5. Frontend: Update TypeScript interfaces in `web/src/domain/types.ts` and test fixtures.
6. Target execution: Build binaries and execute schema migration and backfill on `SinkLayerControlPlane.Code`.

## Implementation Strategy
- Pure Domain: update `SessionFact` (add `User string`) and `InvocationFact` (remove `ParentID`, remove `User`) in `backend/internal/domain/models.go`.
- Domain Aggregations & Anomalies: update `aggregations.go`, `anomalies.go`, and helpers to read `session.User` and `session.ParentID`.
- Boundary Adapters & DTOs: update `schemas.go` (`SessionFactDTO`, `InvocationFactDTO`), `formatters.go`, `opencode/extractor.go`, `codex/rollout.go`, `backfill/engine.go`, `main.go`.
- Migration: update `step_v3_v4.go` and `validator.go` to handle user hoisting and distinct invocation IDs.
- Frontend: update `web/src/domain/types.ts` and test fixtures.
- Test suites: update tests in backend and frontend.
- Target execution: build updated binaries and execute backfill in `SinkLayerControlPlane.Code`.
