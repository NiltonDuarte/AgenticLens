# refactor(git): decouple git analyzer from hardcoded branches and folder prefixes (AL-39)

## Overview

Decouples the AgenticLens Git analyzer and repository adapters from hardcoded branch names (`origin/main`), static folder prefix lists (`src/`, `backend/`, `internal/`, `cmd/`, `web/`, `tests/`), duplicated boolean conditional checks, and static `"workspaces/"` path strings. Upstream base branches are now dynamically resolved across four priority tiers with an explicit failure invariant if unresolvable. Monitored source code folders are dynamically validated against user configuration (`workspaces.src_folders`, `code_metrics.include`, and `code_metrics.prod_patterns`), preventing silent omission of code churn across non-standard repository topologies.

- Contract: [workspaces/AL-39/CONTRACT.md](workspaces/AL-39/CONTRACT.md)
- Telemetry: [workspaces/AL-39/SUMMARY.md](workspaces/AL-39/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- `config.go`: Added `BaseBranch string` to `domain.CodeMetricsConfig` (`base_branch`) and initialized `BaseBranch: ""` in `DefaultProjectConfig()`. Preserved pure value semantics and zero I/O imports.

### Boundary Adapters (`backend/internal/adapters/`)
- `config/loader.go`: Added `BaseBranch *string` to `rawCodeMetricsConfig` and updated `mergeConfig` to merge configured `base_branch` into `ProjectConfig.CodeMetrics`.
- `git/classifier.go`: Enriched `ClassificationConfig` with `BaseBranch`, `SrcFolders`, and `WorkspacesDir`. Decoupled `ClassifyFile` to prioritize `cfg.WorkspacesDir` (fallback `"workspaces"`).
- `git/inference.go`: Decoupled `findWorkspaceDir` and `ComputeSDLCIterations` to accept a dynamic workspaces directory name.
- `git/analyzer.go`:
  - Implemented `ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error)` evaluating configured branch, symbolic reference (`refs/remotes/origin/HEAD`), and candidate probes (`origin/main` > `main` > `origin/master` > `master`). Fails explicitly prompting configuration if unresolved; never falls back to `HEAD`.
  - Replaced all four hardcoded `"origin/main"` references in `GetTicketNumstat` with dynamic `baseBranch`.
  - Implemented `ValidateSrcFolders(srcFolders []string) error`, failing explicitly if empty or whitespace-only.
  - Implemented `IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool` and eliminated duplicate 6-condition checks.
  - Updated `ComputeCodeFacts` signature to return `(domain.CodeFacts, error)` and pass dynamic `WorkspacesDir` to iteration counting.
  - Refactored `GetTicketNumstat` fallback resolution to validate `cfg.BaseBranch` through `ResolveBaseBranch`.
- `backfill/engine.go`: Handled errors from `ComputeCodeFacts` without masking failures.

### Configuration & CI Workflows
- `.agenticlens-config.yaml`: Added `code_metrics.base_branch: origin/main` with explanatory comments.
- `.github/workflows/ci.yml`: Added `fetch-depth: 0` to `backend-check` checkout step to ensure remote branch refs are available in CI test runs.

### CLI Layer (`backend/cmd/agentic-lens/`)
- `main.go`: Updated `runGenerate` to catch errors from `git.ComputeCodeFacts`, log diagnostic prompts to stderr, and exit with status 1.

### Testing & Verification
- Unit and boundary tests covering base branch resolution tiers and explicit failure prompt, empty source folder prompt, custom repository folder layouts (`app/`, `pkg/`, `services/`, root files), custom workspaces directory isolation, and CLI exit code handling.
- Zero `any` types, zero emojis, and zero skipped tests.
