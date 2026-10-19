# Implementation Plan: AL-29 Iteration 02 - Project-Level Configurable Ticket Type Resolution

## 1. Goal and Iteration Scope

Iteration 02 addresses user steering feedback to decouple ticket type classification from hardcoded engine rules and fragile workspace directory heuristics.

In Iteration 01, ticket prefixes PO-* and GEN-* were hardcoded in the Go functional core (`backend/internal/domain/enums.go`) and boundary git adapter (`backend/internal/adapters/git/inference.go`). PO and GEN are repository-specific conventions defined at the project level in `.agenticlens/config.yaml`, not immutable invariants of the AgenticLens domain engine.

Iteration 02 refactors the resolution architecture as follows:
- The primary resolution method is the explicit command-line flag (`--ticket-type`), followed by cached facts from prior runs.
- The secondary resolution method is the git branch prefix (`<type>/<ticket>`).
- The tertiary backup method is project-level configuration: prefix mappings configured under `ticket.prefixes` in `.agenticlens/config.yaml` supporting an optional `ticket_type` field.
- Hardcoded PO-* and GEN-* logic in Go source code is eliminated. The Go domain core becomes 100% parameter-driven, operating purely on caller-supplied prefix configuration models.
- The project configuration `.agenticlens/config.yaml` is updated to assign `ticket_type: docs` to prefix PO and `ticket_type: chore` to prefix GEN.
- Core CLI generate logic loads project configuration early, evaluates the 4-tier resolution chain deterministically, and preserves strict typing and warning boundaries.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation spans the following Go backend files, adapters, and configuration assets:

### Pure Functional Core Layer (`backend/internal/domain/`)
- `backend/internal/domain/config.go`:
  - Update struct `TicketPrefix` to include optional field `TicketType string` with JSON and YAML tags `json:"ticket_type,omitempty" yaml:"ticket_type,omitempty"`.
  - Add pure prefix resolution function `InferTicketTypeFromPrefix(ticketID string, prefixes []TicketPrefix) (ConventionalTicketType, bool)`.
  - Add helper method on `TicketConfig`: `(t TicketConfig) InferTicketType(ticketID string) (ConventionalTicketType, bool)` delegating to `InferTicketTypeFromPrefix`.
  - Update `DefaultProjectConfig()` to provide baseline struct defaults with `ticket_type: "chore"` for GEN and `ticket_type: "docs"` for PO.
- `backend/internal/domain/enums.go`:
  - Remove hardcoded PO-* and GEN-* string checks. Forward calls or replace `InferTicketTypeFromPrefix` to take the configured prefix slice.
  - Maintain canonical enum definitions, `CanonicalTicketTypes`, `NormalizeTicketType`, and `IsValidTicketType`.
- `backend/internal/domain/config_test.go` and `backend/internal/domain/enums_test.go`:
  - Unit tests verifying pure, parameter-driven prefix matching with various custom prefix configurations, casing rules, delimiter boundaries, and non-empty suffix checks.

### Imperative Boundary Adapters Layer (`backend/internal/adapters/`)
- `backend/internal/adapters/config/loader.go`:
  - Extend `rawPrefixItem` struct with field `TicketType string` with JSON and YAML tags.
  - Extend `UnmarshalYAML` on `rawPrefixItem` to decode `ticket_type` from YAML mapping nodes into `r.TicketType`. Ensure scalar nodes leave `TicketType` empty.
  - Extend `UnmarshalJSON` on `rawPrefixItem` to decode `ticket_type` from JSON objects into `r.TicketType`. Ensure scalar strings leave `TicketType` empty.
  - Update `mergeConfig` to propagate `TicketType: strings.TrimSpace(item.TicketType)` into each `domain.TicketPrefix`.
- `backend/internal/adapters/config/loader_test.go`:
  - Unit tests verifying YAML and JSON deserialization of structured prefix entries containing `ticket_type`.
- `backend/internal/adapters/git/inference.go`:
  - Refactor `InferTicketType(ticketID string, prefixes ...domain.TicketPrefix) *string`:
    - Remove hardcoded domain prefix calls.
    - Inspect git branches first via `InferTicketTypeFromBranchList`.
    - If branch inspect does not resolve, and configured `prefixes` are passed, delegate to `domain.InferTicketTypeFromPrefix(cleanTicket, prefixes)`.
    - Return nil if still unclassified.
- `backend/internal/adapters/git/inference_test.go`:
  - Update tests to assert that `InferTicketType` without prefixes returns nil on non-branch tickets (even for PO-* and GEN-*).
  - Verify that passing configured prefixes resolves `PO-5` to `docs`, `GEN-42` to `chore`, and custom prefixes to their configured types.

