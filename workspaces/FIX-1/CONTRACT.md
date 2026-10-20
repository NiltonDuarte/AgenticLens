# Contract: FIX-1 Enforce Non-Null Ticket Title and Ticket ID, Issue Warnings for Missing Ticket Type and Empty Invocations, and Register FIX Prefix

## 1. Problem Statement

The AgenticLens command-line interface generate subcommand currently allows metric ledgers to be generated with null or missing metadata, resulting in data degradation and silent operational anomalies:

1. Null Ticket Title Generation:
When the generate command runs without the ticket-title CLI flag and without an existing on-disk ledger containing a title, activeTicketTitle resolves to nil. This results in ticket_title: null being serialized into agentic_metrics.json. Ticket titles represent a primary human-readable identifier across dashboard views, comparative tables, and markdown summaries. The CLI must strictly fail with an explicit error and non-zero exit code if a title does not already exist in the target ledger and is not provided via the command line. The ticket_title field in newly generated or updated ledgers must never be null.

2. Null Ticket ID Vulnerability:
When ticket-id is omitted and not present in cached ledger facts, the CLI falls back to the literal string UNKNOWN for console logs and allows ticket_id to be stored as null in fact ledgers when the allow-empty flag is enabled. Every metrics ledger must be strictly bound to a known ticket. The CLI must fail immediately if ticket-id is not provided and cannot be resolved from existing ledger facts. The ticket_id field must never be null.

3. Missing Ticket Type Visibility:
Ticket type classification is critical for conventional commit categorization (feat, fix, docs, chore, etc.) and SDLC metrics rollups. When ticket-type is omitted and cannot be inferred from configured prefixes or git branch names, normal non-strict execution silently proceeds with ticket_type: null or prints a generic warning. The contract specifies that ticket_type remains nullable (not a hard fatal failure unless strict-type is set), but the CLI must reliably emit a prominent warning to standard error requesting the calling agent or user to explicitly provide ticket-type.

4. Silent Empty Invocations Anomaly:
When the allow-empty escape hatch is used or when session extraction resolves zero invocations, the CLI generates metrics tables and ledger files with empty invocation arrays without alerting the operator. Invocations are the fundamental unit of LLM telemetry. Generating a ledger with zero invocations is abnormal and typically indicates an extraction failure, misconfigured nonce, or disconnected database. The CLI must emit an explicit warning to standard error alerting the user that an empty invocation list is not expected.

5. Unregistered FIX Ticket Prefix:
The root configuration file .agenticlens-config.yaml currently defines prefixes AL, GEN (type chore), and PO (type docs). The bug fix prefix FIX (e.g., FIX-1) is absent. As a result, tickets starting with FIX- cannot automatically infer their conventional commit classification as fix, requiring manual flag entry or defaulting to unclassified.

Ticket FIX-1 remediates these shortcomings across the Go backend CLI, boundary adapters, configuration file, and regression test suites.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All deliverables must strictly preserve the architectural boundaries documented in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Functional Core (backend/internal/domain/):
- Domain models and value objects remain immutable with value semantics.
- Pure helper functions in domain/config.go (such as InferTicketTypeFromPrefix and NormalizeTicketType) operate solely on immutable structs with zero disk or network I/O.
- Model definitions in domain/models.go retain existing pointer types (TicketID *string, TicketTitle *string, TicketType *string) to preserve backward compatibility with legacy historical ledgers and migration engines (schema versions v1, v2, v3).

2. Boundary Adapters (backend/internal/adapters/):
- Schema validation and persistence boundaries in internal/adapters/repository.go and internal/adapters/schemas.go enforce strict integrity checks before writing fact ledgers to disk.
- SaveFactLedger validates that mandatory fields (ticket_id and ticket_title) are non-nil and contain non-empty, trimmed strings. If violated, persistence returns a descriptive wrapped error.
- Config loader in internal/adapters/config/ reads and validates updated prefix mappings from .agenticlens-config.yaml.

3. CLI Application Layer (backend/cmd/agentic-lens/):
- runGenerate in backend/cmd/agentic-lens/main.go coordinates flag parsing, cache lookup, validation gates, console output, and process termination.
- Pre-execution validation gates in runGenerate evaluate activeTicketID and activeTicketTitle immediately after flag parsing and cached fact loading. If either is missing or whitespace-only, the CLI writes a structured error to standard error and terminates with exit code 1.
- Warning channels write diagnostic messages to os.Stderr when activeTicketType is unresolved (in non-strict mode) or when len(invocations) is zero.

