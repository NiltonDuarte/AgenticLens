h3. Title: AL-30: Fail Telemetry Generation on Unresolved Root Session

*TARGET*: AgenticLens (cli | adapters | main | tests)

*CONTEXT*

Currently, when `agentic-lens generate` is executed, if the telemetry engine fails to resolve a root session from the OpenCode database (via nonce, explicit session ID, ticket title, or workspace path) and no cached facts exist, it silently proceeds with zero invocations, zero agent duration, zero token counts, and zero tools. Having no agentic telemetry completely defeats the purpose of the tool and corrupts the telemetry ledger with empty, misleading records. Generating an empty session must not silently succeed.

*GOAL*

Enforce strict root session resolution in `agentic-lens generate`: the CLI must fail with a non-zero exit code and an informative error message if no root session can be identified from the provided information (or cached in an existing ledger), preventing the creation of empty, zero-invocation metrics.

*VERTICAL SLICE SCOPE*

* CLI & Execution Flow (`cmd/agentic-lens/main.go`):
  - After session resolution and database extraction, check if any root session was identified or any invocations were loaded (either fresh from DB or preserved in existing ledger).
  - If `len(knownThreadIDs) == 0` and `len(invocations) == 0`:
    - Print an actionable error to `os.Stderr`: `Error: unable to identify root session for ticket '<TICKET_ID>'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted.`
    - Abort execution and exit with non-zero exit code (`os.Exit(1)`).
    - Do NOT write or overwrite output files (`--json-output` or `--append`).
  - Support an optional escape hatch flag `--allow-empty` (strictly defaulting to `false`) for isolated synthetic unit tests where a live database is intentionally bypassed.
* Boundary Adapter (`internal/adapters/opencode/resolver.go`):
  - Integrate workspace path matching (`workspaces/<ticket>/`) as an automatic fallback in `ResolveActiveRootSession` when ticket title does not match, maximizing automatic session detection before failing.
* Quality & Tests:
  - Unit/CLI tests in `main_test.go`:
    - Calling `generate` with a non-existent ticket ID and no session match must exit with status 1 and emit the error message on stderr.
    - Verified that no JSON ledger or markdown table is generated or overwritten on failure.
    - Verified that valid tickets with identifiable sessions continue to succeed.

*ACCEPTANCE CRITERIA*

* `agentic-lens generate --ticket-id NON_EXISTENT` exits with status 1 and logs error to stderr.
* No JSON or markdown files are created or modified when root session identification fails.
* Automatic fallback in session resolver checks workspace path (`workspaces/<ticket>/`) before failing.
* `make test-backend`, `go vet ./...`, and `make check` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
