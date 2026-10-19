# Implementation Plan: AL-31 Complete Structural Hierarchy Enforcement and CI Git Test Hermeticity (Iteration 03)

## 1. Goal and Iteration Scope

Iteration 03 addresses user steering to complete the transition to pure structural session hierarchy and resolve flaky CI test failures:
- Completely eliminate `IsSubagentRole` from the domain layer and all test suites. Legacy role string heuristics and deprecation shims are removed.
- Refactor `detectSubagentReworkLoops` in `backend/internal/domain/anomalies.go` to strictly check structural hierarchy via `ParentID != nil` for child invocations and session lookup mapping for events.
- Update all synthetic unit tests in `backend/internal/domain/anomalies_test.go` to provide explicit `Invocations` and `SessionID` fixtures reflecting authentic hierarchy (`ParentID == nil` for root coordination, `ParentID != nil` for delegated subagents).
- Eliminate ambient branch dependencies in git inference test suites (`backend/internal/adapters/git/inference_test.go` and `backend/internal/adapters/git/analyzer_test.go`). Introduce self-contained, transient git branch fixtures with automatic cleanup to ensure deterministic pass rates in fresh CI checkouts.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation touches the following files and symbols across domain and adapter layers:

### Domain Core (backend/internal/domain/)
- `backend/internal/domain/enums.go`:
  - Delete `IsSubagentRole(role string) bool`.
  - Audit `OrderSubagentRoles` and `SubagentRoleOrder` to verify zero remaining dependencies on `IsSubagentRole`.
- `backend/internal/domain/anomalies.go`:
  - In `detectSubagentReworkLoops`:
    - Eliminate `hasHierarchy` flag and fallback branches calling `IsSubagentRole`.
    - Construct `isChildSession` mapping `inv.ID` to boolean `inv.ParentID != nil` from `ledger.Invocations`.
    - Filter events from `ledger.Events`: accept an event if and only if `isChildSession[e.SessionID]` is true.
    - Filter invocations from `ledger.Invocations` (fallback when events produce zero roles): accept an invocation if and only if `inv.ParentID != nil`.
    - Preserve existing 3-4 cycle WARNING and 5+ cycle CRITICAL ping-pong evaluation.
- `backend/internal/domain/enums_test.go`:
  - Delete `TestIsSubagentRole_DeprecationAndBackwardsCompatibility`.
- `backend/internal/domain/aggregations_test.go`:
  - Delete `TestIsSubagentRole`.
- `backend/internal/domain/anomalies_test.go`:
  - In `TestDetectAnomalies_SubagentReworkLoop`:
    - Update test cases lacking invocation hierarchy (`coder <-> tester 3 cycles`, `coder <-> tester 4 cycles`, `coder <-> tester 5 cycles`, `fewer than 3 cycles`, `non-alternating transitions`, `orchestrator role is strictly ignored`, `consecutive identical roles collapsed`) to include root invocations with `ParentID == nil` and child invocations with `ParentID != nil`, mapping events via `SessionID`.
  - In `TestDetectAnomalies_DeterministicSortingAllRules`:
    - Provide `Invocations` defining root session and child session for `s1` (`ParentID != nil`) so `RuleSubagentReworkLoop` triggers deterministically alongside the other three anomaly rules.

### Boundary Adapters (backend/internal/adapters/git/)
- `backend/internal/adapters/git/inference_test.go`:
  - Define a hermetic test branch fixture helper creating transient branches pointing to HEAD and registering deletion via `t.Cleanup`.
  - In `TestInferTicketType_BranchPrefixResolution`: create transient fixture branches for `AL-1` (`feat`), `AL-14` (`ci`), `AL-16` (`refactor`), and `AL-30` (`fix`), verify `InferTicketType` resolves them, and rely on `t.Cleanup` for removal.
  - In `TestInferTicketType_Precedence_BranchOverConfiguredPrefix`: create transient fixture branch for `AL-30` (`fix`), assert it overrides the configured `chore` prefix, and rely on `t.Cleanup` for removal.
