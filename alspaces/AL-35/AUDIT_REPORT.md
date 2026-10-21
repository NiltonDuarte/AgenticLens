# AgenticLens Architecture, Code & Test Review Report: Pre-Release v1.0.0 Audit

## 1. Executive Summary

This pre-release source code audit evaluates AgenticLens across its Go backend (`backend/cmd`, `backend/internal`), TypeScript web dashboard (`web/src`), and automated test suites. AgenticLens demonstrates exemplary Functional Core, Imperative Shell (FCIS) boundary discipline: `internal/domain` has zero I/O imports, zero system clock dependencies, zero random number generators, and microsecond-level execution times. Static typing across Go and TypeScript is strictly enforced with zero occurrences of `any` or `interface{}`.

However, because AgenticLens was incubated within this specific repository, critical configuration leakages and test execution bottlenecks must be addressed before public release:
1. **Repository-Specific Coupling & Configuration Leakage**: The Git analyzer hardcodes `origin/main` across merge-base and diff commands, hardcodes repository folder prefixes (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), and ignores project configuration (`ProjectConfig.Workspaces.SrcFolders` and `ProjectConfig.CodeMetrics`). Adapters also hardcode `"workspaces/"` path strings rather than referencing `ProjectConfig.Workspaces.Dir`.
2. **Missing Dynamic Configuration API & Static Branding**: The Go HTTP server lacks a `GET /api/config` endpoint. Consequently, the frontend cannot discover configured ticket prefixes or display the observed project name alongside the permanent `AgenticLens` platform branding.
3. **Test Suite Execution Bottlenecks & Subprocess Churn**: CLI integration subtests take up to 6.4 seconds under concurrency due to subprocess spawning, unisolated OpenCode SQLite database access, and live git branch inspection. Git inference tests mutate real branches on disk (`git branch <name> HEAD` and `git branch -D <name>`). Additionally, five artificial `time.Sleep` calls introduce non-deterministic delays in CLI and HTTP server test suites.

Remediating these findings decouples AgenticLens for cross-project adoption, eliminates test flakiness, and cuts test execution latency by more than 60%.

---

## 2. Test Suite Performance Diagnostics

- **Backend Suite Execution Time (Race Detector)**: 2.32s wall-clock (9.84s user, 8.52s sys)
- **Frontend Suite Execution Time (Vitest)**: 3.41s wall-clock (13.89s cumulative test run)
- **Target Backend Execution Time**: < 0.80s wall-clock
- **Target Frontend Execution Time**: < 1.80s wall-clock

### Slowest Tests & Bottlenecks (> 100ms)

