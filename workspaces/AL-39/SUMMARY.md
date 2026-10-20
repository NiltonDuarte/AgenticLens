# Ticket AL-39: Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes

## Summary of Completed Work

### Domain Core Layer (`backend/internal/domain`)
- **CodeMetricsConfig Model Extension**: Added `BaseBranch string` with tags `json:"base_branch,omitempty" yaml:"base_branch,omitempty"` to `domain.CodeMetricsConfig` in `backend/internal/domain/config.go`.
- **Pure Default Initialization**: Updated `DefaultProjectConfig()` to initialize `BaseBranch: ""` within `CodeMetricsConfig`.
- **FCIS Invariants Preserved**: Maintained strict isolation in domain with pure value-semantic structs, zero I/O imports, and zero external dependencies.

### Adapters & Boundary Shell Layer (`backend/internal/adapters`)
- **Configuration Loader (`config/loader.go`)**:
  - Added `BaseBranch *string` to `rawCodeMetricsConfig` with JSON and YAML tags.
  - Updated `mergeConfig` to merge non-empty `BaseBranch` into `ProjectConfig.CodeMetrics.BaseBranch`.
- **Git Classification Adapter (`git/classifier.go`)**:
  - Enriched `ClassificationConfig` with `BaseBranch string`, `SrcFolders []string`, and `WorkspacesDir string`.
  - Updated `LoadConfig` to populate these fields from `domain.ProjectConfig`.
  - Decoupled `ClassifyFile` from static `"workspaces/"` strings by respecting `cfg.WorkspacesDir` (fallback `"workspaces"`).
- **Iteration Inference Adapter (`git/inference.go`)**:
  - Decoupled `findWorkspaceDir` and `ComputeSDLCIterations` to accept a dynamic workspaces directory name (fallback `"workspaces"`).
- **Git Analyzer Adapter (`git/analyzer.go`)**:
  - Implemented dynamic upstream base branch resolution in `ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)` with 4-tier strategy:
    1. Explicit configuration (`cfg.CodeMetrics.BaseBranch`) with commit existence verification.
    2. Symbolic reference remote tracking (`refs/remotes/origin/HEAD`) with commit existence verification.
    3. Sequential candidate probing (`origin/main` > `main` > `origin/master` > `master`).
    4. Explicit Failure Invariant: If unresolved, fails immediately prompting `unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml`. Strictly prohibited from falling back to `HEAD`.
  - Replaced all 4 hardcoded `"origin/main"` references in `GetTicketNumstat` with dynamic `baseBranch`.
  - Implemented `ValidateSrcFolders(srcFolders []string) error` with explicit failure invariant: fails if empty or whitespace-only prompting `workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml`.
  - Implemented pure helper `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool` matching configured source folders, `cfg.Include`, and `cfg.ProdPatterns`.
  - Eliminated duplicate 6-condition checks at lines 135-140 and lines 167-172 of `analyzer.go`.
  - Updated `ComputeCodeFacts` signature to return `(domain.CodeFacts, error)`, validating source folders and resolving base branch dynamically.
  - Refactored `GetTicketNumstat` fallback resolution to validate `cfg.BaseBranch` against git commit objects using `ResolveBaseBranch`.
- **Backfill Engine (`backfill/engine.go`)**:
  - Updated `recomputeTicketFacts` to handle errors from `ComputeCodeFacts` without masking failures.

### Repository Configuration & CI Workflow
- **Repository Configuration (`.agenticlens-config.yaml`)**:
  - Added `base_branch: origin/main` under `code_metrics:` with explanatory comments detailing the upstream comparison base branch and dynamic fallback probe sequence.
- **Continuous Integration (`.github/workflows/ci.yml`)**:
  - Configured `fetch-depth: 0` on `actions/checkout@v4` in `backend-check` job, matching `build-binary`, ensuring remote tracking branches are available during CI test runs.

### CLI Layer (`backend/cmd/agentic-lens`)
- **Error Propagation (`main.go`)**: Updated `runGenerate` to catch errors returned by `git.ComputeCodeFacts`, print diagnostic guidance to `os.Stderr`, and exit with code 1.

### Testing & Invariants
- **Unit & Boundary Tests**:
  - `config/loader_test.go`: YAML and JSON unmarshaling, default fallback for `base_branch`, and root config loading verification.
  - `git/analyzer_test.go`: Table-driven tests for `ResolveBaseBranch`, `ValidateSrcFolders`, `IsMonitoredSourcePath` (31 matrix cases), elimination of duplicate conditions, dynamic branch diffing, base branch validation via config, and explicit failure invariant assertions.
  - `git/classifier_test.go`: Custom workspace directory classification tests with normalized boundaries.
  - `git/inference_test.go`: Custom workspace directory iteration counting tests.
  - `cmd/agentic-lens/main_test.go`: CLI exit code 1 tests on unresolvable base branch, empty source folders, and deterministic `-b main` repo initialization.
- **Architectural Gates**: Zero `any` or `interface{}` across Go code, zero emojis, zero race conditions (`go test -race ./...`), and zero skipped tests.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 4h 52m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.2k tokens (84% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 18.4%, Coder 41.7%, Tester 25.7%, Reviewer 14.3%
- **Subagent Turn Velocity & Thrashing:** 49.3 avg turns/run (peak: 108 turns) | Planner 28.3, Coder 63.7, Tester 58.5, Reviewer 50.0
- **Test-to-Production Code Ratio:** 4.0:1 (+1,172 test lines / +291 prod lines)
- **Code Delivered:** +1,465 / -62 lines across 15 code files (excluding docs)
- **Reasoning Effort:** 199,167 reasoning tokens (63.1% of generated output)
- **Tool Health & Self-Correction:** 620 tool calls (13 self-corrections, 97.9% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (4 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.5k tokens) — Context tokens reached 163.5k in turn 34 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 52 (session: ses_ea5e9243effefZDhJYZILu0hkM), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+133.1k tokens) — Context tokens surged by 133.1k tokens in turn 80 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 133.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 85 (session: ses_ea6042444ffeLWtlfXPgcYq2PI), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+40.9k tokens) — Context tokens surged by 40.9k tokens in turn 42 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx, total: 40.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.3k tokens) — Context tokens surged by 55.3k tokens in turn 93 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 55.3k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,115,868 | 27,113 | 7,888,342 | 78.9% | 168.2k | ELEVATED | $2.37 |
| **planner** | 3 | 646,870 | 21,399 | 3,423,940 | 84.1% | 103.9k | ELEVATED | $1.01 |
| **coder** | 3 | 1,496,438 | 34,289 | 10,151,063 | 87.2% | 168.4k | ELEVATED | $2.28 |
| **tester** | 2 | 790,775 | 25,452 | 8,010,698 | 91.0% | 163.5k | ELEVATED | $1.41 |
| **reviewer** | 2 | 472,676 | 8,445 | 4,177,937 | 89.8% | 84.9k | HEALTHY | $0.78 |
| **Total** | **11** | **5,522,627** | **116,698** | **33,651,980** | **85.9%** | **168.4k** | **ELEVATED** | **$7.85** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 11 | 5,522,627 | 116,698 | 33,651,980 | 85.9% | $7.85 |
| **Total** | **11** | **5,522,627** | **116,698** | **33,651,980** | **85.9%** | **$7.85** |
