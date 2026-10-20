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
- **Git Analyzer Adapter (`git/analyzer.go`)**:
  - Implemented dynamic upstream base branch resolution in `ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)` with 4-tier strategy:
    1. Explicit configuration (`cfg.CodeMetrics.BaseBranch`) with commit existence verification.
    2. Symbolic reference remote tracking (`refs/remotes/origin/HEAD`) with commit existence verification.
    3. Sequential candidate probing (`origin/main` > `main` > `origin/master` > `master`).
    4. Explicit Failure Invariant: If unresolved, fails immediately prompting `unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml`. Strictly prohibited from falling back to `HEAD`.
  - Replaced all 4 hardcoded `"origin/main"` references in `GetTicketNumstat` with dynamic `baseBranch`.
  - Implemented `ValidateSrcFolders(srcFolders []string) error` with explicit failure invariant: fails if empty or whitespace-only prompting `workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml`.
  - Implemented pure helper `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool` matching configured source folders, `cfg.Include`, and `cfg.ProdPatterns`.
  - Eliminated duplicate 6-condition checks in `analyzer.go`.
  - Updated `ComputeCodeFacts` signature to return `(domain.CodeFacts, error)`, validating source folders and resolving base branch dynamically.
  - Decoupled `ComputeCodeFacts` from filesystem iteration directories: uses fallback iterations if non-nil and > 0, else defaults to 1.
- **Git Inference Adapter (`git/inference.go`)**:
  - Eliminated fragile filesystem crawling for iteration subdirectories: completely removed `findWorkspaceDir`.
  - Deprecated `ComputeSDLCIterations` to return default count 1 without performing any `os.Stat` or `os.ReadDir` traversals.
- **OpenCode Telemetry Extractor (`opencode/extractor.go`)**:
  - Implemented `ExtractSteeringTurns(db *sql.DB, threadIDs []string) (int64, error)` to derive SDLC iterations directly from human steering turns (`role == "user"`) recorded in the root session for the ticket.
- **Backfill Engine (`backfill/engine.go`)**:
  - Updated `recomputeTicketFacts` to handle errors from `ComputeCodeFacts` without masking failures.
  - Connected `ExtractSteeringTurns` during DB recomputation to populate `factLedger.Code.SDLCIterations`.

### Repository Configuration & CI Workflow
- **Repository Configuration (`.agenticlens-config.yaml`)**:
  - Added `base_branch: origin/main` under `code_metrics:` with explanatory comments detailing the upstream comparison base branch and dynamic fallback probe sequence.
- **Continuous Integration (`.github/workflows/ci.yml`)**:
  - Configured `fetch-depth: 0` on `actions/checkout@v4` in `backend-check` job, matching `build-binary`, ensuring remote tracking branches are available during CI test runs.

### CLI Layer (`backend/cmd/agentic-lens`)
- **Steering Turns Integration (`main.go`)**: Extracted steering turns from session records via `opencode.ExtractSteeringTurns`, assigning `codeFacts.SDLCIterations = extractedSteeringTurns` when >= 1.
- **Error Propagation (`main.go`)**: Updated `runGenerate` to catch errors returned by `git.ComputeCodeFacts`, print diagnostic guidance to `os.Stderr`, and exit with code 1.

### Testing & Invariants
- **Unit & Boundary Tests**:
  - `opencode/extractor_test.go`: Added `TestExtractSteeringTurns` verifying single/multiple user turns, case-insensitive role parsing, child session isolation, and resilient JSON error skipping.
  - `config/loader_test.go`: YAML and JSON unmarshaling, default fallback for `base_branch`, and root config loading verification.
  - `git/analyzer_test.go`: Table-driven tests for `ResolveBaseBranch`, `ValidateSrcFolders`, `IsMonitoredSourcePath` (31 matrix cases), elimination of duplicate conditions, dynamic branch diffing, base branch validation via config, and decoupled iterations.
  - `git/classifier_test.go`: Custom workspace directory classification tests with normalized boundaries.
  - `git/inference_test.go`: Decoupled `ComputeSDLCIterations` tests without filesystem directory fixtures.
  - `cmd/agentic-lens/release_verification_test.go`: Removed non-source `TestYAMLFiles_SyntaxValidation`.
  - `cmd/agentic-lens/main_test.go`: CLI exit code 1 tests on unresolvable base branch, empty source folders, and deterministic `-b main` repo initialization.
- **Architectural Gates**: Zero `any` or `interface{}` across Go code, zero emojis, zero race conditions (`go test -race ./...`), and zero skipped tests.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 8
- **Active Agent Duration:** 6h 0m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.2k tokens (84% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 18.7%, Coder 44.5%, Tester 20.8%, Reviewer 16.1%
- **Subagent Turn Velocity & Thrashing:** 49.5 avg turns/run (peak: 108 turns) | Planner 32.5, Coder 60.4, Tester 48.3, Reviewer 55.0
- **Test-to-Production Code Ratio:** 3.5:1 (+1,207 test lines / +349 prod lines)
- **Code Delivered:** +1,558 / -104 lines across 16 code files (excluding docs)
- **Reasoning Effort:** 279,033 reasoning tokens (63.4% of generated output)
- **Tool Health & Self-Correction:** 934 tool calls (18 self-corrections, 98.1% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (5 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.5k tokens) — Context tokens reached 163.5k in turn 34 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 52 (session: ses_ea5e9243effefZDhJYZILu0hkM), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+133.1k tokens) — Context tokens surged by 133.1k tokens in turn 80 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 133.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 85 (session: ses_ea6042444ffeLWtlfXPgcYq2PI), breaching 160k critical ceiling
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 5 out of 5 calls (100.0% failure rate, threshold: >=50%)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+40.9k tokens) — Context tokens surged by 40.9k tokens in turn 42 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx, total: 40.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.3k tokens) — Context tokens surged by 55.3k tokens in turn 93 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 55.3k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 3,203,977 | 42,062 | 12,492,204 | 79.6% | 168.2k | ELEVATED | $3.64 |
| **planner** | 4 | 882,822 | 27,118 | 5,441,490 | 86.0% | 103.9k | ELEVATED | $1.46 |
| **coder** | 5 | 2,196,072 | 50,562 | 16,874,932 | 88.5% | 168.4k | ELEVATED | $3.47 |
| **tester** | 3 | 925,926 | 28,048 | 9,225,629 | 90.9% | 163.5k | ELEVATED | $1.62 |
| **reviewer** | 3 | 743,802 | 13,580 | 7,046,087 | 90.5% | 84.9k | HEALTHY | $1.26 |
| **Total** | **16** | **7,952,599** | **161,370** | **51,080,342** | **86.5%** | **168.4k** | **ELEVATED** | **$11.45** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 16 | 7,952,599 | 161,370 | 51,080,342 | 86.5% | $11.45 |
| **Total** | **16** | **7,952,599** | **161,370** | **51,080,342** | **86.5%** | **$11.45** |
