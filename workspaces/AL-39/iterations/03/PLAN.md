# Implementation Plan: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes (Iteration 03)

**Artifact Path**: `workspaces/AL-39/iterations/03/PLAN.md`  
**Ticket**: AL-39  
**Scope Reference**: `workspaces/AL-39/CONTRACT.md`, `workspaces/AL-39/iterations/01/PLAN.md`, `workspaces/AL-39/iterations/02/PLAN.md`

---

## 1. Goal and Iteration Scope

Iteration 03 addresses steering feedback regarding fragile filesystem crawling for SDLC iterations and cleans up non-source project file validation tests:

1. **Delete Non-Source YAML Test**: Remove `TestYAMLFiles_SyntaxValidation` from `backend/cmd/agentic-lens/release_verification_test.go`. Project configuration and CI workflow manifests are project setup files, not functional unit components of the source application under test.
2. **Telemetry-Native SDLC Iterations (Option A without Fallback)**:
   - Eliminate fragile filesystem crawling for iteration subdirectories (`workspaces/<TICKET>/iterations/`).
   - Remove `findWorkspaceDir` from `backend/internal/adapters/git/inference.go`.
   - Implement telemetry-native extraction of human steering turns directly from OpenCode SQLite session records in `backend/internal/adapters/opencode/extractor.go`.
   - Update `backend/cmd/agentic-lens/main.go` and `backend/internal/adapters/backfill/engine.go` to populate `codeFacts.SDLCIterations` using the extracted steering turns count (with default 1 when 0 or unknown, preserving explicit fallback if provided).
   - Deprecate `git.ComputeSDLCIterations` and decouple `git.ComputeCodeFacts` from filesystem inspection.
3. **Test Suite Modernization**:
   - Add targeted tests in `backend/internal/adapters/opencode/extractor_test.go` verifying accurate extraction of steering turns.
   - Update `inference_test.go` and `analyzer_test.go` to remove obsolete tests that relied on creating `workspaces/<TICKET>/iterations/01, 02` directories.
   - Maintain quality gates: zero emojis, zero `any` or `interface{}`, and clean verification via `make check` and `make test`.

---

## 2. Grounded Reference Architecture & Impacted Symbols

### Boundary Adapter Layer (OpenCode Telemetry Extraction)
- **File**: `backend/internal/adapters/opencode/extractor.go`
  - **Symbol**: `ExtractSteeringTurns(db *sql.DB, threadIDs []string) (int64, error)` (New function).
  - **Responsibilities**:
    - Validates inputs (`db` non-nil, `threadIDs` non-empty).
    - Sanitizes `threadIDs` via `buildQuotedIDList`.
    - Queries `message` records belonging strictly to the root session `threadIDs`.
    - Unmarshals message payloads into `rawMessageData`.
    - Filters rows where role equals `user`.
    - Returns the count of human steering turns.

### Boundary Adapter Layer (Git Analyzer and Inference)
- **File**: `backend/internal/adapters/git/inference.go`
  - **Symbol**: `findWorkspaceDir(workspacesDir, ticketID, subDir string) (string, bool)`: Remove entirely.
  - **Symbol**: `ComputeSDLCIterations(ticketID string, workspacesDir ...string) int64`: Deprecate and decouple from filesystem. Return default count 1 without performing any `os.Stat`, `os.ReadDir`, or directory traversals.
- **File**: `backend/internal/adapters/git/analyzer.go`
  - **Symbol**: `ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) (domain.CodeFacts, error)`: Remove calls to `ComputeSDLCIterations`. Assign `SDLCIterations` to `fallback.SDLCIterations` when fallback is non-nil and positive; otherwise assign 1. Remove unused local variable `wsDir`.

### Application Layer (CLI Orchestration & Backfill)
- **File**: `backend/cmd/agentic-lens/main.go`
  - **Symbol**: `runGenerate`.
  - In the OpenCode database extraction block, invoke `opencode.ExtractSteeringTurns(dbClient.DB(), knownThreadIDs)`.
  - After `git.ComputeCodeFacts`, override `codeFacts.SDLCIterations` with the extracted steering turn count if greater than or equal to 1. If 0 or omitted, retain fallback if provided, or default to 1.
- **File**: `backend/internal/adapters/backfill/engine.go`
  - **Symbol**: `recomputeTicketFacts`.
  - In the database recomputation step, invoke `opencode.ExtractSteeringTurns` on the resolved root session ID. Update `factLedger.Code.SDLCIterations` if turns are greater than or equal to 1.

### Verification and Test Files
- **File**: `backend/cmd/agentic-lens/release_verification_test.go`
  - **Symbol**: `TestYAMLFiles_SyntaxValidation`: Delete function. Retain `findRepoRoot` for existing release verification tests.
- **File**: `backend/internal/adapters/opencode/extractor_test.go`
  - Add comprehensive unit tests for `ExtractSteeringTurns`.
- **File**: `backend/internal/adapters/git/inference_test.go`
  - Update `TestComputeSDLCIterations_CustomWorkspacesDir` to remove obsolete iteration directory creation and assert default return behavior.
