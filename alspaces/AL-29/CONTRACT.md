# Contract: AL-29 Conventional Commit Ticket Type Classification in Telemetry Engine

## 1. Problem Statement

In the existing implementation, the AgenticLens CLI and git inference engine (`InferTicketType` in `backend/internal/adapters/git/inference.go`) silently default any unmatched ticket to `ticket_type: "feature"`. In addition, `InferTicketType` executes a heuristic git log query (`git log --all --grep=<cleanTicket>`) attempting to grep commit messages for keywords such as "feat:", "fix:", or "chore:".

This legacy heuristic has three fatal flaws:
1. Commit message grepping is brittle and nondeterministic. Arbitrary text in past commit messages frequently misclassifies new or unrelated sessions.
2. Silent defaulting to `feature` falsifies telemetry facts. Product Owner scoping sessions (`PO-#`) and general engineering maintenance tickets (`GEN-#`) are erroneously recorded as feature developments in `agentic_metrics.json` and `SUMMARY.md`.
3. The CLI lacks input normalization and validation for `--ticket-type`, allowing verbose non-standard tokens (`feature`, `bug`, `documentation`) to leak into persistent ledgers rather than enforcing canonical Conventional Commit standards (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).

Ticket AL-29 refactors ticket type resolution into a deterministic 3-tier hierarchy:
- Tier 1: Caller-provided `--ticket-type` command-line flag or cached facts, normalized and validated.
- Tier 2: Git branch prefix inspection (`feat/*`, `fix/*`, `docs/*`, `chore/*`, `refactor/*`, `test/*`, `ci/*`, `perf/*`, `build/*`).
- Tier 3: Workspace ticket identifier prefix mapping (`PO-*` to `docs`, `GEN-*` to `chore`).

Commit message grepping is banished. The fallback chain stops strictly at the workspace ticket prefix. When a ticket type cannot be determined from the 3 tiers, the telemetry engine prints an explicit warning to standard error, persists `nil` (unclassified) rather than fabricating `feature`, and provides a `--strict-type` option that terminates execution with a non-zero exit code.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes across Go backend layers must strictly adhere to the inward dependency principles established in `TECHNICAL_GUIDELINES.md`:

### Pure Functional Core Layer (`backend/internal/domain/`)
- Typed Constants and Enums:
  Define `ConventionalTicketType` string enum in `backend/internal/domain/enums.go` with canonical tokens: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`.
- Pure Normalization and Validation:
  Implement pure mapping functions in `backend/internal/domain/enums.go`:
  - `NormalizeTicketType(raw string) (ConventionalTicketType, bool)`: Trims whitespace, lowercases input, maps recognized verbose aliases (`feature` to `feat`, `bug`/`bugfix` to `fix`, `documentation`/`doc` to `docs`, `task` to `chore`, `performance` to `perf`, `tests` to `test`), and verifies against canonical tokens.
  - `IsValidTicketType(raw string) bool`: Reports whether a normalized token belongs to the canonical set.
  - `CanonicalTicketTypes() []ConventionalTicketType`: Returns an ordered slice of all supported conventional commit types.
  - `InferTicketTypeFromPrefix(ticketID string) (ConventionalTicketType, bool)`: Pure prefix inspector mapping `PO-*` (case-insensitive) to `docs` and `GEN-*` (case-insensitive) to `chore`.
- Invariant: Zero external dependencies, zero disk I/O, zero git invocation, zero system clocks, and zero mutable state. Pure functions operate exclusively on immutable strings and return value objects.

### Imperative Adapters Layer (`backend/internal/adapters/git/`)
- Git Branch Inspection Refactor (`backend/internal/adapters/git/inference.go`):
  - Retain branch search via `runGitCommand("branch", "-a", "--list", "*<cleanTicket>*")`.
  - Extract branch prefix (stripping `remotes/` and `origin/`).
  - Map branch prefix to canonical `ConventionalTicketType` via `domain.NormalizeTicketType`.
  - Fall back to `domain.InferTicketTypeFromPrefix(cleanTicket)`.
  - Completely delete commit message inspection (`git log --grep`).
  - Delete silent fallback returning `feature`. Return `nil` when neither branch prefix nor workspace ticket prefix matches.
- Invariant: Encapsulates all git subprocess execution within the adapter boundary. Never leaks raw git errors into domain logic.

### CLI Application Layer (`backend/cmd/agentic-lens/main.go`)
- Command-line Flag Resolution in `runGenerate`:
  - Register optional boolean flag `--strict-type` (default false): forces command failure if ticket type is unresolved.
  - Evaluate Tier 1: Check `--ticket-type` flag. If supplied, normalize with `domain.NormalizeTicketType`. If invalid, print error message to standard error and terminate with exit status 1.
  - If flag is omitted, evaluate cached facts ticket type if present and valid.
  - Evaluate Tier 2 and Tier 3: Call `git.InferTicketType(activeTicketID)`.
  - If ticket type remains unresolved (`nil`):
    - If `--strict-type` is true, write error to standard error: `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` and terminate with `os.Exit(1)`.
    - If `--strict-type` is false, write warning to standard error: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.` and retain `activeTicketType = nil`.