- `backend/internal/adapters/git/analyzer_test.go`:
  - In `TestInferenceOnActiveWorkspace`: create transient fixture branch for `AL-11` (`feat`) with automatic `t.Cleanup`, guaranteeing `InferTicketType("AL-11")` returns `feat` in CI without reliance on ambient repo state.

---

## 3. Phased Implementation Tasks

### Phase 1: Complete Removal of IsSubagentRole from Domain Engine
- Task 1.1: In `backend/internal/domain/enums.go`, delete `IsSubagentRole(role string) bool` and its associated doc comments.
- Task 1.2: Audit `OrderSubagentRoles` in `backend/internal/domain/enums.go` to confirm it accepts trimmed non-empty candidate roles directly from child invocations without persona blacklists.
- Task 1.3: In `backend/internal/domain/enums_test.go`, remove `TestIsSubagentRole_DeprecationAndBackwardsCompatibility`.
- Task 1.4: In `backend/internal/domain/aggregations_test.go`, remove `TestIsSubagentRole`.
- Task 1.5: Audit the entire `backend/` codebase to verify zero remaining references to `IsSubagentRole`.

### Phase 2: Refactor Anomaly Detection to Pure Structural Hierarchy
- Task 2.1: In `backend/internal/domain/anomalies.go`, locate `detectSubagentReworkLoops`.
- Task 2.2: Remove the `hasHierarchy` boolean flag and all calls to `IsSubagentRole`.
- Task 2.3: Build a child session lookup map `isChildSession` from `ledger.Invocations` indexing non-empty `inv.ID` to `(inv.ParentID != nil)`.
- Task 2.4: Process `ledger.Events`:
  - Check whether `isChildSession[e.SessionID]` is true.
  - If true, normalize `e.Role` (trim whitespace, lowercase). If non-empty and different from the last recorded role, append to the subagent role sequence.
  - If false or unmapped, ignore the event.
- Task 2.5: Process fallback `ledger.Invocations` when events yield zero roles:
  - Check whether `inv.ParentID != nil`.
  - If true, normalize `inv.Role` (trim whitespace, lowercase). If non-empty and different from the last recorded role, append to the subagent role sequence.
  - If false, ignore the invocation.
- Task 2.6: Preserve cycle evaluation and threshold rules:
  - Consecutive alternating pairs triggering WARNING at 3 to 4 cycles.
  - Consecutive alternating pairs triggering CRITICAL at 5 or more cycles.

### Phase 3: Synchronize Anomaly Detection Test Fixtures
- Task 3.1: In `backend/internal/domain/anomalies_test.go`, update `TestDetectAnomalies_SubagentReworkLoop`:
  - Subtest `coder <-> tester 3 cycles triggers WARNING`: Add `Invocations` containing a root invocation with `ParentID == nil` and child invocations for coder and tester with `ParentID != nil`. Set `SessionID` on events matching the child invocation identifiers.
  - Subtest `coder <-> tester 4 cycles triggers WARNING`: Add root and child invocations with `ParentID != nil` and map event `SessionID`.
  - Subtest `coder <-> tester 5 cycles triggers CRITICAL`: In the invocation fallback list, add a root invocation with `ParentID == nil` and supply 10 child invocations with `ParentID != nil` pointing to the root ID, alternating roles.
  - Subtest `fewer than 3 cycles does not trigger`: Add root and child invocations with `ParentID != nil` and map event `SessionID`.
  - Subtest `non-alternating transitions do not trigger`: Add root and child invocations for coder, tester, and reviewer with `ParentID != nil` and map event `SessionID`.
  - Subtest `orchestrator role is strictly ignored in rework loop detection`: Add root invocation with `Role: "orchestrator"` and `ParentID == nil`, and child invocations with `ParentID != nil`. Set `SessionID` on orchestrator events to the root ID, and child session IDs on coder and tester events.
  - Subtest `consecutive identical roles collapsed without false trigger`: Add root and child invocations with `ParentID != nil` and map event `SessionID`.
