# fix(cli): fail telemetry generation on unresolved root session (AL-30)

## Overview
Enforces strict root session resolution in the `agentic-lens generate` CLI subcommand. Previously, when the telemetry engine failed to identify a root session from OpenCode SQLite and lacked cached facts, it silently generated empty metrics with zero invocations, corrupting telemetry ledgers.

The CLI now terminates with exit code 1 and outputs an actionable error message to standard error when no root session or invocations are resolved. Output files (`--json-output` or `--append`) are protected from creation or modification on failure. An escape hatch flag (`--allow-empty`) is introduced for isolated synthetic tests. In addition, the OpenCode resolver integrates automatic workspace path fallback (`workspaces/<ticket>/`) under Priority 3 when ticket title lookup returns no rows, maximizing automatic session detection before failure.

- Contract: [workspaces/AL-30/CONTRACT.md](workspaces/AL-30/CONTRACT.md)
- Telemetry: [workspaces/AL-30/SUMMARY.md](workspaces/AL-30/SUMMARY.md)

## Summary of Changes

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- Enforced session resolution gate in `runGenerate`: aborts with exit code 1 if `len(knownThreadIDs) == 0 && len(invocations) == 0 && !*allowEmptyFlag`.
- Standardized error message to `os.Stderr`:
  `Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`
- Positioned validation prior to code metrics computation, ledger derivation, and disk writes, guaranteeing that output files are neither created nor modified on failure.
- Registered boolean flag `--allow-empty` (default `false`) as an explicit escape hatch for synthetic tests.

### Boundary Adapters Layer (`backend/internal/adapters/opencode/`)
- Enhanced `ResolveActiveRootSession` in `resolver.go` with Priority 3 automatic fallback: calls `ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)` when ticket title lookup yields `sql.ErrNoRows`.
- Wrapped fallback database errors with `resolving session by workspace path fallback: %w`.
- Updated function documentation for `ResolveActiveRootSession`.

### Quality & Tests (`backend/cmd/agentic-lens/`, `backend/internal/adapters/opencode/`)
- Added `TestCLIGenerate_UnresolvedSessionEnforcement` covering non-existent tickets, missing ticket ID (`UNKNOWN`), output file protection, summary file preservation, `--allow-empty` escape hatch, valid ticket resolution, cached invocation bypass, and fallback database errors.
- Added unit tests in `extractor_test.go` covering workspace path fallback resolution, unresolved ticket handling, and error wrapping on database failures.
- Updated existing synthetic unit tests in `main_test.go` to supply `--allow-empty`.
- Fully compliant with FCIS architecture, zero `any` or empty `interface{}`, and zero emojis.
