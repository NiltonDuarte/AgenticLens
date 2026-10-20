# Contract: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes

## 1. Problem Statement & Audit Finding Alignment

The pre-release source code audit (AL-35) identified architectural coupling within the Git diff analysis subsystem in backend/internal/adapters/git/analyzer.go, where upstream comparison branches and source directory prefixes are statically hardcoded. This causes diff extraction to fail in repositories that do not use origin/main as their upstream branch, or when operating in repositories with custom source directory structures or custom workspace locations.

Specific audit findings addressed by this contract:
- Section 3 High Finding: Hardcoded Git Base Branch origin/main. backend/internal/adapters/git/analyzer.go explicitly hardcodes origin/main across four separate Git diff and log invocations (diff --numstat origin/main...HEAD, diff --numstat origin/main, log origin/main --grep=..., and diff --numstat origin/main...branch). This fails when running against repositories with local main without remotes, master, trunk, development, or in detached CI/CD runners lacking an origin remote.
- Section 3 High Finding: Hardcoded Source Folder Prefixes in Git Analyzer. analyzer.go discards diff entries unless at least one modified path matches a static list of hardcoded prefixes (src/, backend/, internal/, cmd/, web/, tests/). This completely ignores user configuration (ProjectConfig.Workspaces.SrcFolders and ProjectConfig.CodeMetrics.Include) and silently drops commits in projects structured with app/, pkg/, services/, client/, ui/, core/, lib/, or root-level files.
- Section 3 Low Finding: Duplicate Directory Prefix Conditions in Git Analyzer. The exact same 6-condition boolean chain is duplicated verbatim twice at lines 135-140 and lines 167-172 of analyzer.go.
- Section 3 Medium Finding: Hardcoded workspaces/ Directory Across Adapters. File classification in classifier.go (line 107) and iteration discovery in inference.go (line 32) hardcode the string workspaces/ instead of inspecting ProjectConfig.Workspaces.Dir.
- Section 4 Recommendation 1: Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes (AL-39).

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must strictly adhere to TECHNICAL_GUIDELINES.md and repository architectural standards.

### Inward Dependency Flow & Layer Separation
1. Functional Core (backend/internal/domain/config.go):
   - Houses pure configuration models and default constructors (ProjectConfig, CodeMetricsConfig, WorkspacesConfig).
   - Invariant: Zero external dependencies, zero filesystem operations, zero subprocess execution (os/exec), and zero clocks. Pure value semantics with immutable structs.
2. Boundary Adapter Layer (backend/internal/adapters/):
   - backend/internal/adapters/config/loader.go: Encapsulates all filesystem I/O, parent directory traversals, format detection, and YAML/JSON unmarshaling for project configuration.
   - backend/internal/adapters/git/analyzer.go: Encapsulates Git subprocess execution (os/exec.Command), base branch resolution, path filtering, and numstat parsing.
   - backend/internal/adapters/git/classifier.go: Pure pattern matching and file category assignment driven by configuration structures.
   - backend/internal/adapters/git/inference.go: Workspace iteration counting and branch prefix type resolution.
   - Invariant: All filesystem and Git subprocess I/O is restricted to adapters. Adapters return pure domain models or explicit wrapped errors.
3. Consuming CLI Layer (backend/cmd/agentic-lens/main.go):
   - Application entrypoint and CLI orchestration. Invokes configuration loading and Git analysis. Must catch and halt execution on invariant violation errors with non-zero exit codes.
4. Presentation Layer (web/):
   - Pure presentation dashboard. The web dashboard consumes pure domain.CodeFacts via JSON payloads. Zero direct awareness of Git branches, Git execution, or local repository layout.

### Strong Typing & Zero Any Invariant
- Zero usage of any or interface{} across Go backend and TypeScript frontend.
- Typed structs, concrete slices, and explicit error returns throughout.

### Zero Emojis Mandate
- Absolute prohibition of emojis across code, comments, docstrings, commit messages, PR drafts, CLI messages, logs, test assertions, and workspace artifacts.

## 3. Scope Boundaries

### In Scope for AL-39
1. Pure Domain Model Extension (backend/internal/domain/config.go):
   - Add BaseBranch string field to domain.CodeMetricsConfig with JSON and YAML tags base_branch.
   - Update DefaultProjectConfig() to initialize BaseBranch as an empty string.