### Repository Configuration (`.agenticlens/config.yaml`)
- `.agenticlens/config.yaml`:
  - Add `ticket_type: chore` to prefix `GEN`.
  - Add `ticket_type: docs` to prefix `PO`.
  - Leave prefix `AL` without a default `ticket_type`.

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- `backend/cmd/agentic-lens/main.go`:
  - In `runGenerate`:
    - Load project configuration via `adaptersconfig.Load("")` early, before ticket type resolution.
    - Implement deterministic 4-tier resolution hierarchy:
      - Tier 1: Evaluate CLI flag `--ticket-type` via `domain.NormalizeTicketType`. If invalid, write error to stderr and exit 1. If omitted, evaluate cached facts ticket type.
      - Tier 2: Evaluate git branch prefix via `git.InferTicketType(activeTicketID)`.
      - Tier 3: Evaluate configured project prefix mapping via `domain.InferTicketTypeFromPrefix(activeTicketID, projCfg.Ticket.Prefixes)`.
      - Tier 4: If unresolved, check `--strict-type` (error and exit 1) or emit warning to stderr and keep `activeTicketType = nil`.
- `backend/cmd/agentic-lens/main_test.go`:
  - Integration tests verifying:
    - CLI flag `--ticket-type` overrides branch and project config.
    - Git branch prefix overrides project config prefix mapping.
    - Configured project prefix backup resolves `PO-5` to `docs` and `GEN-1` to `chore` when flags and branches are absent.
    - Unconfigured prefixes emit stderr warning and persist `null` in JSON ledger, or fail when `--strict-type` is supplied.

---

## 3. Phased Implementation Tasks

### Phase 1: Pure Domain Model and Configurable Prefix Resolution
- Task 1.1: Extend `TicketPrefix` struct in `backend/internal/domain/config.go`:
  - Add field `TicketType string` with struct tags `json:"ticket_type,omitempty"` and `yaml:"ticket_type,omitempty"`.
- Task 1.2: Implement Pure Configurable Prefix Matcher in `backend/internal/domain/config.go`:
  - Define pure function `InferTicketTypeFromPrefix(ticketID string, prefixes []TicketPrefix) (ConventionalTicketType, bool)`.
  - Matching specification:
    - Trim whitespace from `ticketID`. If empty, return empty string and false.
    - Iterate through configured `prefixes` slice.
    - For each item, clean `prefix.TicketType`. If empty, skip to next prefix.
    - Clean `prefix.Prefix`, stripping any trailing delimiter characters (hyphen or underscore).
    - If cleaned prefix token is empty, skip.
    - Check case-insensitive prefix match against `ticketID` followed by delimiter boundary (hyphen or underscore).
    - Check that remaining characters following the delimiter are non-empty and not pure whitespace.
    - Normalize configured `TicketType` via `domain.NormalizeTicketType`.
    - If normalized successfully, return canonical `ConventionalTicketType` and true.
    - If no configured prefix matches, return empty string and false.
- Task 1.3: Add Helper Method on `TicketConfig`:
  - Define `(t TicketConfig) InferTicketType(ticketID string) (ConventionalTicketType, bool)` delegating to `InferTicketTypeFromPrefix(ticketID, t.Prefixes)`.
- Task 1.4: Refactor `backend/internal/domain/enums.go`:
  - Remove hardcoded string literals "PO-" and "GEN-" from `InferTicketTypeFromPrefix`.
  - Update `InferTicketTypeFromPrefix` signature to accept `(ticketID string, prefixes []TicketPrefix)` or relocate definition entirely to `config.go`.
- Task 1.5: Update `DefaultProjectConfig()` in `backend/internal/domain/config.go`:
  - In `DefaultProjectConfig()`, populate `TicketType: "chore"` for the `GEN` prefix entry, `TicketType: "docs"` for the `PO` prefix entry, and empty `TicketType` for the `AL` prefix entry.
- Task 1.6: Unit Tests in `backend/internal/domain/config_test.go` and `backend/internal/domain/enums_test.go`:
  - Test pure prefix resolution with custom prefix configurations (e.g., prefix `DOC` -> `docs`, `TASK` -> `chore`, `BUG` -> `fix`).
  - Test delimiter protection: ensure `POD-1` does not match prefix `PO`, and `GENTLE-1` does not match prefix `GEN`.
  - Test empty suffix rejection: ensure `PO-` and `GEN-` without numbers or identifiers return false.
  - Test case insensitivity: ensure `po-100`, `PO-100`, `Po-100` resolve consistently.
  - Test prefixes without configured `ticket_type` (like `AL`) return false.
  - Test invalid configured `ticket_type` values return false.

