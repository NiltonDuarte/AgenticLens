# Implementation Plan: AL-29 Conventional Commit Ticket Type Classification in Telemetry Engine

## 1. Goal and Iteration Scope

Implement deterministic conventional commit ticket type classification in the AgenticLens Go backend. Replace brittle git log commit heuristics and silent "feature" fallbacks with a strict 3-tier resolution engine (explicit flag -> branch prefix -> workspace prefix). Enforce canonical conventional commit types, provide alias normalization, add the `--strict-type` verification flag, and report explicit warnings when ticket types are unclassified.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation spans the following Go backend files and packages:

### Pure Functional Core Layer (`backend/internal/domain/`)
- `backend/internal/domain/enums.go`:
  - Define `ConventionalTicketType` string enum and constants: `TicketTypeFeat`, `TicketTypeFix`, `TicketTypeDocs`, `TicketTypeChore`, `TicketTypeRefactor`, `TicketTypeTest`, `TicketTypeCI`, `TicketTypePerf`, `TicketTypeBuild`.
  - Implement pure functions: `NormalizeTicketType`, `IsValidTicketType`, `CanonicalTicketTypes`, and `InferTicketTypeFromPrefix`.
- `backend/internal/domain/enums_test.go`:
  - Table-driven unit tests verifying normalization of canonical names and aliases (`feature` to `feat`, `bug` to `fix`, `documentation` to `docs`), rejection of invalid inputs, and case-insensitive prefix classification (`PO-5` to `docs`, `GEN-42` to `chore`, `AL-11` to false).

### Boundary Adapters Layer (`backend/internal/adapters/git/`)
- `backend/internal/adapters/git/inference.go`:
  - Refactor `InferTicketType`: remove git log command execution and commit scanning.
  - Parse branch prefix and normalize via `domain.NormalizeTicketType`.
  - Integrate `domain.InferTicketTypeFromPrefix` for ticket ID prefix resolution.
  - Return `nil` when unmatched (remove fallback defaulting to "feature").
- `backend/internal/adapters/git/inference_test.go`:
  - Dedicated unit tests validating deterministic prefix resolution, branch parsing, commit message ignorance, and unclassified `nil` returns.