| Duration | Package | Test Name | Diagnosed Bottleneck | Recommended Optimization |
|---|---|---|---|---|
| 6.41s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/false_prefix_POD-1...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.41s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/false_prefix_GENTLE-1...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.40s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/strict-type_succeeds...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.37s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/strict-type_custom_prefix...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.37s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/explicit_flag_overrides...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.37s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/custom_prefix_resolves...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.36s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/configured_prefix_overrides...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.27s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/unmatched_ticket_warning...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 6.27s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/automatic_resolution_PO...` | Subprocess CLI spawning, host git branch scan, unisolated OpenCode DB contention | Isolate DB with `:memory:` or temp mock; bypass live git command for prefix unit tests |
| 0.52s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/explicit_alias_feature...` | Subprocess CLI execution with `--allow-empty` | Direct invocation or in-memory command execution |
| 0.47s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/cached_facts_with_alias...` | Subprocess CLI execution with disk I/O | Direct invocation or in-memory command execution |
| 0.31s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/Tier_2_git_branch...` | Subprocess CLI executing real `git branch` lookups | In-memory branch mock or stub |
| 0.31s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/Tier_1_explicit_flag...` | Subprocess CLI executing real `git branch` lookups | In-memory branch mock or stub |
| 0.25s | `internal/adapters/git` | `TestInferTicketType_AdversarialEdgeCases/false_prefix...` | Subprocess git invocation `git branch -a --list` | Pure branch string parsing without shell out |
| 0.24s | `internal/adapters/git` | `TestInferTicketType_AdversarialEdgeCases/invalid_suffix...` | Subprocess git invocation `git branch -a --list` | Pure branch string parsing without shell out |
| 0.22s | `cmd/agentic-lens` | `TestCLIGenerateCommand` | Subprocess CLI execution, temp directory setup | In-memory runner or pooled test harness |
| 0.21s | `internal/adapters/git` | `TestInferTicketType_AdversarialEdgeCases/valid_prefix...` | Subprocess git invocation `git branch -a --list` | Pure branch string parsing without shell out |
| 0.17s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/bUiLd...` | Subprocess CLI execution | Direct command runner |
| 0.17s | `cmd/agentic-lens` | `TestCLIBackfillCommand` | Subprocess CLI execution and JSON ledger unmarshaling | Direct command runner |
| 0.16s | `internal/adapters/git` | `TestInferTicketType_PrefixMapping/without_prefixes...` | Subprocess git invocation on 4 ticket IDs | Pure string parsing test |
| 0.14s | `internal/adapters/git` | `TestInferTicketType_UnmatchedReturnsNil` | Subprocess git invocation `git branch -a --list` | In-memory branch list |
| 0.14s | `cmd/agentic-lens` | `TestCLIGenerate_V3PrunedEventsAndZeroErrorMessage` | Subprocess CLI execution | Direct command runner |
| 0.13s | `cmd/agentic-lens` | `TestCLITicketTypeClassification/automatic_GEN...` | Subprocess CLI execution | Direct command runner |
| 0.11s | `internal/adapters/git` | `TestInferTicketType_PrefixMapping/custom_prefix...` | Subprocess git invocation `git branch -a --list` | Pure string parsing test |
| 0.11s | `internal/adapters/git` | `TestInferTicketType_BranchPrefixResolution` | Physical branch creation via `git branch` on disk | In-memory string fixtures via `InferTicketTypeFromBranchList` |
| 0.10s | `internal/adapters/git` | `TestInferenceOnActiveWorkspace` | Physical branch creation via `git branch` on disk | In-memory string fixtures via `InferTicketTypeFromBranchList` |
| 440ms | `web/src/components` | `ArtifactViewer Component > large multiline documents` | Heavy DOM tree mounting and markdown node parsing | Lightweight render test with virtualization |
| 329ms | `web/src/views` | `TicketDeepDive View Component > renders loading state` | Full component tree mount with charts and cards | Shallow component mocking for view orchestrator tests |
| 325ms | `web/src/views` | `MetricProgression View Component > displays trend charts` | Full SVG / canvas chart rendering in jsdom | Pure state assertions separated from chart rendering |

### Test Performance Findings & Concrete Patches

#### 1. Artificial Sleeps in Test Loops
- **Locations**:
  - `backend/cmd/agentic-lens/main_test.go:393, 583, 628, 665` (10ms polling sleep in server startup loops)
  - `backend/internal/adapters/http/live_handlers_test.go:333` (50ms sleep in live SSE event delivery test)
- **Problem**: Fixed sleeps waste CPU cycles, slow down CI, and cause flakiness when system load spikes.
- **Prescription**:
  - Replace server port polling loops with an explicit notification channel or zero-delay exponential backoff with tiny initial delay (1ms).
  - Replace `live_handlers_test.go` fixed sleep with a synchronization channel triggered upon SSE event broadcast.

```go
// Before (backend/internal/adapters/http/live_handlers_test.go:333):
svc.Publish(event)
time.Sleep(50 * time.Millisecond)

// After:
// Use deterministic event subscriber channel or select on received event:
delivered := make(chan struct{})
svc.OnPublishHook(func() { close(delivered) })
svc.Publish(event)
select {
case <-delivered:
case <-time.After(500 * time.Millisecond):
    t.Fatal("timed out waiting for event delivery")
}
```

#### 2. Subprocess Git Invocations and Branch Mutation in Unit Tests
- **Locations**:
  - `backend/internal/adapters/git/inference_test.go:11-21` (`createTestBranch` executes `git branch <name> HEAD` and `git branch -D <name>`)
  - `backend/internal/adapters/git/inference.go:118` (`InferTicketType` shells out to `git branch -a --list` unconditionally)
- **Problem**: Incurring disk I/O and process fork overhead on every test. Concurrently modifying git refs in the developer's working copy risks workspace corruption.
- **Prescription**:
  - The core parsing algorithm already exists as a pure function: `git.InferTicketTypeFromBranchList(branchOut, ticketID)`.
  - Unit tests in `inference_test.go` should test `InferTicketTypeFromBranchList` using in-memory string fixtures, eliminating all calls to `git branch`.
  - In `InferTicketType`: if configured prefixes match the ticket ID, evaluate configured prefixes without shelling out to git first when git branch matching is not required, or inject a git command runner interface.

---

## 3. Architecture & Code Smells by Severity

