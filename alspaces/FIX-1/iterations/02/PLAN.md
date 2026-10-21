# Delta Plan: FIX-1 Iteration 02 - Remove Redundant Root Config Test in Config Loader Test Suite

## 1. Goal and Iteration Scope

This delta iteration addresses user feedback concerning a redundant test case in the configuration adapter test suite.

The test subtest "Root config file parses FIX prefix with fix ticket_type and infers correctly" in backend/internal/adapters/config/loader_test.go under TestLoad_CustomExtensionAndEmptyPathEdgeCases hardcodes an assertion against the project root configuration file (.agenticlens-config.yaml) located at ../../.. .

This subtest introduces unnecessary coupling between unit tests and the live project root configuration file. Generic parsing and inference for structured prefix mappings (including the FIX prefix and conventional type fix) is already comprehensively verified using isolated directory fixtures in loader_test.go (lines 1050 to 1095). Furthermore, end-to-end integration inference for the root configuration file is verified in backend/cmd/agentic-lens/main_test.go (lines 4073 to 4104).

The goal of this iteration is to remove this redundant subtest from backend/internal/adapters/config/loader_test.go, eliminating test fragility and maintenance overhead while preserving full unit and integration test coverage.

---

## 2. Context and Architectural Rationale

### Existing Coverage Grounding
1. Isolated Fixture Unit Test:
   - File: backend/internal/adapters/config/loader_test.go
   - Function: TestLoad_TicketPrefixesStructuredFormats
   - Subtest: YAML structured prefix with ticket_type
   - Scope: Uses t.TempDir() and a synthetic config.yaml containing prefixes AL, GEN, PO, and FIX with whitespace-padded ticket_type values. Asserts that config.Load correctly parses the prefix list, normalizes whitespace, and populates the domain configuration struct.
2. CLI Integration Verification:
   - File: backend/cmd/agentic-lens/main_test.go
   - Function: TestCLIGenerateMandatoryTitleAndID
   - Subtest: generate with FIX-1 infers ticket_type fix
   - Scope: Runs runGenerate with ticket ID FIX-1 against an isolated temporary directory, verifying that the actual application environment resolves the ticket prefix from .agenticlens-config.yaml and outputs ticket_type fix.
3. Redundant Subtest:
   - File: backend/internal/adapters/config/loader_test.go
   - Function: TestLoad_CustomExtensionAndEmptyPathEdgeCases
   - Subtest: Root config file parses FIX prefix with fix ticket_type and infers correctly (lines 2069 to 2099)
   - Problem: Hardcodes relative directory traversal to the repository root directory. Unit tests in internal/adapters/config must validate loader parsing mechanisms against synthetic fixtures rather than binding against live repository configuration.

---

## 3. Grounded File Paths and Symbols

### Target Modification
- File: backend/internal/adapters/config/loader_test.go
- Enclosing Function: TestLoad_CustomExtensionAndEmptyPathEdgeCases
- Target Lines: 2069 to 2099
- Symbol to Remove: The subtest invocation t.Run with name "Root config file parses FIX prefix with fix ticket_type and infers correctly" and its entire closure body.

### Unaffected Core and Adapter Modules
- backend/internal/domain/config.go: Retains prefix lookup and inference logic.
- backend/internal/adapters/config/loader.go: Retains YAML and JSON loading functions.
- backend/cmd/agentic-lens/main.go: Retains generate command validations and inference.
- backend/cmd/agentic-lens/main_test.go: Retains end-to-end FIX-1 CLI inference verification.

---

## 4. Phased Implementation Tasks

### Phase 1: Test Suite Refactoring
- Task 1.1: Remove Redundant Subtest in loader_test.go:
  - In backend/internal/adapters/config/loader_test.go, locate TestLoad_CustomExtensionAndEmptyPathEdgeCases.
  - Delete lines 2069 through 2099 containing t.Run("Root config file parses FIX prefix with fix ticket_type and infers correctly", func(t *testing.T) { ... }).
  - Verify that TestLoad_CustomExtensionAndEmptyPathEdgeCases cleanly closes immediately following the previous subtest for invalid configuration extensions.
  - Check that no unused imports or syntax issues are introduced by the deletion.

### Phase 2: Verification and Quality Assurance
- Task 2.1: Run Config Adapter Unit Tests:
  - Execute go test -v ./backend/internal/adapters/config/...
  - Confirm all remaining tests in loader_test.go pass with zero failures.
- Task 2.2: Run Backend Full Test Suite:
  - Execute go test -race ./backend/...
  - Confirm domain, adapter, and CLI tests continue to pass with zero failures.
- Task 2.3: Static Analysis and Linter Checks:
  - Run go vet ./backend/...
  - Run make check from repository root.
- Task 2.4: Build Gate:
  - Run make build from repository root.
  - Confirm binary builds cleanly with zero errors.

---

## 5. Acceptance Criteria

1. Subtest Absence:
   - backend/internal/adapters/config/loader_test.go contains no subtest named "Root config file parses FIX prefix with fix ticket_type and infers correctly".
   - loader_test.go contains no direct file system loading targeting ../../.. .
2. Test Suite Stability:
   - All tests in backend/internal/adapters/config pass cleanly.
   - Integration test TestCLIGenerateMandatoryTitleAndID in backend/cmd/agentic-lens/main_test.go continues to verify FIX prefix inference in an end-to-end scenario.
3. Clean Tooling and Builds:
   - go test -race ./backend/... completes with zero failures.
   - make check passes with zero linter errors.
4. Compliance Invariants:
   - Zero code blocks across all markdown workspace deliverables.
   - Zero emojis in files, logs, comments, or documentation.

---

## 6. Execution Dependencies and Delivery Sequence

1. Coder Subagent: Remove the redundant subtest from backend/internal/adapters/config/loader_test.go.
2. Tester Subagent: Execute go test ./backend/internal/adapters/config/... and full regression suite via go test -race ./backend/... .
3. Reviewer Subagent: Audit diff in loader_test.go to confirm no other tests or imports were unintentionally affected.
