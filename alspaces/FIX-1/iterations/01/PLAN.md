# Implementation Plan: FIX-1 Enforce Non-Null Ticket Title and Ticket ID, Issue Warnings for Missing Ticket Type and Empty Invocations, and Register FIX Prefix

## 1. Goal and Iteration Scope

This iteration delivers the complete implementation of FIX-1. It adds the FIX ticket prefix to the repository configuration, adds strict CLI validation gates for ticket title and ticket ID, implements diagnostic warnings for missing ticket types and empty invocations, enforces boundary validation during ledger persistence, and updates test suites to verify compliance and prevent regressions.

---

## 2. Grounded Reference Architecture and Concrete File Paths

### Configuration Layer
- .agenticlens-config.yaml: Add prefix entry FIX with description and ticket_type: fix under ticket.prefixes.

### CLI Application Layer
- backend/cmd/agentic-lens/main.go:
  - Inside runGenerate, validate activeTicketID immediately after flag parsing and cached fact extraction. If empty, print error to os.Stderr and exit with code 1.
  - Inside runGenerate, validate activeTicketTitle immediately after flag parsing and cached fact extraction. If nil or empty after trimming, print error to os.Stderr and exit with code 1.
  - Review activeTicketType resolution. Ensure non-strict fallback prints a warning to os.Stderr requesting the agent to explicitly provide --ticket-type.
  - After database/cache fact extraction, check if len(invocations) is zero. If zero, print a warning to os.Stderr alerting the user that an empty invocation list is not expected.

### Boundary Adapters Layer
- backend/internal/adapters/repository.go:
  - In SaveFactLedger, validate that ledger.TicketID is non-nil and not whitespace, and ledger.TicketTitle is non-nil and not whitespace. Return a formatted error if validation fails.
- backend/internal/adapters/repository_test.go:
  - Add unit tests verifying SaveFactLedger returns an error when TicketID or TicketTitle is nil or empty.

### Configuration and Domain Tests
- backend/internal/adapters/config/loader_test.go:
  - Add test asserting that .agenticlens-config.yaml parses the FIX prefix with ticket_type fix.
- backend/internal/domain/config_test.go:
  - Add unit test verifying domain.InferTicketTypeFromPrefix accurately resolves FIX-1 to conventional type fix.

### CLI Integration Tests
- backend/cmd/agentic-lens/main_test.go:
  - Add integration tests for runGenerate validating mandatory ticket title, mandatory ticket ID, cached title fallback, ticket type warning, empty invocations warning, and FIX-1 prefix inference.
  - Update existing CLI tests in main_test.go that invoke generate without an existing ledger to pass --ticket-title "Test Fixture Ticket" to satisfy the new validation gate.

---

## 3. Phased Implementation Tasks

### Phase 1: Configuration Update for FIX Prefix
- Task 1.1: Edit .agenticlens-config.yaml:
  - Under ticket.prefixes, append a new mapping object:
    - prefix: FIX
    - description: Bug fixes, defect remediation, and telemetry validation repairs
    - ticket_type: fix
- Task 1.2: Validate Configuration Parsing:
  - Verify that adaptersconfig.Load successfully loads the new prefix.
  - Add an assertion in backend/internal/adapters/config/loader_test.go verifying the FIX prefix entry and ticket_type mapping.
  - Add an assertion in backend/internal/domain/config_test.go verifying that InferTicketTypeFromPrefix maps FIX-101 to ConventionalTicketTypeFix.

### Phase 2: CLI Validation Gates and Warnings in main.go
- Task 2.1: Enforce Mandatory Ticket Identifier:
  - In runGenerate in backend/cmd/agentic-lens/main.go:
  - Following the resolution of activeTicketID (from ticketIDFlag and cachedFacts.TicketID):
  - Check if activeTicketID is empty.
  - If empty: write to os.Stderr:
    Error: missing ticket ID. Specify --ticket-id explicitly.
  - Call os.Exit(1).
- Task 2.2: Enforce Mandatory Ticket Title:
  - In runGenerate in backend/cmd/agentic-lens/main.go:
  - Following the resolution of activeTicketTitle (from ticketTitleFlag and cachedFacts.TicketTitle):
  - Check if activeTicketTitle is nil or strings.TrimSpace(*activeTicketTitle) is empty.
  - If nil or empty: write to os.Stderr:
    Error: missing ticket title for ticket '<activeTicketID>'. Specify --ticket-title explicitly.
  - Call os.Exit(1).
  - Clean activeTicketTitle by trimming whitespace, ensuring the pointer passed into domain.AgenticFactLedger references the trimmed non-empty string.