2. Configuration Loader Adaptation (backend/internal/adapters/config/loader.go):
   - Add BaseBranch *string field to rawCodeMetricsConfig.
   - Update mergeConfig to map and merge configured base_branch values into domain.ProjectConfig.
3. Classification Configuration Synchronization (backend/internal/adapters/git/classifier.go):
   - Add BaseBranch string, SrcFolders []string, and WorkspacesDir string fields to git.ClassificationConfig.
   - Update LoadConfig(repoRoot string) to populate these fields from the loaded domain.ProjectConfig.
   - Decouple ClassifyFile from hardcoded workspaces/ strings by respecting cfg.WorkspacesDir (with fallback to workspaces).
4. Dynamic Upstream Base Branch Resolution (backend/internal/adapters/git/analyzer.go):
   - Implement ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error).
   - Implement multi-tier resolution strategy.
   - Enforce Explicit Failure Invariant 1: fail immediately with an explicit error prompting configuration if unresolved; never fall back to HEAD.
   - Replace all four hardcoded origin/main occurrences in GetTicketNumstat with dynamic base branch resolution.
5. Dynamic Monitored Source Folder Filtering (backend/internal/adapters/git/analyzer.go):
   - Implement ValidateSrcFolders(srcFolders []string) error.
   - Enforce Explicit Failure Invariant 2: fail immediately with an explicit error prompting configuration if src_folders is empty or unconfigured; never fall back or permit all files.
   - Implement pure helper IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool.
   - Eliminate duplicate 6-condition checks at lines 135-140 and lines 167-172 of analyzer.go.
6. Dynamic Workspace Directory Resolution (backend/internal/adapters/git/inference.go):
   - Update findWorkspaceDir to accept a workspaces directory parameter with fallback to workspaces.
   - Update ComputeSDLCIterations to accept an optional workspaces directory parameter.
   - Pass cfg.WorkspacesDir from ComputeCodeFacts into iteration calculation.
7. CLI and Backfill Error Handling:
   - Update ComputeCodeFacts signature in analyzer.go to return (domain.CodeFacts, error).
   - Update main.go and backfill/engine.go to halt execution on fatal configuration invariant violations.
8. Testing and Quality Verification:
   - Unit tests covering base branch resolution, candidate probing, and unresolvable branch failure prompts.
   - Unit tests covering source folder validation, empty folder failure prompts, and custom folder structures (app/, pkg/, services/, root files).
   - Unit tests covering custom workspaces directories in file classification and iteration discovery.
   - Quality gates passing (make check, make test, make build) with zero any and zero emojis.

### Out of Scope for AL-39
1. Exposing /api/config REST endpoint and frontend branding changes (governed downstream by AL-40).
2. Refactoring artificial sleeps and subprocess churning in tests (governed downstream by AL-41).
3. Modifying SQLite schema or OpenCode session fact extraction.
4. Altering frontend UI layout, charts, or views in web/src/.

## 4. Architectural Contracts & Explicit Failure Invariants

### Invariant 1: Dynamic Upstream Base Branch Resolution Contract
Function signature: ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)

Resolution strategy executes in strict priority order:
1. Explicit Configuration Tier:
   - Inspect cfg.CodeMetrics.BaseBranch after trimming whitespace.
   - If non-empty, verify branch existence in git via git rev-parse --verify --quiet branch^{commit}.
   - If reachable, return the configured branch string immediately.
   - If configured but unreachable or invalid in git, fail execution immediately with an explicit error indicating that the configured branch cannot be found.
2. Symbolic Reference Remote Tracking Tier:
   - Execute git symbolic-ref refs/remotes/origin/HEAD.
   - If command succeeds, strip any refs/remotes/ prefix and verify commit existence via git rev-parse --verify --quiet branch^{commit}.
   - If verified, return the remote tracking branch (for example, origin/main).
3. Candidate Probing Tier:
   - Sequentially probe candidate branch names in this exact order:
     Priority 1: origin/main
     Priority 2: main
     Priority 3: origin/master
     Priority 4: master
   - For each candidate, probe existence via git rev-parse --verify --quiet candidate^{commit}.
   - The first candidate that resolves to an existing commit object is returned immediately.
4. Explicit Failure Invariant:
   - If no upstream branch is reachable or resolved across steps 1 through 3, execution must fail immediately.
   - Return an empty string and an explicit error with the prompt: unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml.
   - Mandatory Prohibition: Execution must NEVER silently fall back to HEAD.

