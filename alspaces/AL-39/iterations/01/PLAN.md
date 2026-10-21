# Implementation Plan: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes

## 1. Goal and Iteration Scope

The goal of iteration 01 is to eliminate hardcoded Git branch names (origin/main), hardcoded directory prefix filters (src/, backend/, internal/, cmd/, web/, tests/), duplicated boolean conditional chains, and hardcoded workspace directory references from backend/internal/adapters/git/.

The implementation enforces two explicit failure invariants:
1. Base Branch Invariant: If no upstream branch can be resolved (via explicit configuration, origin/HEAD symbolic reference, or candidate probes), fail immediately with a clear error prompting the user or agent to configure code_metrics.base_branch in .agenticlens-config.yaml. Never silently fall back to HEAD.
2. Monitored Source Folders Invariant: If src_folders is empty or unconfigured in ProjectConfig, fail immediately with a clear error prompting the user or agent to configure workspaces.src_folders in .agenticlens-config.yaml. Never silently fall back or permit all files.

## 2. Grounded Reference Architecture & Impacted Symbols

### Domain Layer (Functional Core)
- File: backend/internal/domain/config.go
  - Symbol: CodeMetricsConfig struct. Add BaseBranch string with tags json:"base_branch,omitempty" and yaml:"base_branch,omitempty".
  - Symbol: DefaultProjectConfig() function. Initialize BaseBranch: "" in the default CodeMetricsConfig.

### Boundary Adapters Layer (Imperative Shell)
- File: backend/internal/adapters/config/loader.go
  - Symbol: rawCodeMetricsConfig struct. Add BaseBranch *string with tags json:"base_branch" and yaml:"base_branch".
  - Symbol: mergeConfig(dto rawProjectConfigDTO) function. Add assignment for BaseBranch when dto.CodeMetrics.BaseBranch is non-nil and non-empty.
- File: backend/internal/adapters/git/classifier.go
  - Symbol: ClassificationConfig struct. Add BaseBranch string, SrcFolders []string, and WorkspacesDir string fields.
  - Symbol: LoadConfig(repoRoot string) function. Populate BaseBranch, SrcFolders, and WorkspacesDir from loaded domain.ProjectConfig.
  - Symbol: ClassifyFile(rawPath string, cfg *ClassificationConfig) function. Replace static workspaces/ check with dynamic prefix derived from cfg.WorkspacesDir (fallback workspaces).
- File: backend/internal/adapters/git/inference.go
  - Symbol: findWorkspaceDir(workspacesDir, ticketID, subDir string) function. Update signature to accept workspacesDir string, defaulting to workspaces when empty.
  - Symbol: ComputeSDLCIterations(ticketID string, workspacesDir ...string) function. Accept optional workspacesDir parameter, defaulting to workspaces.
- File: backend/internal/adapters/git/analyzer.go
  - Symbol: ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error) function. Implement multi-tier upstream base branch resolution with explicit error on failure.
  - Symbol: ValidateSrcFolders(srcFolders []string) error function. Validate non-empty, non-whitespace source folder slice.
  - Symbol: IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool pure helper function. Validate candidate paths against srcFolders, cfg.Include, and cfg.ProdPatterns.
  - Symbol: GetTicketNumstat(ticketID string, baseBranch string, srcFolders []string, cfg *ClassificationConfig) ([]NumstatEntry, error) function. Replace hardcoded origin/main strings with baseBranch. Replace duplicate conditional chains with IsMonitoredSourcePath.
  - Symbol: ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) (domain.CodeFacts, error) function. Validate configuration, orchestrate base branch resolution, invoke GetTicketNumstat, and pass dynamic WorkspacesDir to ComputeSDLCIterations.

### Application Layer
- File: backend/cmd/agentic-lens/main.go
  - Symbol: runGenerate. Handle error returned from git.ComputeCodeFacts. Print diagnostic prompt to standard error and terminate with exit status 1 upon failure.
