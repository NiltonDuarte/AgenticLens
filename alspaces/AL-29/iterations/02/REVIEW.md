# Review Report: AL-29 Iteration 02

- Ticket: AL-29 (Conventional Commit Ticket Type Classification in Telemetry Engine)
- Iteration: 02 (Project-Level Configurable Ticket Type Resolution)
- Scope: `backend/internal/domain/`, `backend/internal/adapters/`, `backend/cmd/agentic-lens/`, `.agenticlens/config.yaml`
- Reviewer: Code Review Subagent
- Date: 2026-10-19
- Overall Verdict: APPROVED

---

## Axis 1: Spec & Contract Alignment

### 1. User Steering Mandates & Decoupling
- **Zero Hardcoded PO/GEN in Go Domain Core**:
  - `backend/internal/domain/config.go`: `InferTicketTypeFromPrefix(ticketID string, prefixes []TicketPrefix)` is purely parameter-driven. It operates on caller-supplied prefix configurations without embedded "PO" or "GEN" literals.
  - `backend/internal/domain/enums.go`: Contains canonical enums, validation predicates, and normalization maps. Hardcoded prefix inspection was purged.
  - Baseline domain defaults in `DefaultProjectConfig()` map `GEN` to `chore` and `PO` to `docs` only as project configuration seed defaults.
- **Purely Data-Driven Prefix Inference**:
  - `InferTicketTypeFromPrefix` defends against false prefix collisions (`POD-1` does not match `PO`, `GENTLE-1` does not match `GEN`). Delimiter checks enforce `-` or `_` followed by non-empty identifiers.
  - `TicketConfig.InferTicketType(ticketID)` provides a clean receiver method delegating directly to `InferTicketTypeFromPrefix`.
- **Project Configuration Prefix Backup**:
  - `.agenticlens/config.yaml` configures `ticket_type: chore` for `GEN` and `ticket_type: docs` for `PO`.
  - Prefix `AL` intentionally leaves `ticket_type` blank.
- **CLI Flag as Primary Resolution Method**:
  - CLI `--ticket-type` input takes absolute precedence. It evaluates first and normalizes aliases (`feature` to `feat`) before inspecting branches or project configuration.

### 2. 4-Tier Resolution Precedence Hierarchy
The resolution sequence implemented in `backend/cmd/agentic-lens/main.go` strictly enforces the four tiers:
1. **Tier 1 (Explicit CLI / Cache)**: Evaluates `--ticket-type` via `domain.NormalizeTicketType`. Invalid types fail immediately with exit status 1 and list all canonical types. Falls back to cached ledger facts if the flag is omitted.
2. **Tier 2 (Git Branch Prefix)**: Evaluates git branch names via `git.InferTicketType(activeTicketID)`. Strips remote prefixes (`remotes/`, `origin/`) and normalizes category segments (`feat/*`, `fix/*`).
3. **Tier 3 (Project-Configured Prefix Mapping)**: Evaluates `projCfg.Ticket.Prefixes` via `git.InferTicketType(activeTicketID, projCfg.Ticket.Prefixes...)`. Acts strictly as a backup when no branch matches.
4. **Tier 4 (Fallback & Gating)**: If unresolved:
   - With `--strict-type`: writes `Error: missing ticket type for ticket '<TICKET_ID>'. Specify --ticket-type explicitly.` to stderr and terminates with `os.Exit(1)`.
   - Without `--strict-type`: writes `WARNING: Missing ticket type for ticket <TICKET_ID>. Specify --ticket-type explicitly.` to stderr and retains `activeTicketType = nil`, persisting `"ticket_type": null` in JSON.

### 3. Scope Boundaries & Zero Scope Creep
- No frontend files in `web/` were modified.
- No schema version bumps or backfill modifications were introduced.
- Early configuration loading via `adaptersconfig.Load("")` in `main.go` is reused downstream for metrics derivation (`DeriveMetricsLedger`), eliminating redundant disk reads without collateral changes.

