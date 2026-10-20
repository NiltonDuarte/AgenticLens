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
- **Backfill Engine (`backfill/engine.go`)**:
  - Updated `recomputeTicketFacts` to handle errors from `ComputeCodeFacts` without masking failures.

### CLI Layer (`backend/cmd/agentic-lens`)
- **Error Propagation (`main.go`)**: Updated `runGenerate` to catch errors returned by `git.ComputeCodeFacts`, print diagnostic guidance to `os.Stderr`, and exit with code 1.

### Testing & Invariants
- **Unit & Boundary Tests**:
  - `config/loader_test.go`: YAML and JSON unmarshaling and default fallback for `base_branch`.
  - `git/analyzer_test.go`: Table-driven tests for `ResolveBaseBranch`, `ValidateSrcFolders`, `IsMonitoredSourcePath` (31 matrix cases), elimination of duplicate conditions, dynamic branch diffing, and explicit failure invariant assertions.
  - `git/classifier_test.go`: Custom workspace directory classification tests with normalized boundaries.
  - `git/inference_test.go`: Custom workspace directory iteration counting tests.
  - `cmd/agentic-lens/main_test.go`: CLI exit code 1 tests on unresolvable base branch and empty source folders.
- **Architectural Gates**: Zero `any` or `interface{}` across Go code, zero emojis, zero race conditions (`go test -race ./...`), and zero skipped tests.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 1h 20m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 115.0k tokens (58% of 200k healthy ceiling, 1 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 17.0%, Coder 44.0%, Tester 26.7%, Reviewer 12.3%
- **Subagent Turn Velocity & Thrashing:** 66.0 avg turns/run (peak: 108 turns) | Planner 44.0, Coder 108.0, Tester 60.0, Reviewer 52.0
- **Test-to-Production Code Ratio:** 3.6:1 (+1,040 test lines / +288 prod lines)
- **Code Delivered:** +1,328 / -60 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 134,383 reasoning tokens (61.1% of generated output)
- **Tool Health & Self-Correction:** 318 tool calls (6 self-corrections, 98.1% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (2 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.5k tokens) — Context tokens reached 163.5k in turn 34 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 52 (session: ses_ea5e9243effefZDhJYZILu0hkM), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+40.9k tokens) — Context tokens surged by 40.9k tokens in turn 42 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx, total: 40.9k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 757,213 | 17,093 | 3,084,356 | 80.3% | 115.0k | ELEVATED | $0.90 |
| **planner** | 1 | 369,677 | 17,105 | 2,072,304 | 84.9% | 103.9k | ELEVATED | $0.61 |
| **coder** | 1 | 951,901 | 26,289 | 7,403,197 | 88.6% | 168.4k | ELEVATED | $1.59 |
| **tester** | 1 | 522,762 | 20,471 | 5,381,420 | 91.1% | 163.5k | ELEVATED | $0.96 |
| **reviewer** | 1 | 263,426 | 4,484 | 2,425,176 | 90.2% | 84.9k | HEALTHY | $0.44 |
| **Total** | **5** | **2,864,979** | **85,442** | **20,366,453** | **87.7%** | **168.4k** | **ELEVATED** | **$4.50** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 2,864,979 | 85,442 | 20,366,453 | 87.7% | $4.50 |
| **Total** | **5** | **2,864,979** | **85,442** | **20,366,453** | **87.7%** | **$4.50** |