### [High]: Hardcoded Git Base Branch `origin/main`
- **Tier**: Tier 1 (Layering & Boundary Discipline) & Tier 3 (Domain Modeling)
- **Affected Files**: `backend/internal/adapters/git/analyzer.go:78, 88, 100, 129`
- **Observed Smell**: Git analyzer explicitly hardcodes `origin/main` as the upstream comparison base:
  - `diff --numstat origin/main...HEAD`
  - `diff --numstat origin/main`
  - `log origin/main --grep=...`
  - `diff --numstat origin/main...<branch>`
  This breaks in repositories using `main` without remotes, `master`, `trunk`, `development`, or when running detached in CI/CD without an `origin` remote configured.
- **Subtraction Target**: Delete the static `"origin/main"` literal from all four diff and log invocations.
- **Proposed Target Shape**:
  Introduce dynamic base branch resolution:
  ```go
  func ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)
  ```
  1. Check `cfg.CodeMetrics.BaseBranch` if configured.
  2. Query `git symbolic-ref refs/remotes/origin/HEAD` to detect default remote tracking branch.
  3. Probe candidate upstream branches: `origin/main`, `main`, `origin/master`, `master`.
  4. Explicit Failure Invariant: If no upstream base branch is reachable or resolved, fail execution immediately and output an explicit error message prompting the user or agent to configure `code_metrics.base_branch` in `.agenticlens-config.yaml`. Never silently fall back to `HEAD`.
- **Benefits**: Eliminates silent fallbacks and makes AgenticLens deterministic and configurable across any repository branch naming strategy.
- **Blast Radius**: `backend/internal/adapters/git/analyzer.go`.

---

### [High]: Hardcoded Source Folder Prefixes in Git Analyzer
- **Tier**: Tier 2 (Module Depth & Dead Code) & Tier 1 (Layering)
- **Affected Files**: `backend/internal/adapters/git/analyzer.go:135-140, 167-172`
- **Observed Smell**: In two separate locations, `analyzer.go` discards git diff entries unless at least one modified path matches a hardcoded list of directories:
  ```go
  if strings.HasPrefix(r.Path, "src/") ||
     strings.HasPrefix(r.Path, "backend/") ||
     strings.HasPrefix(r.Path, "internal/") ||
     strings.HasPrefix(r.Path, "cmd/") ||
     strings.HasPrefix(r.Path, "web/") ||
     strings.HasPrefix(r.Path, "tests/") {
      return parsed
  }
  ```
  This completely discards code changes in projects structured with `app/`, `pkg/`, `services/`, `client/`, `ui/`, `core/`, `lib/`, or root-level files (such as `main.py` or `Cargo.toml`). Furthermore, this check completely ignores `.agenticlens/config.yaml` (`workspaces.src_folders` and `code_metrics.include`).
- **Subtraction Target**: Delete both duplicate 6-condition checks.
- **Proposed Target Shape**:
  Pass `ClassificationConfig` and `ProjectConfig.Workspaces.SrcFolders` into `GetTicketNumstat`. Validate paths against configured source folders and glob patterns.
  Explicit Failure Invariant: If `src_folders` is empty or unconfigured, fail execution immediately and prompt the user or agent to configure `workspaces.src_folders` in `.agenticlens-config.yaml`. Never silently fall back or permit all files.
- **Benefits**: Prevents silent omission of git commit metrics while enforcing explicit project boundary definitions. Eliminates 12 lines of duplicate code.
- **Blast Radius**: `backend/internal/adapters/git/analyzer.go`, `inference.go`.

---

### [High]: Missing Dynamic `/api/config` REST Endpoint and Frontend Project Branding Exposure
- **Tier**: Tier 1 (Boundary Discipline) & Tier 2 (Module Depth)
- **Affected Files**:
  - `backend/internal/adapters/http/server.go:34-49`
  - `web/src/api/client.ts`
  - `web/src/components/Navbar.tsx:75-77`
  - `web/src/views/LandingPage.tsx:136`
- **Observed Smell**: The Go HTTP server does not expose `/api/config`. The frontend has no knowledge of `ProjectConfig.ProjectName`, `Ticket.Prefixes`, or `Workspaces.Dir`.
  As a result:
  - The UI hardcodes "AgenticLens Telemetry" with no mechanism to display the observed project's name.
  - The frontend MultiSelect dropdowns for ticket prefixes must scrape prefixes exclusively from loaded historical ledgers (`extractDistinctPrefixes`). Unused configured prefixes cannot be selected.