- Invariant: Enforces execution gating and user warnings at the CLI boundary prior to persisting ledgers. Prevents fabricated data from entering `domain.AgenticFactLedger`.

### Presentation Layer (`web/`)
- Frontend contracts in `web/src/domain/types.ts` already support nullable string types (`readonly ticket_type: string | null`).
- Presentation components display the resolved type or fallback cleanly.
- Web layer remains untouched in this backend ticket.

---

## 3. Scope Boundaries: Go Backend and Tests

### In Scope
1. Domain Enums and Normalization (`backend/internal/domain/enums.go`):
   - Canonical types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`.
   - Normalization mapping covering aliases: `feature` -> `feat`, `bug`/`bugfix` -> `fix`, `doc`/`documentation` -> `docs`, `task` -> `chore`, `perf`/`performance` -> `perf`, `test`/`tests` -> `test`.
   - Prefix classifier: `PO-*` -> `docs`, `GEN-*` -> `chore`.
   - Unit tests covering canonical normalization, alias resolution, prefix parsing, and invalid strings.
2. Git Boundary Adapter Refactor (`backend/internal/adapters/git/inference.go`):
   - Removal of `runGitCommand("log", "--all", "--grep=...")`.
   - Branch prefix extraction with canonical token normalization.
   - Elimination of silent `"feature"` fallback.
   - Dedicated unit tests in `backend/internal/adapters/git/inference_test.go` and updates in `backend/internal/adapters/git/analyzer_test.go`.
3. CLI Flags and Error Reporting (`backend/cmd/agentic-lens/main.go`):
   - Registration and handling of `--strict-type`.
   - Normalization and validation of `--ticket-type`.
   - Stderr warning logging for unresolved types: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.`.
   - Stderr error logging and non-zero exit status for `--strict-type` violations.
4. CLI End-to-End Tests (`backend/cmd/agentic-lens/main_test.go`):
   - Validation that `PO-5` produces `docs` without explicit flag.
   - Validation that `GEN-1` produces `chore` without explicit flag.
   - Validation that `feat/*` branch produces `feat`.
   - Validation that unmatched tickets without prefix log stderr warning and record `ticket_type: null`.
   - Validation that unmatched tickets with `--strict-type` exit with status 1 and error message.
   - Validation that `--ticket-type feature` normalizes to `feat`.

### Out of Scope
1. Web frontend dashboard modifications or new UI components (`web/`).
2. Telemetry schema version bump (the ledger field remains `ticket_type *string`).
3. Historical ledger backfill or rewriting of past workspace files (`internal/adapters/backfill/`).
4. SQLite schema changes or modifications to OpenCode queries.

---

## 4. Architectural Specifications

### Resolution Precedence Hierarchy
1. Tier 1 (Explicit Input):
   - Read `--ticket-type` CLI argument.
   - If provided: invoke `domain.NormalizeTicketType`. If valid, assign normalized token. If invalid, abort with error and exit status 1.
   - If omitted: inspect cached facts from `--json-output`. If non-empty, normalize.
2. Tier 2 (Git Branch Prefix):
   - Run `git branch -a --list *<cleanTicket>*`.
   - Parse branch prefix preceding slash.
   - Normalize prefix via `domain.NormalizeTicketType`.
   - If valid conventional type, assign token.
3. Tier 3 (Workspace Ticket Prefix):
   - Inspect ticket identifier pattern.
   - Case-insensitive `PO-*` -> `docs`.
   - Case-insensitive `GEN-*` -> `chore`.
   - Other prefixes (e.g. `AL-*`) do not map to fixed types.
4. Fallback Termination:
   - If unresolved, return `nil`.
   - If `--strict-type` is active: print error to `os.Stderr` and call `os.Exit(1)`.
   - If `--strict-type` is inactive: print warning to `os.Stderr` and persist `nil`.

---

## 5. Acceptance Criteria

1. Commit message grepping (`git log --grep`) is completely deleted from `InferTicketType`.
2. `git.InferTicketType("PO-5")` resolves deterministically to `"docs"`.
3. `git.InferTicketType("GEN-42")` resolves deterministically to `"chore"`.
4. Branch prefix inspection returns canonical tokens (`feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `ci`, `perf`, `build`).
5. Unmatched tickets on branches lacking conventional prefixes return `nil` and are never defaulted to `"feature"`.
6. Invoking `agentic-lens generate` with an unresolvable ticket type emits `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.` to stderr and writes `null` for `ticket_type` in the JSON ledger.
7. Invoking `agentic-lens generate` with `--strict-type` on an unresolvable ticket exits with status 1 and writes `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` to stderr.
8. Supplying `--ticket-type feature` normalizes the stored fact to `"feat"`.
9. `make test-backend`, `go vet ./...`, and `make check` pass with zero errors.
10. Strict adherence to technical guidelines: zero `any` or empty `interface{}`, and zero emojis across code, docstrings, error messages, and tests.