### Phase 2: Configuration Adapter Schema and Loader Extension
- Task 2.1: Extend `rawPrefixItem` in `backend/internal/adapters/config/loader.go`:
  - Add field `TicketType string` with struct tags `json:"ticket_type" yaml:"ticket_type"`.
- Task 2.2: Update YAML Unmarshaling in `backend/internal/adapters/config/loader.go`:
  - In `rawPrefixItem.UnmarshalYAML`:
    - When node is `yaml.ScalarNode`, assign `r.TicketType = ""` along with existing fields.
    - When node is `yaml.MappingNode`, add `TicketType string yaml:"ticket_type"` to the temporary decoding struct.
    - Assign `r.TicketType = strings.TrimSpace(item.TicketType)`.
- Task 2.3: Update JSON Unmarshaling in `backend/internal/adapters/config/loader.go`:
  - In `rawPrefixItem.UnmarshalJSON`:
    - When data is JSON string, assign `r.TicketType = ""`.
    - When data is JSON object, add `TicketType string json:"ticket_type"` to the temporary decoding struct.
    - Assign `r.TicketType = strings.TrimSpace(item.TicketType)`.
- Task 2.4: Update Configuration Merging in `backend/internal/adapters/config/loader.go`:
  - In `mergeConfig`, when iterating `dto.Ticket.Prefixes`, map `TicketType: strings.TrimSpace(item.TicketType)` to the constructed `domain.TicketPrefix` instance.
- Task 2.5: Loader Unit Tests in `backend/internal/adapters/config/loader_test.go`:
  - Test loading YAML configuration where prefixes define `ticket_type` mappings.
  - Test loading JSON configuration where prefixes define `ticket_type` mappings.
  - Test scalar string prefixes leave `TicketType` as empty string.
  - Test structured prefixes without `ticket_type` leave `TicketType` as empty string.

### Phase 3: Project Configuration File Update
- Task 3.1: Update `.agenticlens/config.yaml`:
  - Under `ticket.prefixes`, update the `GEN` item to include `ticket_type: chore`.
  - Under `ticket.prefixes`, update the `PO` item to include `ticket_type: docs`.
  - Keep `AL` prefix description intact with no `ticket_type` property.

### Phase 4: Git Adapter Decoupling from Hardcoded Domain Constants
- Task 4.1: Refactor `InferTicketType` in `backend/internal/adapters/git/inference.go`:
  - Update signature: `func InferTicketType(ticketID string, prefixes ...domain.TicketPrefix) *string`.
  - Implementation details:
    - Clean input `ticketID`. If empty, return nil.
    - Tier 2 (Branch inspection): Query git branches via `runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*")` and invoke `InferTicketTypeFromBranchList`. If valid conventional commit type returned, return it immediately.
    - Tier 3 (Project-configured prefix backup): If branch inspection returns nil and `len(prefixes) > 0`, invoke `domain.InferTicketTypeFromPrefix(cleanTicket, prefixes)`. If match found, return pointer to string.
    - Delete any hardcoded calls to PO or GEN.
    - Return nil if unresolved.
- Task 4.2: Update Git Inference Tests in `backend/internal/adapters/git/inference_test.go`:
  - Test calling `InferTicketType("PO-5")` without prefixes returns nil when no branch exists for `PO-5`.
  - Test calling `InferTicketType("PO-5", configuredPrefixes...)` returns "docs".
  - Test calling `InferTicketType("GEN-42", configuredPrefixes...)` returns "chore".
  - Test calling `InferTicketType("CUSTOM-1", customPrefixes...)` returns expected mapped type.
  - Maintain existing branch resolution tests in `TestInferTicketTypeFromBranchList_Adversarial`.

### Phase 5: CLI Early Config Ingestion and Resolution Hierarchy
- Task 5.1: Early Configuration Ingestion in `backend/cmd/agentic-lens/main.go`:
  - In `runGenerate`, move `projCfg, _ := adaptersconfig.Load("")` upwards so it executes immediately after parsing CLI flags and before evaluating ticket identifiers and types.
- Task 5.2: Implement Deterministic 4-Tier Resolution Hierarchy:
  - Tier 1 (Explicit CLI input or cached facts):
    - If `--ticket-type` flag is non-empty, normalize using `domain.NormalizeTicketType`. If invalid, write error to stderr with supported canonical types and terminate with `os.Exit(1)`.
    - If flag is empty, inspect cached facts `cachedFacts.TicketType`. If present and valid, assign normalized string pointer.
  - Tier 2 (Git branch inspection):
    - If `activeTicketType` is nil and `activeTicketID` is non-empty, call `git.InferTicketType(activeTicketID)`.
  - Tier 3 (Project configuration prefix backup):
    - If `activeTicketType` is nil and `activeTicketID` is non-empty, call `domain.InferTicketTypeFromPrefix(activeTicketID, projCfg.Ticket.Prefixes)`. If valid, assign normalized string pointer.
  - Tier 4 (Termination and warning):
    - If `activeTicketType` remains nil:
      - If `--strict-type` is true, write error to stderr: `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` and terminate with `os.Exit(1)`.
      - If `--strict-type` is false, write warning to stderr: `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.` and retain `activeTicketType = nil`.