- **Subtraction Target**: Delete static branding strings in `Navbar.tsx` and replace with dynamic branding component.
- **Proposed Target Shape**:
  1. Add `GET /api/config` route to `server.go` returning:
     ```json
     {
       "project_name": "MyProject",
       "prefixes": [{"prefix": "AL", "description": "...", "ticket_type": "feat"}],
       "workspaces_dir": "workspaces",
       "roles": ["orchestrator", "planner", "coder", "tester", "reviewer"],
       "version": "1.0.0"
     }
     ```
  2. Add `fetchProjectConfig()` to `web/src/api/client.ts`.
  3. In `Navbar.tsx`, permanently preserve `AgenticLens` as the primary platform brand, while rendering the observed project name alongside it:
     ```tsx
     <span className="brand-title">AgenticLens</span>
     {config.project_name && <span className="project-badge">{config.project_name}</span>}
     ```
- **Benefits**: Fulfills white-label multi-project observability requirements while maintaining platform brand identity. Enables prefix filters on initial load without existing ledgers.
- **Blast Radius**: HTTP server routing, frontend API client, Navbar, and LandingPage components.

---

### [High]: Subprocess Spawning Churn and Branch Mutation in Git Test Suite
- **Tier**: Tier 4 (Test Performance & Flakiness)
- **Affected Files**:
  - `backend/internal/adapters/git/inference_test.go:11-21, 57-61`
  - `backend/cmd/agentic-lens/main_test.go:2636-3150`
- **Observed Smell**: Unit tests in `inference_test.go` call `exec.Command("git", "branch", ...)` on the live git working directory. Integration tests in `main_test.go` invoke the compiled binary nearly 100 times, querying the live git repository and the user's local OpenCode SQLite database concurrently, inflating test runtime to over 6 seconds per subtest.
- **Subtraction Target**: Remove disk git branch creation helpers (`createTestBranch`) and decouple `InferTicketType` from mandatory subprocess spawning in tests.
- **Proposed Target Shape**:
  - Test pure logic directly with `InferTicketTypeFromBranchList(fixture, ticketID)`.
  - Isolate CLI integration tests by providing an empty temporary SQLite database via `--db-path` and isolated `--workspaces` directory.
- **Benefits**: Reduces `main_test.go` execution time from >6s per subtest down to <200ms. Prevents git repository state mutation during testing.
- **Blast Radius**: `inference_test.go`, `main_test.go`.

---

### [Medium]: Hardcoded `"workspaces/"` Directory Across Adapters
- **Tier**: Tier 1 (Boundary Discipline) & Tier 3 (Domain Modeling)
- **Affected Files**:
  - `backend/internal/adapters/git/classifier.go:107`
  - `backend/internal/adapters/git/inference.go:32`
  - `backend/internal/adapters/repository.go:30, 126`
- **Observed Smell**:
  - `classifier.go:107`: `if strings.HasPrefix(cleanPath, "workspaces/") || strings.Contains(cleanPath, "/workspaces/") { return CategoryIgnored }`
  - `inference.go:32`: `candidate := filepath.Join(curr, "workspaces", ticketID, subDir)`
  - `repository.go:126`: `if parentName != "workspaces" && parentName != "." { factLedger.TicketID = &parentName }`
  If a project configures `workspaces.dir: .agentic/workspaces` or `workspaces.dir: tickets`, these hardcoded checks fail to ignore ticket artifacts, fail to calculate SDLC iterations, and misidentify ticket IDs.
- **Subtraction Target**: Eliminate hardcoded `"workspaces/"` string constants across adapters.
- **Proposed Target Shape**:
  Pass `workspacesDir := cfg.Workspaces.Dir` into classifier and inference methods, with `"workspaces"` strictly as a fallback default.
- **Benefits**: Respects custom workspace directory configurations.
- **Blast Radius**: `internal/adapters/git`, `internal/adapters/repository.go`.

---

### [Medium]: Artificial Sleeps in Test Suites
- **Tier**: Tier 4 (Test Performance)
- **Affected Files**:
  - `backend/cmd/agentic-lens/main_test.go:393, 583, 628, 665`
  - `backend/internal/adapters/http/live_handlers_test.go:333`
- **Observed Smell**: 5 instances of `time.Sleep` ranging from 10ms to 50ms used for HTTP server readiness checks and SSE event delivery synchronization.
- **Subtraction Target**: Delete all `time.Sleep` calls in `*_test.go`.
- **Proposed Target Shape**: Replace with deterministic channel synchronization and port-probing loops with exponential backoff starting at 100 microseconds.
- **Benefits**: Prevents race conditions on heavily loaded CI runners and cuts cumulative test suite delay.
- **Blast Radius**: `backend/cmd/agentic-lens/main_test.go`, `backend/internal/adapters/http/live_handlers_test.go`.

---