- `backend/internal/adapters/git/analyzer_test.go`:
  - Update `TestInferenceOnActiveWorkspace` to assert `nil` or conventional token rather than expecting legacy `"feature"` on unmatched workspaces.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- `backend/cmd/agentic-lens/main.go`:
  - In `runGenerate`, declare `--strict-type` boolean flag.
  - Validate and normalize `--ticket-type` via `domain.NormalizeTicketType`. Terminate with error on invalid value.
  - If flag is missing, evaluate cached facts or call `git.InferTicketType`.
  - If still `nil`: check `--strict-type`. If true, log error and exit with status 1. If false, log warning to stderr: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.`.
- `backend/cmd/agentic-lens/main_test.go`:
  - Integration tests for CLI generate: alias normalization, invalid type rejection, warning on unclassified ticket, `--strict-type` failure, and automatic `PO-*`/`GEN-*` resolution.

---

## 3. Phased Implementation Tasks

### Phase 1: Pure Domain Enums, Aliases, and Normalization
- Task 1.1: Define Conventional Commit Enums in `backend/internal/domain/enums.go`:
  - Declare type `ConventionalTicketType string`.
  - Declare constants:
    - `TicketTypeFeat` = "feat"
    - `TicketTypeFix` = "fix"
    - `TicketTypeDocs` = "docs"
    - `TicketTypeChore` = "chore"
    - `TicketTypeRefactor` = "refactor"
    - `TicketTypeTest` = "test"
    - `TicketTypeCI` = "ci"
    - `TicketTypePerf` = "perf"
    - `TicketTypeBuild` = "build"
  - Declare slice `CanonicalTicketTypes` containing all 9 tokens in logical order.
- Task 1.2: Implement `NormalizeTicketType` in `backend/internal/domain/enums.go`:
  - Receive raw string parameter.
  - Trim whitespace and convert to lowercase.
  - Switch on trimmed value:
    - "feat", "feature" -> return `TicketTypeFeat`, true
    - "fix", "bug", "bugfix" -> return `TicketTypeFix`, true
    - "docs", "doc", "documentation" -> return `TicketTypeDocs`, true
    - "chore", "task" -> return `TicketTypeChore`, true
    - "refactor" -> return `TicketTypeRefactor`, true
    - "test", "tests" -> return `TicketTypeTest`, true
    - "ci" -> return `TicketTypeCI`, true
    - "perf", "performance" -> return `TicketTypePerf`, true
    - "build" -> return `TicketTypeBuild`, true
    - default -> return empty string, false
- Task 1.3: Implement `IsValidTicketType` in `backend/internal/domain/enums.go`:
  - Calls `NormalizeTicketType(raw)` and returns the boolean validity.
- Task 1.4: Implement `InferTicketTypeFromPrefix` in `backend/internal/domain/enums.go`:
  - Receive `ticketID string`.
  - Trim whitespace and convert to uppercase.
  - If string has prefix "PO-", return `TicketTypeDocs`, true.
  - If string has prefix "GEN-", return `TicketTypeChore`, true.
  - Otherwise return empty string, false.
- Task 1.5: Unit Tests in `backend/internal/domain/enums_test.go`:
  - Test all canonical types return themselves and true.
  - Test aliases ("feature", "bug", "bugfix", "documentation", "task") normalize correctly.
  - Test invalid strings ("unknown", "epic", "story", "") return false.
  - Test prefix inference: "PO-1", "po-42", "GEN-100", "gen-5" match expected types; "AL-1", "PROJ-1" return false.
  - Execute `go test -v -race ./backend/internal/domain/...`.

### Phase 2: Git Boundary Adapter Refactoring
- Task 2.1: Refactor `InferTicketType` in `backend/internal/adapters/git/inference.go`:
  - Clean input ticket ID. Return `nil` if empty.
  - Tier 2 (Branch inspection):
    - Run `git branch -a --list *<cleanTicket>*`.
    - If command succeeds, iterate lines, strip prefix asterisks and spaces, split on slash.
    - Filter out "remotes". If first segment is "origin" and second exists, use second segment; otherwise use first segment.
    - Invoke `domain.NormalizeTicketType(prefix)`.
    - If valid, convert to string pointer and return immediately.
  - Tier 3 (Workspace ticket prefix):
    - Invoke `domain.InferTicketTypeFromPrefix(cleanTicket)`.
    - If valid, convert to string pointer and return immediately.
  - Delete git log search block (`git log --all --grep...`).
  - Delete fallback creating "feature".
  - Return `nil`.
- Task 2.2: Add Dedicated Adapter Tests in `backend/internal/adapters/git/inference_test.go`:
  - Unit test `InferTicketType("PO-5")` asserts returned string pointer equals "docs".
  - Unit test `InferTicketType("GEN-42")` asserts returned string pointer equals "chore".
  - Unit test `InferTicketType` with empty ticket ID asserts returned value is nil.
  - Unit test verifying unmatched ticket ID with no matching branch returns nil (no silent "feature" fallback).
- Task 2.3: Update Existing Tests in `backend/internal/adapters/git/analyzer_test.go`:
  - Inspect `TestInferenceOnActiveWorkspace`: update assertion on `InferTicketType("AL-11")` to account for branch state or assert nil when not on a `feat/*` branch.
  - Run `go test -v -race ./backend/internal/adapters/git/...`.

### Phase 3: CLI Ticket Type Validation, Strict Enforcement, and Warning Logging
- Task 3.1: Register `--strict-type` Flag in `backend/cmd/agentic-lens/main.go`:
  - In `runGenerate`, declare `strictTypeFlag := genFlags.Bool("strict-type", false, "Require a valid ticket type, failing if omitted and cannot be inferred")`.
- Task 3.2: Refactor Ticket Type Resolution in `runGenerate`:
  - Priority 1: Check `--ticket-type` flag.
    - If non-empty: invoke `domain.NormalizeTicketType(*ticketTypeFlag)`.
    - If invalid: write formatted error to `os.Stderr`: `Error: invalid ticket type '<VALUE>'. Must be one of: feat, fix, docs, chore, refactor, test, ci, perf, build` and terminate via `os.Exit(1)`.
    - If valid: assign pointer to normalized string representation.
  - If flag is empty:
    - Check `cachedFacts != nil && cachedFacts.TicketType != nil`.
    - If present, attempt `domain.NormalizeTicketType(*cachedFacts.TicketType)`. If valid, assign normalized string.
  - Priority 2 & 3: If still `nil`, call `git.InferTicketType(activeTicketID)`.
    - If non-nil, assign returned pointer.
  - Priority 4: If still `nil`:
    - If `*strictTypeFlag`: write error to `os.Stderr`: `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` and terminate via `os.Exit(1)`.
    - If not strict: write warning to `os.Stderr`: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.`.
    - Leave `activeTicketType = nil`.
- Task 3.3: Ensure Fact Ledger Assignment:
  - Assign `factLedger.TicketType = activeTicketType`. Ensure JSON marshaling produces `null` when nil, never fabricated `"feature"`.

### Phase 4: CLI Integration and Adversarial Edge Case Tests
- Task 4.1: Add CLI Tests in `backend/cmd/agentic-lens/main_test.go`:
  - Test explicit `--ticket-type feature` normalizes to `"feat"` in generated JSON ledger.
  - Test explicit `--ticket-type invalid-type` exits with status 1 and prints list of valid conventional commit types.
  - Test automatic resolution of `PO-5` records `ticket_type: "docs"` without `--ticket-type` flag.
  - Test automatic resolution of `GEN-1` records `ticket_type: "chore"` without `--ticket-type` flag.
  - Test unmatched ticket on detached HEAD or untracked branch:
    - Without `--strict-type`: exits with status 0 (with `--allow-empty`), logs warning containing `WARNING: Missing ticket type for ticket`, and output JSON contains `ticket_type: null`.
    - With `--strict-type`: exits with status 1 and error message containing `Error: missing ticket type for ticket`.

### Phase 5: Verification, Quality Gates, and Standards Compliance
- Task 5.1: Execute `cd backend && go vet ./...`.
- Task 5.2: Execute `cd backend && go test -v -race ./...`.
- Task 5.3: Run repository quality gate: `make check`.
- Task 5.4: Verify zero `any` or empty `interface{}` across Go backend.
- Task 5.5: Verify zero emojis in code, logs, comments, and tests.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Enums and Normalization) builds the pure functional foundation with zero dependencies.
2. Phase 2 (Git Boundary Adapter Refactoring) updates `InferTicketType` to use domain normalization and prefix rules while removing commit log inspection.
3. Phase 3 (CLI Ticket Type Validation and Enforcement) integrates flag normalization, `--strict-type` gating, and stderr warning reporting into `cmd/agentic-lens/main.go`.
4. Phase 4 (CLI Integration Tests) validates end-to-end command execution across all acceptance criteria and adversarial edge cases.
5. Phase 5 (Verification and Quality Gates) validates compile cleanliness, test coverage, and guideline compliance.

---

## 5. Verification Plan

### Test Commands
- Domain unit tests: `cd backend && go test -v -race ./internal/domain/...`
- Git adapter tests: `cd backend && go test -v -race ./internal/adapters/git/...`
- CLI command tests: `cd backend && go test -v -race -run TestCLITicketType ./cmd/agentic-lens/...`
- Full backend suite: `make test-backend`
- Static analysis: `make check`

### Manual Verification Scenarios
- Run generate with alias: `./bin/agentic-lens generate --ticket-id AL-29 --ticket-type feature --allow-empty --json-output test.json` -> verify `ticket_type` in `test.json` is `"feat"`.
- Run generate with invalid type: `./bin/agentic-lens generate --ticket-id AL-29 --ticket-type unknown --allow-empty` -> verify status 1 and stderr message listing valid types.
- Run generate for PO ticket: `./bin/agentic-lens generate --ticket-id PO-5 --allow-empty --json-output test.json` -> verify `ticket_type` in `test.json` is `"docs"`.
- Run generate for GEN ticket: `./bin/agentic-lens generate --ticket-id GEN-42 --allow-empty --json-output test.json` -> verify `ticket_type` in `test.json` is `"chore"`.
- Run generate for unmatched ticket without flag: `./bin/agentic-lens generate --ticket-id UNKNOWN-99 --allow-empty --json-output test.json` -> verify stderr warning and `ticket_type: null` in `test.json`.
- Run generate with strict type on unmatched ticket: `./bin/agentic-lens generate --ticket-id UNKNOWN-99 --strict-type --allow-empty` -> verify status 1 and stderr error.
