# Code Review Audit: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes

**Workspace Target**: `workspaces/AL-39/`  
**Iteration**: `01`  
**Scope Reference**: `workspaces/AL-39/CONTRACT.md`, `workspaces/AL-39/iterations/01/PLAN.md`, `TECHNICAL_GUIDELINES.md`  
**Review Type**: Read-only verification against Axis 1 (Spec & Contract Alignment) and Axis 2 (Engineering Standards & Architecture)

---

## Axis 1: Spec & Contract Alignment

### 1. Dynamic Upstream Base Branch Resolution
- **Resolution Pipeline**: Implemented in `backend/internal/adapters/git/analyzer.go:ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)`.
  - **Tier 1 (Explicit Configuration)**: Checks `cfg.CodeMetrics.BaseBranch` after trimming whitespace. Verifies commit existence via `git rev-parse --verify --quiet <branch>^{commit}`. Returns configured branch if reachable; returns explicit error if configured but not reachable.
  - **Tier 2 (Symbolic Reference)**: Executes `git symbolic-ref refs/remotes/origin/HEAD`. Trims `refs/remotes/` and verifies commit existence via `rev-parse`. Returns tracked branch if reachable; falls through if broken.
  - **Tier 3 (Candidate Probes)**: Sequentially probes `origin/main`, `main`, `origin/master`, and `master` via `rev-parse --verify --quiet <candidate>^{commit}` in exact priority order.
  - **Tier 4 (Failure Invariant)**: Returns an empty string and explicit error: `unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml`.
  - **Prohibition Invariant**: Confirmed zero fallback to `HEAD`.
- **Diff / Log Invocations in `GetTicketNumstat`**:
  - Replaced all 4 hardcoded `origin/main` strings with dynamic `baseBranch`:
    - `diff --numstat <baseBranch>...HEAD`
    - `diff --numstat <baseBranch>`
    - `log <baseBranch> --grep=...`
    - `diff --numstat <baseBranch>...<branch>`

### 2. Monitored Source Folders Validation & Pure Path Matching
- **Input Validation**: `ValidateSrcFolders(srcFolders []string) error` in `analyzer.go`.
  - Fails if `len(srcFolders) == 0` or if all slice elements are empty/whitespace-only.
  - Returns exact prompt: `workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml`.
  - Confirmed zero silent fallback or indiscriminate acceptance of all files.
- **Pure Path Matching**: `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool`.
  - Normalizes path using forward slashes.
  - Evaluates folder boundaries (`cleanPath == normFolder || strings.HasPrefix(cleanPath, normFolder+"/")`).
  - Evaluates `cfg.Include` and `cfg.ProdPatterns` using `matchPattern`.
  - Eliminates false positive boundary matches (e.g. `app_helper.go` does not falsely match `app`).
- **Duplicate Boolean Chains**: The two identical 6-condition chains previously duplicated at lines 135-140 and 167-172 of `analyzer.go` have been removed and replaced with calls to `IsMonitoredSourcePath(r.Path, srcFolders, cfg)`.

### 3. Dynamic Workspaces Directory Decoupling
- **Classification**: `ClassifyFile(rawPath string, cfg *ClassificationConfig)` in `classifier.go` uses `cfg.WorkspacesDir` (defaulting to `"workspaces"` when unconfigured or whitespace). Ignores paths matching `wsDir`, `wsDir+"/"`, and `"/"+wsDir+"/"`.
- **Iteration Inference**: `findWorkspaceDir` and `ComputeSDLCIterations(ticketID string, workspacesDir ...string)` in `inference.go` accept custom workspace directory paths.
- **Fact Derivation**: `ComputeCodeFacts` passes `cfg.WorkspacesDir` into `ComputeSDLCIterations`.

### 4. Upward Error Propagation & CLI Handling
- `ComputeCodeFacts` updated signature returns `(domain.CodeFacts, error)`.
- `main.go:runGenerate` checks the returned error, writes configuration diagnostic prompt to `os.Stderr`, and terminates with `os.Exit(1)`.
- `backfill/engine.go:recomputeTicketFacts` inspects errors from `ComputeCodeFacts` and wraps them with `%w` into the migration result instead of masking failures.

