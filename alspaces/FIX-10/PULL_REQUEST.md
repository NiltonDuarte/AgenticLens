# fix: support multi-root backfill and gate heuristic fallbacks behind explicit flag

## Summary
Resolves two telemetry backfill integrity issues in AgenticLens:
1. Multi-root session truncation: Discovers and extracts facts across all root sessions present in a ticket ledger rather than stopping at the first root session found.
2. Unsafe implicit heuristic fallbacks: Strict exact session ID matching is now enforced by default. Heuristic title substring and workspace path resolution are gated behind `--allow-fallbacks`. Missing SQLite sessions preserve existing historical invocations and events intact and are marked `partial` pending subsequent host migration.

## Changes
- `backend/internal/adapters/backfill/options.go`: Added `AllowFallbacks bool` to `BackfillOptions`.
- `backend/cmd/agentic-lens/main.go`: Added `--allow-fallbacks` CLI flag and usage documentation.
- `backend/internal/adapters/backfill/engine.go`:
  - Resolved root sessions across all invocations and sessions in the ledger.
  - Gated heuristic matching behind `opts.AllowFallbacks`.
  - Preserved historical invocations when sessions cannot be resolved in the local SQLite database.
- `backend/internal/adapters/backfill/engine_test.go`: Added tests verifying multi-root extraction, strict rejection of untrusted fallbacks, and explicit fallback behavior.
- `backend/cmd/agentic-lens/main_test.go`: Updated help output test to include `--allow-fallbacks`.

## Verification
- Unit test suite: all backend and frontend unit tests passing.
- Validated backfill against 10 real ticket workspaces:
  - `DPS-11170`: successfully extracted multiple root sessions (12 invocations, 25 SDLC iterations).
  - `DPS-11173`, `DPS-11187`, `DPS-11197`, `DPS-11199`: preserved historical invocations intact without title hijacking.