- Task 3.2: In `backend/internal/domain/anomalies_test.go`, update `TestDetectAnomalies_DeterministicSortingAllRules`:
  - Add `Invocations` containing a root invocation with `ParentID == nil` and a child invocation with `ID: "s1"` and `ParentID != nil` pointing to root, ensuring session `s1` is recognized as a subagent so `RuleSubagentReworkLoop` triggers as expected.
- Task 3.3: Verify existing adversarial tests in `anomalies_test.go` (`structural hierarchy ignores root session events`, `structural hierarchy ignores root invocations in fallback sequence`, `adversarial: root session labeled coder interleaved`, `adversarial: root session events alternating subagent roles do not trigger`, `adversarial: invocations fallback ignores root invocation labeled reviewer`) pass cleanly.

### Phase 4: Make Git Inference Tests Hermetic and Deterministic in CI
- Task 4.1: In `backend/internal/adapters/git/inference_test.go`:
  - Implement a test helper function `createTestBranch(t *testing.T, branchName string)`:
    - Execute `git branch <branchName>` pointing to current HEAD.
    - If branch already exists, delete it first via `git branch -D <branchName>`.
    - Register automatic cleanup with `t.Cleanup` executing `git branch -D <branchName>`.
  - Refactor `TestInferTicketType_BranchPrefixResolution`:
    - Remove reliance on ambient developer branches.
    - Create transient fixture branches: `feat/AL-1-ci-fixture`, `ci/AL-14-ci-fixture`, `refactor/AL-16-ci-fixture`, and `fix/AL-30-ci-fixture`.
    - Call `git.InferTicketType(tc.ticketID)` and assert resolution to expected conventional types (`feat`, `ci`, `refactor`, `fix`).
    - Confirm all fixture branches are deleted on test exit.
  - Refactor `TestInferTicketType_Precedence_BranchOverConfiguredPrefix`:
    - Create transient fixture branch `fix/AL-30-ci-fixture` pointing to HEAD with automatic `t.Cleanup`.
    - Call `git.InferTicketType("AL-30", conflictingPrefixes...)` where `conflictingPrefixes` maps `AL` to `chore`.
    - Assert `fix` is returned, proving Tier 2 branch search overrides Tier 3 configured prefixes.
    - Retain unbranched assertions (`PO-5`, `CUSTOM-99`) verifying Tier 3 fallback and nil return.
- Task 4.2: In `backend/internal/adapters/git/analyzer_test.go`:
  - Implement or share `createTestBranch(t *testing.T, branchName string)`.
  - In `TestInferenceOnActiveWorkspace`:
    - Create transient fixture branch `feat/AL-11-ci-fixture` pointing to HEAD with automatic `t.Cleanup`.
    - Call `git.InferTicketType("AL-11")` and assert resolution to `feat`.
    - Retain existing title, iteration, and git user assertions.
- Task 4.3: Ensure tests that create and delete git branch refs do not use `t.Parallel()` to avoid git ref lock contention.

### Phase 5: Verification and Quality Gates
- Task 5.1: Run `go vet ./...` across all packages to confirm static analysis clean state.
- Task 5.2: Run `go test -race -count=1 ./...` across all Go packages to verify absence of data races and confirm 100% test pass rate in simulated isolated checkout.
- Task 5.3: Run `backend/internal/adapters/parity_test.go` to confirm zero regressions across reference workspace ledgers.
- Task 5.4: Execute `make build` to verify clean single-binary compilation.
- Task 5.5: Confirm zero emojis and zero code blocks across all modified files and workspace documents.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 completely removes `IsSubagentRole` from `enums.go` and deletes the corresponding unit tests.
2. Phase 2 refactors `detectSubagentReworkLoops` in `anomalies.go` to use pure structural hierarchy (`ParentID != nil` and `isChildSession`).
3. Phase 3 synchronizes `anomalies_test.go` fixtures to supply root and child invocations with explicit parent pointers and session identifiers.
4. Phase 4 introduces transient fixture branches in `inference_test.go` and `analyzer_test.go`, fixing CI failures.
5. Phase 5 executes the verification suite (`go vet`, `go test -race`, `parity_test.go`, `make build`).