- File: backend/internal/adapters/backfill/engine.go
  - Symbol: recomputeTicketFacts. Check error returned from git.ComputeCodeFacts and handle safely during batch migrations.

### Test Files
- backend/internal/adapters/config/loader_test.go: Add tests for code_metrics.base_branch unmarshaling and fallback merging.
- backend/internal/adapters/git/analyzer_test.go: Add tests for ResolveBaseBranch, ValidateSrcFolders, IsMonitoredSourcePath, duplicate chain elimination, and explicit failure errors.
- backend/internal/adapters/git/classifier_test.go: Add tests for custom workspaces_dir ignore rules.
- backend/internal/adapters/git/inference_test.go: Add tests for custom workspaces directory iteration counting.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Pure Domain Model & Configuration Adapter Schema Extension
- Task 1.1: Extend domain.CodeMetricsConfig in backend/internal/domain/config.go
  - Add BaseBranch string with tags json:"base_branch,omitempty" and yaml:"base_branch,omitempty".
  - In DefaultProjectConfig(), set BaseBranch: "" inside CodeMetricsConfig.
- Task 1.2: Update Configuration Loader in backend/internal/adapters/config/loader.go
  - Add BaseBranch *string to rawCodeMetricsConfig with tags json:"base_branch" and yaml:"base_branch".
  - In mergeConfig, check if dto.CodeMetrics.BaseBranch is non-nil and trimmed non-empty, and assign to cfg.CodeMetrics.BaseBranch.
- Task 1.3: Update ClassificationConfig in backend/internal/adapters/git/classifier.go
  - Add BaseBranch string, SrcFolders []string, and WorkspacesDir string fields.
  - In LoadConfig(repoRoot string), map BaseBranch: cfg.CodeMetrics.BaseBranch, SrcFolders: cfg.Workspaces.SrcFolders, and WorkspacesDir: cfg.Workspaces.Dir.

### Phase 2: Dynamic Upstream Base Branch Resolution & Failure Invariant
- Task 2.1: Implement ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error) in backend/internal/adapters/git/analyzer.go
  - Tier 1 (Configured branch): Check cfg.CodeMetrics.BaseBranch. If non-empty, check validity via git rev-parse --verify --quiet branch^{commit}. If valid, return it. If configured but invalid, return error stating that the configured base branch cannot be resolved.
  - Tier 2 (Symbolic ref): Execute git symbolic-ref refs/remotes/origin/HEAD. If successful, parse the target branch, trim refs/remotes/, verify via git rev-parse, and return it.
  - Tier 3 (Candidate probes): Sequentially probe candidates origin/main, main, origin/master, master using git rev-parse --verify --quiet candidate^{commit}. Return the first matching candidate.
  - Tier 4 (Failure Invariant): If all tiers fail to resolve, return empty string and fmt.Errorf("unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml"). Never silently fall back to HEAD.
- Task 2.2: Replace static origin/main in GetTicketNumstat
  - Replace origin/main...HEAD with dynamic baseBranch...HEAD.
  - Replace diff --numstat origin/main with dynamic diff --numstat baseBranch.
  - Replace log origin/main with dynamic log baseBranch.
  - Replace origin/main...branch with dynamic baseBranch...branch.

### Phase 3: Monitored Source Folders Validation & Helper Implementation
- Task 3.1: Implement ValidateSrcFolders(srcFolders []string) error in backend/internal/adapters/git/analyzer.go
  - If len(srcFolders) == 0, return fmt.Errorf("workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml").
  - Iterate through srcFolders; if every element is whitespace-only, return the exact same error.
  - Return nil if at least one non-empty source folder is present.
- Task 3.2: Implement IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool in backend/internal/adapters/git/analyzer.go
  - Clean path via forward-slash normalization.
  - Check path against normalized items in srcFolders (exact match or path starts with folder plus forward slash).
  - If not matched, check cfg.Include patterns via matchPattern.
  - If not matched, check cfg.ProdPatterns patterns via matchPattern.
  - Return true on any match, false otherwise.