### Invariant 2: Dynamic Monitored Source Folders Contract
Function signature: ValidateSrcFolders(srcFolders []string) error
Helper signature: IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool

Validation rules:
1. Input Examination:
   - Inspect srcFolders slice from ProjectConfig.Workspaces.SrcFolders (or ClassificationConfig.SrcFolders).
   - If srcFolders is empty (len(srcFolders) == 0), or if all elements consist solely of empty strings or whitespace:
     Fail execution immediately and return an explicit error with the prompt: workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml.
   - Mandatory Prohibition: Execution must NEVER silently fall back or permit all files.
2. Monitored Path Evaluation:
   - Clean the candidate file path using forward slashes and trimmed whitespace.
   - Match against configured srcFolders: For each folder in srcFolders, normalize by trimming leading and trailing slashes. If the file path equals the folder name or starts with the folder name followed by a forward slash, return true.
   - Match against explicit inclusion patterns: If cfg is non-nil and cfg.Include contains glob patterns matching the file path via matchPattern, return true.
   - Match against production patterns: If cfg is non-nil and cfg.ProdPatterns contains glob patterns matching the file path via matchPattern, return true.
   - If none of the above conditions match, return false.

### Invariant 3: Configured Workspaces Directory Contract
1. Classification Ignore Rules:
   - In ClassifyFile, resolve the active workspace directory name from cfg.WorkspacesDir. If empty or unconfigured, default strictly to workspaces.
   - Construct normalized prefix patterns (such as dir/ and /dir/).
   - If candidate path matches the workspace prefix, classify immediately as CategoryIgnored.
2. Iteration Calculation:
   - In ComputeSDLCIterations, accept the configured workspaces directory name (defaulting to workspaces when empty).
   - Discover iteration directories using the dynamic directory path instead of hardcoded workspaces.

### Invariant 4: Upward Error Propagation Contract
- When GetTicketNumstat or ComputeCodeFacts encounters an invariant violation (unresolved base branch or unconfigured source folders), the error must be returned up the call stack.
- backend/cmd/agentic-lens/main.go must catch this error, log the exact configuration guidance message to standard error, and exit with status 1. It must never generate or persist corrupt or partial telemetry ledgers under an unverified baseline.

## 5. Go/TypeScript Layer Contracts and Symbol Boundaries

- Domain Core (backend/internal/domain/config.go): CodeMetricsConfig, DefaultProjectConfig. Adds BaseBranch string to CodeMetricsConfig. Pure struct definition with zero I/O.
- Boundary Adapter (backend/internal/adapters/config/loader.go): rawCodeMetricsConfig, mergeConfig. Unmarshals base_branch from YAML/JSON and merges into ProjectConfig.CodeMetrics.
- Boundary Adapter (backend/internal/adapters/git/classifier.go): ClassificationConfig, LoadConfig, ClassifyFile. Enriches ClassificationConfig with BaseBranch, SrcFolders, and WorkspacesDir. Decouples workspace folder ignore logic from static strings.
- Boundary Adapter (backend/internal/adapters/git/inference.go): findWorkspaceDir, ComputeSDLCIterations. Accepts dynamic workspaces directory name with fallback to workspaces.
- Boundary Adapter (backend/internal/adapters/git/analyzer.go): ResolveBaseBranch, ValidateSrcFolders, IsMonitoredSourcePath, GetTicketNumstat, ComputeCodeFacts. Implements dynamic base branch resolution, failure invariants, monitored path helper, eliminates duplicate conditional checks, and updates numstat comparisons.
- CLI Entrypoint (backend/cmd/agentic-lens/main.go): runGenerate. Handles errors returned by ComputeCodeFacts and exits immediately with code 1 upon invariant violation.
- Backfill Adapter (backend/internal/adapters/backfill/engine.go): recomputeTicketFacts. Handles errors from ComputeCodeFacts during batch backfill runs.
- Presentation (web/src/): None (Decoupled). Consumes pure domain.CodeFacts from Go backend. No changes required.

## 6. Quality Gates and Verification Criteria

Every deliverable must satisfy the following automated quality checks:
- Static analysis passing: make check (including go vet ./..., TypeScript lint, and typecheck).
- Full test suite passing with race detector: make test (including go test -race ./...).
- Single binary compilation passing: make build (compiled to bin/agentic-lens).
- Zero any or interface{} across all Go and TypeScript files.
- Zero emojis across all code, tests, docstrings, and workspace files.
