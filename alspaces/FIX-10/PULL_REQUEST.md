# fix(backfill): multi-root session backfill, explicit fallbacks, and hierarchy normalization (FIX-10)

## Summary
Resolves telemetry backfill integrity issues and normalizes session-invocation hierarchy:
1. Multi-root session truncation: Discovers and extracts facts across all root sessions present in a ticket ledger rather than stopping at the first root session found.
2. Unsafe implicit heuristic fallbacks: Strict exact session ID matching is enforced by default. Heuristic title substring and workspace path resolution are gated behind `--allow-fallbacks`. Missing SQLite sessions preserve existing historical invocations and events intact and are marked `partial` pending host migration.
3. Schema v4 session-invocation normalization: Hoisted `user` and `parent_id` to container-level `SessionFact`, removed `user` and `parent_id` from `InvocationFact`, and assigned distinct `<session_id>-inv-<index>` invocation IDs across OpenCode, Codex, migration, and frontend types.

## Changes
- `backend/internal/domain/models.go`: Added `User string` to `SessionFact`; removed `ParentID` and `User` from `InvocationFact`.
- `backend/internal/domain/aggregations.go` & `anomalies.go`: Updated aggregation and anomaly detection routines to evaluate session-level metadata.
- `backend/internal/domain/migration/`: Updated `step_v3_v4.go` and `validator.go` for user hoisting, distinct invocation IDs, and schema v4 invariants.
- `backend/internal/adapters/backfill/engine.go`: Multi-root candidate discovery, strict fallback gating behind `AllowFallbacks`, and session-level hierarchy preservation.
- `backend/cmd/agentic-lens/main.go`: Added `--allow-fallbacks` CLI flag and updated extraction ingestion.
- `backend/internal/adapters/opencode/extractor.go`: Added `ExtractSessions` producing normalized session containers with distinct invocation IDs.
- `backend/internal/adapters/codex/rollout.go`: Implemented `SessionFacts()` and thread parent propagation on `Transcript`.
- `web/src/domain/types.ts`: Synchronized TypeScript interfaces (`user?: string` on `SessionFact`, optional fields on `InvocationFact`).

## Documentation & Verification
- Technical Contract: `alspaces/FIX-10/CONTRACT.md`
- Implementation Summary: `alspaces/FIX-10/SUMMARY.md`
- Unit test suite: `make` passed with 100% success across all backend packages and frontend tests.
- Backfill validation: Validated against 10 real ticket workspaces (`DPS-11170`, `DPS-11173`, `DPS-11187`, `DPS-11197`, `DPS-11199`).
