# FIX-10 Contract: Backfill Multi-Root Sessions and Require Explicit Command Line Fallbacks

## Problem Statement
The AgenticLens `backfill` engine currently possesses two data integrity bugs:
1. Multi-root session truncation: When backfilling an OpenCode ticket ledger with multiple root sessions (e.g. iterative human steering turns or restart sessions where `parent_id` is null), `engine.go` loops through invocations and performs a `break` upon finding the first root session. Consequently, only that first root session is passed to `opencode.ExtractFacts`, dropping all subsequent root sessions and their child subagent trees, causing massive data loss.
2. Unsafe implicit heuristic fallbacks: If the session ID recorded in the ledger does not exist in the local SQLite database (common in multi-developer or multi-machine environments), `engine.go` automatically invokes `opencode.ResolveActiveRootSession` searching for root sessions with `WHERE title LIKE '%<TICKET_ID>%'`, followed by commit range / workspace path heuristics. This can latch onto unrelated review sessions (such as code review or linting branches) that mention the ticket identifier, corrupting valid historical invocation counts.

## Intended Behavior
1. Backfill discovers and extracts ALL unique root sessions present in the existing ledger.
2. Backfill strictly trusts exact session IDs by default. Heuristic fallbacks (title substring match, commit date/path match) are completely disabled unless an explicit CLI flag (`--allow-fallbacks`) is specified.
3. If session IDs from the ledger are not present in the local database and `--allow-fallbacks` is not provided, the backfill engine treats local DB extraction as unresolved for those sessions, preserving the existing session facts and invocations without corruption, and evaluating migration status as pending host migration.
4. If multiple root sessions exist and all (or a subset) are resolved, facts are cleanly refreshed without dropping multi-root data.

## Boundaries and Invariants
- FCIS architecture: `domain` remains pure, while `internal/adapters/backfill` manages orchestration.
- CLI consistency: Flags added to `backfill` must follow existing flag conventions in `main.go`.
- Zero data loss: Existing invocations and tool facts must never be overwritten by unrelated sessions.
- Zero emojis across code, comments, summaries, and commit messages.

## Acceptance Criteria
- [ ] `BackfillOptions` includes `AllowFallbacks bool`.
- [ ] CLI flag `--allow-fallbacks` is supported by `agentic-lens backfill`.
- [ ] `engine.go` extracts facts for all root session IDs present in the ledger.
- [ ] Heuristic session resolution (title matching and commit date range) is skipped when `AllowFallbacks` is false.
- [ ] Missing session IDs in local database without `--allow-fallbacks` preserve existing facts intact.
- [ ] Comprehensive unit tests verify multi-root resolution, missing session ID preservation, and explicit fallback behavior.
