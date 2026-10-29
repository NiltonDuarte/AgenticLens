# Iteration 01 Plan: Strict Session ID Resolution and Multi-Root Backfill

## Overview
Implement multi-root session discovery and gate heuristic fallbacks behind an explicit `--allow-fallbacks` CLI option.

## Proposed Changes

### 1. `backend/internal/adapters/backfill/options.go`
- Add `AllowFallbacks bool` field to `BackfillOptions`.

### 2. `backend/cmd/agentic-lens/main.go`
- Add `--allow-fallbacks` flag to `backfillFlags` in `runBackfill`.
- Pass `AllowFallbacks` into `backfill.BackfillOptions`.
- Document flag in `backfillFlags.Usage()`.

### 3. `backend/internal/adapters/backfill/engine.go`
- Modify OpenCode session resolution:
  - Collect all candidate root session IDs across `openCodeSessions` and `existingOpenCodeInvocations`.
  - For each candidate ID, check if it exists in the SQLite database and trace up to its root if it has a parent.
  - Dedup resolved root session IDs into a slice `resolvedRootIDs`.
  - Only execute Strategy 2 (`opencode.ResolveActiveRootSession`) and Strategy 3 (`ResolveSessionByWorkspacePath`) if `len(resolvedRootIDs) == 0 && opts.AllowFallbacks`.
  - When `len(resolvedRootIDs) > 0`, query facts across `resolvedRootIDs` simultaneously using `opencode.ExtractFacts` and `opencode.ExtractSteeringTurns`.
  - Safely replace or merge refreshed session trees without discarding unresolved sessions from other hosts.
  - Set `dbSessionFound = true` only when resolved sessions are extracted.

### 4. `backend/internal/adapters/backfill/engine_test.go`
- Add tests:
  - `TestBackfillEngine_MultiRootSessionsExtracted`: verify that multiple root sessions in a single ledger are all extracted and preserved without truncation.
  - `TestBackfillEngine_MissingSessionIDPreservesFactsWithoutFallbacks`: verify that if a ledger contains a session ID not present in SQLite DB, and `--allow-fallbacks` is false, it does NOT latch onto another session matching the ticket title.
  - `TestBackfillEngine_ExplicitFallbacksEnabled`: verify that when `--allow-fallbacks` is true, title matching resolves as a fallback.

### 5. Verification and Integration
- Run `go test ./...` in `backend/`.
- Compile multi-platform binaries (`make build` or `go build`).
- Test in target repository (`/Users/guedesduarte/workspace/git/SinkLayerControlPlane.Code`) on `workspaces/`.