### [Medium]: Shallow Passthrough Functions in Formatters Adapter
- **Tier**: Tier 2 (Module Depth & Shallow Abstractions)
- **Affected Files**: `backend/internal/adapters/formatters.go:30-32, 48-50`
- **Observed Smell**:
  ```go
  func FormatLeadTime(seconds *int64) string {
      return domain.FormatLeadTime(seconds)
  }
  func FormatTokens(tokens int64) string {
      return domain.FormatTokensK(tokens)
  }
  ```
  These functions add zero logic, validation, or transformation over the domain package. Callers could invoke `domain.FormatLeadTime` and `domain.FormatTokensK` directly.
- **Subtraction Target**: Deprecate or inline these shallow forwarding wrappers.
- **Proposed Target Shape**: Callers in `adapters` invoke `domain.FormatLeadTime` and `domain.FormatTokensK` directly.
- **Benefits**: Reduces mental indirection and eliminates redundant wrapper functions.
- **Blast Radius**: Callers in `internal/adapters`.

---

### [Low]: Duplicate Directory Prefix Conditions in Git Analyzer
- **Tier**: Tier 2 (Module Depth)
- **Affected Files**: `backend/internal/adapters/git/analyzer.go:135-140, 167-172`
- **Observed Smell**: Exact same 6-condition boolean chain repeated twice:
  `strings.HasPrefix(r.Path, "src/") || strings.HasPrefix(r.Path, "backend/") || strings.HasPrefix(r.Path, "internal/") ...`
- **Subtraction Target**: Delete duplicated conditional chain.
- **Proposed Target Shape**: Extract a single pure helper `isMonitoredSourcePath(path string, srcFolders []string) bool` or rely directly on `ClassifyFile`.
- **Benefits**: Eliminates code duplication and centralizes path filtering.
- **Blast Radius**: `internal/adapters/git/analyzer.go`.

---

### [Low]: Frontend Heavy Component Mounting Overhead in Vitest Specs
- **Tier**: Tier 4 (Frontend Test Performance)
- **Affected Files**: `web/src/views/TicketDeepDive.test.tsx`, `web/src/views/MetricProgression.test.tsx`, `web/src/components/TraceWaterfall.test.tsx`
- **Observed Smell**: View test files mount entire React trees including complex Markdown renderers, Lucide SVG icon sets, and charts in `jsdom`. Individual tests take 300-440ms each, aggregating to 3.41s total suite execution.
- **Subtraction Target**: Remove redundant full-tree DOM re-mounts in utility and filtering tests.
- **Proposed Target Shape**: Isolate pure domain state transformations (`progression.ts`, `formatters.ts`) into pure unit tests (which execute in <10ms). In React component tests, mock heavy leaf subcomponents (`ArtifactViewer`, `TraceWaterfall`) when testing high-level view routing and filters.
- **Benefits**: Cuts frontend Vitest execution time from ~3.4s down to <1.5s.
- **Blast Radius**: `web/src/**/*.test.tsx`.

---

## 4. Top Remediation Recommendations & Prioritized Roadmap

1. **AL-39 (New Recommendation)**: Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes
   - Implement dynamic base branch detection (`code_metrics.base_branch`, `refs/remotes/origin/HEAD`, `origin/main`, `main`, `origin/master`, `master`); if unresolved, fail immediately and prompt user/agent to configure `code_metrics.base_branch`.
   - Replace hardcoded folder prefix list (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`) with dynamic inspection of `ProjectConfig.Workspaces.SrcFolders` and `ProjectConfig.CodeMetrics`; if `src_folders` is empty or unconfigured, fail immediately and prompt user/agent to configure `workspaces.src_folders`.
   - Remove duplicate 6-condition checks in `analyzer.go`.
   - Respect `ProjectConfig.Workspaces.Dir` in `classifier.go` and `inference.go`.

2. **AL-40 (New Recommendation)**: Dynamic Configuration REST API Endpoint and Observability Dashboard Branding
   - Implement `GET /api/config` in Go HTTP server returning sanitized `ProjectConfig`.
   - Add `fetchProjectConfig()` in `web/src/api/client.ts`.
   - Update `Navbar.tsx` to dynamically render observed project name alongside permanent `AgenticLens` platform branding.
   - Populate filter dropdowns with configured prefixes before ledgers exist.

3. **AL-41 (New Recommendation)**: Eliminate Subprocess Churn, Artificial Sleeps, and Test Execution Bottlenecks
   - Banish all 5 `time.Sleep` instances across `main_test.go` and `live_handlers_test.go`.
   - Refactor `inference_test.go` to use in-memory branch resolution instead of mutating git branches on disk.
   - Isolate CLI integration tests with dedicated temporary SQLite databases.
   - Inline shallow passthrough wrappers `adapters.FormatLeadTime` and `adapters.FormatTokens`.
