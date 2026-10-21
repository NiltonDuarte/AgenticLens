# Pull Request: AL-11 Native Go Telemetry Extractor and Active Session Metric Generation

## Description
This pull request implements ticket AL-11, delivering a native Go telemetry extraction engine and active session generator directly within `agentic-lens` CLI (`generate` subcommand). This eliminates the dependency on the external Python script (`scripts/extract_agentic_metrics.py`), moving the project toward single-binary operational independence.

For detailed architecture boundaries and delivery summary, refer to:
- Contract: `workspaces/AL-11/CONTRACT.md`
- Summary & Metrics: `workspaces/AL-11/SUMMARY.md`

## Summary of Changes
1. **Pure Go SQLite Driver Integration**:
   - Integrated `modernc.org/sqlite` into `go.mod` to guarantee `CGO_ENABLED=0` static binary compilation.
2. **OpenCode SQLite Adapter (`internal/adapters/opencode/`)**:
   - `client.go`: Pure Go read-only SQLite connection management with `mode=ro` and `query_only=1`.
   - `resolver.go`: 4-tier root session resolution hierarchy (Nonce -> Session ID -> Ticket ID -> Most recent fallback).
   - `extractor.go`: Extracted invocations, token breakdowns (input, output, cache read, reasoning, peak context), duration bounds, financial costs, compactions, and tool execution facts.
   - `extractor_test.go`: In-memory SQLite fixture tests covering prioritized resolution and fact extraction parity.
3. **Multi-Language Git Adapter (`internal/adapters/git/`)**:
   - `classifier.go`: Universal file classification distinguishing production code, test code, and ignored files, with optional configuration overrides from `.agenticlens/config.json`.
   - `analyzer.go`: Git numstat resolution across merge-base diffs, branch inspection, and commit grep.
   - `inference.go`: Title and ticket type inference from workspace artifacts and git log, SDLC iteration counting, and git user resolution.
   - `analyzer_test.go`: Unit tests for classifier and inference logic.
4. **Repository Adapter Extension (`internal/adapters/repository.go`)**:
   - Implemented `LoadExistingFactLedger` to read existing `agentic_metrics.json` as fallback or baseline.
5. **CLI Subcommand Engine (`cmd/agentic-lens/main.go`)**:
   - Wired `generate` subcommand supporting `--ticket-id`, `--nonce`, `--session-id`, `--ticket-title`, `--ticket-type`, `--started-at`, `--completed-at`, `--json-output`, `--append`, and `--db-path`.
   - Added CLI integration tests in `cmd/agentic-lens/main_test.go` verifying JSON output and markdown table idempotency.
6. **Build Hardening (`Makefile`)**:
   - Hardened Makefile frontend targets to execute only when `web/node_modules` exists.

## Verification
- `go vet ./...` passed.
- `CGO_ENABLED=0 go test ./...` passed.
- `go test -race ./...` passed.
- `make check && make test && make build` passed.
- Single-binary execution `./bin/agentic-lens generate --ticket-id AL-11 --json-output workspaces/AL-11/agentic_metrics.json --append workspaces/AL-11/SUMMARY.md` ran and succeeded natively.