### Strict Invariants
- Zero Code Blocks Mandate: All plans and specifications are documented exclusively in prose, architectural descriptions, and structured task lists.
- Zero Emojis Mandate: Zero emojis are permitted in code, docstrings, error messages, warning messages, console output, test names, or markdown documents.
- Value Semantics and Zero Any: Strong typing across all Go packages. Zero any or empty interface{} declarations.

---

## 3. Scope Boundaries

### In Scope
1. Configuration Update (.agenticlens-config.yaml):
- Add the FIX prefix under ticket.prefixes with description Bug fixes, defect remediation, and telemetry validation repairs, and conventional commit ticket_type fix.

2. CLI Validation Gates and Warnings (backend/cmd/agentic-lens/main.go):
- Enforce mandatory activeTicketID: fail with exit code 1 if missing from both CLI flags and cached facts.
- Enforce mandatory activeTicketTitle: fail with exit code 1 if missing from both CLI flags and cached facts, ensuring ticket_title is never null in newly written ledgers.
- Emit explicit warning to os.Stderr when activeTicketType is nil (in non-strict mode) requesting the calling agent to provide --ticket-type.
- Emit explicit warning to os.Stderr when len(invocations) is zero alerting the user that an empty invocation list is not expected.

3. Boundary Adapter Defensive Validation (backend/internal/adapters/repository.go):
- Add defensive validation in SaveFactLedger ensuring fact ledgers cannot be serialized to disk with nil or whitespace-only TicketID or TicketTitle.

4. Test Suite Alignment and Regression Tests (backend/cmd/agentic-lens/main_test.go):
- Add explicit integration test cases verifying:
  - Exit code 1 when --ticket-title is omitted on a new ticket.
  - Success when --ticket-title is omitted but present in an existing fact ledger.
  - Exit code 1 when --ticket-id is omitted on a new ticket.
  - Warning emission to standard error when ticket-type cannot be resolved.
  - Warning emission to standard error when invocations list is empty under --allow-empty.
  - Automatic ticket type inference of fix for ticket FIX-1 via configured prefix.
- Update existing integration tests in main_test.go that invoke generate on isolated empty databases to supply --ticket-title so they do not fail the new mandatory title validation.

5. Domain and Adapter Unit Tests:
- Verify prefix inference for FIX in backend/internal/domain/config_test.go and backend/internal/adapters/config/loader_test.go.
- Verify boundary rejection of nil title and nil ID in backend/internal/adapters/repository_test.go.

### Out of Scope
1. Modifying historical schema migrations (backend/internal/domain/migration/) or changing schema version numbers (remains schema version 3).
2. Modifying TypeScript frontend models (web/src/domain/types.ts already handles nullable or string ticket_title gracefully).
3. Modifying OpenCode SQLite extraction logic or database tables.

---

## 4. Verification and Acceptance Criteria

1. Mandatory Title Gate:
Executing the generate subcommand without --ticket-title against a ticket with no existing agentic_metrics.json file must terminate with exit code 1 and output an error message containing Error: ticket title is required.

2. Existing Title Preservation:
Executing generate without --ticket-title against a workspace that already contains an agentic_metrics.json with a valid ticket_title must successfully preserve the existing title and succeed with exit code 0.

3. Mandatory ID Gate:
Executing generate without --ticket-id when no existing ledger is loaded must terminate with exit code 1 and output an error message containing Error: ticket identifier is required.

4. Nullable Ticket Type with Warning:
Executing generate for an unclassified ticket ID without --ticket-type and without --strict-type must exit with code 0 while writing a warning to standard error containing WARNING: Missing ticket type for ticket.

5. Empty Invocations Warning:
Executing generate with --allow-empty where zero invocations are resolved must exit with code 0 while writing a warning to standard error containing WARNING: No invocations found for ticket.

6. FIX Prefix Resolution:
Executing generate --ticket-id FIX-1 (with a valid --ticket-title) must automatically classify the ledger ticket_type as fix without requiring --ticket-type on the command line.

7. Automated Test Suite Health:
Running make test and make check must pass with zero failures, zero race conditions, and zero linter warnings.