- Task 3.3: Eliminate duplicate 6-condition checks in GetTicketNumstat
  - Delete lines 135-140 in step 3. Replace with check against IsMonitoredSourcePath(r.Path, srcFolders, cfg).
  - Delete lines 167-172 in step 4. Replace with check against IsMonitoredSourcePath(r.Path, srcFolders, cfg).

### Phase 4: Dynamic Workspaces Directory Decoupling
- Task 4.1: Update ClassifyFile in backend/internal/adapters/git/classifier.go
  - Extract workspace directory name from cfg.WorkspacesDir (fallback workspaces).
  - Construct prefix strings dir/ and /dir/.
  - If candidate path matches workspace prefix, return CategoryIgnored.
- Task 4.2: Update findWorkspaceDir and ComputeSDLCIterations in backend/internal/adapters/git/inference.go
  - Update findWorkspaceDir(workspacesDir, ticketID, subDir string) (string, bool) to use workspacesDir (defaulting to workspaces if empty).
  - Update ComputeSDLCIterations(ticketID string, workspacesDir ...string) int64 to pass workspacesDir to findWorkspaceDir.
- Task 4.3: Update ComputeCodeFacts in backend/internal/adapters/git/analyzer.go
  - Pass cfg.WorkspacesDir to ComputeSDLCIterations.

### Phase 5: CLI Entrypoint Integration & Error Handling
- Task 5.1: Update ComputeCodeFacts in backend/internal/adapters/git/analyzer.go
  - Signature: ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) (domain.CodeFacts, error).
  - Validate cfg.SrcFolders with ValidateSrcFolders. If invalid, return zeroed domain.CodeFacts and error.
  - Resolve base branch if cfg.BaseBranch is empty. If unresolvable, return zeroed domain.CodeFacts and error.
  - Call GetTicketNumstat(ticketID, baseBranch, cfg.SrcFolders, cfg). If error, return zeroed domain.CodeFacts and error.
- Task 5.2: Update CLI entrypoint in backend/cmd/agentic-lens/main.go
  - Check error returned by git.ComputeCodeFacts.
  - If error is non-nil, print Error: err to standard error and call os.Exit(1).
- Task 5.3: Update Backfill engine in backend/internal/adapters/backfill/engine.go
  - Handle error from git.ComputeCodeFacts. If error occurs, fail migration step or record error without masking.

### Phase 6: Unit, Adversarial, and Integration Tests
- Task 6.1: Test Base Branch Resolution in backend/internal/adapters/git/analyzer_test.go
  - Test explicit configuration precedence (code_metrics.base_branch).
  - Test symbolic ref detection (refs/remotes/origin/HEAD).
  - Test sequential candidate probing (origin/main, main, origin/master, master).
  - Test unresolvable base branch asserting failure with the prompt: unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml.
  - Assert that execution never falls back to HEAD.
- Task 6.2: Test Monitored Source Folders & Path Helper in backend/internal/adapters/git/analyzer_test.go
  - Test empty src_folders asserting failure with the prompt: workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml.
  - Test whitespace-only src_folders asserting identical failure error.
  - Table test for IsMonitoredSourcePath asserting matching behavior for custom folders (app/, pkg/, services/), root files matching cfg.Include, patterns matching cfg.ProdPatterns, and non-matching paths.
  - Test removal of duplicate conditional chains: verify that diffs in custom source directories are captured and diffs outside are ignored.
- Task 6.3: Test Workspaces Directory Decoupling in backend/internal/adapters/git/classifier_test.go and inference_test.go
  - Test ClassifyFile with custom workspaces_dir (such as .tickets or custom_workspaces) ignored, while default workspaces remains ignored when unconfigured.
  - Test ComputeSDLCIterations discovering iterations inside a custom workspaces directory.