- **File**: `backend/internal/adapters/git/analyzer_test.go`
  - Update `TestInferenceOnActiveWorkspace` and `TestGitAnalyzer_AdversarialEdgeCases`.
  - Update subtest `custom workspaces_dir resolves iterations in ComputeCodeFacts` to remove directory creation and assert that `ComputeCodeFacts` does not perform filesystem lookups for iterations.

---

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Remove Non-Source YAML Test
- **Task 1.1**: Open `backend/cmd/agentic-lens/release_verification_test.go`.
- **Task 1.2**: Remove function `TestYAMLFiles_SyntaxValidation(t *testing.T)` (lines 349-378).
- **Task 1.3**: Retain `findRepoRoot() (string, error)` at line 380, verifying it remains accessible to `TestBinaryVersion_MatchesTag`, `TestGitCommitProvenance_EmbeddedInBinary`, `TestChecksumVerification_FailsOnMismatchedDigest`, and `TestChecksumVerification_FailsOnCorruptedArchive`.

### Phase 2: Telemetry-Native Steering Turns Extraction
- **Task 2.1**: Implement `ExtractSteeringTurns(db *sql.DB, threadIDs []string) (int64, error)` in `backend/internal/adapters/opencode/extractor.go`.
  - Check if `db == nil`; return 0 and `errors.New("database connection is nil")`.
  - Check if `len(threadIDs) == 0`; return 0 and nil.
  - Build sanitized quoted ID list using `buildQuotedIDList(threadIDs)`. Return formatted error if validation fails.
  - Execute parameterized/quoted query selecting `m.data` from `message m` where `m.session_id IN (<quotedList>)` ordered by `m.time_created ASC`.
  - Iterate through rows:
    - Scan `m.data` into `sql.NullString`.
    - If `!data.Valid` or trimmed string is empty, skip.
    - Unmarshal into `rawMessageData`. If unmarshaling fails, skip to ensure adversarial resilience against corrupted message rows.
    - If `rawMD.Role != nil && strings.EqualFold(strings.TrimSpace(*rawMD.Role), "user")`, increment turn counter.
  - Check `rows.Err()`; return formatted error on failure.
  - Return accumulated count and nil.

### Phase 3: Decouple Git Inference and Analyzer from Filesystem Crawling
- **Task 3.1**: Refactor `backend/internal/adapters/git/inference.go`.
  - Delete `findWorkspaceDir(workspacesDir, ticketID, subDir string) (string, bool)`.
  - Refactor `ComputeSDLCIterations(ticketID string, workspacesDir ...string) int64`:
    - Add deprecation comment indicating that SDLC iterations are telemetry-native from session steering turns.
    - Return `int64(1)` directly.
    - Remove imports of `os` and `path/filepath` if no longer used by other functions in `inference.go`.
- **Task 3.2**: Refactor `backend/internal/adapters/git/analyzer.go`.
  - In `ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) (domain.CodeFacts, error)`:
    - Remove local variable `wsDir := cfg.WorkspacesDir`.
    - In error/default branches (lines 345 and 362), set `SDLCIterations: 1` (or `fallback.SDLCIterations` if non-nil).
    - In normal return branch (line 424), replace `ComputeSDLCIterations(ticketID, cfg.WorkspacesDir)` with:
      - If `fallback != nil && fallback.SDLCIterations > 0`, use `fallback.SDLCIterations`.
      - Else, use `int64(1)`.
    - Populate `CodeFacts.SDLCIterations` with this resolved value.

### Phase 4: CLI and Backfill Telemetry Integration
- **Task 4.1**: Update `backend/cmd/agentic-lens/main.go`.
  - In `runGenerate`:
    - Declare `var extractedSteeringTurns int64`.
    - In the session database block (lines 173-192), call `opencode.ExtractSteeringTurns(dbClient.DB(), knownThreadIDs)`.
    - If error is nil and returned turns > 0, set `extractedSteeringTurns = turns`.
    - After calling `git.ComputeCodeFacts`:
      - If `extractedSteeringTurns >= 1`: set `codeFacts.SDLCIterations = extractedSteeringTurns`.
      - Else if `fallbackCode != nil && fallbackCode.SDLCIterations > 0`: set `codeFacts.SDLCIterations = fallbackCode.SDLCIterations`.
      - Else if `codeFacts.SDLCIterations <= 0`: set `codeFacts.SDLCIterations = 1`.
- **Task 4.2**: Update `backend/internal/adapters/backfill/engine.go`.
  - In `recomputeTicketFacts`:
    - In Section 2 (OpenCode DB Recomputation), when a root session is resolved and `opts.RecomputeDB` is true:
    - Call `turns, err := opencode.ExtractSteeringTurns(opencodeDB.DB(), []string{rootID})`.
    - If `err == nil && turns >= 1`: set `factLedger.Code.SDLCIterations = turns`.