---

## Axis 2: Engineering Standards & Architecture

### 1. Functional Core, Imperative Shell (FCIS)
- **Domain Purity**:
  - `backend/internal/domain/config.go` and `enums.go` import only standard library string and formatting utilities (`fmt`, `regexp`, `sort`, `strings`).
  - Zero imports of `internal/adapters`, `cmd/`, `os`, `net/http`, or system clocks (`time.Now()`).
  - Functions are deterministic and side-effect free.
- **Value Semantics**:
  - `TicketPrefix`, `TicketConfig`, and `ProjectConfig` use immutable value structures. Receiver methods operate on value copies.
- **Boundary Isolation**:
  - Git subprocess execution remains encapsulated in `backend/internal/adapters/git/`.
  - Configuration parsing and YAML/JSON unmarshaling remain encapsulated in `backend/internal/adapters/config/`.

### 2. Typing Discipline & Boundary Serialization
- **Zero Any & Interface{}**:
  - Strict type safety verified across the Go codebase. Zero occurrences of `any` or `interface{}` used as types.
  - Serialization structures in `rawPrefixItem` define explicit scalar and mapping handlers for YAML and JSON decoding.
- **Descriptive Error Wrapping**:
  - All loader decoding failures wrap low-level errors using `%w` with context strings (e.g. `decoding ticket prefix mapping: %w`, `decoding yaml configuration %s: %w`).

### 3. Test Suite Quality & Resilience
- **Adversarial Boundary Coverage**:
  - `backend/internal/domain/config_test.go`: Tests delimiter protection, empty suffixes (`PO-`, `GEN-`), case insensitivity, trailing hyphens/underscores in configuration, and custom prefixes.
  - `backend/internal/adapters/config/loader_test.go`: Tests YAML and JSON unmarshaling of structured prefixes with `ticket_type`, scalar strings, empty values, and malformed structures.
  - `backend/internal/adapters/git/inference_test.go`: Tests branch vs config precedence, unbranched tickets returning nil without prefixes, and custom prefix mappings.
  - `backend/cmd/agentic-lens/main_test.go`: Tests Tier 1 overriding Tier 2, Tier 2 overriding Tier 3, project config backup for PO and GEN, stderr warning logging, and `--strict-type` exit code 1.
- **Zero Mocks in Domain**:
  - Pure domain unit tests instantiate concrete structs directly with zero mock frameworks or stubs.
- **Zero Silent Skips**:
  - Verified zero occurrences of `t.Skip` across all test files.

### 4. Zero Emojis
- Zero emojis found across all code, comments, docstrings, error messages, and tests.

---

## Verification Matrix

| Check | Requirement | Result |
|---|---|---|
| 1 | Domain decoupled: zero hardcoded PO/GEN in Go domain core | PASS |
| 2 | Purely data-driven prefix inference via `InferTicketTypeFromPrefix` | PASS |
| 3 | `.agenticlens/config.yaml` maps PO to `docs` and GEN to `chore` | PASS |
| 4 | Tier 1: CLI `--ticket-type` overrides branch and project config | PASS |
| 5 | Tier 2: Git branch prefix overrides project config mapping | PASS |
| 6 | Tier 3: Config prefix mapping acts as backup for unbranched tickets | PASS |
| 7 | Tier 4: Missing type without `--strict-type` logs stderr warning and outputs `null` | PASS |
| 8 | Tier 4: Missing type with `--strict-type` logs stderr error and exits 1 | PASS |
| 9 | Go FCIS: zero I/O imports or outer layer dependencies in `internal/domain/` | PASS |
| 10 | Zero `any` or empty `interface{}` in Go backend | PASS |
| 11 | Zero emojis in code, docstrings, warnings, and tests | PASS |
| 12 | Zero mocks in domain tests and zero silent test skips | PASS |
