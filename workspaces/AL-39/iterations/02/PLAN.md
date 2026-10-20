# Implementation Plan: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes (Iteration 02)

## 1. Goal and Iteration Scope

Iteration 02 addresses review feedback and hardening items identified after iteration 01:
1. Explicitly document and configure `code_metrics.base_branch: origin/main` in `.agenticlens-config.yaml` with clear comments explaining its purpose as the upstream comparison base branch (e.g. origin/main or main) and automatic fallback behavior when omitted or empty.
2. Ensure full git history checkout (`fetch-depth: 0`) in `.github/workflows/ci.yml` under `backend-check` to match `build-binary`, ensuring Git integration tests have full history available during CI backend verification.
3. Validate configured base branch in `GetTicketNumstat` (`backend/internal/adapters/git/analyzer.go`): When `baseBranch` is empty and `cfg != nil && cfg.BaseBranch != ""`, resolve and validate the branch via `ResolveBaseBranch("", projCfg)`.
4. Enforce deterministic default branch initialization (`git init -b main`) in temporary repository setup within `backend/cmd/agentic-lens/main_test.go`.
5. Maintain all invariants: Zero emojis, zero `any`/`interface{}`, and clean compilation and test verification.

## 2. Step-by-Step Implementation Tasks

### Task 1: Update Global Project Configuration (.agenticlens-config.yaml)
- File: `.agenticlens-config.yaml`
- Under `code_metrics:`, add `base_branch: origin/main` with explanatory comments.
- Clarify that empty value triggers automatic multi-tier fallback (symbolic-ref origin/HEAD -> probe candidates origin/main, main, origin/master, master).

### Task 2: Update CI Workflow Configuration (.github/workflows/ci.yml)
- File: `.github/workflows/ci.yml`
- In job `backend-check`, update step `Check out repository` to include `with: fetch-depth: 0`, matching line 76 in `build-binary`.

### Task 3: Validate Configured Base Branch in GetTicketNumstat
- File: `backend/internal/adapters/git/analyzer.go`
- In `GetTicketNumstat(ticketID string, baseBranch string, srcFolders []string, cfg *ClassificationConfig)`:
  - If `strings.TrimSpace(baseBranch) == ""`:
    - Construct `projCfg` if `cfg != nil && strings.TrimSpace(cfg.BaseBranch) != ""`.
    - Invoke `ResolveBaseBranch("", projCfg)` so configured branches are validated against git commits via Tier 1, returning an error if unreachable.
- File: `backend/internal/adapters/git/analyzer_test.go`
  - Add test asserting that `GetTicketNumstat` with empty `baseBranch` and an invalid `cfg.BaseBranch` returns an explicit validation error.
  - Add test asserting that `GetTicketNumstat` with empty `baseBranch` and a valid `cfg.BaseBranch` succeeds.

### Task 4: Ensure Deterministic Branch Initialization in CLI Integration Tests
- File: `backend/cmd/agentic-lens/main_test.go`
  - Inspect `git init` invocations in `TestCLI_TicketTypePrecedence` (lines 3464 and 3523).
  - Use `git init -b main <dir>` so the initial branch is deterministically `main`.

### Task 5: Invariants & Verification
- Verify zero emojis across all modified files.
- Verify zero `any` or `interface{}` across Go and TypeScript.
- Verify `make check` and `make test`.