### Phase 5: Test Suite Updates & Adversarial Hardening
- **Task 5.1**: Add `TestExtractSteeringTurns` in `backend/internal/adapters/opencode/extractor_test.go`.
  - Test single root user message returns 1.
  - Test three root user messages across turns return 3.
  - Test assistant messages are ignored.
  - Test user messages in child sessions (`parent_id != nil`) are ignored when querying root thread IDs.
  - Test empty thread IDs return 0 with nil error.
  - Test nil database returns error.
  - Test corrupted JSON message data is gracefully ignored without error.
- **Task 5.2**: Update `backend/internal/adapters/git/inference_test.go`.
  - In `TestComputeSDLCIterations_CustomWorkspacesDir`:
    - Remove temporary directory structure creation for `iterations/01, 02, 03, 04`.
    - Assert `git.ComputeSDLCIterations("ANY-TICKET")` returns 1.
- **Task 5.3**: Update `backend/internal/adapters/git/analyzer_test.go`.
  - Update `TestInferenceOnActiveWorkspace` to assert `git.ComputeSDLCIterations("AL-11") == 1`.
  - Update `TestGitAnalyzer_AdversarialEdgeCases` to assert `git.ComputeSDLCIterations("") == 1`.
  - In subtest `custom workspaces_dir resolves iterations in ComputeCodeFacts`:
    - Remove creation of `iterations/01, 02, 03` directories.
    - Assert that `ComputeCodeFacts` returns `SDLCIterations: 1` when fallback is nil, and preserves `fallback.SDLCIterations` when fallback is provided.

### Phase 6: Verification & Quality Gates
- **Task 6.1**: Run `make check` to verify formatting, vet, and static analysis.
- **Task 6.2**: Run `make test` to verify backend and frontend test suites pass with race detection.
- **Task 6.3**: Verify zero emojis and zero `any` across touched files.

---

## 4. Verification Strategy & Adversarial Assertions

### Verification Test Matrix

| Test Identifier | Target File | Preconditions | Input / Action | Expected Outcome |
| :--- | :--- | :--- | :--- | :--- |
| `TestExtractSteeringTurns_SingleUserTurn` | `extractor_test.go` | SQLite DB with root session | 1 user message in root session | Returns count = 1, error = nil |
| `TestExtractSteeringTurns_MultipleUserTurns` | `extractor_test.go` | SQLite DB with root session | 3 user messages in root session | Returns count = 3, error = nil |
| `TestExtractSteeringTurns_IgnoreAssistant` | `extractor_test.go` | SQLite DB with root session | 5 assistant messages, 0 user | Returns count = 0, error = nil |
| `TestExtractSteeringTurns_ExcludeChildSessions` | `extractor_test.go` | SQLite DB with root and child session | 1 user message in root, 2 user messages in child | Querying root ID returns count = 1 |
| `TestExtractSteeringTurns_EmptyThreads` | `extractor_test.go` | SQLite DB | `threadIDs = []string{}` | Returns count = 0, error = nil |
| `TestExtractSteeringTurns_NilDB` | `extractor_test.go` | None | `db = nil` | Returns count = 0, error != nil |
| `TestExtractSteeringTurns_MalformedJSON` | `extractor_test.go` | SQLite DB | Message row with invalid JSON string | Skips malformed row, returns count = 0, error = nil |
| `TestComputeSDLCIterations_Decoupled` | `inference_test.go` | No filesystem workspace created | Call `ComputeSDLCIterations("AL-39")` | Returns 1 without filesystem inspection |
| `TestComputeCodeFacts_DecoupledIterations` | `analyzer_test.go` | Valid git repo, no workspace iterations directory | Call `ComputeCodeFacts("AL-39", cfg, nil)` | `CodeFacts.SDLCIterations` == 1 |
| `TestComputeCodeFacts_PreserveFallbackIterations` | `analyzer_test.go` | Valid git repo | Fallback `CodeFacts` with `SDLCIterations: 4` | `CodeFacts.SDLCIterations` == 4 |
| `ReleaseVerification_NoYAMLTest` | `release_verification_test.go` | Repository root | Run release verification tests | `TestYAMLFiles_SyntaxValidation` is absent; remaining verification tests pass |

### Quality Gate Commands
- Static Analysis: `make check`
- Unit and Race Tests: `make test`
- Build Verification: `make build`
- Zero Emojis Check: Verify no Unicode emoji symbols in touched files.
- Zero Any Check: Verify no `any` or `interface{}` in modified Go sources.

---

## 5. Summary of Deliverables

- **Plan Artifact**: `workspaces/AL-39/iterations/03/PLAN.md`
- **Core Changes**:
  - Deleted non-source `TestYAMLFiles_SyntaxValidation`.
  - Added telemetry-native `ExtractSteeringTurns` in `opencode/extractor.go`.
  - Eliminated `findWorkspaceDir` and filesystem iteration crawling in `git/inference.go` and `git/analyzer.go`.
  - Connected steering turns extraction in `main.go` and `backfill/engine.go`.
  - Updated unit test fixtures in `extractor_test.go`, `inference_test.go`, and `analyzer_test.go`.