- Task 5.3: Fact Ledger Assignment and Reuse:
  - Assign `factLedger.TicketType = activeTicketType`.
  - Reuse the already-loaded `projCfg` later in `runGenerate` when deriving metrics ledger rollups and role mappings, avoiding redundant config loads.

### Phase 6: End-to-End Test Suite Updates and Verification
- Task 6.1: Update CLI Integration Tests in `backend/cmd/agentic-lens/main_test.go`:
  - Verify that `--ticket-type` CLI flag takes absolute precedence over branch names and configuration files.
  - Verify that git branch prefix takes precedence over project configuration prefix mappings.
  - Verify that `agentic-lens generate --ticket-id PO-5 --allow-empty` resolves to `docs` using `.agenticlens/config.yaml`.
  - Verify that `agentic-lens generate --ticket-id GEN-1 --allow-empty` resolves to `chore` using `.agenticlens/config.yaml`.
  - Verify that `agentic-lens generate --ticket-id AL-29 --allow-empty` (which has no configured `ticket_type`) emits warning to stderr and produces `ticket_type: null` when not on a conventional branch.
  - Verify that `agentic-lens generate --ticket-id AL-29 --strict-type --allow-empty` exits with code 1 and logs error to stderr.
- Task 6.2: Quality Gate Execution:
  - Run `go vet ./...` in `backend`.
  - Run `go test -v -race ./...` across all packages in `backend`.
  - Run `make check`.
  - Verify zero `any` or empty `interface{}` across Go packages.
  - Verify zero emojis in code, docstrings, error messages, and tests.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Model and Configurable Matcher) creates the pure data-driven foundation with zero external dependencies.
2. Phase 2 (Config Loader Extension) updates adapter serialization so configuration files can declare `ticket_type` per prefix.
3. Phase 3 (Project Config Update) configures PO and GEN in `.agenticlens/config.yaml`.
4. Phase 4 (Git Adapter Decoupling) cleans `git/inference.go` of hardcoded domain constants and accepts caller-supplied prefix mappings as backup.
5. Phase 5 (CLI Resolution Hierarchy) connects early configuration loading and enforces the 4-tier resolution chain in `main.go`.
6. Phase 6 (Integration Tests & Quality Gates) verifies end-to-end correctness, regression safety, and standard compliance.

---

## 5. Verification Plan

### Test Commands
- Domain unit tests:
  `cd backend && go test -v -race ./internal/domain/...`
- Config loader tests:
  `cd backend && go test -v -race ./internal/adapters/config/...`
- Git adapter tests:
  `cd backend && go test -v -race ./internal/adapters/git/...`
- CLI end-to-end integration tests:
  `cd backend && go test -v -race -run TestCLITicketType ./cmd/agentic-lens/...`
- Full backend suite:
  `make test-backend`
- Static analysis and formatting check:
  `make check`

### Verification Scenarios
- Scenario 1: Run CLI with explicit flag on PO ticket:
  `./bin/agentic-lens generate --ticket-id PO-5 --ticket-type fix --allow-empty --json-output test.json`
  Assert `ticket_type` in `test.json` is "fix", proving CLI flag overrides project config backup.
- Scenario 2: Run CLI without flag on PO ticket:
  `./bin/agentic-lens generate --ticket-id PO-5 --allow-empty --json-output test.json`
  Assert `ticket_type` in `test.json` is "docs", proving project config backup works when flag and branch are omitted.
- Scenario 3: Run CLI without flag on GEN ticket:
  `./bin/agentic-lens generate --ticket-id GEN-42 --allow-empty --json-output test.json`
  Assert `ticket_type` in `test.json` is "chore", proving project config backup works for GEN.
- Scenario 4: Run CLI without flag on unconfigured ticket prefix (e.g. AL-999) without branch:
  `./bin/agentic-lens generate --ticket-id AL-999 --allow-empty --json-output test.json`
  Assert stderr emits `WARNING: Missing ticket type for ticket AL-999. Specify --ticket-type explicitly.` and `ticket_type` in `test.json` is `null`.
- Scenario 5: Run CLI with `--strict-type` on unconfigured ticket prefix without branch:
  `./bin/agentic-lens generate --ticket-id AL-999 --strict-type --allow-empty`
  Assert exit code is 1 and stderr emits `Error: missing ticket type for ticket 'AL-999'. Specify --ticket-type explicitly.`.