- Task 6.4: Test Configuration Loader in backend/internal/adapters/config/loader_test.go
  - Test unmarshaling code_metrics.base_branch from YAML and JSON files.
  - Test fallback to empty string when omitted.
- Task 6.5: Verify Existing Test Suites
  - Run all existing tests across analyzer_test.go, classifier_test.go, inference_test.go, and loader_test.go to ensure zero regressions.

### Phase 7: Verification & Quality Gates
- Task 7.1: Run make check to verify static analysis, formatting, and linting.
- Task 7.2: Run make test to verify backend and frontend unit tests with race detection.
- Task 7.3: Run make build to verify single binary compilation.
- Task 7.4: Audit working tree for zero any and zero emojis.

## 4. Verification Strategy & Adversarial Assertions

### Verification Test Matrix

- TestResolveBaseBranch_Configured (analyzer.go): Config specifies code_metrics.base_branch: "release-1.0" which exists in git. Resolves "release-1.0" with nil error.
- TestResolveBaseBranch_Configured_Missing (analyzer.go): Config specifies code_metrics.base_branch: "nonexistent-branch". Fails with error indicating configured branch not found.
- TestResolveBaseBranch_SymbolicRef (analyzer.go): Config omitted, git symbolic-ref points to origin/trunk. Resolves "origin/trunk" with nil error.
- TestResolveBaseBranch_ProbeCandidates (analyzer.go): Config omitted, repo default branch is local master. Resolves "master" with nil error.
- TestResolveBaseBranch_Unresolvable_FailureInvariant (analyzer.go): Repo has only branch dev-temp, no candidate branches, unconfigured. Fails immediately with error containing unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml. Does NOT return HEAD.
- TestValidateSrcFolders_Empty_FailureInvariant (analyzer.go): src_folders: []. Fails immediately with error containing workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml.
- TestValidateSrcFolders_Whitespace_FailureInvariant (analyzer.go): src_folders: ["   ", "\t"]. Fails immediately with exact error prompt.
- TestIsMonitoredSourcePath_CustomLayout (analyzer.go): src_folders: ["app", "pkg"], paths: app/main.go, pkg/util.go, docs/spec.md. Returns true for app and pkg files, false for docs.
- TestIsMonitoredSourcePath_IncludeOverride (analyzer.go): src_folders: ["backend"], include: ["main.py"], path: main.py. Returns true for main.py.
- TestClassifyFile_CustomWorkspacesDir (classifier.go): workspaces_dir: "tickets", path: tickets/AL-39/CONTRACT.md. Returns CategoryIgnored.
- TestComputeSDLCIterations_CustomWorkspacesDir (inference.go): Custom directory my_workspaces/AL-39/iterations/ with 3 folders. Returns 3.
- TestConfigLoader_BaseBranch (loader.go): YAML manifest containing code_metrics.base_branch: "develop". cfg.CodeMetrics.BaseBranch equals "develop".
- TestGetTicketNumstat_DynamicBranch (analyzer.go): Resolves branch master, ticket commits diffed against master. Diffs correctly computed against master. Zero references to origin/main.
- TestMain_FatalOnBaseBranchUnresolved (main.go): CLI invoked in repo with unresolvable base branch. Exits with status 1, writes configuration guidance to stderr.

### Quality Gate Invariant Verification Commands
- Static Analysis: make check
- Backend Unit & Race Tests: go test -race ./internal/domain/... ./internal/adapters/... ./cmd/...
- Full Automated Quality Suite: make
- Invariant Grep Check (Zero Any): grep -rn "any" backend/internal/domain/ backend/internal/adapters/git/ (must return zero hits in implementation code)
- Invariant Grep Check (Zero Hardcoded origin/main in analyzer): grep -rn "origin/main" backend/internal/adapters/git/analyzer.go (must return zero hits)
- Invariant Grep Check (Zero Emojis): Verify no Unicode emoji codepoints exist in touched files.
