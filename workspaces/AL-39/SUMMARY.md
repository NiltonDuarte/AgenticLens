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
  - Decoupled `ClassifyFile` from static `"workspaces/"` strings by respecting `cfg.WorkspacesDir` strictly when configured. Removed invalid default workspace directory.
  - Removed universal ignore rules for `.md`, `.json`, and `.lock` files. Preserved specific lockfiles (`go.sum`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`) and dotfiles.
- **Git Analyzer Adapter (`git/analyzer.go`)**:
  - Implemented dynamic upstream base branch resolution in `ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)` with 4-tier strategy:
    1. Explicit configuration (`cfg.CodeMetrics.BaseBranch`) with commit existence verification.
    2. Symbolic reference remote tracking (`refs/remotes/origin/HEAD`) with commit existence verification.
    3. Sequential candidate probing (`origin/main` > `main` > `origin/master` > `master`).
    4. Explicit Failure Invariant: If unresolved, fails immediately prompting `unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml`. Strictly prohibited from falling back to `HEAD`.
  - Replaced all 4 hardcoded `"origin/main"` references in `GetTicketNumstat` with dynamic `baseBranch`.
  - Removed pathspec exclusions (`:(exclude)*.md`, `:(exclude)*.json`, `:(exclude)*.lock`) from `GetTicketNumstat`.
  - Implemented `ValidateSrcFolders(srcFolders []string) error` with explicit failure invariant: fails if empty or whitespace-only prompting `workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml`.
  - Implemented pure helper `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool` matching configured source folders, `cfg.Include`, and `cfg.ProdPatterns`.
  - Eliminated duplicate 6-condition checks in `analyzer.go`.
  - Updated `ComputeCodeFacts` signature to return `(domain.CodeFacts, error)`. Removed fallback to hardcoded `backend`, `internal`, `cmd`, `web` directories. If configuration or source folders are unconfigured or missing, returns zeroes or caller fallback with nil error.
- **Git Inference Adapter (`git/inference.go`)**:
  - Eliminated fragile filesystem crawling for iteration subdirectories: completely removed `findWorkspaceDir`.
  - Completely deleted `ComputeSDLCIterations` function.
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
  - `git/analyzer_test.go`: Table-driven tests for `ResolveBaseBranch`, `ValidateSrcFolders`, `IsMonitoredSourcePath`, elimination of duplicate conditions, dynamic branch diffing, base branch validation via config, absence of hardcoded folder fallback in `ComputeCodeFacts`, and inclusion of markdown/JSON in numstat.
  - `git/classifier_test.go`: Adversarial boundary tests for unconfigured vs configured workspaces directories, and removal of universal `.md`, `.json`, `.lock` ignore rules.
  - `git/inference_test.go`: AST-based validation in `TestComputeSDLCIterations_ZeroCallSitesOrExports` enforcing zero occurrences, exports, or call sites for `ComputeSDLCIterations` across the Go backend.
  - `cmd/agentic-lens/release_verification_test.go`: Removed non-source `TestYAMLFiles_SyntaxValidation`.
  - `cmd/agentic-lens/main_test.go`: CLI exit code 1 tests on unresolvable base branch, empty source folders, and deterministic `-b main` repo initialization.
- **Architectural Gates**: Zero `any` or `interface{}` across Go code, zero emojis, zero race conditions (`go test -race ./...`), and zero skipped tests.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 13
- **Active Agent Duration:** 17h 58m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.8k tokens (84% of 200k healthy ceiling, 4 compactions) | Subagents peak 168.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 13.6%, Coder 52.9%, Tester 18.3%, Reviewer 15.2%
- **Subagent Turn Velocity & Thrashing:** 52.3 avg turns/run (peak: 141 turns) | Planner 32.4, Coder 62.7, Tester 51.8, Reviewer 54.5
- **Test-to-Production Code Ratio:** 4.3:1 (+1,563 test lines / +362 prod lines)
- **Code Delivered:** +1,927 / -138 lines across 16 code files (excluding docs)
- **Reasoning Effort:** 494,619 reasoning tokens (67.5% of generated output)
- **Tool Health & Self-Correction:** 1,456 tool calls (24 self-corrections, 98.4% success rate)

### Anomalies & Health Flags
- **Status:** 9 execution anomalies detected (7 CRITICAL, 2 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.5k tokens) — Context tokens reached 163.5k in turn 34 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 39 (session: ses_ea4c6a475ffew6wxUwyEZ1btbT), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k in turn 52 (session: ses_ea5e9243effefZDhJYZILu0hkM), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+133.1k tokens) — Context tokens surged by 133.1k tokens in turn 80 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 133.1k tokens, threshold: 80000)
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 85 (session: ses_ea6042444ffeLWtlfXPgcYq2PI), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+101.5k tokens) — Context tokens surged by 101.5k tokens in turn 289 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 101.5k tokens, threshold: 80000)
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 6 out of 6 calls (100.0% failure rate, threshold: >=50%)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+40.9k tokens) — Context tokens surged by 40.9k tokens in turn 42 (session: ses_ea5f7156bffe8KXEp1pZQvBjNx, total: 40.9k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+55.3k tokens) — Context tokens surged by 55.3k tokens in turn 93 (session: ses_ea6042444ffeLWtlfXPgcYq2PI, total: 55.3k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 5,134,085 | 57,137 | 24,673,364 | 82.8% | 168.8k | ELEVATED | $6.13 |
| **planner** | 5 | 1,032,405 | 35,817 | 6,656,910 | 86.6% | 103.9k | ELEVATED | $1.78 |
| **coder** | 9 | 3,978,918 | 90,195 | 36,236,428 | 90.1% | 168.4k | ELEVATED | $6.94 |
| **tester** | 4 | 1,458,946 | 37,654 | 12,771,146 | 89.7% | 163.5k | ELEVATED | $2.41 |
| **reviewer** | 4 | 1,495,496 | 17,260 | 8,779,258 | 85.4% | 84.9k | HEALTHY | $2.00 |
| **Total** | **23** | **13,099,850** | **238,063** | **89,117,106** | **87.2%** | **168.8k** | **ELEVATED** | **$19.26** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 23 | 13,099,850 | 238,063 | 89,117,106 | 87.2% | $19.26 |
| **Total** | **23** | **13,099,850** | **238,063** | **89,117,106** | **87.2%** | **$19.26** |
