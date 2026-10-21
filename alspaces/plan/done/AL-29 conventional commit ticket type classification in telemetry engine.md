h3. Title: AL-29: Conventional Commit Ticket Type Classification in Telemetry Engine

*TARGET*: AgenticLens (cli | adapters | domain | tests)

*CONTEXT*

Currently, the AgenticLens CLI and git inference engine (`InferTicketType`) silently default unmatched tickets to `ticket_type: "feature"`. Furthermore, the existing heuristic attempts to grep git commit logs, which is brittle and produces false classifications. Consequently, non-feature sessions—such as Product Owner scoping (`PO-#`) and generic maintenance tasks (`GEN-#`)—are erroneously recorded as `feature` in telemetry ledgers. The telemetry engine must eliminate unreliable commit message heuristics, stop fallbacks at workspace ticket prefixes, and enforce a strict resolution hierarchy with explicit warnings or errors for missing ticket types.

*GOAL*

Refactor ticket type resolution to use a deterministic 3-tier hierarchy (explicit flag -> branch prefix -> workspace ticket prefix), banish commit message inspection, eliminate silent fallbacks to "feature", and return explicit warnings/errors when a ticket type cannot be resolved.

*VERTICAL SLICE SCOPE*

* Domain & Model:
  - Add typed constants/enums for standard Conventional Commit types in `internal/domain`: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`.
  - Provide normalization helpers mapping verbose aliases to conventional tokens (`feature` -> `feat`, `bug`/`bugfix` -> `fix`, `documentation` -> `docs`).
* Boundary Adapter & Git Inference:
  - Remove git commit message inspection (`git log --grep`) from `InferTicketType`. Do not rely on git commit logs for classification.
  - Establish deterministic 3-tier resolution:
    1. Caller-provided `--ticket-type` flag (highest priority).
    2. Git branch prefix (`feat/*`, `fix/*`, `docs/*`, `chore/*`, `refactor/*`, `test/*`, `ci/*`, `perf/*`, `build/*`).
    3. Workspace ticket ID prefix:
       - `PO-*` (case-insensitive) -> `docs`
       - `GEN-*` (case-insensitive) -> `chore`
  - Stop the fallback chain at workspace ticket prefix. Zero silent defaulting to `feature` or arbitrary guesses.
* CLI Behavior & Warning/Error Reporting:
  - If a ticket type cannot be determined from the 3 tiers:
    - Emit an explicit warning to stderr: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.`
    - Do not fabricate `ticket_type: "feature"`. Persist `nil` / unclassified.
    - Support `--strict-type` flag (or default validation) returning non-zero error when `--ticket-type` is omitted and cannot be inferred.
* Quality:
  - Unit tests in `inference_test.go` and `main_test.go`:
    - `PO-5` resolves to `docs`.
    - `GEN-1` resolves to `chore`.
    - `feat/AL-1` branch resolves to `feat`.
    - Unmatched ticket on branch without prefix yields `nil` with warning/error (no silent "feature" fallback).
    - Commit message contents are ignored and do not affect classification.

*ACCEPTANCE CRITERIA*

* Commit message inspection is eliminated from `InferTicketType`.
* `git.InferTicketType("PO-5")` resolves deterministically to `"docs"`.
* `git.InferTicketType("GEN-42")` resolves deterministically to `"chore"`.
* Branch prefix mapping returns canonical Conventional Commit tokens (`feat`, `fix`, `refactor`, `chore`, `docs`).
* Tickets that cannot be resolved via branch or prefix emit an explicit warning/error and are NOT defaulted to `feature`.
* CLI `agentic-lens generate --ticket-type <TYPE>` normalizes and validates conventional commit inputs.
* `make test-backend`, `go vet ./...`, and `make check` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
