# FIX-10: Backfill Multi-Root Sessions and Require Explicit Command Line Fallbacks

## Subsystem
Backend / Adapters / Backfill / CLI / Tests

## Phase
Phase 7 (v1 Defect Remediation)

## Status
In Progress

## Blocked by
None

## Blocks
None

## Description
AgenticLens `backfill` engine currently exhibits two critical data integrity defects:
1. Truncates multi-root session ledgers: in `engine.go`, the backfill loop stops at the first root session found (`break`), omitting secondary and later root sessions. This causes significant data and cost loss (e.g. dropping root sessions and their subagents).
2. Unsafe implicit heuristic fallbacks: when session IDs are not found in the local database (such as when working across multiple developer machines), the backfill automatically falls back to title fuzzy matching (`ResolveActiveRootSession`) and workspace path matching, inadvertently matching unrelated review or auxiliary sessions and overwriting valid historical invocations.

## Acceptance Criteria
1. The backfill engine discovers and queries ALL root sessions present in the existing ledger rather than breaking on the first root session.
2. The backfill engine strictly trusts exact session IDs by default. Heuristic title matching and commit date/path fallbacks are disabled by default.
3. An explicit CLI flag (`--allow-fallbacks`) is added to `agentic-lens backfill` to gate heuristic fallback matching.
4. When exact session IDs do not exist in the local SQLite database and `--allow-fallbacks` is not passed, existing invocations and events in the ledger are preserved without corruption, and migration status is evaluated as pending host migration.
5. Unit tests cover multi-root session backfill and verify that unflagged runs do not trigger heuristic fallbacks when session IDs are missing.
