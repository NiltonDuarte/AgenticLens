# Implementation Plan: AL-39 Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes (Iteration 04)

**Artifact Path**: `workspaces/AL-39/iterations/04/PLAN.md`  
**Ticket**: AL-39  
**Scope Reference**: `workspaces/AL-39/CONTRACT.md`, `workspaces/AL-39/iterations/03/PLAN.md`

---

## 1. Goal and Iteration Scope

Iteration 04 finalizes the complete decoupling of the Git analyzer and classifier from hardcoded values:

1. **Remove Hardcoded Folder Fallbacks in `ComputeCodeFacts`**:
   - In `backend/internal/adapters/git/analyzer.go`: remove hardcoded fallback to `SrcFolders: []string{"backend", "internal", "cmd", "web"}`.
   - If `cfg == nil`, attempt `LoadConfig("")`.
   - If `cfg == nil || len(cfg.SrcFolders) == 0 || ValidateSrcFolders(cfg.SrcFolders) != nil`: return zeroes or fallback (`*fallback, nil` if `fallback != nil`; otherwise `domain.CodeFacts{SDLCIterations: 1}, nil`).
2. **Remove Git Diff Pathspec Exclusions in `GetTicketNumstat`**:
   - Remove `excludeArgs` (`:(exclude)*.md`, `:(exclude)*.json`, `:(exclude)*.lock`). Git diff invocations should not pass these pathspecs so that markdown, json, etc. are passed to `ClassifyFile`.
3. **Decouple Workspace Directory and File Classification**:
   - In `backend/internal/adapters/git/classifier.go`: remove `wsDir := "workspaces"`. Workspace directory checking must only run when `cfg != nil && strings.TrimSpace(cfg.WorkspacesDir) != ""`. There is no valid default workspace directory.
   - In `ClassifyFile`: remove universal ignore rules `ext == ".md" || ext == ".json" || ext == ".lock"`. Specific lockfiles (`go.sum`, `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`) and dotfiles remain ignored (`CategoryIgnored`). Arbitrary markdown/JSON/lock files fall through to `CategoryOther`.
4. **Remove Obsolete `ComputeSDLCIterations`**:
   - In `backend/internal/adapters/git/inference.go`: delete `ComputeSDLCIterations` function entirely.
5. **Update Test Suites**:
   - `backend/internal/adapters/git/inference_test.go`: Delete `TestComputeSDLCIterations_CustomWorkspacesDir` and `TestComputeSDLCIterations_Decoupled`.
   - `backend/internal/adapters/git/analyzer_test.go`: Remove calls to `ComputeSDLCIterations` in `TestInferenceOnActiveWorkspace` and `TestGitAnalyzer_AdversarialEdgeCases`. Update `TestClassifyFile` expectations for `README.md`, `docs/ARCH.md`, `package.json`, and workspaces files with nil config (expect `CategoryOther`). Add cases with configured `WorkspacesDir` expecting `CategoryIgnored`. Update `TestComputeCodeFacts_Invariants` subtest for empty src_folders to assert zeroes or fallback are returned with nil error.
   - `backend/internal/adapters/git/classifier_test.go`: Update `default workspaces directory is ignored when unconfigured` test and `TestClassifyFile_TestAndProd` to align with the removal of default workspaces directory and universal md/json/lock ignoring.
6. **Invariants**:
   - Zero emojis in any code, comments, or tests.
   - Strict static typing, zero `any` or `interface{}`.
   - All tests passing with `go test -race ./...`.