### 5. Scope Invariants & Zero Scope Creep
- Inspection strictly confined to `workspaces/AL-39/` and ticket deliverables.
- Zero changes to frontend views in `web/src/`.
- Zero changes to SQLite database schemas or OpenCode session tables.
- Zero premature implementations of deferred tickets (AL-40 configuration REST endpoint and AL-41 test runtime optimization).

---

## Axis 2: Engineering Standards & Architecture

### 1. Functional Core, Imperative Shell (FCIS)
- **Domain Core (`backend/internal/domain/config.go`)**:
  - `CodeMetricsConfig` extended with `BaseBranch string` with JSON and YAML tags.
  - `DefaultProjectConfig()` sets `BaseBranch: ""`.
  - Domain imports strictly restricted to pure standard libraries: `fmt`, `regexp`, `sort`, `strings`. Zero I/O, zero network, zero `os/exec`, zero clocks, and zero outer-layer imports. Pure value-semantic immutable structs preserved.
- **Boundary Adapters (`backend/internal/adapters/`)**:
  - All Git command executions isolated in `adapters/git/`.
  - All configuration file reads and unmarshaling isolated in `adapters/config/`.

### 2. Typing Discipline & Zero Any Invariant
- Zero `any` or `interface{}` instances introduced across `domain/config.go`, `adapters/config/loader.go`, `adapters/git/analyzer.go`, `adapters/git/classifier.go`, `adapters/git/inference.go`, `cmd/agentic-lens/main.go`, and `adapters/backfill/engine.go`.
- Concrete slices (`[]string`, `[]NumstatEntry`) and strongly-typed structs throughout.

### 3. Zero Emojis Mandate
- Grepped all touched source files, comments, docstrings, error messages, test cases, and workspace documents for non-ASCII codepoints and emojis.
- Total emoji count: 0.

### 4. Test Coverage & Quality Gates
- **Loader Tests (`loader_test.go`)**:
  - `TestLoad_CodeMetricsBaseBranch`: Verifies YAML unmarshaling, JSON unmarshaling, and default empty string fallback.
- **Analyzer Tests (`analyzer_test.go`)**:
  - `TestResolveBaseBranch_Configured`: Configured branch existing and trimmed.
  - `TestResolveBaseBranch_Configured_Missing`: Configured branch missing triggers explicit failure error (no fallback to HEAD).
  - `TestResolveBaseBranch_SymbolicRef`: Valid symbolic ref detection and broken symbolic ref fallthrough.
  - `TestResolveBaseBranch_ProbeCandidates`: Sequential candidate priority probes (`origin/main` > `main` > `origin/master` > `master`).
  - `TestResolveBaseBranch_Unresolvable_FailureInvariant`: Unresolvable base branch returns exact configuration prompt and never HEAD.
  - `TestValidateSrcFolders`: Nil, empty, whitespace-only slices fail; valid slices succeed.
  - `TestIsMonitoredSourcePath`: Table-driven tests verifying custom directories (`app/`, `pkg/`, `services/`), root includes (`main.py`, `vite.config.ts`), prod patterns (`*.vue`, `core/**`), and false prefix boundaries.
  - `TestGetTicketNumstat_DynamicBranchAndDuplicateConditionsEliminated`: Verifies elimination of duplicate conditions and dynamic branch filtering.
  - `TestComputeCodeFacts_Invariants`: Verifies invariant failure propagation and custom workspace iteration resolution.
- **Classifier & Inference Tests (`classifier_test.go`, `inference_test.go`)**:
  - `TestClassifyFile_WorkspacesDir`: Custom workspace folder ignored; boundary prefix paths not falsely ignored.
  - `TestComputeSDLCIterations_CustomWorkspacesDir`: Correct count excluding non-directory files; slash and whitespace trimming tested.
- **CLI Integration Tests (`main_test.go`)**:
  - `TestCLI_Invariants/fatal on unresolvable base branch`: CLI exits 1 with base branch guidance.
  - `TestCLI_Invariants/fatal on empty src_folders`: CLI exits 1 with source folder guidance.

---

## Verdict

Status: **APPROVED**
