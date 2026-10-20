h3. Title: AL-39: Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes

*TARGET*: AgenticLens (adapters | git | domain | config | tests)

*CONTEXT*

The pre-release source code audit (AL-35) revealed that `backend/internal/adapters/git/analyzer.go` explicitly hardcodes `origin/main` across four separate git diff and log invocations (`origin/main...HEAD`, `origin/main`, `log origin/main`, `origin/main...<branch>`). This causes git diff analysis to fail when operating on repositories where the default branch is `main` (local without remotes), `master`, `trunk`, or `development`, or when running detached in CI/CD without an upstream `origin`. Furthermore, `analyzer.go` filters out all diff entries unless at least one modified file path matches a hardcoded directory list (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), completely ignoring user configuration (`ProjectConfig.Workspaces.SrcFolders` and `ProjectConfig.CodeMetrics.Include`) and discarding diffs in repositories with alternative folder structures (such as `app/`, `pkg/`, `services/`, `client/`, `ui/`, or root files). Finally, adapters hardcode `"workspaces/"` path strings rather than referencing `ProjectConfig.Workspaces.Dir`.

*GOAL*

Decouple the Git analyzer from hardcoded branch names and directory prefixes by implementing dynamic upstream base branch detection, respecting `ProjectConfig.Workspaces.SrcFolders` and `CodeMetrics` configuration for file diff inclusion, removing duplicated conditional chains, and resolving workspace directories dynamically.

*VERTICAL SLICE SCOPE*

* Dynamic Upstream Base Branch Resolution (`backend/internal/adapters/git/analyzer.go`):
  - Implement `ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) string`.
  - Strategy:
    1. Check `cfg.CodeMetrics.BaseBranch` if explicitly specified.
    2. Check `git symbolic-ref refs/remotes/origin/HEAD` for default remote tracking branch.
    3. Probe candidates: `origin/main`, `main`, `origin/master`, `master`.
    4. Fall back to `HEAD` if no upstream candidate exists.
  - Replace all hardcoded `"origin/main"` occurrences in `GetTicketNumstat` with dynamic base branch resolution.
* Dynamic Monitored Folder Filtering (`backend/internal/adapters/git/analyzer.go`):
  - Remove duplicate 6-condition checks at lines 135-140 and lines 167-172.
  - Implement pure helper `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool`.
  - Match modified files against `srcFolders` and `cfg.Include`/`cfg.ProdPatterns`. If `srcFolders` is empty or unconfigured, permit all non-ignored project files.
* Configured Workspaces Directory Resolution (`backend/internal/adapters/git/classifier.go`, `inference.go`):
  - Update `classifier.go` to use `cfg.Workspaces.Dir` (fallback: `"workspaces"`) when checking ignored paths.
  - Update `inference.go` `findWorkspaceDir` to accept workspaces directory name instead of hardcoded `"workspaces"`.
* Testing & Verification:
  - Unit tests verifying base branch resolution against local-only repos, master-based repos, and custom-configured base branches.
  - Table-driven tests in `analyzer_test.go` verifying diff inclusion across custom folder layouts (`app/`, `pkg/`, root files).
  - Verify zero regressions in existing ticket metrics extraction.

*ACCEPTANCE CRITERIA*

* Base branch is dynamically resolved without hardcoded `origin/main` requirements.
* Configured source code folders (`src_folders`) and pattern configurations are respected by git diff analyzer.
* Duplicate folder prefix conditional checks in `analyzer.go` are completely removed.
* Classifier and inference adapters respect custom workspace directory configurations.
* Quality gates pass (`make check`, `make test`, `make build`) with zero any and zero emojis.

*DEFINITION OF DONE*

* Code changes merged to target branch.
* Quality gates passing (`make check`, `make test`, `make build`).
* Zero any and zero emojis across all files.