- Task 2.3: Ensure Ticket Type Warning in Non-Strict Mode:
  - In runGenerate in backend/cmd/agentic-lens/main.go:
  - Following activeTicketType inference (from ticketTypeFlag, cachedFacts.TicketType, and git.InferTicketType):
  - If activeTicketType is nil and strictTypeFlag is false:
  - Write to os.Stderr:
    WARNING: Missing ticket type for ticket <activeTicketID>. Specify --ticket-type explicitly.
  - Confirm ticket_type in factLedger remains nil without terminating the process.
- Task 2.4: Implement Empty Invocations Warning:
  - In runGenerate in backend/cmd/agentic-lens/main.go:
  - After database extraction and cached facts merging (following line 192):
  - Check if len(invocations) is zero.
  - If zero: write to os.Stderr:
    WARNING: Zero invocations recorded for ticket <activeTicketID>. This is not expected.
  - Allow execution to proceed when allowEmptyFlag is true or when knownThreadIDs exist.

### Phase 3: Defensive Boundary Validation in Adapters
- Task 3.1: Add Pre-Write Validation in SaveFactLedger:
  - In backend/internal/adapters/repository.go, update SaveFactLedger:
  - Check if ledger.TicketID is nil or strings.TrimSpace(*ledger.TicketID) is empty. If so, return an error wrapping a descriptive message indicating ticket_id cannot be null or empty.
  - Check if ledger.TicketTitle is nil or strings.TrimSpace(*ledger.TicketTitle) is empty. If so, return an error wrapping a descriptive message indicating ticket_title cannot be null or empty.
- Task 3.2: Unit Test Adapter Validation:
  - In backend/internal/adapters/repository_test.go:
  - Add TestSaveFactLedger_MandatoryFieldsValidation:
    - Attempt to save a ledger with nil TicketID: verify error is returned and file is not written.
    - Attempt to save a ledger with empty string TicketID: verify error is returned.
    - Attempt to save a ledger with nil TicketTitle: verify error is returned.
    - Attempt to save a ledger with whitespace-only TicketTitle: verify error is returned.
    - Attempt to save a ledger with valid TicketID and TicketTitle: verify file is written successfully.

### Phase 4: CLI Integration Tests and Existing Test Suite Maintenance
- Task 4.1: Add Integration Tests for New Validations in main_test.go:
  - Add TestCLIGenerateMandatoryTitleAndID with subtests:
    - Subtest 1: generate fails with exit code 1 and error message on stderr when --ticket-title is omitted and no cached ledger exists.
    - Subtest 2: generate succeeds when --ticket-title is omitted but the destination json-output points to an existing ledger containing a non-empty ticket_title.
    - Subtest 3: generate fails with exit code 1 and error message on stderr when --ticket-id is omitted and no cached ledger exists.
    - Subtest 4: generate succeeds and logs a warning on stderr when ticket-type cannot be resolved in non-strict mode.
    - Subtest 5: generate with --allow-empty logs a warning on stderr when invocations list is empty.
    - Subtest 6: generate with --ticket-id FIX-1 resolves ticket_type to fix in generated JSON without --ticket-type flag.
- Task 4.2: Update Existing Integration Tests in main_test.go:
  - Review all existing test functions in backend/cmd/agentic-lens/main_test.go that execute the generate CLI command against temporary databases or directories.
  - For tests where generate was previously called without --ticket-title and without cached ledger facts, add --ticket-title "Test Fixture Ticket".
  - Verify that tests asserting failure on missing root sessions or invalid ticket types continue to test their specific failure conditions without being obscured by title validation.

### Phase 5: Verification, Quality Gates, and Compliance Audit
- Task 5.1: Execute Backend Test Suites:
  - Run go test -race ./backend/... and verify all unit and integration tests pass cleanly.
- Task 5.2: Execute Static Analysis:
  - Run go vet ./... across the Go backend.
  - Run make check from repository root.
- Task 5.3: End-to-End Build and Verification:
  - Run make build from repository root.
  - Verify the compiled agentic-lens binary executes generate --help successfully.
- Task 5.4: Compliance Audits:
  - Verify zero code blocks in CONTRACT.md and PLAN.md.
  - Verify zero emojis across all code files, comments, error strings, warning messages, and markdown files.
  - Verify zero any or empty interface{} declarations across Go packages.

---

## 4. Execution Dependencies and Delivery Sequence

1. Configuration First: Update .agenticlens-config.yaml with prefix FIX and verify config loader and domain prefix tests.
2. Adapter Defensive Shell: Implement validation in repository.go and add adapter unit tests.
3. CLI Imperative Shell: Implement title/ID validation gates, type warning, and empty invocations warning in main.go.
4. Test Suite Alignment: Update existing main_test.go generate invocations to include --ticket-title and add new comprehensive integration tests.
5. Final Verification: Run make check and make test to verify zero regressions across all layers.
